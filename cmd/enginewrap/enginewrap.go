// This is a simple wrapper for the Chess engine. For some reason the pipes
// from wine to the engine break but when we wrap the engine in a
// Go win binary they work. The order of operations is then
// kingworker (go) --> Wine --> enginewrap.exe (go win) --> engine

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--kingtc-benchmark" {
		runBenchmarkMode()
		return
	}

	autoScale, err := configureKingTCAutoScale()
	if err != nil {
		fmt.Fprintln(os.Stderr, "kingtc autoscale setup failed:", err)
		os.Exit(1)
	}

	shouldPostInput := os.Getenv("SHOULD_POST_INPUT")
	cmd := exec.Command("./TheKing350.exe")

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)
	engine, err := cmd.StdinPipe()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = cmd.Start()
	if err != nil {
		fmt.Printf("cmd.Run() failed with %s\n", err)
		os.Exit(1)
	}

	fmt.Println("Engine wrapper started, waiting for commands")
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		line := s.Text()
		if shouldPostInput == "true" {
			fmt.Println("In: " + line)
		}
		if line == "kingtc-benchmark" {
			report, err := runKingTCBenchmark()
			if err != nil {
				fmt.Fprintln(os.Stderr, "kingtc-benchmark failed:", err)
				continue
			}
			printBenchmarkReport(report)
			continue
		}
		if line == "go" && autoScale.enabled {
			fmt.Println(autoScale.logLine())
			engine.Write([]byte(autoScale.command() + "\n"))
		}
		engine.Write([]byte(line + "\n"))
		if line == "quit" {
			fmt.Println("quit received, waiting for engine to quit")
			break
		}
	}
	if err := s.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "failed to read engine commands:", err)
	}

	// engine proces cleanup happens here
	// close the pipe
	if err := engine.Close(); err != nil {
		fmt.Fprintln(os.Stderr, "failed to close engine input:", err)
	}
	// wait for the system to reap the process
	if err := cmd.Wait(); err != nil {
		fmt.Fprintln(os.Stderr, "engine exited with error:", err)
	}
}
