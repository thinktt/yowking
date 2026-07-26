package engine

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestApplyRandomSettingForcesZero(t *testing.T) {
	log = logrus.New().WithField("test", true)
	settings := Settings{RandomIsOff: true, RandomIsForced: true}

	applyRandomSetting(&settings)

	if settings.CmpVals.Rnd != "0" {
		t.Fatalf("expected rnd=0, got %q", settings.CmpVals.Rnd)
	}
}

func TestApplyRandomSettingReplacesZeroWithFifty(t *testing.T) {
	log = logrus.New().WithField("test", true)
	settings := Settings{RandomIsForced: true}
	settings.CmpVals.Rnd = "0"

	applyRandomSetting(&settings)

	if settings.CmpVals.Rnd != "50" {
		t.Fatalf("expected rnd=50, got %q", settings.CmpVals.Rnd)
	}
}

func TestApplyRandomSettingPreservesNonzeroValue(t *testing.T) {
	log = logrus.New().WithField("test", true)
	settings := Settings{RandomIsForced: true}
	settings.CmpVals.Rnd = "65"

	applyRandomSetting(&settings)

	if settings.CmpVals.Rnd != "65" {
		t.Fatalf("expected rnd=65, got %q", settings.CmpVals.Rnd)
	}
}
