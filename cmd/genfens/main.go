package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	chess "github.com/corentings/chess/v2"
	"github.com/thinktt/yowking/pkg/models"
)

type testGame struct {
	ID       string   `json:"id"`
	Opponent string   `json:"opponent"`
	PlayedAs string   `json:"playedAs"`
	Moves    []string `json:"moves"`
}

type testFENCase struct {
	GameID   string `json:"gameId"`
	CmpName  string `json:"cmpName"`
	Book     string `json:"book"`
	PlayedAs string `json:"playedAs,omitempty"`
	Ply      int    `json:"ply"`
	FEN      string `json:"fen"`
	Move     string `json:"move"`
}

type testFENOutput struct {
	SourceGames       string        `json:"sourceGames"`
	SourcePersonalities string      `json:"sourcePersonalities"`
	GamesCount        int           `json:"gamesCount"`
	MaxPlyPerGame     int           `json:"maxPlyPerGame"`
	CasesCount        int           `json:"casesCount"`
	Cases             []testFENCase `json:"cases"`
}

func main() {
	gamesPath := flag.String("games", "notes/testGames.json", "path to test games JSON")
	personalitiesPath := flag.String("personalities", "dist/personalities.json", "path to personalities.json")
	outPath := flag.String("out", "notes/testFens.json", "output JSON path")
	maxPly := flag.Int("plies", 6, "number of plies to export per game")
	flag.Parse()

	if *maxPly <= 0 {
		fatalf("plies must be > 0")
	}

	games := mustLoadGames(*gamesPath)
	cmps := mustLoadPersonalities(*personalitiesPath)

	cases := make([]testFENCase, 0, len(games)*(*maxPly))
	for _, g := range games {
		cmp, ok := cmps[g.Opponent]
		if !ok {
			fatalf("personality %q (game %s) not found in %s", g.Opponent, g.ID, *personalitiesPath)
		}
		if cmp.Book == "" {
			fatalf("personality %q (game %s) has empty book", g.Opponent, g.ID)
		}

		game := chess.NewGame()
		limit := min(*maxPly, len(g.Moves))
		for ply := 0; ply < limit; ply++ {
			move := g.Moves[ply]
			if err := game.PushMove(move, nil); err != nil {
				fatalf("game %s ply %d move %q invalid: %v", g.ID, ply+1, move, err)
			}

			cases = append(cases, testFENCase{
				GameID:   g.ID,
				CmpName:  g.Opponent,
				Book:     cmp.Book,
				PlayedAs: g.PlayedAs,
				Ply:      ply + 1,
				FEN:      game.FEN(),
				Move:     move,
			})
		}
	}

	out := testFENOutput{
		SourceGames:         *gamesPath,
		SourcePersonalities: *personalitiesPath,
		GamesCount:          len(games),
		MaxPlyPerGame:       *maxPly,
		CasesCount:          len(cases),
		Cases:               cases,
	}

	if err := os.MkdirAll(filepath.Dir(*outPath), 0o755); err != nil {
		fatalf("create output dir: %v", err)
	}

	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		fatalf("marshal output: %v", err)
	}
	b = append(b, '\n')
	if err := os.WriteFile(*outPath, b, 0o644); err != nil {
		fatalf("write %s: %v", *outPath, err)
	}

	fmt.Printf("wrote %d cases from %d games to %s\n", len(cases), len(games), *outPath)
}

func mustLoadGames(path string) []testGame {
	b, err := os.ReadFile(path)
	if err != nil {
		fatalf("read games file %s: %v", path, err)
	}
	var games []testGame
	if err := json.Unmarshal(b, &games); err != nil {
		fatalf("parse games file %s: %v", path, err)
	}
	return games
}

func mustLoadPersonalities(path string) map[string]models.Cmp {
	b, err := os.ReadFile(path)
	if err != nil {
		fatalf("read personalities file %s: %v", path, err)
	}
	var cmps map[string]models.Cmp
	if err := json.Unmarshal(b, &cmps); err != nil {
		fatalf("parse personalities file %s: %v", path, err)
	}
	return cmps
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
