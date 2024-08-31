package games

import (
	"fmt"
	"strings"

	"github.com/thinktt/yowking/pkg/models"
	"github.com/thinktt/yowking/pkg/moveque"
)

func CheckForEngineMove(game models.Game2) {

	// the game is over, get out of here
	if game.Status != "started" {
		return
	}

	// look for any cmp playing this game
	var cmpName string
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

	chessGame, err := ParseGame(game)
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
	engineMove, err := moveque.GetMove(moveReq)
	if err != nil {
		fmt.Println("Error getting engine move: ", err.Error())
		return
	}

	// if there is not an Algebra move we will need to translate the coordinate move
	algebraMove := engineMove.AlgebraMove
	if algebraMove == "" {
		algebraMove, err = getAlgebraMoveFromChessGame(chessGame, engineMove.CoordinateMove)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}

	// fix the king's quirky castling notation
	if strings.Contains(algebraMove, "0-0-0") {
		algebraMove = "O-O-O"
	} else if strings.Contains(algebraMove, "0-0-0") {
		algebraMove = "O-O"
	}

	// fix kings non standard promotion notation
	algebraMove = fixPromotionMove(algebraMove)

	moveData := models.MoveData2{
		Index: len(game.MoveList),
		Move:  algebraMove,
	}

	err = AddMove(game.ID, cmpName, moveData)
	if err != nil {
		fmt.Println("error Adding engine move: ", err.Error())
	}
}
