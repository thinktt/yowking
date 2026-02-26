package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	chess "github.com/corentings/chess/v2"
	"github.com/thinktt/yowking/pkg/books"
)

var ExtraBooks = []string{
	"AnandV.bin",
	"AnderssenA.bin",
	"BlackburneJ.bin",
	"BogoljubowE.bin",
	"BotvinnikM.bin",
	"CaptureBook.bin",
	"ChigorinM.bin",
	"EuweM.bin",
	"EvansL.bin",
	"FastBook.bin",
	"FineR.bin",
	"FlohrS.bin",
	"GellerE.bin",
	"IvanchukV.bin",
	"KamskyG.bin",
	"KarpovA.bin",
	"KashdanI.bin",
	"KeresP.bin",
	"KorchnoiV.bin",
	"KramnikV.bin",
	"LarsenB.bin",
	"LaskerE.bin",
	"LekoP.bin",
	"MarshallF.bin",
	"NajdorfM.bin",
	"NimzowitschA.bin",
	"PaulsenL.bin",
	"PawnMoves.bin",
	"PetrosianT.bin",
	"PillsburyH.bin",
	"PolgarJ.bin",
	"ReshevskyS.bin",
	"RetiR.bin",
	"SeirawanY.bin",
	"ShirovA.bin",
	"ShortN.bin",
	"SlowBook.bin",
	"SmyslovV.bin",
	"SpasskyB.bin",
	"SteinitzW.bin",
	"Strong.bin",
	"TarraschS.bin",
	"TartakowerS.bin",
	"TimmanJ.bin",
	"Trapper.bin",
	"Unorthodox.bin",
}

type fenCaseFile struct {
	Cases []fenCase `json:"cases"`
}

type fenCase struct {
	GameID   string `json:"gameId"`
	CmpName  string `json:"cmpName"`
	Book     string `json:"book"`
	PlayedAs string `json:"playedAs"`
	Ply      int    `json:"ply"`
	FEN      string `json:"fen"`
	Move     string `json:"move"`
}

type runCase struct {
	ID       string  `json:"id"`
	Scenario string  `json:"scenario"`
	Source   fenCase `json:"source"`
	TestBook string  `json:"testBook"`
}

type moveOut struct {
	Move   string `json:"move"`
	Weight uint16 `json:"weight"`
}

type caseResult struct {
	ID          string    `json:"id"`
	Scenario    string    `json:"scenario"`
	GameID      string    `json:"gameId"`
	Ply         int       `json:"ply"`
	CmpName     string    `json:"cmpName"`
	SourceBook  string    `json:"sourceBook"`
	TestBook    string    `json:"testBook"`
	FEN         string    `json:"fen"`
	DurationMS  float64   `json:"durationMs"`
	Err         string    `json:"err,omitempty"`
	Moves       []moveOut `json:"moves,omitempty"`
	SourceMove  string    `json:"sourceMove,omitempty"`
	SourceSide  string    `json:"sourceSide,omitempty"`
	SourceIndex int       `json:"sourceIndex"`
}

type runSummary struct {
	Engine            string       `json:"engine"`
	GeneratedAt       time.Time    `json:"generatedAt"`
	Input             string       `json:"input"`
	BooksDir          string       `json:"booksDir"`
	ExtraBooks        []string     `json:"extraBooks"`
	NativeCases       int          `json:"nativeCases"`
	RotatedCases      int          `json:"rotatedCases"`
	TotalCases        int          `json:"totalCases"`
	Errors            int          `json:"errors"`
	CasesWithMoves    int          `json:"casesWithMoves"`
	CasesNoMoves      int          `json:"casesNoMoves"`
	BooksWithMoves    []string     `json:"booksWithMoves"`
	BooksWithNoMoves  []string     `json:"booksWithNoMoves"`
	TotalDurationMS   float64      `json:"totalDurationMs"`
	AverageDurationMS float64      `json:"averageDurationMs"`
	Results           []caseResult `json:"results"`
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "fens", "run-go":
		// "run-go" kept as a compatibility alias.
	case "mem":
		if err := runMem(os.Args[2:]); err != nil {
			fatal(err)
		}
		return
	default:
		usage()
		os.Exit(2)
	}
	if err := runGo(os.Args[2:]); err != nil {
		fatal(err)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: booktester <fens|mem> [flags]")
}

func runGo(args []string) error {
	fs := flag.NewFlagSet("fens", flag.ExitOnError)
	in := fs.String("in", "notes/testFens.json", "input test fens json")
	out := fs.String("out", "notes/booktest-go.json", "output results json")
	booksDir := fs.String("books-dir", "dist/books", "books directory")
	extraBooks := fs.String("extra-books", strings.Join(ExtraBooks, ","), "comma-separated extra books for rotated pass")
	_ = fs.Parse(args)

	absBooksDir, err := filepath.Abs(*booksDir)
	if err != nil {
		return fmt.Errorf("resolve books dir: %w", err)
	}

	cases, native, rotated, extras, err := buildRunCases(*in, parseCSV(*extraBooks))
	if err != nil {
		return err
	}

	startAll := time.Now()
	results := make([]caseResult, 0, len(cases))
	errCount := 0
	withMoves := 0
	noMoves := 0
	bookSeenOK := map[string]bool{}
	bookHadMove := map[string]bool{}

	for i, c := range cases {
		logCaseStart(c)
		start := time.Now()
		moves, runErr := books.GetAllBookMovesFromDir(c.Source.FEN, absBooksDir, c.TestBook)
		res := caseResult{
			ID:          c.ID,
			Scenario:    c.Scenario,
			GameID:      c.Source.GameID,
			Ply:         c.Source.Ply,
			CmpName:     c.Source.CmpName,
			SourceBook:  c.Source.Book,
			TestBook:    c.TestBook,
			FEN:         c.Source.FEN,
			DurationMS:  msSince(start),
			SourceMove:  c.Source.Move,
			SourceSide:  c.Source.PlayedAs,
			SourceIndex: i,
		}
		if runErr != nil {
			res.Err = runErr.Error()
			errCount++
		} else {
			res.Moves = make([]moveOut, 0, len(moves))
			for _, m := range moves {
				res.Moves = append(res.Moves, moveOut{Move: m.Move, Weight: m.Weight})
			}
			bookSeenOK[res.TestBook] = true
			if len(res.Moves) == 0 {
				noMoves++
			} else {
				withMoves++
				bookHadMove[res.TestBook] = true
			}
		}
		logCaseResult(res)
		results = append(results, res)
	}

	totalMS := msSince(startAll)
	sum := runSummary{
		Engine:            "go",
		GeneratedAt:       time.Now().UTC(),
		Input:             *in,
		BooksDir:          absBooksDir,
		ExtraBooks:        extras,
		NativeCases:       native,
		RotatedCases:      rotated,
		TotalCases:        len(results),
		Errors:            errCount,
		CasesWithMoves:    withMoves,
		CasesNoMoves:      noMoves,
		BooksWithMoves:    sortedBookKeys(bookHadMove),
		BooksWithNoMoves:  booksWithNoMoves(bookSeenOK, bookHadMove),
		TotalDurationMS:   totalMS,
		AverageDurationMS: safeDiv(totalMS, float64(len(results))),
		Results:           results,
	}
	if err := writeJSON(*out, sum); err != nil {
		return err
	}
	printRunSummary(sum)
	return nil
}

func runMem(args []string) error {
	fs := flag.NewFlagSet("mem", flag.ExitOnError)
	booksDir := fs.String("books-dir", "dist/books", "books directory")
	_ = fs.Parse(args)

	absBooksDir, err := filepath.Abs(*booksDir)
	if err != nil {
		return fmt.Errorf("resolve books dir: %w", err)
	}

	entries, err := os.ReadDir(absBooksDir)
	if err != nil {
		return fmt.Errorf("read books dir: %w", err)
	}

	type loadedBook struct {
		name string
		book *chess.PolyglotBook
	}
	loaded := make([]loadedBook, 0)
	var diskBytes int64

	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".bin") {
			continue
		}
		p := filepath.Join(absBooksDir, e.Name())
		info, err := e.Info()
		if err != nil {
			return fmt.Errorf("stat %s: %w", p, err)
		}
		diskBytes += info.Size()

		f, err := os.Open(p)
		if err != nil {
			return fmt.Errorf("open %s: %w", p, err)
		}
		b, err := chess.LoadFromReader(f)
		f.Close()
		if err != nil {
			return fmt.Errorf("load %s: %w", p, err)
		}
		loaded = append(loaded, loadedBook{name: e.Name(), book: b})
	}

	runtime.GC()
	var after runtime.MemStats
	runtime.ReadMemStats(&after)

	heapDelta := memDelta(after.HeapAlloc, before.HeapAlloc)
	allocDelta := memDelta(after.Alloc, before.Alloc)
	sysDelta := memDelta(after.Sys, before.Sys)

	fmt.Printf("books=%d disk_bytes=%d disk_mb=%.2f\n", len(loaded), diskBytes, float64(diskBytes)/(1024*1024))
	fmt.Printf("heap_alloc_delta=%d (%.2f MB)\n", heapDelta, float64(heapDelta)/(1024*1024))
	fmt.Printf("alloc_delta=%d (%.2f MB)\n", allocDelta, float64(allocDelta)/(1024*1024))
	fmt.Printf("sys_delta=%d (%.2f MB)\n", sysDelta, float64(sysDelta)/(1024*1024))

	// Keep parsed books alive until after stats are printed.
	if len(loaded) == 0 {
		return errors.New("no .bin books loaded")
	}
	return nil
}

func buildRunCases(in string, extraBooks []string) ([]runCase, int, int, []string, error) {
	var input fenCaseFile
	if err := readJSON(in, &input); err != nil {
		return nil, 0, 0, nil, fmt.Errorf("read input %s: %w", in, err)
	}
	if len(input.Cases) == 0 {
		return nil, 0, 0, nil, errors.New("no cases in input")
	}

	out := make([]runCase, 0, len(input.Cases)*2)
	for _, c := range input.Cases {
		out = append(out, runCase{
			ID:       fmt.Sprintf("native:%s:%03d", c.GameID, c.Ply),
			Scenario: "native",
			Source:   c,
			TestBook: c.Book,
		})
	}
	nativeCount := len(out)

	if len(extraBooks) == 0 {
		return out, nativeCount, 0, nil, nil
	}

	gameBook := make(map[string]string)
	gameOrder := make([]string, 0)
	for _, c := range input.Cases {
		if _, ok := gameBook[c.GameID]; ok {
			continue
		}
		book := extraBooks[len(gameOrder)%len(extraBooks)]
		gameBook[c.GameID] = book
		gameOrder = append(gameOrder, c.GameID)
	}

	rotatedCount := 0
	for _, c := range input.Cases {
		testBook := gameBook[c.GameID]
		out = append(out, runCase{
			ID:       fmt.Sprintf("rotated:%s:%03d:%s", c.GameID, c.Ply, testBook),
			Scenario: "rotated",
			Source:   c,
			TestBook: testBook,
		})
		rotatedCount++
	}
	return out, nativeCount, rotatedCount, extraBooks, nil
}

func parseCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func readJSON(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func writeJSON(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(path, b, 0o644)
}

func msSince(t time.Time) float64 {
	return float64(time.Since(t).Microseconds()) / 1000.0
}

func safeDiv(a, b float64) float64 {
	if b == 0 {
		return 0
	}
	return a / b
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func logCaseStart(c runCase) {
	fmt.Printf("test fen: %s\n", c.Source.FEN)
}

func logCaseResult(res caseResult) {
	if res.Err != "" {
		fmt.Printf("error: %s\n", res.Err)
		return
	}
	fmt.Printf("moves %s\n", formatMoves(res.Moves))
}

func formatMoves(moves []moveOut) string {
	if len(moves) == 0 {
		return "(none)"
	}
	parts := make([]string, 0, len(moves))
	for _, m := range moves {
		parts = append(parts, fmt.Sprintf("%s:%d", m.Move, m.Weight))
	}
	return strings.Join(parts, " ")
}

func printRunSummary(sum runSummary) {
	fmt.Printf(
		"summary %s cases=%d native=%d rotated=%d errors=%d with_moves=%d no_moves=%d books_with_moves=%d books_with_no_moves=%d total_ms=%.2f avg_ms=%.2f\n",
		sum.Engine,
		sum.TotalCases,
		sum.NativeCases,
		sum.RotatedCases,
		sum.Errors,
		sum.CasesWithMoves,
		sum.CasesNoMoves,
		len(sum.BooksWithMoves),
		len(sum.BooksWithNoMoves),
		sum.TotalDurationMS,
		sum.AverageDurationMS,
	)
	if len(sum.BooksWithNoMoves) > 0 {
		fmt.Printf("books_with_no_moves %s\n", strings.Join(sum.BooksWithNoMoves, " "))
	}
}

func sortedBookKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func booksWithNoMoves(bookSeenOK, bookHadMove map[string]bool) []string {
	out := make([]string, 0)
	for book := range bookSeenOK {
		if !bookHadMove[book] {
			out = append(out, book)
		}
	}
	sort.Strings(out)
	return out
}

func memDelta(a, b uint64) uint64 {
	if a >= b {
		return a - b
	}
	return 0
}
