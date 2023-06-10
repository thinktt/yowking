// This is a simple wrapper for the Chess engine. For some reason the pipes
// from Node to wine to the engine break but when we wrap the engine in a
// Go win binary they work. The order of operations is then
// Node in linux --> Wine --> enginewrap.exe --> engine

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func main() {
	isWsl := os.Getenv("IS_WSL")
	shouldPostInput := os.Getenv("SHOULD_POST_INPUT")
	fmt.Println("shouldPostInput: " + shouldPostInput)

	var cmd *exec.Cmd
	if isWsl == "true" {
		cmd = exec.Command("./assets/TheKing350noOpk.exe")
	} else {
		cmd = exec.Command("wine", "enginewrap.exe")
	}

	engine, err := cmd.StdinPipe()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	engineOut, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	engineErr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// handle the engine steams in real time
	go pipeOut(engineOut)
	go pipeOut(engineErr)
	go forwardUserCommands(engine)

	// start the engine
	err = cmd.Start()
	if err != nil {
		fmt.Printf("cmd.Run() failed with %s\n", err)
		os.Exit(1)
	}

	// marshall the settings json data
	var settings Settings
	err = json.Unmarshal([]byte(testJson), &settings)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Success marshalling json")

	// prepare all the personality setting commands to be sent to the engine
	cmpLoaderTemplate := `cm_parm default
	cm_parm opp={{.Opp}} opn={{.Opn}} opb={{.Opb}} opr={{.Opr}} opq={{.Opq}}
	cm_parm myp={{.Myp}} myn={{.Myn}} myb={{.Myb}} myr={{.Myr}} myq={{.Myq}}
	cm_parm mycc={{.Mycc}} mymob={{.Mymob}} myks={{.Myks}}  mypp={{.Mypp}} mypw={{.Mypw}}
	cm_parm opcc={{.Opcc}} opmob={{.Opmob}} opks={{.Opks}} oppp={{.Oppp}} oppw={{.Oppw}}
	cm_parm cfd={{.Cfd}} sop={{.Sop}} avd={{.Avd}} rnd={{.Rnd}} sel={{.Sel}} md={{.Md}}
	cm_parm tts={{.Tts}}
	easy
	`
	t := template.Must(template.New("pValsTemplate").Parse(cmpLoaderTemplate))
	buf := &bytes.Buffer{}
	if err := t.Execute(buf, settings.PVals); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	timeStr := fmt.Sprintf("time %d\n", settings.ClockTime)
	otimStr := fmt.Sprintf("otim %d\n", settings.ClockTime)

	// send settings to the engine
	engine.Write([]byte("xboard\n"))
	engine.Write([]byte("post\n"))
	engine.Write([]byte(timeStr))
	engine.Write([]byte(otimStr))
	engine.Write([]byte(buf.String()))

	// send all the moves to the engine
	for _, move := range settings.Moves {
		moveStr := fmt.Sprintf("%s\n", move)
		engine.Write([]byte(moveStr))
	}

	// start the engine
	engine.Write([]byte("go\n"))

	// keep this app going while the egine is running
	cmd.Wait()
}

func pipeOut(r io.Reader) {
	s := bufio.NewScanner(r)
	moveCandidate := MoveData{}
	for s.Scan() {
		engineLine := s.Text()
		fmt.Println(engineLine)

		// check if the engine line final move result
		words := strings.Fields(engineLine)
		if strings.Contains(engineLine, "move") && len(words) == 2 {
			moveCandidate.CoordinateMove = strings.Fields(engineLine)[1]
			break
		}

		// parese the enginLine if it is a move line
		moveData, err := parseMoveLine(words)
		if err != nil {
			continue
		}
		moveCandidate = moveData
	}

	fmt.Println("moveCandidate: ", moveCandidate)
}

func parseMoveLine(words []string) (MoveData, error) {
	if len(words) < 5 {
		return MoveData{}, errors.New("Not enough words for move line")
	}

	// only a valid move line if first 4 words are numbers
	var err error
	var numbers [4]int
	for i := 0; i < 4; i++ {
		numbers[i], err = strconv.Atoi(words[i])
		if err != nil {
			return MoveData{}, errors.New("First 4 words of engine line are not numbers")
		}
	}

	moveData := MoveData{
		Depth:          numbers[0],
		Eval:           numbers[1],
		Time:           numbers[2],
		Id:             numbers[3],
		AlgebraMove:    words[4],
		CoordinateMove: "",
	}

	return moveData, nil
}

func getDrawEval(currentEval int, settings Settings) bool {
	contemtForDraw, err := strconv.Atoi(settings.PVals.Cfd)
	if err != nil {
		return false
	}

	if len(settings.Moves) <= 30 {
		return false
	}

	return (currentEval + contemtForDraw) < 0
}

func forwardUserCommands(engine io.WriteCloser) {
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		line := s.Text()
		engine.Write([]byte(line + "\n"))
	}
}
