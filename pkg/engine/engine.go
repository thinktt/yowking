package engine

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/thinktt/yowking/pkg/models"
)

type MoveData = models.MoveData
type Settings = models.MoveReq

var isVerboseMode = false
var logger = logrus.New()
var log *logrus.Entry

func GetMove(settings Settings) (MoveData, error) {
	// fmt.Println(settings)
	log = logger.WithFields(logrus.Fields{
		"gameId": settings.GameId,
		"moveNo": len(settings.Moves),
	})

	isWsl := os.Getenv("IS_WSL")
	// shouldPostInput := os.Getenv("SHOULD_POST_INPUT")
	// log.Println("shouldPostInput: " + shouldPostInput)

	var cmd *exec.Cmd
	if isWsl == "true" {
		cmd = exec.Command("./TheKing350noOpk.exe")
	} else {
		cmd = exec.Command("wine", "enginewrap.exe")
	}

	engine, err := cmd.StdinPipe()
	if err != nil {
		log.Println(err)
		return MoveData{}, err
	}

	engineOut, err := cmd.StdoutPipe()
	if err != nil {
		log.Println(err)
		return MoveData{}, err
	}

	engineErr, err := cmd.StderrPipe()
	if err != nil {
		log.Println(err)
		return MoveData{}, err
	}

	moveChan := make(chan MoveData)
	errChan := make(chan error)
	defer close(moveChan)
	defer close(errChan)

	// handle the engine streams in real time
	go readEngineOut(engineOut, moveChan, settings.StopId)
	go readEngineErrs(engineErr)
	go forwardUserCommands(engine)

	// start the engine
	err = cmd.Start()
	if err != nil {
		errToGo := fmt.Errorf("cmd.Run() failed with %s", err)
		return MoveData{}, errToGo
	}

	// from here if getMove() errors or completes be sure to stop the engine
	defer func() {
		go stopEngine(engine, cmd, log)
	}()

	if settings.RandomIsOff {
		settings.CmpVals.Rnd = "0"
		log.Info("randomIsOff is set, setting cmp rnd val to 0")
	}

	// log all the cmpVals with keys
	// fmt.Printf("%+v\n", settings.CmpVals)

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
	if err := t.Execute(buf, settings.CmpVals); err != nil {
		return MoveData{}, err
	}

	// log.Println("clockTime: ", settings.ClockTime)
	timeStr := fmt.Sprintf("time %d\n", settings.ClockTime)
	otimStr := fmt.Sprintf("otim %d\n", settings.ClockTime)

	// send settings to the engine
	engine.Write([]byte("xboard\n"))
	engine.Write([]byte("post\n"))
	engine.Write([]byte(timeStr))
	engine.Write([]byte(otimStr))
	engine.Write(buf.Bytes())

	// send all the moves to the engine
	for _, move := range settings.Moves {
		moveStr := fmt.Sprintf("%s\n", move)
		engine.Write([]byte(moveStr))
	}

	select {
	case moveData := <-moveChan:
		// having move Data right now means the engine didn't like the settings
		return moveData, nil
	default:
		log.Println("engine accepted the settings witouth error")
	}

	// start the engine
	engine.Write([]byte("go\n"))

	// wait for the engine to send back a move
	moveData := <-moveChan
	return moveData, nil
}

func stopEngine(engine io.WriteCloser, cmd *exec.Cmd, log *logrus.Entry) {
	// send a quilt command to the engine
	engine.Write([]byte("quit\n"))
	engine.Close()
	cmd.Wait()
	log.Println("engine closed")
}

func readEngineOut(r io.Reader, moveChan chan MoveData, stopId int) {
	s := bufio.NewScanner(r)
	moveCandidate := MoveData{}

	for s.Scan() {
		engineLine := s.Text()
		if isVerboseMode {
			log.Println(engineLine)
		}

		// if the engine finds a setting error send empty move response with an error
		if strings.Contains(engineLine, "Error") ||
			strings.Contains(engineLine, "Illegal") {
			errStr := "callout by engine: " + engineLine
			moveCandidate = MoveData{Err: &errStr}
			break
		}

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

		// if the move line is the stopId move line, break and send this move
		if moveData.Id == stopId {
			log.Println("engine found stopId, move:", moveData.AlgebraMove)
			break
		}
	}

	// moveCandidate.Err = errStr
	moveChan <- moveCandidate
}

func readEngineErrs(r io.Reader) {
	s := bufio.NewScanner(r)
	for s.Scan() {
		engineLine := s.Text()
		log.Error("Engine ERR:", engineLine)
	}
}

func parseMoveLine(words []string) (MoveData, error) {
	if len(words) < 5 {
		return MoveData{}, errors.New("not enough words for move line")
	}

	// only a valid move line if first 4 words are numbers
	var err error
	var numbers [4]int
	for i := 0; i < 4; i++ {
		numbers[i], err = strconv.Atoi(words[i])
		if err != nil {
			return MoveData{}, errors.New("first 4 words of engine line are not numbers")
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

func forwardUserCommands(engine io.WriteCloser) {
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		line := s.Text()
		engine.Write([]byte(line + "\n"))
	}
}
