package main

import (
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/thinktt/yowking/pkg/books"
	"github.com/thinktt/yowking/pkg/engine"
	"github.com/thinktt/yowking/pkg/models"
	"github.com/thinktt/yowking/pkg/personalities"
)

func main() {
	// Connect to NATS server
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatalf("Error connecting to NATS: %v", err)
	}

	// Create a JetStream Context
	js, err := nc.JetStream()
	if err != nil {
		log.Fatalf("Error creating JetStream context: %v", err)
	}

	// Subscribe
	log.Print("Subscribing to stream queue...")
	_, err = js.QueueSubscribe(
		"move-req",
		"kingworkers",
		handleMoveReq,
		nats.MaxAckPending(1),
		nats.ManualAck(),
		nats.AckWait(30*time.Second),
	)
	if err != nil {
		log.Printf("Error subscribing to stream queue: %v", err)
	}
	runtime.Goexit()
}

func handleMoveReq(m *nats.Msg) {

	meta, err := m.Metadata()
	if err != nil {
		log.Fatalf("Error retrieving message metadata: %v", err)
	}
	log.Println("Received message seq:", meta.Sequence.Stream, "msgId:", meta.Sequence.Consumer)

	// Create a new instance of engine.Settings
	var moveReq models.MoveReq

	// Unmarshal the JSON data into settings
	err = json.Unmarshal(m.Data, &moveReq)
	if err != nil {
		log.Printf("Error unmarshaling data: %v", err)
		return
	}

	cmp, ok := personalities.CmpMap[moveReq.CmpName]
	if !ok {
		errMsg := fmt.Sprintf("%s is not a valid personality", moveReq.CmpName)
		log.Print(errMsg)
		return
	}

	bookMove, err := books.GetMove(moveReq.Moves, cmp.Book)
	// if no err we have a book move and can just return the move
	if err == nil {
		bookMove.GameId = moveReq.GameId
		log.Println(bookMove)
		m.Ack()
		return
	}

	// we were unable to get a book move, let's try the engine
	settings := models.Settings{
		Moves:     moveReq.Moves,
		CmpVals:   cmp.Vals,
		ClockTime: personalities.GetClockTime(cmp),
	}

	moveData, err := engine.GetMove(settings)
	if err != nil {
		log.Println("There was ane error getting the move: ", err)
		return
	}

	// engine didn't accept the input, return a 400 error
	if moveData.Err != nil {
		log.Println(*moveData.Err)
		return
	}

	moveData.WillAcceptDraw = personalities.GetDrawEval(moveData.Eval, settings)
	moveData.Type = "engine"
	moveData.GameId = moveReq.GameId
	m.Ack()

}
