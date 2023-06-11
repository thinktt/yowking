package personalities

import (
	"encoding/json"
	"fmt"
	"os"

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
