// This is a simple wrapper for the Chess engine. For some reason the pipes
// from Node to wine to the engine break but when we wrap the engine in a
// Go win binary they work. The order of operations is then
// Node in linux --> Wine --> enginewrap.exe --> engine

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"os"
	"os/exec"
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
	for s.Scan() {
		fmt.Println(s.Text())
	}
}

func forwardUserCommands(engine io.WriteCloser) {
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		line := s.Text()
		engine.Write([]byte(line + "\n"))
	}
}
