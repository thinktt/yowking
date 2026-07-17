package main

import (
	"testing"

	"github.com/thinktt/yowking/pkg/models"
)

func TestApplyWorkerOverridesForcesRandomOff(t *testing.T) {
	moveReq := applyWorkerOverrides(models.MoveReq{}, true)
	if !moveReq.RandomIsOff {
		t.Fatal("expected worker override to disable personality randomness")
	}
}

func TestApplyWorkerOverridesPreservesRequestWithoutOverride(t *testing.T) {
	moveReq := applyWorkerOverrides(models.MoveReq{RandomIsOff: false}, false)
	if moveReq.RandomIsOff {
		t.Fatal("expected worker to preserve request randomness setting")
	}
}
