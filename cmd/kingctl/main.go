package main

import (
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
	moveFlags := flag.NewFlagSet("move", flag.ContinueOnError)
	moveFlags.SetOutput(os.Stderr)
	var skipBook bool
	moveFlags.BoolVar(&skipBook, "s", false, "skip book lookup")
	moveFlags.BoolVar(&skipBook, "skip-book", false, "skip book lookup")
	if err := moveFlags.Parse(args); err != nil {
		return err
	}

	if moveFlags.NArg() != 1 {
		return errors.New("usage: kingctl move [-s|--skip-book] <json>")
	}

	moveReq, err := readMoveReq(moveFlags.Arg(0))
	if err != nil {
		return err
	}
	if skipBook {
		moveReq.ShouldSkipBook = true
	}
	if moveReq.GameId == "" {
		moveReq.GameId = fmt.Sprintf("kingctl-%d", time.Now().UnixNano())
	}

	baseDir, err := binaryDir()
	if err != nil {
		return err
	}
	if err := os.Chdir(baseDir); err != nil {
		return fmt.Errorf("change dir to %q: %w", baseDir, err)
	}

	cmpMap, err := loadCmps("personalities.json")
	if err != nil {
		return err
	}
	calibratedClockTimes, err := loadClockTimes("calibrations/clockTimes.json")
	if err != nil {
		return err
	}

	moveResponse, err := resolveMoveLocal(moveReq, cmpMap, calibratedClockTimes)
	if err != nil {
		return err
	}
	return writeJSON(moveResponse)
}

type clockTimes struct {
	Easy int `json:"Easy"`
	Hard int `json:"Hard"`
	Gm   int `json:"Gm"`
}

func resolveMoveLocal(moveReq models.MoveReq, cmpMap map[string]models.Cmp, clocks clockTimes) (models.MoveData, error) {
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
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	personalitiesMap := make(map[string]models.Cmp)
	if err := json.NewDecoder(file).Decode(&personalitiesMap); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	return personalitiesMap, nil
}

func loadClockTimes(path string) (clockTimes, error) {
	file, err := os.Open(path)
	if err != nil {
		return clockTimes{}, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	var loadedClockTimes clockTimes
	if err := json.NewDecoder(file).Decode(&loadedClockTimes); err != nil {
		return clockTimes{}, fmt.Errorf("decode %s: %w", path, err)
	}
	return loadedClockTimes, nil
}

func getClockTime(cmp models.Cmp, clocks clockTimes) int {
	if cmp.Ponder == "easy" {
		return clocks.Easy
	}
	if cmp.Rating >= 2700 {
		return clocks.Gm
	}
	return clocks.Hard
}

func getDrawEval(currentEval int, settings models.MoveReq) bool {
	contemptForDraw, err := strconv.Atoi(settings.CmpVals.Cfd)
	if err != nil {
		return false
	}
	if len(settings.Moves) <= 30 {
		return false
	}
	return (currentEval + contemptForDraw) < 0
}

func readMoveReq(rawJSON string) (models.MoveReq, error) {
	payload := []byte(rawJSON)
	var moveRequest models.MoveReq
	if err := json.Unmarshal(payload, &moveRequest); err != nil {
		return models.MoveReq{}, fmt.Errorf("parse move request json: %w", err)
	}
	if moveRequest.CmpName == "" {
		return models.MoveReq{}, errors.New("cmpName is required")
	}
	return moveRequest, nil
}

func writeJSON(value any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(value)
}

func binaryDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolve executable path: %w", err)
	}
	return filepath.Dir(exe), nil
}
