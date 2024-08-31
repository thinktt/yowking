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
	Status      string   `json:"status" bson:"status" binding:"required,oneof=created started mate resign stalemate draw"`
	Winner      string   `json:"winner,omitempty" bson:"winner" binding:"required,oneof=pending none white black"`
	Moves       string   `json:"moves" bson:"moves,omitempty" binding:"required"`
	MoveList    []string `json:"moveList,omitempty" bson:"moveList"`
	WhitePlayer Player   `json:"whitePlayer" bson:"whitePlayer" binding:"required"`
	BlackPlayer Player   `json:"blackPlayer" bson:"blackPlayer" binding:"required"`
	// DrawType    string   `json:"drawType,omitempty" bson:"drawType" binding:"omitempty,oneof=material stalemate threefold fiftyMove mutual"`
}

func (g *Game2) TurnColor() string {
	if len(g.MoveList)%2 == 0 {
		return "white"
	}
	return "black"
}

type Game2MutableFields struct {
	ID         string `json:"id" binding:"required,alphanum,min=8,max=8"`
	LastMoveAt int64  `json:"lastMoveAt" binding:"required"`
	Moves      string `json:"moves" binding:"required"`
	Status     string `json:"status,omitempty" binding:"required,oneof=started mate resign draw"`
	Winner     string `json:"winner,omitempty" binding:"required,oneof=white black"`
}

// DrawType   string `json:"drawType" bson:"drawType,omitempty" binding:"required,oneof=mutual fifty stalemate material"`
// LichessID   string   `json:"lichessId" bson:"lichessId"  binding:"alphanum,min=8,max=8"`
// MoveList    []string `json:"moveList,omitempty" bson:"moveList"`

type MoveData2 struct {
	Index int    `json:"index"`
	Move  string `json:"move" binding:"required"`
}
