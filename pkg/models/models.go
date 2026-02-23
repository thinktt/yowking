package models

// MoveReq is the worker request contract used by kingworker.
type MoveReq struct {
	Moves          []string `json:"moves" binding:"required,dive,alphanum,min=4,max=5"`
	CmpName        string   `json:"cmpName" binding:"required,alphanum,max=15"`
	GameId         string   `json:"gameId" binding:"required,alphanum,max=15"`
	StopId         int      `json:"stopId" binding:"omitempty,alphanum,max=15"`
	ClockTime      int      `json:"clockTime" binding:"omitempty,alphanum,max=15"`
	RandomIsOff    bool     `json:"randomIsOff"`
	ShouldSkipBook bool     `json:"shouldSkipBook"`
	CmpVals        CmpVals  `json:"-"`
}

// CmpVals are the King engine personality tuning parameters.
type CmpVals struct {
	Opp   string `json:"opp"`
	Opn   string `json:"opn"`
	Opb   string `json:"opb"`
	Opr   string `json:"opr"`
	Opq   string `json:"opq"`
	Myp   string `json:"myp"`
	Myn   string `json:"myn"`
	Myb   string `json:"myb"`
	Myr   string `json:"myr"`
	Myq   string `json:"myq"`
	Mycc  string `json:"mycc"`
	Mymob string `json:"mymob"`
	Myks  string `json:"myks"`
	Mypp  string `json:"mypp"`
	Mypw  string `json:"mypw"`
	Opcc  string `json:"opcc"`
	Opmob string `json:"opmob"`
	Opks  string `json:"opks"`
	Oppp  string `json:"oppp"`
	Oppw  string `json:"oppw"`
	Cfd   string `json:"cfd"`
	Sop   string `json:"sop"`
	Avd   string `json:"avd"`
	Rnd   string `json:"rnd"`
	Sel   string `json:"sel"`
	Md    string `json:"md"`
	Tts   string `json:"tts"`
}

// Cmp is a named King personality, backed by personalities.json.
type Cmp struct {
	Vals   CmpVals `json:"out"`
	Name   string  `json:"name"`
	Ponder string  `json:"ponder"`
	Book   string  `json:"book"`
	Rating int     `json:"rating"`
}

// MoveData is the kingworker response payload.
type MoveData struct {
	Depth          int     `json:"depth,omitempty"`
	Eval           int     `json:"eval,omitempty"`
	Time           int     `json:"time,omitempty"`
	Id             int     `json:"id,omitempty"`
	AlgebraMove    string  `json:"algebraMove,omitempty"`
	CoordinateMove string  `json:"coordinateMove,omitempty"`
	WillAcceptDraw bool    `json:"willAcceptDraw"`
	Err            *string `json:"err,omitempty"`
	Type           string  `json:"type"`
	GameId         string  `json:"gameId,omitempty"`
}
