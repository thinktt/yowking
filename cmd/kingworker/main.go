package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/thinktt/yowking/pkg/books"
	"github.com/thinktt/yowking/pkg/engine"
	"github.com/thinktt/yowking/pkg/models"
	"github.com/thinktt/yowking/pkg/personalities"
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

		moveRes, err := handleMoveReq(moveReq)
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

func handleMoveReq(moveReq models.MoveReq) (models.MoveData, error) {
	// set logger context
	logContext := logrus.WithFields(logrus.Fields{
		"gameId": moveReq.GameId,
		"moveNo": len(moveReq.Moves),
	})

	// Get the personality info, errors will be relayed to via move response
	cmp, ok := personalities.CmpMap[moveReq.CmpName]
	if !ok {
		errMsg := fmt.Sprintf("%s is not a valid personality", moveReq.CmpName)
		logContext.Error(errMsg)
		return models.MoveData{Err: &errMsg}, nil
	}
	logContext.Println("playing as", cmp.Name, "using book", cmp.Book)

	// attemt to get a book move, if successful return it as the move
	if moveReq.ShouldSkipBook {
		logContext.Println("shouldSkipBook is set, skipping book check")
	} else {
		bookMove, err := books.GetMove(moveReq.Moves, cmp.Book)
		if err == nil {
			bookMove.GameId = moveReq.GameId
			logContext.Println("book move found:", bookMove.CoordinateMove)
			return bookMove, nil
		}
		logContext.Println("no book move found, sending move to engine")
	}

	settings := moveReq
	settings.CmpVals = cmp.Vals
	if moveReq.ClockTime == 0 {
		settings.ClockTime = personalities.GetClockTime(cmp)
		logContext.Println("using calibrated clock time:", settings.ClockTime)
	} else {
		logContext.Println("using manual clock time:", settings.ClockTime)
	}

	// Get the move data from the engine,
	moveData, err := engine.GetMove(settings)
	if err != nil {
		logContext.Error("There was ane error getting the move: ", err)
		return models.MoveData{}, err
	}

	// if movdData has inbbeded err relay it back to the client
	if moveData.Err != nil {
		logContext.Error(*moveData.Err)
		return moveData, nil
	}

	moveData.WillAcceptDraw = personalities.GetDrawEval(moveData.Eval, settings)
	moveData.Type = "engine"
	moveData.GameId = moveReq.GameId

	logContext.Println("move received from engine:", moveData.CoordinateMove)
	return moveData, nil

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
