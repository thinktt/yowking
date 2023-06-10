package main

import (
	"encoding/json"
	"fmt"

	"github.com/thinktt/yowking/pkg/engine"
	"github.com/thinktt/yowking/pkg/models"
)

type MoveData = models.MoveData
type Settings = models.Settings

var testJson = models.TestJson

func main() {

	// marshall the settings json data
	var settings Settings
	err := json.Unmarshal([]byte(testJson), &settings)
	if err != nil {
		fmt.Println("Error unmarshalling settings json")
		return
	}
	fmt.Println("Success marshalling json")

	moveData, err := engine.GetMove(settings)
	if err != nil {
		fmt.Println("There was ane error getting the move: ", err)
		return
	}
	fmt.Println(moveData)
	// wait 10 seconds before exiting
	// time.Sleep(10 * time.Second)
}
