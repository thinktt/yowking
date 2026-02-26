package books

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	chess "github.com/corentings/chess/v2"
	"github.com/thinktt/yowking/pkg/models"
)

// BookMove mirrors the useful polyglot fields for regression/testing.
type BookMove struct {
	Move   string
	Weight uint16
}

var ErrNoBookMove = errors.New("no book move")

// GetMove applies the provided move list and returns a weighted-random move from the named book.
func GetMove(moves []string, bookName string) (models.MoveData, error) {
	move, err := HeavyMoveFromMoves(moves, bookName)
	if err != nil {
		errStr := err.Error()
		return models.MoveData{Err: &errStr}, err
	}

	return models.MoveData{
		CoordinateMove: move,
		WillAcceptDraw: false,
		Type:           "book",
	}, nil
}

func HeavyMoveFromMoves(moves []string, bookName string) (string, error) {
	fen, err := FENFromMoves(moves)
	if err != nil {
		return "", err
	}
	return HeavyMoveFromFEN(fen, bookName)
}

// FENFromMoves applies a move list from the initial position and returns the resulting FEN.
func FENFromMoves(moves []string) (string, error) {
	g := chess.NewGame()
	for i, s := range moves {
		if err := pushSloppy(g, s); err != nil {
			return "", fmt.Errorf("apply move %d (%q): %w", i+1, s, err)
		}
	}
	return g.FEN(), nil
}

// HeavyMoveFromFEN selects a weighted-random move, ignoring zero-weight entries.
func HeavyMoveFromFEN(fen, bookName string) (string, error) {
	bookMoves, err := GetAllBookMoves(fen, bookName)
	if err != nil {
		return "", err
	}
	if len(bookMoves) == 0 {
		return "", ErrNoBookMove
	}

	weighted := make([]string, 0)
	for _, m := range bookMoves {
		for i := uint16(0); i < m.Weight; i++ {
			weighted = append(weighted, m.Move)
		}
	}
	if len(weighted) == 0 {
		return "", ErrNoBookMove
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return weighted[r.Intn(len(weighted))], nil
}

// GetAllBookMoves loads a polyglot book from ./books/<bookName> and returns all moves for the FEN.
func GetAllBookMoves(fen, bookName string) ([]BookMove, error) {
	return GetAllBookMovesFromDir(fen, "books", bookName)
}

// GetAllBookMovesFromDir loads a polyglot book from <booksDir>/<bookName> and returns all moves for the FEN.
func GetAllBookMovesFromDir(fen, booksDir, bookName string) ([]BookMove, error) {
	bookPath := filepath.Join(booksDir, bookName)

	f, err := os.Open(bookPath)
	if err != nil {
		return nil, fmt.Errorf("open polyglot book %q: %w", bookPath, err)
	}
	defer f.Close()

	book, err := chess.LoadFromReader(f)
	if err != nil {
		return nil, fmt.Errorf("load polyglot book %q: %w", bookPath, err)
	}

	hasher := chess.NewChessHasher()
	hashHex, err := hasher.HashPosition(fen)
	if err != nil {
		return nil, fmt.Errorf("hash fen: %w", err)
	}

	entries := book.FindMoves(chess.ZobristHashToUint64(hashHex))
	moves := make([]BookMove, 0, len(entries))
	for _, entry := range entries {
		moveStr, err := polyglotEntryToUCIMove(entry)
		if err != nil {
			return nil, err
		}
		moves = append(moves, BookMove{
			Move:   moveStr,
			Weight: entry.Weight,
		})
	}

	return moves, nil
}

func polyglotEntryToUCIMove(entry chess.PolyglotEntry) (string, error) {
	pm := chess.DecodeMove(entry.Move)
	move := pm.ToMove()

	from := move.S1().String()
	to := move.S2().String()
	if from == "" || to == "" {
		return "", fmt.Errorf("decode polyglot move %#x", entry.Move)
	}

	uci := from + to
	switch move.Promo() {
	case chess.Knight:
		uci += "n"
	case chess.Bishop:
		uci += "b"
	case chess.Rook:
		uci += "r"
	case chess.Queen:
		uci += "q"
	}

	return uci, nil
}

func pushSloppy(g *chess.Game, s string) error {
	if err := g.PushMove(s, nil); err == nil {
		return nil
	}
	if err := g.PushNotationMove(s, chess.UCINotation{}, nil); err == nil {
		return nil
	}
	if err := g.PushNotationMove(s, chess.LongAlgebraicNotation{}, nil); err == nil {
		return nil
	}
	if err := g.PushNotationMove(s, chess.AlgebraicNotation{}, nil); err == nil {
		return nil
	}
	return fmt.Errorf("invalid move")
}
