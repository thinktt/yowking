//go:build windows

package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"sort"
	"syscall"
	"unsafe"
)

const (
	benchmarkWarmups = 2
	benchmarkRuns    = 5
	// A pass is roughly half a second under Wine on Bee. GetThreadTimes is
	// quantized to 10 ms here, so a longer pass keeps that granularity small.
	benchmarkLimit  = 1_500_000
	benchmarkPrimes = 114_155
)

type benchmarkReport struct {
	PrimeCount  int
	DurationsNs []uint64
	MedianNs    uint64
}

type filetime struct {
	LowDateTime  uint32
	HighDateTime uint32
}

var (
	kernel32              = syscall.NewLazyDLL("kernel32.dll")
	getCurrentThread      = kernel32.NewProc("GetCurrentThread")
	getThreadTimes        = kernel32.NewProc("GetThreadTimes")
	benchmarkResultSink   uint64
)

func runBenchmarkMode() {
	fmt.Println("kingtc-benchmark-ready")
	if !bufio.NewScanner(os.Stdin).Scan() {
		fmt.Fprintln(os.Stderr, "kingtc-benchmark requires a start line")
		os.Exit(2)
	}
	report, err := runKingTCBenchmark()
	if err != nil {
		fmt.Fprintln(os.Stderr, "kingtc-benchmark failed:", err)
		os.Exit(1)
	}
	printBenchmarkReport(report)
}

func runKingTCBenchmark() (benchmarkReport, error) {
	// GetThreadTimes measures the Windows OS thread, while Go may otherwise move
	// this goroutine between OS threads between the start and end samples.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	for i := 0; i < benchmarkWarmups; i++ {
		if err := runPrimeBenchmark(); err != nil {
			return benchmarkReport{}, err
		}
	}

	durations := make([]uint64, 0, benchmarkRuns)
	for i := 0; i < benchmarkRuns; i++ {
		start, err := currentThreadCPUTimeNs()
		if err != nil {
			return benchmarkReport{}, err
		}
		if err := runPrimeBenchmark(); err != nil {
			return benchmarkReport{}, err
		}
		end, err := currentThreadCPUTimeNs()
		if err != nil {
			return benchmarkReport{}, err
		}
		if end < start {
			return benchmarkReport{}, fmt.Errorf("thread CPU clock moved backwards: start=%d end=%d", start, end)
		}
		durations = append(durations, end-start)
	}

	sorted := append([]uint64(nil), durations...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	return benchmarkReport{
		PrimeCount:  benchmarkPrimes,
		DurationsNs: durations,
		MedianNs:    sorted[len(sorted)/2],
	}, nil
}

func runPrimeBenchmark() error {
	count := 0
	for candidate := 2; candidate <= benchmarkLimit; candidate++ {
		prime := true
		for divisor := 2; divisor*divisor <= candidate; divisor++ {
			if candidate%divisor == 0 {
				prime = false
				break
			}
		}
		if prime {
			count++
		}
	}
	if count != benchmarkPrimes {
		return fmt.Errorf("unexpected prime count: got=%d want=%d", count, benchmarkPrimes)
	}
	benchmarkResultSink += uint64(count)
	return nil
}

func currentThreadCPUTimeNs() (uint64, error) {
	thread, _, _ := getCurrentThread.Call()
	var creation, exit, kernel, user filetime
	ok, _, callErr := getThreadTimes.Call(
		thread,
		uintptr(unsafe.Pointer(&creation)),
		uintptr(unsafe.Pointer(&exit)),
		uintptr(unsafe.Pointer(&kernel)),
		uintptr(unsafe.Pointer(&user)),
	)
	if ok == 0 {
		return 0, callErr
	}
	// FILETIME values are 100-nanosecond units; sum user and kernel CPU time.
	return (filetimeValue(kernel) + filetimeValue(user)) * 100, nil
}

func filetimeValue(value filetime) uint64 {
	return uint64(value.HighDateTime)<<32 | uint64(value.LowDateTime)
}

func printBenchmarkReport(report benchmarkReport) {
	fmt.Printf(
		"kingtc-benchmark primes=%d limit=%d warmups=%d runs=%d cpu-ns=%v median-cpu-ns=%d\n",
		report.PrimeCount,
		benchmarkLimit,
		benchmarkWarmups,
		benchmarkRuns,
		report.DurationsNs,
		report.MedianNs,
	)
}
