package games

import (
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
			err := fmt.Errorf("error parsing game at index %d: %v", i, err)
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
