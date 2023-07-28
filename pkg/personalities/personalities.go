package personalities

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/thinktt/yowking/pkg/models"
)

var CmpMap = make(map[string]models.Cmp)

type Clocktimes struct {
	Easy int
	Hard int
	Gm   int
}

var clockTimes Clocktimes

func init() {
	loadCmps()
	loadClockTimes()
}

func loadCmps() {
	file, err := os.Open("personalities.json")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&CmpMap)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
	}
}

func loadClockTimes() {
	clockTimesFile, err := os.Open("calibrations/clockTimes.json")
	if err != nil {
		fmt.Println(err)
	}
	defer clockTimesFile.Close()
	err = json.NewDecoder(clockTimesFile).Decode(&clockTimes)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("clockTimes loaded: %+v\n", clockTimes)
}

func GetDrawEval(currentEval int, settings models.MoveReq) bool {
	contemtForDraw, err := strconv.Atoi(settings.CmpVals.Cfd)
	if err != nil {
		return false
	}

	if len(settings.Moves) <= 30 {
		return false
	}

	return (currentEval + contemtForDraw) < 0
}

func GetClockTime(cmp models.Cmp) int {
	if cmp.Ponder == "easy" {
		return clockTimes.Easy
	}

	if cmp.Rating >= 2700 {
		return clockTimes.Gm
	}

	return clockTimes.Hard
}
