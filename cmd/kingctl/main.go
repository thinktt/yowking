package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/thinktt/yowking/internal/booktester"
	"github.com/thinktt/yowking/pkg/books"
	"github.com/thinktt/yowking/pkg/engine"
	"github.com/thinktt/yowking/pkg/models"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "help", "-h", "--help":
		printUsage()
	case "move":
		if err := runMove(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "book":
		baseDir, err := binaryDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := booktester.Run(os.Args[2:], baseDir); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(2)
	}
}

func printUsage() {
	fmt.Println("kingctl - yowking debug CLI")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  kingctl move <json>")
	fmt.Println("  kingctl book <fens|mem>")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  move    Run move resolution directly (book + engine), no NATS")
	fmt.Println("  book    Run book tests/memory checks")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println(`  kingctl move '{"cmpName":"Wizard","gameId":"g1","moves":["e2e4"]}'`)
	fmt.Println(`  kingctl move --skip-book '{"cmpName":"Wizard","gameId":"g1","moves":["e2e4"]}'`)
	fmt.Println(`  kingctl book fens`)
	fmt.Println(`  kingctl book mem`)
}

func runMove(args []string) error {
	fs := flag.NewFlagSet("move", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	skipBook := fs.Bool("s", false, "skip book lookup")
	fs.BoolVar(skipBook, "skip-book", false, "skip book lookup")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() != 1 {
		return errors.New("usage: kingctl move [-s|--skip-book] <json>")
	}

	req, err := readMoveReq(fs.Arg(0))
	if err != nil {
		return err
	}
	if *skipBook {
		req.ShouldSkipBook = true
	}
	if req.GameId == "" {
		req.GameId = fmt.Sprintf("kingctl-%d", time.Now().UnixNano())
	}

	baseDir, err := binaryDir()
	if err != nil {
		return err
	}
	loadDotEnvDefaults()
	if err := os.Chdir(baseDir); err != nil {
		return fmt.Errorf("change dir to %q: %w", baseDir, err)
	}

	cmpMap, err := loadCmps("personalities.json")
	if err != nil {
		return err
	}
	clockTimes, err := loadClockTimes("calibrations/clockTimes.json")
	if err != nil {
		return err
	}

	moveRes, err := handleMoveReqLocal(req, cmpMap, clockTimes)
	if err != nil {
		return err
	}
	return writeJSON(moveRes)
}

type clocktimes struct {
	Easy int `json:"Easy"`
	Hard int `json:"Hard"`
	Gm   int `json:"Gm"`
}

func handleMoveReqLocal(moveReq models.MoveReq, cmpMap map[string]models.Cmp, clocks clocktimes) (models.MoveData, error) {
	cmp, ok := cmpMap[moveReq.CmpName]
	if !ok {
		errMsg := fmt.Sprintf("%s is not a valid personality", moveReq.CmpName)
		return models.MoveData{Err: &errMsg}, nil
	}

	if !moveReq.ShouldSkipBook {
		bookMove, err := books.GetMove(moveReq.Moves, cmp.Book)
		if err == nil {
			bookMove.GameId = moveReq.GameId
			return bookMove, nil
		}
	}

	settings := moveReq
	settings.CmpVals = cmp.Vals
	if moveReq.ClockTime == 0 {
		settings.ClockTime = getClockTime(cmp, clocks)
	}

	moveData, err := engine.GetMove(settings)
	if err != nil {
		return models.MoveData{}, err
	}
	if moveData.Err != nil {
		return moveData, nil
	}

	moveData.WillAcceptDraw = getDrawEval(moveData.Eval, settings)
	moveData.Type = "engine"
	moveData.GameId = moveReq.GameId
	return moveData, nil
}

func loadCmps(path string) (map[string]models.Cmp, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	cmps := make(map[string]models.Cmp)
	if err := json.NewDecoder(f).Decode(&cmps); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	return cmps, nil
}

func loadClockTimes(path string) (clocktimes, error) {
	f, err := os.Open(path)
	if err != nil {
		return clocktimes{}, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	var out clocktimes
	if err := json.NewDecoder(f).Decode(&out); err != nil {
		return clocktimes{}, fmt.Errorf("decode %s: %w", path, err)
	}
	return out, nil
}

func getClockTime(cmp models.Cmp, clocks clocktimes) int {
	if cmp.Ponder == "easy" {
		return clocks.Easy
	}
	if cmp.Rating >= 2700 {
		return clocks.Gm
	}
	return clocks.Hard
}

func getDrawEval(currentEval int, settings models.MoveReq) bool {
	contemtForDraw, err := strconv.Atoi(settings.CmpVals.Cfd)
	if err != nil {
		return false
	}
	if len(settings.Moves) <= 30 {
		return false
	}
	return (currentEval + contemtForDraw) < 0
}

func readMoveReq(raw string) (models.MoveReq, error) {
	payload := []byte(raw)
	var req models.MoveReq
	if err := json.Unmarshal(payload, &req); err != nil {
		return models.MoveReq{}, fmt.Errorf("parse move request json: %w", err)
	}
	if req.CmpName == "" {
		return models.MoveReq{}, errors.New("cmpName is required")
	}
	return req, nil
}

func writeJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func binaryDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolve executable path: %w", err)
	}
	return filepath.Dir(exe), nil
}

func loadDotEnvDefaults() {
	cwd, err := os.Getwd()
	if err != nil {
		return
	}
	p := filepath.Join(cwd, ".env")
	if _, err := os.Stat(p); err == nil {
		loadEnvFile(p)
	}
}

func loadEnvFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if line == "" || line[0] == '#' {
			continue
		}
		eq := -1
		for i := 0; i < len(line); i++ {
			if line[i] == '=' {
				eq = i
				break
			}
		}
		if eq <= 0 {
			continue
		}
		key := line[:eq]
		val := line[eq+1:]
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, val)
		}
	}
}
