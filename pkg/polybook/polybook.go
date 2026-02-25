package polybook

import (
	"fmt"
	"os"
	"path/filepath"

	chess "github.com/corentings/chess/v2"
)

// BookMove mirrors the useful fields from the JS polyglot reader output so we
// can compare results during migration.
type BookMove struct {
	Move   string
	Weight uint16
	Learn  uint32
}

// GetAllBookMoves loads a polyglot book from ./books/<bookName> and returns all
// moves available for the given FEN.
func GetAllBookMoves(fen, bookName string) ([]BookMove, error) {
	return GetAllBookMovesFromDir(fen, "books", bookName)
}

// GetAllBookMovesFromDir loads a polyglot book from <booksDir>/<bookName> and
// returns all moves available for the given FEN.
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
			Learn:  entry.Learn,
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
