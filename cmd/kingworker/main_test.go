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

func TestPrepareMoveResponseCopiesRequestIdentity(t *testing.T) {
	moveReq := models.MoveReq{
		GameId:    "game1234",
		WorkerTag: "kingWC",
		Moves:     []string{"e2e4", "e7e5", "g1f3"},
	}
	moveRes := prepareMoveResponse(moveReq, models.MoveData{CoordinateMove: "b8c6"})

	if moveRes.Index != 3 {
		t.Fatalf("expected index 3, got %d", moveRes.Index)
	}
	if moveRes.GameId != moveReq.GameId {
		t.Fatalf("expected game ID %q, got %q", moveReq.GameId, moveRes.GameId)
	}
	if moveRes.WorkerTag != moveReq.WorkerTag {
		t.Fatalf("expected worker tag %q, got %q", moveReq.WorkerTag, moveRes.WorkerTag)
	}
}

func TestGetMoveResSubjectUsesWorkerTagWhenPresent(t *testing.T) {
	moveRes := models.MoveData{GameId: "game1234", WorkerTag: "kingWC"}
	if got := getMoveResSubject(moveRes); got != "move-res.kingWC" {
		t.Fatalf("getMoveResSubject() = %q, want %q", got, "move-res.kingWC")
	}
}

func TestGetMoveResSubjectPreservesLegacyGameSubject(t *testing.T) {
	moveRes := models.MoveData{GameId: "game1234"}
	if got := getMoveResSubject(moveRes); got != "move-res.game1234" {
		t.Fatalf("getMoveResSubject() = %q, want %q", got, "move-res.game1234")
	}
}
