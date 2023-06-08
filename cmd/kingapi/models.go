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
	"moves": [
		"e2e4", "c7c6", "d2d4", "d7d5", "b1c3", "d5e4", "c3e4", "g8f6", "e4f6", 
		"e7f6", "g2g3", "f8d6"
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
