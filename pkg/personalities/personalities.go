package personalities

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/thinktt/yowking/pkg/models"
)

type Cmp = models.Cmp
type PVals = models.PVals

var CmpMap = make(map[string]Cmp)

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
		fmt.Printf("Error decoding JSON: %v", err)
	}

	fmt.Println(CmpMap)
}

// pull json from the file personalities.json parse it into an slice
// of Cmp structs

// func getPersonality(cmpName string) Cmp {
