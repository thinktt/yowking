package books

import (
	"fmt"
	"os"
	"testing"
)

// anything outside dependencies across the app will be in this folder
// this is for tests with dependecies, any other unit tests shouldn't
// this shouldn't matter
func init() {
	rootDir := os.Getenv("ROOT_DIR")
	err := os.Chdir(rootDir + "/dist")
	if err != nil {
		fmt.Println(err)
	}
	pwd, _ := os.Getwd()
	fmt.Println("working in", pwd)
}

func TestGetMove_NoBookMove(t *testing.T) {
	testMoves := []string{"d2d4", "e7e6", "e2e4", "d7d5", "a2a3"}
	move, err := GetMove(testMoves, "EarlyQueen.bin")
	fmt.Println(err)
	if move.Err != nil && *move.Err != "no book move" {
		t.Errorf("error message should be no book string instead %s", *move.Err)
	}
}

func TestGetMove_BookHasMoove(t *testing.T) {
	testMoves := []string{}
	move, err := GetMove(testMoves, "TalM.bin")
	fmt.Println(err)
	if move.CoordinateMove == "" {
		t.Errorf("expected a move instead got none")
	}
}
