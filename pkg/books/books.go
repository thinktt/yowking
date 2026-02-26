package books

import (
	"github.com/thinktt/yowking/pkg/models"
	"github.com/thinktt/yowking/pkg/polybook"
)

func GetMove(moves []string, bookName string) (models.MoveData, error) {
	move, err := polybook.GetMove(moves, bookName)
	if err != nil {
		errStr := err.Error()
		return models.MoveData{Err: &errStr}, err
	}

	moveData := models.MoveData{
		CoordinateMove: move,
		WillAcceptDraw: false,
		Type:           "book",
	}

	return moveData, nil
}
