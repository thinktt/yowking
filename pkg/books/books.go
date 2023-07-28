package books

import (
	"encoding/json"
	"os/exec"

	"github.com/thinktt/yowking/pkg/models"
)

type BookQuery struct {
	Moves []string `json:"moves"`
	Book  string   `json:"book"`
}

func GetMove(moves []string, bookName string) (models.MoveData, error) {
	// fmt.Printf("moves %v", moves)

	bookQuery := BookQuery{
		Moves: moves,
		Book:  bookName,
	}

	// fmt.Printf("query %v\n", bookQuery)

	bookQueryJson, err := json.Marshal(bookQuery)
	if err != nil {
		return models.MoveData{}, err
	}

	cmd := exec.Command("./runbook", string(bookQueryJson))
	out, err := cmd.CombinedOutput()
	if err != nil {
		// fmt.Println("book:", string(out))
		errStr := string(out)
		return models.MoveData{Err: &errStr}, err
	}

	moveData := models.MoveData{
		CoordinateMove: string(out),
		WillAcceptDraw: false,
		Type:           "book",
	}

	return moveData, nil
}
