package models

type Player struct {
	ID   string `json:"id" binding:"required,alphanum"`
	Type string `json:"type" binding:"required,oneof=cmp lichess"`
}
type Game2 struct {
	ID          string   `json:"id" bson:"id" binding:"required,alphanum,min=8,max=8"`
	LichessID   string   `json:"lichessId,omitempty" bson:"lichessId"`
	CreatedAt   int64    `json:"createdAt" bson:"createdAt" binding:"required"`
	LastMoveAt  int64    `json:"lastMoveAt" bson:"lastMoveAt" binding:"required"`
	Status      string   `json:"status" bson:"status" binding:"required,oneof=created started mate resign stalemate draw"`
	Winner      string   `json:"winner,omitempty" bson:"winner,omitempty" binding:"oneof=white black"`
	Moves       string   `json:"moves" bson:"moves,omitempty"`
	MoveList    []string `json:"moveList,omitempty" bson:"moveList"`
	WhitePlayer Player   `json:"whitePlayer" bson:"whitePlayer" binding:"required"`
	BlackPlayer Player   `json:"blackPlayer" bson:"blackPlayer" binding:"required"`
}
