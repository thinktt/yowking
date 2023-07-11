package moveque

import (
	"fmt"
	"sync/atomic"

	"github.com/thinktt/yowking/pkg/engine"
)

type MoveRequest struct {
	Settings engine.Settings
	Result   chan MoveResult
}

type MoveResult struct {
	Data engine.MoveData
	Err  error
}

var moveRequests = make(chan MoveRequest, 1) // Buffer size of 1
var waitingRequests int64 = 0                // number of requests waiting to be processed
var totalRequests int64 = 0                  // total number of requests

func init() {
	go processMoves()
}

func processMoves() {
	for req := range moveRequests {
		data, err := engine.GetMove(req.Settings)
		req.Result <- MoveResult{Data: data, Err: err}
		atomic.AddInt64(&waitingRequests, -1)
	}
}

func GetMove(settings engine.Settings) (engine.MoveData, error) {
	result := make(chan MoveResult)
	wating := atomic.AddInt64(&waitingRequests, 1)
	total := atomic.AddInt64(&totalRequests, 1)
	fmt.Println("total request: ", total, "waiting request: ", wating)
	moveRequests <- MoveRequest{Settings: settings, Result: result}
	res := <-result
	return res.Data, res.Err
}

func GetQueLength() int64 {
	return atomic.LoadInt64(&waitingRequests)
}
