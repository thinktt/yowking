package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/thinktt/yowking/pkg/engine"
	"github.com/thinktt/yowking/pkg/models"
)

type MoveData = models.MoveData
type Settings = models.Settings

var testJson = models.TestJson

func main() {

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "API is healthy",
		})
	})

	r.POST("/move", func(c *gin.Context) {
		var settings Settings
		if err := c.ShouldBindJSON(&settings); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		moveData, err := engine.GetMove(settings)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, moveData)
	})

	r.Run()

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
}
