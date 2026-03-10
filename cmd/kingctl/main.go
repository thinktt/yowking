package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/thinktt/yowking/internal/booktester"
	"github.com/thinktt/yowking/internal/moves"
	"github.com/thinktt/yowking/pkg/models"
	"github.com/thinktt/yowking/pkg/personalities"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(2)
	}

	command := os.Args[1]
	commandArgs := os.Args[2:]

	switch command {
	case "help", "-h", "--help":
		printUsage()
	case "move":
		if err := runMoveCommand(commandArgs); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "book":
		if err := runBookCommand(commandArgs); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", command)
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

func runMoveCommand(commandArgs []string) error {
	moveRequestJSON, skipBookLookup, err := parseMoveCommandArgs(commandArgs)
	if err != nil {
		return err
	}

	moveReq, err := readMoveReq(moveRequestJSON)
	if err != nil {
		return err
	}
	moveReq = applyMoveRequestDefaults(moveReq, skipBookLookup)

	binaryDirectoryPath, err := binaryDir()
	if err != nil {
		return err
	}
	if err := prepareLocalRuntime(binaryDirectoryPath); err != nil {
		return err
	}

	moveResponse, err := moves.HandleMoveReq(moveReq)
	if err != nil {
		return err
	}
	return writeJSON(moveResponse)
}

func parseMoveCommandArgs(commandArgs []string) (string, bool, error) {
	// Define supported flags for the move command.
	moveFlags := flag.NewFlagSet("move", flag.ContinueOnError)
	moveFlags.SetOutput(os.Stderr)

	var skipBook bool
	moveFlags.BoolVar(&skipBook, "s", false, "skip book lookup")
	moveFlags.BoolVar(&skipBook, "skip-book", false, "skip book lookup")

	// Parse flags and remaining positional args.
	if err := moveFlags.Parse(commandArgs); err != nil {
		return "", false, err
	}

	// Require exactly one positional argument: the move request JSON.
	if moveFlags.NArg() != 1 {
		return "", false, errors.New("usage: kingctl move [-s|--skip-book] <json>")
	}

	// Return decomposed values to the caller.
	moveRequestJSON := moveFlags.Arg(0)

	return moveRequestJSON, skipBook, nil
}

func applyMoveRequestDefaults(moveReq models.MoveReq, skipBookLookup bool) models.MoveReq {
	if skipBookLookup {
		moveReq.ShouldSkipBook = true
	}
	gameIDMissing := moveReq.GameId == ""
	if gameIDMissing {
		moveReq.GameId = fmt.Sprintf("kingctl-%d", time.Now().UnixNano())
	}
	return moveReq
}

func prepareLocalRuntime(binaryDirectoryPath string) error {
	if err := os.Chdir(binaryDirectoryPath); err != nil {
		return fmt.Errorf("change dir to %q: %w", binaryDirectoryPath, err)
	}
	personalities.Reload()
	return nil
}

func runBookCommand(commandArgs []string) error {
	bookSubcommand, err := parseBookCommandArgs(commandArgs)
	if err != nil {
		return err
	}

	binaryDirectoryPath, err := binaryDir()
	if err != nil {
		return err
	}
	booktesterArgs := []string{bookSubcommand}
	return booktester.Run(booktesterArgs, binaryDirectoryPath)
}

func parseBookCommandArgs(commandArgs []string) (string, error) {
	if len(commandArgs) != 1 {
		return "", errors.New("usage: kingctl book <fens|mem>")
	}
	bookSubcommand := commandArgs[0]
	isFens := bookSubcommand == "fens"
	isMem := bookSubcommand == "mem"
	if !isFens && !isMem {
		return "", errors.New("usage: kingctl book <fens|mem>")
	}
	return bookSubcommand, nil
}

func readMoveReq(rawJSON string) (models.MoveReq, error) {
	var moveRequest models.MoveReq
	if err := json.Unmarshal([]byte(rawJSON), &moveRequest); err != nil {
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
