package engine

import (
	"strings"
	"testing"
	"time"

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

func TestReadEngineOutPublishesLatestCandidate(t *testing.T) {
	log = logrus.New().WithField("test", true)
	moveChan := make(chan engineOutput, 1)

	readEngineOut(strings.NewReader(
		"8 12 20 100 e4 e5\n9 18 35 200 Nf3 Nc6\n",
	), moveChan, 0)

	result := <-moveChan
	if !result.final {
		t.Fatal("expected final engine output")
	}
	if result.moveData.AlgebraMove != "Nf3" {
		t.Fatalf("latest move = %q, want %q", result.moveData.AlgebraMove, "Nf3")
	}
	if result.moveData.Id != 200 {
		t.Fatalf("latest ID = %d, want 200", result.moveData.Id)
	}
}

func TestGetMoveTimeout(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		t.Setenv("ENGINE_MOVE_TIMEOUT", "")
		timeout, err := getMoveTimeout()
		if err != nil {
			t.Fatal(err)
		}
		if timeout != 15*time.Minute {
			t.Fatalf("timeout = %s, want 15m", timeout)
		}
	})

	t.Run("configured", func(t *testing.T) {
		t.Setenv("ENGINE_MOVE_TIMEOUT", "90s")
		timeout, err := getMoveTimeout()
		if err != nil {
			t.Fatal(err)
		}
		if timeout != 90*time.Second {
			t.Fatalf("timeout = %s, want 90s", timeout)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		t.Setenv("ENGINE_MOVE_TIMEOUT", "soon")
		if _, err := getMoveTimeout(); err == nil {
			t.Fatal("expected invalid timeout error")
		}
	})
}
