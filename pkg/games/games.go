package games

import (
	"errors"
	"fmt"

	"github.com/notnil/chess"
)

func CheckMoves(moves []string) error {
	gameParser := chess.NewGame()
	for i, move := range moves {
		err := gameParser.MoveStr(move)
		if err != nil {
			errMsg := fmt.Sprintf("Invalid move at index %d: %v", i, err)
			return errors.New(errMsg)
		}
	}
	return nil
}
