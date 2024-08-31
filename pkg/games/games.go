package games

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/akamensky/base58"
	"github.com/notnil/chess"
	"github.com/thinktt/yowking/pkg/db"
	"github.com/thinktt/yowking/pkg/events"
	"github.com/thinktt/yowking/pkg/models"
	"github.com/thinktt/yowking/pkg/moveque"
)

func PublishGameUpdates(gameID string) error {
	game, err := db.GetGame2(gameID)
	if err != nil {
		return err
	}

	// if status hasn't change omit status and winner from update
	status := ""
	winner := ""
	if game.Status != "started" {
		status = game.Status
		winner = game.Winner
	}

	gameMuation := models.Game2MutableFields{
		ID:         game.ID,
		LastMoveAt: game.LastMoveAt,
		Moves:      game.Moves,
		Status:     status,
		Winner:     winner,
	}

	jsonData, _ := json.Marshal(gameMuation)
	events.Pub.PublishMessage(game.ID, string(jsonData))

	go CheckForEngineMove(game)

	return nil
}

func CheckForEngineMove(game models.Game2) {
	var cmpName string
	// turnColor := game.TurnColor()
	if game.TurnColor() == "white" && game.WhitePlayer.Type == "cmp" {
		cmpName = game.WhitePlayer.ID
	}
	if game.TurnColor() == "black" && game.BlackPlayer.Type == "cmp" {
		cmpName = game.BlackPlayer.ID
	}

	// no cmp found for this turn, nothing needs to be done
	if cmpName == "" {
		fmt.Println("not a cmp turn")
		return
	}

	chessGame, err := parseToChessGame(game.MoveList)
	if err != nil {
		fmt.Println("Error parsing game: ", err.Error())
	}

	uciMoves, err := getUCIMovesFromChessGame(chessGame)
	if err != nil {
		fmt.Println("Error parsing UCI moves: ", err.Error())
	}

	moveReq := models.MoveReq{
		Moves:   uciMoves,
		CmpName: cmpName,
		GameId:  game.ID,
	}

	// get the next move from the engine workers
	moveData, err := moveque.GetMove(moveReq)
	if err != nil {
		fmt.Println("Error getting engine move: ", err.Error())
		return
	}

	fmt.Println(moveData)

	// if there is not an Algebra move we will need to translate the coordinate move
	algebraMove := moveData.AlgebraMove
	if algebraMove == "" {
		algebraMove, err = getAlgebraMoveFromChessGame(chessGame, moveData.CoordinateMove)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}

	// fix the king's quirky castling notation
	if algebraMove == "0-0" {
		algebraMove = "O-O"
	}
	if algebraMove == "0-0-0" {
		algebraMove = "O-O-O"
	}

	// fix kings non standard promotion notation
	algebraMove = fixPromotionMove(algebraMove)

	fmt.Println(algebraMove)
	// moves := append(game.MoveList, algebraMove)
	err = AddMove(game.ID, cmpName,
		models.MoveData2{Index: len(game.Moves), Move: algebraMove})
	if err != nil {
		fmt.Println("error Adding engine move: ", err.Error())
	}
}

func fixPromotionMove(algebraMove string) string {
	if len(algebraMove) == 0 {
		return algebraMove
	}

	lastChar := algebraMove[len(algebraMove)-1]

	switch lastChar {
	case 'N', 'Q', 'B', 'R':
		return algebraMove[:len(algebraMove)-1] + "=" + string(lastChar)
	default:
		return algebraMove
	}
}

// var uciRegex = regexp.MustCompile(`[a-h][1-8][a-h][1-8][qrbn]?`)

func getAlgebraMoveFromChessGame(chessGame *chess.Game, newUciMove string) (string, error) {
	err := chessGame.MoveStr(newUciMove)
	if err != nil {
		err := fmt.Errorf("error adding coordinate move to chessGame: %s", err.Error())
		return "", err
	}

	chess.UseNotation(chess.AlgebraicNotation{})(chessGame)
	pgn := chessGame.String()
	pgnSlice := strings.Fields(pgn)
	if len(pgnSlice) < 3 {
		err := fmt.Errorf("unable parse algebra move from chessGame: PGN too short")
		return "", err
	}

	algebraMove := pgnSlice[len(pgnSlice)-2]
	return algebraMove, nil
}

func getUCIMovesFromChessGame(chessGame *chess.Game) ([]string, error) {
	chess.UseNotation(chess.UCINotation{})(chessGame)
	moves := make([]string, 0, len(chessGame.Moves()))

	for _, move := range chessGame.Moves() {
		moves = append(moves, move.String())
	}

	return moves, nil
}

func parseToChessGame(moves []string) (*chess.Game, error) {
	chessGame := chess.NewGame()
	for i, move := range moves {
		err := chessGame.MoveStr(move)
		if err != nil {
			errMsg := fmt.Sprintf("Invalid move at index %d: %v", i, err)
			return nil, errors.New(errMsg)
		}
	}
	return chessGame, nil
}

func GetAlgebraicNotation(uciMoves []string) ([]string, error) {
	gameParser := chess.NewGame(chess.UseNotation(chess.UCINotation{}))
	algebraicMoves := make([]string, 0, len(uciMoves))

	for i, move := range uciMoves {
		err := gameParser.MoveStr(move)
		if err != nil {
			errMsg := fmt.Sprintf("Invalid UCI move at index %d: %v", i, err)
			return nil, errors.New(errMsg)
		}
		algebraicMove := gameParser.Moves()[i].String()
		algebraicMoves = append(algebraicMoves, algebraicMove)
	}

	return algebraicMoves, nil
}

func GetGameID() (string, error) {
	// Create a byte array of length 6
	b := make([]byte, 6)

	// Fill the byte array with random bytes
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// Encode the byte array to a Base58 string
	encoded := base58.Encode(b)

	// If the encoded length is 9, chop off the first letter
	if len(encoded) == 9 {
		encoded = encoded[1:]
	}
	return encoded, nil
}
