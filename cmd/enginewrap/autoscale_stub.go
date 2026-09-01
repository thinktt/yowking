//go:build !windows

package main

type kingTCAutoScale struct {
	enabled bool
}

func configureKingTCAutoScale() (kingTCAutoScale, error) {
	return kingTCAutoScale{}, nil
}

func (scale kingTCAutoScale) command() string { return "" }

func (scale kingTCAutoScale) logLine() string { return "" }
