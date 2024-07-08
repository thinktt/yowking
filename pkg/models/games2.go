package models

type Player struct {
	ID   string `json:"id" binding:"required,alphanum"`
	Type string `json:"type" binding:"required,oneof=cmp lichess"`
}
type Game2 struct {
	ID          string   `json:"id" bson:"id" binding:"required,alphanum,min=8,max=8"`
	LichessID   string   `json:"lichessId" bson:"lichessId"`
	CreatedAt   int64    `json:"createdAt" bson:"createdAt" `
	LastMoveAt  int64    `json:"lastMoveAt" bson:"lastMoveAt" `
	Status      string   `json:"status" bson:"status"`           // binding:"oneof=created started mate resign stalemate draw"`
	Winner      string   `json:"winner,omitempty" bson:"winner"` //  binding:"oneof=pending none white black"`
	Moves       string   `json:"moves" bson:"moves,omitempty"`
	MoveList    []string `json:"moveList,omitempty" bson:"moveList"`
	WhitePlayer Player   `json:"whitePlayer" bson:"whitePlayer" binding:"required"`
	BlackPlayer Player   `json:"blackPlayer" bson:"blackPlayer" binding:"required"`
}

type MoveData2 struct {
	Index int    `json:"index"`
	Move  string `json:"move" binding:"required"`
}
