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

const (
	moveAckWait          = 30 * time.Second
	moveProgressInterval = 15 * time.Second
	defaultWorkerTag     = "default"
	defaultAPITag        = "default"
)

func main() {
	token := os.Getenv("NATS_TOKEN")
	if token == "" {
		log.Fatal("NATS_TOKEN environment variable is not set")
	}

	workerTag := workerTagFromEnv(os.Getenv("WORKER_TAG"))
	moveReqSubject := getMoveReqSubject(workerTag)
	consumerName := getConsumerName(workerTag)

	natsUrl := os.Getenv("NATS_URL")
	if natsUrl == "" {
		log.Println("NATS_URL not set, using:", nats.DefaultURL)
		natsUrl = nats.DefaultURL
	} else {
		log.Println("NATS_URL set to:", natsUrl)
	}

	log.Printf("worker configuration: tag=%q", workerTag)

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
		Subjects: []string{"move-req.*"},
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
		moveReqSubject,
		consumerName,
		nats.ManualAck(),
		nats.AckWait(moveAckWait),
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
			if ackErr := m.Ack(); ackErr != nil {
				log.Errorf("Error acknowledging malformed move request: %v", ackErr)
			}
			continue
		}
		moveReq = normalizeMoveRequest(moveReq)

		// since we have move-req data we can now log with context
		logContext := logrus.WithFields(logrus.Fields{
			"gameId":    moveReq.GameId,
			"moveNo":    len(moveReq.Moves),
			"workerTag": moveReq.WorkerTag,
		})

		stopProgress := startProgressHeartbeat(m, logContext)
		moveRes, err := moves.HandleMoveReq(moveReq)
		stopProgress()
		if err != nil {
			logContext.Errorf("Error handling move request: %v", err)
			errMsg := err.Error()
			moveRes.Err = &errMsg
		}
		moveRes = prepareMoveResponse(moveReq, moveRes)

		err = PubMoveRes(js, moveReq, moveRes)
		if err != nil {
			logContext.Errorf("Error publishing move response: %v", err)
			continue
		}
		logContext.Println("succesfully published move response")

		if err := m.Ack(); err != nil {
			logContext.Errorf("Error acknowledging move request: %v", err)
		}
	}
}

func startProgressHeartbeat(msg *nats.Msg, logContext *logrus.Entry) func() {
	done := make(chan struct{})
	stopped := make(chan struct{})

	go func() {
		defer close(stopped)
		ticker := time.NewTicker(moveProgressInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := msg.InProgress(); err != nil {
					logContext.WithError(err).Warn("failed to extend move request acknowledgement")
				}
			case <-done:
				return
			}
		}
	}()

	return func() {
		close(done)
		<-stopped
	}
}

func prepareMoveResponse(moveReq models.MoveReq, moveRes models.MoveData) models.MoveData {
	moveReq = normalizeMoveRequest(moveReq)
	moveRes.Index = len(moveReq.Moves)
	moveRes.GameId = moveReq.GameId
	moveRes.WorkerTag = moveReq.WorkerTag
	return moveRes
}

func workerTagFromEnv(value string) string {
	if value == "" {
		return defaultWorkerTag
	}
	return value
}

func normalizeMoveRequest(moveReq models.MoveReq) models.MoveReq {
	moveReq.WorkerTag = workerTagFromEnv(moveReq.WorkerTag)
	moveReq.ApiTag = apiTagFromRequest(moveReq.ApiTag)
	return moveReq
}

func apiTagFromRequest(value string) string {
	if value == "" {
		return defaultAPITag
	}
	return value
}

func getMoveReqSubject(workerTag string) string {
	return fmt.Sprintf("move-req.%s", workerTagFromEnv(workerTag))
}

func getConsumerName(workerTag string) string {
	return fmt.Sprintf("kingworkers-%s", workerTagFromEnv(workerTag))
}

// PubMoveRes publishes a response to the API tag from the move request.
// The response payload carries the worker tag and game identity separately.
func PubMoveRes(js nats.JetStreamContext, moveReq models.MoveReq, moveData models.MoveData) error {
	moveReq = normalizeMoveRequest(moveReq)
	moveData.WorkerTag = workerTagFromEnv(moveData.WorkerTag)

	// Convert your moveData to JSON
	data, err := json.Marshal(moveData)
	if err != nil {
		return err
	}

	subject := getMoveResSubject(moveReq.ApiTag)

	// Publish the data
	_, err = js.Publish(subject, data)
	if err != nil {
		return err
	}

	return nil
}

func getMoveResSubject(apiTag string) string {
	return fmt.Sprintf("move-res.%s", apiTagFromRequest(apiTag))
}
