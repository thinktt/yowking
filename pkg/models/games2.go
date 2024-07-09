package models

type Player struct {
	ID   string `json:"id" binding:"required"` //needs lichess id validation
	Type string `json:"type" binding:"required,oneof=cmp lichess"`
}
type Game2New struct {
	LichessID   string `json:"lichessId" bson:"lichessId" binding:"omitempty,alphanum,min=8,max=8"`
	WhitePlayer Player `json:"whitePlayer" bson:"whitePlayer" binding:"required"`
	BlackPlayer Player `json:"blackPlayer" bson:"blackPlayer" binding:"required"`
}

type Game2 struct {
	ID          string   `json:"id" bson:"id" binding:"required,alphanum,min=8,max=8"`
	LichessID   string   `json:"lichessId" bson:"lichessId"  binding:"alphanum,min=8,max=8"`
	CreatedAt   int64    `json:"createdAt" bson:"createdAt" binding:"required"`
	LastMoveAt  int64    `json:"lastMoveAt" bson:"lastMoveAt" binding:"required"`
	Status      string   `json:"status" bson:"status" binding:"required,oneof=created mate resign stalemate draw"`
	Winner      string   `json:"winner,omitempty" bson:"winner" binding:"required,oneof=none white black"`
	Moves       string   `json:"moves" bson:"moves,omitempty" binding:"required"`
	MoveList    []string `json:"moveList,omitempty" bson:"moveList"`
	WhitePlayer Player   `json:"whitePlayer" bson:"whitePlayer" binding:"required"`
	BlackPlayer Player   `json:"blackPlayer" bson:"blackPlayer" binding:"required"`
}

type MoveData2 struct {
	Index int    `json:"index"`
	Move  string `json:"move" binding:"required"`
}
