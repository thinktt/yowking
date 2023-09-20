package moveque

import (
	"os"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/thinktt/yowking/pkg/models"
)

var log = logrus.New()

func init() {
	token := os.Getenv("NATS_TOKEN")
	if token == "" {
		log.Fatal("NATS_TOKEN environment variable is not set")
	}

	natsUrl := os.Getenv("NATS_URL")
	if natsUrl == "" {
		log.Println("NATS_URL not set, using:", nats.DefaultURL)
		natsUrl = nats.DefaultURL
	} else {
		log.Println("NATS_URL set to:", natsUrl)
	}

	nc, err := nats.Connect(natsUrl, nats.Token(token))
	if err != nil {
		log.Fatalf("Error connecting to NATS: %v", err)
	}

	// Create a JetStream Context
	js, err := nc.JetStream()
	if err != nil {
		log.Fatalf("Error creating JetStream context: %v", err)
	}

	// Create move-req-stream
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     "move-req-stream",
		Subjects: []string{"move-req"},
	})
	if err != nil {
		log.Printf("Failed to create stream: %v", err)
	} else {
		log.Println("move-req-stream found or created")
	}

	// Create move-res-stream
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     "move-res-stream",
		Subjects: []string{"move-res.*"},
	})
	if err != nil {
		log.Printf("Failed to create stream: %v", err)
	} else {
		log.Println("move-res-stream found or created")
	}
}

func GetMove(moveReq models.MoveReq) (models.MoveData, error) {
	moveRes := models.MoveData{}

	return moveRes, nil
}

// result := make(chan MoveResult)
// wating := atomic.AddInt64(&waitingRequests, 1)
// total := atomic.AddInt64(&totalRequests, 1)
// fmt.Println("total request: ", total, "waiting request: ", wating)
// moveRequests <- MoveRequest{Settings: settings, Result: result}
// res := <-result

// func GetQueLength() int64 {
// 	return atomic.LoadInt64(&waitingRequests)
// }

// type MoveRequest struct {
// 	Settings engine.Settings
// 	Result   chan MoveResult
// }

// type MoveResult struct {
// 	Data engine.MoveData
// 	Err  error
// }

// var moveRequests = make(chan MoveRequest, 1) // Buffer size of 1
// var waitingRequests int64 = 0                // number of requests waiting to be processed
// var totalRequests int64 = 0                  // total number of requests

// func init() {
// 	go processMoves()
// }

// func processMoves() {
// 	for req := range moveRequests {
// 		data, err := engine.GetMove(req.Settings)
// 		req.Result <- MoveResult{Data: data, Err: err}
// 		atomic.AddInt64(&waitingRequests, -1)
// 	}
// }
