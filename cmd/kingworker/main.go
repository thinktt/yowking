package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/thinktt/yowking/pkg/books"
	"github.com/thinktt/yowking/pkg/engine"
	"github.com/thinktt/yowking/pkg/models"
	"github.com/thinktt/yowking/pkg/personalities"
)

func main() {
	token := os.Getenv("NATS_TOKEN")
	if token == "" {
		log.Fatal("NATS_TOKEN environment variable is not set")
	}

	natsUrl := os.Getenv("NATS_URL")
	if natsUrl == "" {
		log.Println("NATS_URL not set, useing:", nats.DefaultURL)
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
			log.Printf("Error fetching messages: %v", err)
			continue
		}

		if len(msgs) == 0 {
			log.Println("an empty message slice was returned")
			continue
		}
		m := msgs[0]
		moveRes, err := handleMoveReq(m)
		if err != nil {
			log.Printf("Error handling move request: %v", err)
			continue
		}

		err = PubMoveRes(js, moveRes)
		if err != nil {
			log.Printf("Error publishing move response: %v", err)
			// continue
		}

		m.Ack()
	}

}

func handleMoveReq(m *nats.Msg) (models.MoveData, error) {

	meta, err := m.Metadata()
	if err != nil {
		log.Printf("Error retrieving message metadata: %v", err)
	}
	log.Println("Received message seq:", meta.Sequence.Stream, "msgId:", meta.Sequence.Consumer)

	// Create a new instance of engine.Settings
	var moveReq models.MoveReq

	// Unmarshal the JSON data errors will be relayed to via move response
	err = json.Unmarshal(m.Data, &moveReq)
	if err != nil {
		errMsg := fmt.Sprintf("Error unmarshaling data: %v", err)
		log.Print(errMsg)
		return models.MoveData{Err: &errMsg}, nil
	}

	// Get the personality info, errors will be relayed to via move response
	cmp, ok := personalities.CmpMap[moveReq.CmpName]
	if !ok {
		errMsg := fmt.Sprintf("%s is not a valid personality", moveReq.CmpName)
		log.Print(errMsg)
		return models.MoveData{Err: &errMsg}, nil
	}

	// attemt to get a book move, if successful return it as the move
	bookMove, err := books.GetMove(moveReq.Moves, cmp.Book)
	if err == nil {
		bookMove.GameId = moveReq.GameId
		log.Println(bookMove)
		m.Ack()
		return bookMove, nil
	}

	// we were unable to get a book move, let's try the engine
	settings := models.Settings{
		Moves:     moveReq.Moves,
		CmpVals:   cmp.Vals,
		ClockTime: personalities.GetClockTime(cmp),
	}

	// Get the move data from the engine,
	moveData, err := engine.GetMove(settings)
	if err != nil {
		log.Println("There was ane error getting the move: ", err)
		return models.MoveData{}, err
	}

	// if movdData has inbbeded err relay it back to the client
	if moveData.Err != nil {
		log.Println(*moveData.Err)
		return moveData, nil
	}

	moveData.WillAcceptDraw = personalities.GetDrawEval(moveData.Eval, settings)
	moveData.Type = "engine"
	moveData.GameId = moveReq.GameId

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
