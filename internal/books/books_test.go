package books

import "testing"

func TestFENFromMovesAppliesUCIMoves(t *testing.T) {
	moves := []string{"d2d4", "b8c6", "c2c4", "b7b5", "c4b5"}

	fen, err := FENFromMoves(moves)
	if err != nil {
		t.Fatal(err)
	}

	want := "r1bqkbnr/p1pppppp/2n5/1P6/3P4/8/PP2PPPP/RNBQKBNR b KQkq - 0 3"
	if fen != want {
		t.Fatalf("FENFromMoves() = %q, want %q", fen, want)
	}
}

func TestFENFromMovesRejectsAlgebraicMoves(t *testing.T) {
	_, err := FENFromMoves([]string{"e4"})
	if err == nil {
		t.Fatal("expected algebraic move to be rejected")
	}
}

func TestFENFromMovesRejectsIllegalUCIMove(t *testing.T) {
	_, err := FENFromMoves([]string{"e2e5"})
	if err == nil {
		t.Fatal("expected illegal UCI move to be rejected")
	}
}
