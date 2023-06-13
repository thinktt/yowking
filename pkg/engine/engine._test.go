package engine

import (
	"fmt"
	"os"
	"testing"

	"github.com/thinktt/yowking/pkg/models"
	"github.com/thinktt/yowking/pkg/personalities"
)

// anything outside dependencies across the app will be in this folder
// this is for tests with dependecies, for any other unit tests this
// shouldn't matter
func init() {
	rootDir := os.Getenv("ROOT_DIR")
	err := os.Chdir(rootDir + "/dist")
	if err != nil {
		fmt.Println(err)
	}
	pwd, _ := os.Getwd()
	fmt.Println("working in", pwd)
}

func TestGetMove_BadMoves(t *testing.T) {
	testMoves := []string{"d2d4", "e7e6", "e2e4", "d7d5", "a2a3"}
	cmp := personalities.CmpMap["Tal"]
	settings := models.Settings{
		CmpVals: cmp.Vals,
		Moves:   testMoves,
	}
	move, err := GetMove(settings)
	if err != nil {
		t.Error(err)
	}

	if move.CoordinateMove == "" {
		t.Errorf("expected a move instead got none")
	}

}

func TestGetMove_BadMove(t *testing.T) {
	testMoves := []string{"d2d4", "e7e7", "e2e4", "d7d5", "a2a3"}
	cmp := personalities.CmpMap["Tal"]
	settings := models.Settings{
		CmpVals: cmp.Vals,
		Moves:   testMoves,
	}
	move, _ := GetMove(settings)
	if move.Err == nil {
		t.Errorf("error message instead got none")
	}

	if move.CoordinateMove != "" {
		t.Errorf("expected empty move instead got %v", move.CoordinateMove)
	}

}
