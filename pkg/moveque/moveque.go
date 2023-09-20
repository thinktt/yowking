package moveque

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/thinktt/yowking/pkg/models"
)

var log = logrus.New()
var moveStream nats.JetStreamContext
var nc *nats.Conn

func init() {
	var err error

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

	nc, err = nats.Connect(natsUrl, nats.Token(token))
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

	moveStream = js
}

func GetMove(moveReq models.MoveReq) (models.MoveData, error) {
	moveRes := models.MoveData{}

	// Serialize moveReq to JSON
	data, err := json.Marshal(moveReq)
	if err != nil {
		return moveRes, err
	}

	subject := fmt.Sprintf("move-res.%s", moveReq.GameId)

	// Set up a subscription
	sub, err := nc.SubscribeSync(subject)
	if err != nil {
		log.Errorf("Error subscribing to subject: %v", err)
		return moveRes, err
	}
	defer sub.Unsubscribe()

	// Publish it to move-req-stream
	_, err = moveStream.Publish("move-req", data)
	if err != nil {
		// log.Error(err)
		return moveRes, err
	}

	// Wait for a single message
	msg, err := sub.NextMsg(time.Second * 10) // Waits up to 10 seconds
	if err != nil {
		log.Errorf("Error receiving message: %v", err)
		return moveRes, err
	}

	// ack this message so we will not get it again
	msg.Ack()

	// Process your message here, e.g.,:
	// fmt.Printf("Received message: %s\n", msg.Data)

	// Parse the message into moveRes
	if err := json.Unmarshal(msg.Data, &moveRes); err != nil {
		log.Errorf("Error parsing message data: %v", err)
		return moveRes, err
	}

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
