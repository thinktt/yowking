package games

import (
	"errors"
	"fmt"
	"strings"

	"github.com/notnil/chess"
	"github.com/thinktt/yowking/pkg/models"
)

// GetTurnColor given a move index returns the color that move belongs to
func GetTurnColor(moveIndex int) string {
	if moveIndex%2 == 0 {
		return "white"
	}
	return "black"
}

func GetUsercolor(game models.Game2, user string) string {
	if user == game.WhitePlayer.ID {
		return "white"
	}

	if user == game.BlackPlayer.ID {
		return "black"
	}

	// this user is not playing this game
	return ""
}

func ParseGame(game models.Game2) (*chess.Game, error) {
	chessGame := chess.NewGame()
	for i, move := range game.MoveList {
		err := chessGame.MoveStr(move)
		if err != nil {
			err := fmt.Errorf("Error parsing game at index %d: %v", i, err)
			return nil, err
		}
	}

	return chessGame, nil

}

func GetGameStatus(chessGame *chess.Game) (status string, winner string) {

	chessGame.Draw(chess.FiftyMoveRule)
	chessGame.Draw(chess.ThreefoldRepetition)

	ending := chessGame.Method()

	status = "started"
	switch ending {
	case chess.Checkmate:
		status = "mate"
		if chessGame.Outcome() == chess.WhiteWon {
			winner = "white"
		} else {
			winner = "black"
		}
	case chess.Stalemate:
		status = "stalemate"
		winner = "none"
	case chess.InsufficientMaterial:
		status = "draw"
		winner = "none"
	case chess.FiftyMoveRule:
		status = "draw"
		winner = "none"
	case chess.ThreefoldRepetition:
		status = "draw"
		winner = "none"
	}

	return
}

func GetProperLastMove(chessGame *chess.Game) (string, error) {
	pgn := chessGame.String()
	pgnSlice := strings.Fields(pgn)
	if len(pgnSlice) < 3 {
		err := fmt.Errorf("unable parse move from chessGame: PGN too short")
		return "", err
	}

	move := pgnSlice[len(pgnSlice)-2]
	return move, nil
}

func CheckMoves(moves []string) (string, error) {
	chessGame := chess.NewGame()
	for i, move := range moves {
		err := chessGame.MoveStr(move)
		if err != nil {
			errMsg := fmt.Sprintf("Invalid move at index %d: %v", i, err)
			return "", errors.New(errMsg)
		}
	}

	drawMethods := chessGame.EligibleDraws()
	ending := chessGame.Method()

	status := "started"
	switch ending {
	case chess.Checkmate:
		status = "mate"
	case chess.Stalemate:
		status = "stalemate"
	case chess.InsufficientMaterial:
		status = "draw"
	}

	fmt.Println("Ways to draw:", drawMethods)
	fmt.Println("Game ended as: ", ending)

	moveList := chessGame.String()
	algebraMoves := strings.Fields(moveList)
	// lastMove := algebraMoves[len(algebraMoves)-1]
	fmt.Println(moves)
	fmt.Println()
	fmt.Println(algebraMoves)

	// // find the last move in the list of moves, this loop fixes issue
	// // when there are extra spaces
	// lastMove := ""
	// for i := len(algebraMoves) - 2; i >= 0; i-- {
	// 	if algebraMoves[i] != "" {
	// 		lastMove = algebraMoves[i]
	// 		break
	// 	}
	// }

	// lastOriginalMove := moves[len(moves)-1]

	// if lastMove != lastOriginalMove {
	// 	err := fmt.Errorf(
	// 		"strict move notation enforced: wanted %s, got %s",
	// 		lastMove, lastOriginalMove,
	// 	)
	// 	return "", err
	// }

	return status, nil

}
