package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/thinktt/yowking/pkg/polybook"
)

var ExtraBooks = []string{
	"PawnMoves.bin",
	"AnandV.bin",
	"AnderssenA.bin",
	"BlackburneJ.bin",
	"BogoljubowE.bin",
	"BotvinnikM.bin",
	"ChigorinM.bin",
	"EuweM.bin",
	"EvansL.bin",
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
	"PetrosianT.bin",
	"PillsburyH.bin",
	"PolgarJ.bin",
	"ReshevskyS.bin",
	"RetiR.bin",
	"SeirawanY.bin",
	"ShirovA.bin",
	"ShortN.bin",
	"SmyslovV.bin",
	"SpasskyB.bin",
	"SteinitzW.bin",
	"TarraschS.bin",
	"TartakowerS.bin",
	"TimmanJ.bin",
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
	Scenario string  `json:"scenario"` // native | rotated
	Source   fenCase `json:"source"`
	TestBook string  `json:"testBook"`
}

type moveOut struct {
	Move   string `json:"move"`
	Weight uint16 `json:"weight"`
	Learn  uint32 `json:"learn"`
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
	JSRunbookDir      string       `json:"jsRunbookDir,omitempty"`
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

type compareSummary struct {
	GeneratedAt      time.Time        `json:"generatedAt"`
	JSFile           string           `json:"jsFile"`
	GoFile           string           `json:"goFile"`
	TotalCases       int              `json:"totalCases"`
	ExactMatches     int              `json:"exactMatches"`
	Mismatches       int              `json:"mismatches"`
	JSErrors         int              `json:"jsErrors"`
	GoErrors         int              `json:"goErrors"`
	JSOnlyErrors     int              `json:"jsOnlyErrors"`
	GoOnlyErrors     int              `json:"goOnlyErrors"`
	BothErrors       int              `json:"bothErrors"`
	JSTotalMS        float64          `json:"jsTotalDurationMs"`
	GoTotalMS        float64          `json:"goTotalDurationMs"`
	Speedup          float64          `json:"goSpeedupVsJs"`
	IssueBreakdown   map[string]int   `json:"issueBreakdown"`
	SampleMismatches []compareProblem `json:"sampleMismatches"`
}

type compareProblem struct {
	ID       string     `json:"id"`
	Scenario string     `json:"scenario"`
	TestBook string     `json:"testBook"`
	FEN      string     `json:"fen"`
	JS       caseResult `json:"js"`
	Go       caseResult `json:"go"`
	Reason   string     `json:"reason"`
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "run-js":
		if err := runJS(os.Args[2:]); err != nil {
			fatal(err)
		}
	case "run-go":
		if err := runGo(os.Args[2:]); err != nil {
			fatal(err)
		}
	case "compare":
		if err := runCompare(os.Args[2:]); err != nil {
			fatal(err)
		}
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: booktester <run-js|run-go|compare> [flags]\n")
}

func runJS(args []string) error {
	fs := flag.NewFlagSet("run-js", flag.ExitOnError)
	in := fs.String("in", "notes/testFens.json", "input test fens json")
	out := fs.String("out", "notes/booktest-js.json", "output results json")
	booksDir := fs.String("books-dir", "dist/books", "books directory")
	runbookDir := fs.String("runbook-dir", "assets/runbook", "runbook directory")
	extraBooks := fs.String("extra-books", strings.Join(ExtraBooks, ","), "comma-separated extra books for rotated pass")
	_ = fs.Parse(args)

	absRunbookDir, err := filepath.Abs(*runbookDir)
	if err != nil {
		return fmt.Errorf("resolve runbook dir: %w", err)
	}
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
		logCaseStart("js", c)
		res, err := jsCase(c, i, absRunbookDir, absBooksDir)
		if err != nil {
			return err
		}
		logCaseResult(res)
		if res.Err != "" {
			errCount++
		} else if len(res.Moves) == 0 {
			noMoves++
			bookSeenOK[res.TestBook] = true
		} else {
			withMoves++
			bookSeenOK[res.TestBook] = true
			bookHadMove[res.TestBook] = true
		}
		results = append(results, res)
	}
	totalMS := msSince(startAll)

	sum := runSummary{
		Engine:            "js",
		GeneratedAt:       time.Now().UTC(),
		Input:             *in,
		BooksDir:          absBooksDir,
		JSRunbookDir:      absRunbookDir,
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

func runGo(args []string) error {
	fs := flag.NewFlagSet("run-go", flag.ExitOnError)
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
		logCaseStart("go", c)
		start := time.Now()
		moves, runErr := polybook.GetAllBookMovesFromDir(c.Source.FEN, absBooksDir, c.TestBook)
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
				res.Moves = append(res.Moves, moveOut{Move: m.Move, Weight: m.Weight, Learn: m.Learn})
			}
			if len(res.Moves) == 0 {
				noMoves++
				bookSeenOK[res.TestBook] = true
			} else {
				withMoves++
				bookSeenOK[res.TestBook] = true
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

func runCompare(args []string) error {
	fs := flag.NewFlagSet("compare", flag.ExitOnError)
	jsPath := fs.String("js", "notes/booktest-js.json", "js results file")
	goPath := fs.String("go", "notes/booktest-go.json", "go results file")
	out := fs.String("out", "notes/booktest-compare.json", "comparison output json")
	sample := fs.Int("sample", 20, "max mismatches to include")
	_ = fs.Parse(args)

	var jsRun, goRun runSummary
	if err := readJSON(*jsPath, &jsRun); err != nil {
		return fmt.Errorf("read js run: %w", err)
	}
	if err := readJSON(*goPath, &goRun); err != nil {
		return fmt.Errorf("read go run: %w", err)
	}
	if len(jsRun.Results) != len(goRun.Results) {
		return fmt.Errorf("result length mismatch: js=%d go=%d", len(jsRun.Results), len(goRun.Results))
	}

	jsErrors, goErrors := 0, 0
	jsOnlyErr, goOnlyErr, bothErr := 0, 0, 0
	exact := 0
	mismatches := make([]compareProblem, 0)
	issueCounts := map[string]int{}

	for i := range jsRun.Results {
		jr := jsRun.Results[i]
		gr := goRun.Results[i]
		if jr.Err != "" {
			jsErrors++
		}
		if gr.Err != "" {
			goErrors++
		}
		switch {
		case jr.Err != "" && gr.Err != "":
			bothErr++
		case jr.Err != "" && gr.Err == "":
			jsOnlyErr++
		case jr.Err == "" && gr.Err != "":
			goOnlyErr++
		}

		if jr.ID != gr.ID {
			mismatches = append(mismatches, compareProblem{
				ID:       jr.ID,
				Scenario: jr.Scenario,
				TestBook: jr.TestBook,
				FEN:      jr.FEN,
				JS:       jr,
				Go:       gr,
				Reason:   "result ordering/id mismatch",
			})
			issueCounts["result ordering/id mismatch"]++
			continue
		}

		reason := compareCase(jr, gr)
		if reason == "" {
			exact++
			continue
		}
		issueCounts[reason]++
		if len(mismatches) < *sample {
			mismatches = append(mismatches, compareProblem{
				ID:       jr.ID,
				Scenario: jr.Scenario,
				TestBook: jr.TestBook,
				FEN:      jr.FEN,
				JS:       jr,
				Go:       gr,
				Reason:   reason,
			})
		}
	}

	total := len(jsRun.Results)
	sum := compareSummary{
		GeneratedAt:      time.Now().UTC(),
		JSFile:           *jsPath,
		GoFile:           *goPath,
		TotalCases:       total,
		ExactMatches:     exact,
		Mismatches:       total - exact,
		JSErrors:         jsErrors,
		GoErrors:         goErrors,
		JSOnlyErrors:     jsOnlyErr,
		GoOnlyErrors:     goOnlyErr,
		BothErrors:       bothErr,
		JSTotalMS:        jsRun.TotalDurationMS,
		GoTotalMS:        goRun.TotalDurationMS,
		Speedup:          safeDiv(jsRun.TotalDurationMS, goRun.TotalDurationMS),
		IssueBreakdown:   issueCounts,
		SampleMismatches: mismatches,
	}

	if err := writeJSON(*out, sum); err != nil {
		return err
	}

	fmt.Printf("cases=%d exact=%d mismatches=%d js_ms=%.2f go_ms=%.2f speedup=%.2fx\n",
		sum.TotalCases, sum.ExactMatches, sum.Mismatches, sum.JSTotalMS, sum.GoTotalMS, sum.Speedup)
	printCompareSummary(sum)
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
	for i, c := range input.Cases {
		out = append(out, runCase{
			ID:       fmt.Sprintf("native:%s:%03d", c.GameID, c.Ply),
			Scenario: "native",
			Source:   c,
			TestBook: c.Book,
		})
		_ = i
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

func jsCase(c runCase, idx int, runbookDir, booksDir string) (caseResult, error) {
	payload := map[string]string{
		"fen":      c.Source.FEN,
		"book":     c.TestBook,
		"booksDir": booksDir,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return caseResult{}, err
	}

	cmd := exec.Command("node", "dumpBookMoves.js", string(b))
	cmd.Dir = runbookDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	start := time.Now()
	out, err := cmd.Output()
	duration := msSince(start)

	res := caseResult{
		ID:          c.ID,
		Scenario:    c.Scenario,
		GameID:      c.Source.GameID,
		Ply:         c.Source.Ply,
		CmpName:     c.Source.CmpName,
		SourceBook:  c.Source.Book,
		TestBook:    c.TestBook,
		FEN:         c.Source.FEN,
		DurationMS:  duration,
		SourceMove:  c.Source.Move,
		SourceSide:  c.Source.PlayedAs,
		SourceIndex: idx,
	}

	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		res.Err = msg
		return res, nil
	}

	var moves []moveOut
	if err := json.Unmarshal(out, &moves); err != nil {
		return caseResult{}, fmt.Errorf("parse js output for %s: %w; raw=%q", c.ID, err, string(out))
	}
	res.Moves = moves
	return res, nil
}

func compareCase(jsR, goR caseResult) string {
	if jsR.Err != "" || goR.Err != "" {
		if jsR.Err == goR.Err {
			return ""
		}
		return "error mismatch"
	}
	jsMoves := normalizeMoves(jsR.Moves)
	goMoves := normalizeMoves(goR.Moves)
	if len(jsMoves) != len(goMoves) {
		return "move count mismatch"
	}
	for i := range jsMoves {
		if jsMoves[i] != goMoves[i] {
			return "move set/weight mismatch"
		}
	}
	return ""
}

func normalizeMoves(in []moveOut) []moveOut {
	out := append([]moveOut(nil), in...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Move != out[j].Move {
			return out[i].Move < out[j].Move
		}
		if out[i].Weight != out[j].Weight {
			return out[i].Weight < out[j].Weight
		}
		return out[i].Learn < out[j].Learn
	})
	return out
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

func logCaseStart(engine string, c runCase) {
	_ = engine
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

func printCompareSummary(sum compareSummary) {
	if sum.Mismatches == 0 {
		fmt.Println("issues none")
		return
	}
	keys := make([]string, 0, len(sum.IssueBreakdown))
	for k := range sum.IssueBreakdown {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s:%d", k, sum.IssueBreakdown[k]))
	}
	fmt.Printf("issues %s\n", strings.Join(parts, " | "))
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
