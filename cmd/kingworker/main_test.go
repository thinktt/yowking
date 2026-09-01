package main

import (
	"testing"

	"github.com/thinktt/yowking/pkg/models"
)

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

func TestWorkerTagDefaultsWhenMissing(t *testing.T) {
	if got := workerTagFromEnv(""); got != "default" {
		t.Fatalf("workerTagFromEnv(\"\") = %q, want %q", got, "default")
	}
}

func TestNormalizeMoveRequestAddsDefaultWorkerTag(t *testing.T) {
	moveReq := normalizeMoveRequest(models.MoveReq{GameId: "game1234"})
	if moveReq.WorkerTag != "default" {
		t.Fatalf("expected default worker tag, got %q", moveReq.WorkerTag)
	}
}

func TestNormalizeMoveRequestAddsDefaultAPITag(t *testing.T) {
	moveReq := normalizeMoveRequest(models.MoveReq{GameId: "game1234"})
	if moveReq.ApiTag != "default" {
		t.Fatalf("expected default API tag, got %q", moveReq.ApiTag)
	}
}

func TestGetMoveReqSubjectUsesDefaultWorkerTag(t *testing.T) {
	if got := getMoveReqSubject(""); got != "move-req.default" {
		t.Fatalf("getMoveReqSubject(\"\") = %q, want %q", got, "move-req.default")
	}
}

func TestGetMoveResSubjectUsesAPITagWhenPresent(t *testing.T) {
	if got := getMoveResSubject("arenaRunner"); got != "move-res.arenaRunner" {
		t.Fatalf("getMoveResSubject() = %q, want %q", got, "move-res.arenaRunner")
	}
}

func TestGetMoveResSubjectUsesDefaultAPITag(t *testing.T) {
	if got := getMoveResSubject(""); got != "move-res.default" {
		t.Fatalf("getMoveResSubject() = %q, want %q", got, "move-res.default")
	}
}
