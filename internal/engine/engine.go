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
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/thinktt/yowking/pkg/models"
	"golang.org/x/sys/unix"
)

type MoveData = models.MoveData
type Settings = models.MoveReq

var isVerboseMode = false
var logger = logrus.New()
var log *logrus.Entry

const (
	defaultMoveTimeout = 15 * time.Minute
	engineStopTimeout  = 5 * time.Second
)

type engineOutput struct {
	moveData MoveData
	final    bool
}

func GetMove(settings Settings) (MoveData, error) {
	// Collect any Wine helpers that exited after the previous move completed.
	reapExitedWineChildren()

	// fmt.Println(settings)
	log = logger.WithFields(logrus.Fields{
		"gameId": settings.GameId,
		"moveNo": len(settings.Moves),
	})
	isVerboseMode = strings.EqualFold(os.Getenv("SHOULD_LOG_ENGINE"), "true")

	isWsl := strings.EqualFold(os.Getenv("IS_WSL"), "true")
	// shouldPostInput := os.Getenv("SHOULD_POST_INPUT")
	// log.Println("shouldPostInput: " + shouldPostInput)

	var cmd *exec.Cmd
	if isWsl {
		cmd = exec.Command("./TheKing350.exe")
	} else {
		cmd = exec.Command("wine", "enginewrap.exe")
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

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

	moveChan := make(chan engineOutput, 1)

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
		stopEngine(engine, cmd, log)
	}()

	// RandomOverride sets the random value for engine moves in this request.
	// It is intended for admin diagnostic tests and may later be replaced by
	// a broader cmpOverride feature.
	if settings.RandomOverride != nil {
		settings.CmpVals.Rnd = strconv.Itoa(*settings.RandomOverride)
		log.WithField("randomOverride", *settings.RandomOverride).Info("using random override")
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
	case result := <-moveChan:
		// having move Data right now means the engine didn't like the settings
		if result.final {
			return result.moveData, nil
		}
	default:
		log.Println("engine accepted the settings witouth error")
	}

	// start the engine
	engine.Write([]byte("go\n"))

	moveTimeout, err := getMoveTimeout()
	if err != nil {
		return MoveData{}, err
	}
	timer := time.NewTimer(moveTimeout)
	defer timer.Stop()

	// Keep the newest post-line move available as a fallback if the final move
	// does not arrive before the worker's deadline.
	moveCandidate := MoveData{}
	for {
		select {
		case result := <-moveChan:
			if hasMove(result.moveData) {
				moveCandidate = result.moveData
			}
			if result.final {
				return result.moveData, nil
			}
		case <-timer.C:
			if hasMove(moveCandidate) {
				warning := fmt.Sprintf(
					"engine timed out after %s; using latest analysis move",
					moveTimeout,
				)
				moveCandidate.Warning = &warning
				moveCandidate.Type = "engine"
				return moveCandidate, nil
			}

			errMsg := fmt.Sprintf(
				"engine timed out after %s without a usable move",
				moveTimeout,
			)
			return MoveData{Err: &errMsg}, nil
		}
	}
}

func getMoveTimeout() (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv("ENGINE_MOVE_TIMEOUT"))
	if value == "" {
		return defaultMoveTimeout, nil
	}

	timeout, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("ENGINE_MOVE_TIMEOUT must be a duration: %w", err)
	}
	if timeout <= 0 {
		return 0, errors.New("ENGINE_MOVE_TIMEOUT must be greater than zero")
	}
	return timeout, nil
}

func hasMove(moveData MoveData) bool {
	return moveData.CoordinateMove != "" || moveData.AlgebraMove != ""
}

func stopEngine(engine io.WriteCloser, cmd *exec.Cmd, log *logrus.Entry) {
	// Wine can reparent PE child processes to kingworker after its launcher exits.
	// Finish the command first, then reap any exited children before another move starts.
	if _, err := engine.Write([]byte("quit\n")); err != nil {
		log.WithError(err).Debug("failed to send quit to engine")
	}
	if err := engine.Close(); err != nil {
		log.WithError(err).Debug("failed to close engine input")
	}
	waitDone := make(chan error, 1)
	go func() {
		waitDone <- cmd.Wait()
	}()

	select {
	case err := <-waitDone:
		if err != nil {
			log.WithError(err).Debug("engine launcher exited with error")
		}
	case <-time.After(engineStopTimeout):
		log.Warn("engine launcher did not exit after quit; killing process group")
		if err := unix.Kill(-cmd.Process.Pid, unix.SIGKILL); err != nil {
			log.WithError(err).Debug("failed to kill engine process group")
			_ = cmd.Process.Kill()
		}
		select {
		case err := <-waitDone:
			if err != nil {
				log.WithError(err).Debug("engine launcher exited after kill")
			}
		case <-time.After(engineStopTimeout):
			log.Error("engine launcher did not exit after kill")
		}
	}

	reaped := reapExitedWineChildren()
	log.WithField("reapedChildren", reaped).Println("engine closed")
}

func reapExitedWineChildren() int {
	reaped := 0

	for {
		var status unix.WaitStatus

		// Wait4 returns at most one exited child process per call.
		// -1 means "any direct child of the current kingworker process."
		// WNOHANG makes this non-blocking, so it checks for exited children
		// and returns immediately instead of waiting for one to exit.
		pid, err := unix.Wait4(-1, &status, unix.WNOHANG, nil)

		// ECHILD means kingworker has no child processes left to wait on.
		// pid == 0 means child processes still exist, but none have exited yet.
		// In either case, there is nothing ready to reap right now.
		if err == unix.ECHILD || pid == 0 {
			return reaped
		}

		// Stop on any unexpected wait error.
		if err != nil {
			return reaped
		}

		reaped++
	}
}

func readEngineOut(r io.Reader, moveChan chan engineOutput, stopId int) {
	s := bufio.NewScanner(r)
	moveCandidate := MoveData{}
	final := false

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
			final = true
			break
		}

		// check if the engine line final move result
		words := strings.Fields(engineLine)
		if strings.Contains(engineLine, "move") && len(words) == 2 {
			moveCandidate.CoordinateMove = strings.Fields(engineLine)[1]
			final = true
			break
		}

		// parse the enginLine if it is a move line
		moveData, err := parseMoveLine(words)
		if err != nil {
			continue
		}
		moveCandidate = moveData
		sendLatestEngineOutput(moveChan, engineOutput{moveData: moveCandidate})

		// if the move line is the stopId move line, break and send this move
		if moveData.Id == stopId {
			log.Println("engine found stopId, move:", moveData.AlgebraMove)
			final = true
			break
		}
	}
	if err := s.Err(); err != nil {
		log.WithError(err).Error("failed to read engine output")
	}

	if !final && !hasMove(moveCandidate) {
		errStr := "engine output closed before producing a move"
		moveCandidate.Err = &errStr
	}
	sendLatestEngineOutput(moveChan, engineOutput{
		moveData: moveCandidate,
		final:    true,
	})
}

func sendLatestEngineOutput(moveChan chan engineOutput, result engineOutput) {
	select {
	case moveChan <- result:
		return
	default:
	}

	select {
	case <-moveChan:
	default:
	}
	moveChan <- result
}

func readEngineErrs(r io.Reader) {
	s := bufio.NewScanner(r)
	for s.Scan() {
		engineLine := s.Text()
		log.Error("Engine ERR:", engineLine)
	}
	if err := s.Err(); err != nil {
		log.WithError(err).Error("failed to read engine error output")
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
	if err := s.Err(); err != nil {
		logger.WithError(err).Error("failed to read user commands")
	}
}
