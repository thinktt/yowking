package models

type Player struct {
	ID   string `json:"id" binding:"required,alphanum"`
	Type string `json:"type" binding:"required,oneof=cmp lichess"`
}

type Game2 struct {
	ID          string `json:"id" binding:"required,alphanum,min=8,max=8"`
	LichessID   string `json:"lichessId,omitempty"`
	CreatedAt   int64  `json:"createdAt" binding:"required"`
	LastMoveAt  int64  `json:"lastMoveAt" binding:"required"`
	Status      string `json:"status" binding:"required,oneof=created started mate resign stalemate draw"`
	Winner      string `json:"winner" binding:"required,oneof=white black"`
	Moves       string `json:"moves" binding:"required"`
	WhitePlayer Player `json:"whitePlayer" binding:"required"`
	BlackPlayer Player `json:"blackPlayer" binding:"required"`
}
