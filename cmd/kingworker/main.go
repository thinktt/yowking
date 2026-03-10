package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/thinktt/yowking/internal/moves"
	"github.com/thinktt/yowking/pkg/models"
)

var log = logrus.New()

func main() {
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

	sub, err := js.PullSubscribe(
		"move-req",
		"kingworkers",
		nats.ManualAck(),
		nats.AckWait(30*time.Second),
	)
	if err != nil {
		log.Fatalf("Error subscribing to stream queue: %v", err)
	}

	for {
		msgs, err := sub.Fetch(1)
		if err != nil && err == nats.ErrTimeout {
			continue
		}

		if err != nil {
			log.Errorf("Error fetching messages: %v", err)
			continue
		}

		if len(msgs) == 0 {
			log.Println("an empty message slice was returned")
			continue
		}
		m := msgs[0]

		meta, err := m.Metadata()
		if err != nil {
			log.Errorf("Error retrieving message metadata: %v", err)
		}
		log.Println("Received message seq:", meta.Sequence.Stream, "msgId:", meta.Sequence.Consumer)

		// Create a new instance of engine.Settings
		var moveReq models.MoveReq

		// Unmarshal the JSON data errors will be relayed to via move response
		err = json.Unmarshal(m.Data, &moveReq)
		if err != nil {
			errMsg := fmt.Sprintf("Error unmarshaling data: %v", err)
			log.Error(errMsg)
			continue
		}

		// since we have move-req data we can now log with context
		logContext := logrus.WithFields(logrus.Fields{
			"gameId": moveReq.GameId,
			"moveNo": len(moveReq.Moves),
		})

		moveRes, err := moves.HandleMoveReq(moveReq)
		if err != nil {
			logContext.Errorf("Error handling move request: %v", err)
			continue
		}

		err = PubMoveRes(js, moveRes)
		if err != nil {
			logContext.Errorf("Error publishing move response: %v", err)
			// continue
		}
		logContext.Println("succesfully published move response")

		m.Ack()
	}
}

// PubMoveRes publishes the move data to the move_res.<gameId> subject
func PubMoveRes(js nats.JetStreamContext, moveData models.MoveData) error {
	// Convert your moveData to JSON
	data, err := json.Marshal(moveData)
	if err != nil {
		return err
	}

	// Generate the subject name
	subject := fmt.Sprintf("move-res.%s", moveData.GameId)

	// Publish the data
	_, err = js.Publish(subject, data)
	if err != nil {
		return err
	}

	return nil
}
