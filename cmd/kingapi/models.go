package main

type Settings struct {
	Moves          []string `json:"moves"`
	PVals          PVals    `json:"pVals"`
	ClockTime      int      `json:"clockTime"`
	SecondsPerMove *int     `json:"secondsPerMove"`
}

type PVals struct {
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

// this is the Tal personality along with some moves
var testJson = `{
	"moves" : [
		"h2h4",  "d7d5", "f2f3",
		"d5d4",  "a2a4", "d4d3",
		"b2b4",  "d3c2", "h1h2",
		"c2b1q", "e2e3", "b1a1",
		"f1d3",  "d8d3", "h2h1",
		"c7c5",  "g2g3", "c8e6",
		"g1h3",  "b8c6", "d1b3",
		"e8c8",  "b3e6", "f7e6",
		"b8d8"
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
