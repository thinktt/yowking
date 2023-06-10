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
	moveData, err := getMove()
	if err != nil {
		fmt.Println("There was ane error getting the move: ", err)
		return
	}
	fmt.Println(moveData)
}

func getMove() (MoveData, error) {
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
		return MoveData{}, err
	}

	engineOut, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println(err)
		return MoveData{}, err
	}

	engineErr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Println(err)
		return MoveData{}, err
	}

	moveChan := make(chan MoveData)
	defer close(moveChan)

	// handle the engine streams in real time
	go pipeOut(engineOut, moveChan)
	go pipeOutErr(engineErr)
	go forwardUserCommands(engine)

	// start the engine
	err = cmd.Start()
	if err != nil {
		errToGo := fmt.Errorf("cmd.Run() failed with %s\n", err)
		return MoveData{}, errToGo
	}
	defer cmd.Wait()

	// marshall the settings json data
	var settings Settings
	err = json.Unmarshal([]byte(testJson), &settings)
	if err != nil {
		err := errors.New("Error unmarshalling settings json")
		return MoveData{}, err
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
		return MoveData{}, err
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

	// wait for the engine to send back a move
	moveData := <-moveChan
	engine.Write([]byte("quit\n"))
	if cmd.Process != nil {
		// cmd.Process.Kill()
		fmt.Println("it looks like the process is still running")
	}

	return moveData, nil

}

func pipeOut(r io.Reader, moveChan chan MoveData) {
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

		// parse the enginLine if it is a move line
		moveData, err := parseMoveLine(words)
		if err != nil {
			continue
		}
		moveCandidate = moveData
	}

	moveChan <- moveCandidate
}

func pipeOutErr(r io.Reader) {
	s := bufio.NewScanner(r)
	for s.Scan() {
		engineLine := s.Text()
		fmt.Println(engineLine)
	}
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
