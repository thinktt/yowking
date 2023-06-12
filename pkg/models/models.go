package models

type Settings struct {
	Moves          []string `json:"moves"`
	CmpVals        CmpVals  `json:"cmpVals"`
	ClockTime      int      `json:"clockTime"`
	SecondsPerMove *int     `json:"secondsPerMove"`
}

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

type MoveReq struct {
	Moves   []string `json:"moves" binding:"required,dive,alphanum,min=4,max=5"`
	CmpName string   `json:"cmpName" binding:"required,alphanum,max=15"`
}

type Cmp struct {
	Vals   CmpVals `json:"out"`
	Name   string  `json:"name"`
	Ponder string  `json:"ponder"`
	Book   string  `json:"book"`
}

// var engineLinesString = `
// 3001  +144      0       417 Nf5 Nc3 Rad8+ Kc1 Nxe3
// 4001  +144      1      1740 Nf5 Nc3 Rad8+ Kc1 Nxe3
// 5002  +156      1      6187 Nf5 Re1 Rad8+ Kc3 c5 Nd2 Nxe3
// 6002  +156      3     20899 Nf5 Re1 Rad8+ Kc3 c5 Nd2 Nxe3
// 7002  +153      7     58359 Nf5 Re1 Rad8+ Kc3 Nxe3 Nd2 c5 Nc4
// 8003  +169     13    133468 Nf5 Nc3 Rad8+ Kc1 Nxe3 Re1 Nxg2 Re4 Rxe4 Nxe4
// 9003  +171     25    314861 Nf5 Re1 Rad8+ Kc3 Nxe3 Nd2 Nxg2 Rxe8+ Rxe8 Nc4 Ne3 Nxe3 Rxe3+ Kd4
// 10003  +164     47    657961 Nf5 Re1 Rad8+ Kc3 Nxe3 Na3 Re4 Nb5 Nd5+ Kd2 Rb4 a4 Rxb2 Nxa7
// 11004  +154    147   2430138 Nf5 Re1 Rad8+ Kc3 Nxe3 Na3 Nxg2 Rxe8+ Rxe8 Nc4 Ne3 Nxe3 Rxe3+ Kd4 Rxg3 Re1
// `

// {
//   "depth": 11004,
//   "eval": 154,
//   "time": 147,
//   "id": 2430138,
//   "algebraMove": "Nf5",
//   "coordinateMove": "d4f5",
//   "willAcceptDraw": false,
//   "timeForMove": 2177,
//   "engineMove": "d4f5"
// }

type MoveData struct {
	Depth          int     `json:"depth"`
	Eval           int     `json:"eval"`
	Time           int     `json:"time"`
	Id             int     `json:"id"`
	AlgebraMove    string  `json:"algebraMove"`
	CoordinateMove string  `json:"coordinateMove"`
	WillAcceptDraw bool    `json:"willAcceptDraw"`
	Err            *string `json:"err,omitempty"`
	// TimeForMove    int    `json:"timeForMove"`
	// EngineMove     string `json:"engineMove"`
}

// 10003   +27    205   2205499 e4 dxc2 Qxc2 e5 f4 exf4 d4 Bd6 Nf3 Nf6 Ne5 c5

// this is the Tal personality along with some moves
var TestJson = `{
	"moves" : [
		"h2h4",  "d7d5", "f2f3",
		"d5d4",  "a2a4", "d4d3"
	],
	"pVals": {
		"opp": "107",
		"opn": "107",
		"opb": "107",
		"opr": "107",
		"opq": "112",
		"myp": "107",
		"myn": "102",
		"myb": "110",
		"myr": "107",
		"myq": "112",
		"mycc": "93",
		"mymob": "126",
		"myks": "70",
		"mypp": "89",
		"mypw": "84",
		"opcc": "93",
		"opmob": "126",
		"opks": "70",
		"oppp": "89",
		"oppw": "84",
		"cfd": "0",
		"sop": "100",
		"avd": "-40",
		"rnd": "0",
		"sel": "9",
		"md": "99",
		"tts": "16777216"
	},
	"clockTime": 8550,
	"secondsPerMove": null
}`
