//go:build windows

package main

import (
	"fmt"
	"os"
	"strconv"
)

const (
	kingTCAutoScaleReferenceEnv = "KINGTC_AUTOSCALE_REFERENCE_CPU_NS"
	kingTCAutoScaleUnit         = uint64(1_000_000)
	kingTCAutoScaleMin          = uint64(250_000)
	kingTCAutoScaleMax          = uint64(4_000_000)
)

type kingTCAutoScale struct {
	enabled      bool
	referenceNs  uint64
	benchmarkNs  uint64
	scaleMillion uint64
}

func configureKingTCAutoScale() (kingTCAutoScale, error) {
	referenceText := os.Getenv(kingTCAutoScaleReferenceEnv)
	if referenceText == "" {
		return kingTCAutoScale{}, nil
	}

	referenceNs, err := strconv.ParseUint(referenceText, 10, 64)
	if err != nil || referenceNs == 0 {
		return kingTCAutoScale{}, fmt.Errorf("%s must be a positive integer", kingTCAutoScaleReferenceEnv)
	}

	report, err := runKingTCBenchmark()
	if err != nil {
		return kingTCAutoScale{}, err
	}
	if report.MedianNs == 0 {
		return kingTCAutoScale{}, fmt.Errorf("benchmark returned zero CPU time")
	}

	// Round to the nearest millionth before applying the binary's fixed-point
	// KingTC multiplier. The bounds match the command parser's accepted range.
	scaleMillion := (referenceNs*kingTCAutoScaleUnit + report.MedianNs/2) / report.MedianNs
	if scaleMillion < kingTCAutoScaleMin {
		scaleMillion = kingTCAutoScaleMin
	}
	if scaleMillion > kingTCAutoScaleMax {
		scaleMillion = kingTCAutoScaleMax
	}

	return kingTCAutoScale{
		enabled:      true,
		referenceNs:  referenceNs,
		benchmarkNs:  report.MedianNs,
		scaleMillion: scaleMillion,
	}, nil
}

func (scale kingTCAutoScale) command() string {
	return fmt.Sprintf("cm_parm kingtc_scale=%d", scale.scaleMillion)
}

func (scale kingTCAutoScale) logLine() string {
	return fmt.Sprintf(
		"kingtc-autoscale reference-cpu-ns=%d benchmark-cpu-ns=%d scale-millionths=%d",
		scale.referenceNs,
		scale.benchmarkNs,
		scale.scaleMillion,
	)
}
