package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/thinktt/yowking/internal/moves"
	"github.com/thinktt/yowking/pkg/models"
)

var log = logrus.New()

const (
	moveAckWait          = time.Minute
	moveProgressInterval = 15 * time.Second
)

func main() {
	token := os.Getenv("NATS_TOKEN")
	if token == "" {
		log.Fatal("NATS_TOKEN environment variable is not set")
	}

	// if WORKER_TAG exist then modify the subject and consumer names accordingly
	workerTag := os.Getenv("WORKER_TAG")
	forceRandomOff, err := boolEnv("FORCE_RANDOM_OFF")
	if err != nil {
		log.Fatal(err)
	}
	forceRandomOn, err := boolEnv("FORCE_RANDOM_ON")
	if err != nil {
		log.Fatal(err)
	}
	if forceRandomOff && forceRandomOn {
		log.Fatal("FORCE_RANDOM_OFF and FORCE_RANDOM_ON cannot both be true")
	}
	moveReqSubject := "move-req"
	consumerName := "kingworkers"
	if workerTag != "" {
		moveReqSubject += "." + workerTag
		consumerName += "-" + workerTag
	}

	natsUrl := os.Getenv("NATS_URL")
	if natsUrl == "" {
		log.Println("NATS_URL not set, using:", nats.DefaultURL)
		natsUrl = nats.DefaultURL
	} else {
		log.Println("NATS_URL set to:", natsUrl)
	}

	log.Printf(
		"worker configuration: tag=%q forceRandomOff=%t forceRandomOn=%t",
		workerTag,
		forceRandomOff,
		forceRandomOn,
	)

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
		Subjects: []string{"move-req", "move-req.*"},
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
		moveReq = applyWorkerOverrides(moveReq, forceRandomOff, forceRandomOn)

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

		err = PubMoveRes(js, moveRes)
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
	moveRes.Index = len(moveReq.Moves)
	moveRes.GameId = moveReq.GameId
	moveRes.WorkerTag = moveReq.WorkerTag
	return moveRes
}

func boolEnv(name string) (bool, error) {
	value := os.Getenv(name)
	if value == "" {
		return false, nil
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean: %w", name, err)
	}
	return parsed, nil
}

func applyWorkerOverrides(moveReq models.MoveReq, forceRandomOff, forceRandomOn bool) models.MoveReq {
	if forceRandomOff {
		moveReq.RandomIsOff = true
	}
	if forceRandomOn {
		moveReq.RandomIsOff = false
		moveReq.RandomIsForced = true
	}
	return moveReq
}

// PubMoveRes publishes legacy responses by game ID and tagged responses by
// worker tag. The response payload carries the game identity in both cases.
func PubMoveRes(js nats.JetStreamContext, moveData models.MoveData) error {
	// Convert your moveData to JSON
	data, err := json.Marshal(moveData)
	if err != nil {
		return err
	}

	subject := getMoveResSubject(moveData)

	// Publish the data
	_, err = js.Publish(subject, data)
	if err != nil {
		return err
	}

	return nil
}

func getMoveResSubject(moveData models.MoveData) string {
	if moveData.WorkerTag != "" {
		return fmt.Sprintf("move-res.%s", moveData.WorkerTag)
	}
	return fmt.Sprintf("move-res.%s", moveData.GameId)
}
