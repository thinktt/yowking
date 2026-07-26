package main

import (
	"testing"

	"github.com/thinktt/yowking/pkg/models"
)

func TestApplyWorkerOverridesForcesRandomOff(t *testing.T) {
	moveReq := applyWorkerOverrides(models.MoveReq{}, true, false)
	if !moveReq.RandomIsOff {
		t.Fatal("expected worker override to disable personality randomness")
	}
}

func TestApplyWorkerOverridesPreservesRequestWithoutOverride(t *testing.T) {
	moveReq := applyWorkerOverrides(models.MoveReq{RandomIsOff: false}, false, false)
	if moveReq.RandomIsOff {
		t.Fatal("expected worker to preserve request randomness setting")
	}
}

func TestApplyWorkerOverridesForcesRandomOn(t *testing.T) {
	moveReq := applyWorkerOverrides(models.MoveReq{RandomIsOff: true}, false, true)
	if moveReq.RandomIsOff {
		t.Fatal("expected force-random worker to clear randomIsOff")
	}
	if !moveReq.RandomIsForced {
		t.Fatal("expected force-random worker to mark randomness as forced")
	}
}
