package personalities

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/thinktt/yowking/pkg/models"
)

var CmpMap = make(map[string]models.Cmp)

func init() {
	loadCmps()
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

func GetDrawEval(currentEval int, settings models.Settings) bool {
	contemtForDraw, err := strconv.Atoi(settings.CmpVals.Cfd)
	if err != nil {
		return false
	}

	if len(settings.Moves) <= 30 {
		return false
	}

	return (currentEval + contemtForDraw) < 0
}
