package games

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/akamensky/base58"
	"github.com/notnil/chess"
	"github.com/thinktt/yowking/pkg/db"
	"github.com/thinktt/yowking/pkg/models"
)

var streams = make(map[string][]chan string)

func GetStream(gameID string) chan string {
	ch := make(chan string)
	if _, exists := streams[gameID]; !exists {
		streams[gameID] = []chan string{}
	}
	streams[gameID] = append(streams[gameID], ch)
	return ch
}

func RemoveStream(gameID string, ch chan string) {
	if channels, exists := streams[gameID]; exists {
		for i, c := range channels {
			if c == ch {
				close(c)
				// rebuild list of this game's streams removing this channel
				streams[gameID] = append(channels[:i], channels[i+1:]...)
				if len(streams[gameID]) == 0 {
					delete(streams, gameID)
				}
				break
			}
		}
	}
}

func SendStreamUpdate(gameID string) error {
	game, err := db.GetGame2(gameID)
	if err != nil {
		return err
	}

	gameMuation := models.Game2MutableFields{
		ID:         game.ID,
		LastMoveAt: game.LastMoveAt,
		Moves:      game.Moves,
		// Status:     game.Status,
		// Winner:     game.Winner,
	}

	jsonData, _ := json.Marshal(gameMuation)
	if channels, exists := streams[gameID]; exists {
		for _, ch := range channels {
			ch <- string(jsonData)
		}
	}

	return nil
}

func CheckMoves(moves []string) (string, error) {
	gameParser := chess.NewGame()
	for i, move := range moves {
		err := gameParser.MoveStr(move)
		if err != nil {
			errMsg := fmt.Sprintf("Invalid move at index %d: %v", i, err)
			return "", errors.New(errMsg)
		}
	}

	// fmt.Print(gameParser.Moves())
	// fmt.Println(gameParser.MoveHistory())

	moveList := gameParser.String()
	algebraMoves := strings.Split(moveList, " ")

	// find the last move in the list of moves, this loop fixes issue
	// when there are extra spaces
	lastMove := ""
	for i := len(algebraMoves) - 2; i >= 0; i-- {
		if algebraMoves[i] != "" {
			lastMove = algebraMoves[i]
			break
		}
	}

	lastOriginalMove := moves[len(moves)-1]

	// fmt.Println(moveList)
	// fmt.Println(lastOriginalMove)
	// fmt.Println(lastMove)

	if lastMove != lastOriginalMove {
		err := fmt.Errorf(
			"strict move notation enforced: wanted %s, got %s",
			lastMove, lastOriginalMove,
		)
		return "", err
	}

	return lastOriginalMove, nil

}

func GetGameID() (string, error) {
	// Create a byte array of length 6
	b := make([]byte, 6)

	// Fill the byte array with random bytes
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// Encode the byte array to a Base58 string
	encoded := base58.Encode(b)

	// If the encoded length is 9, chop off the first letter
	if len(encoded) == 9 {
		encoded = encoded[1:]
	}
	return encoded, nil
}
