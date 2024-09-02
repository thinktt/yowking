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
)

func PublishGameUpdates(gameID string) error {
	game, err := db.GetGame2(gameID)
	if err != nil {
		return err
	}

	// if winner hasn't change omit winner and method from update
	winner := ""
	method := ""
	if game.Winner != "pending" {
		winner = game.Winner
		method = game.Method
	}

	gameMuation := models.Game2MutableFields{
		ID:         game.ID,
		LastMoveAt: game.LastMoveAt,
		Moves:      game.Moves,
		Winner:     winner,
		Method:     method,
	}

	jsonData, _ := json.Marshal(gameMuation)
	events.Pub.PublishMessage(game.ID, string(jsonData))

	go PlayEngineMove(game)

	return nil
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
