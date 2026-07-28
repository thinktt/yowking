//go:build !windows

package main

import "fmt"

type benchmarkReport struct{}

func runBenchmarkMode() {
	panic("kingtc-benchmark is supported only by the Windows engine wrapper")
}

func runKingTCBenchmark() (benchmarkReport, error) {
	return benchmarkReport{}, fmt.Errorf("kingtc-benchmark is supported only by the Windows engine wrapper")
}

func printBenchmarkReport(benchmarkReport) {}
