package games

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/akamensky/base58"
	"github.com/notnil/chess"
	"github.com/thinktt/yowking/pkg/db"
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

	jsonData, _ := json.Marshal(game)
	if channels, exists := streams[gameID]; exists {
		for _, ch := range channels {
			ch <- string(jsonData)
		}
	}

	return nil
}

func CheckMoves(moves []string) error {
	gameParser := chess.NewGame()
	for i, move := range moves {
		err := gameParser.MoveStr(move)
		if err != nil {
			errMsg := fmt.Sprintf("Invalid move at index %d: %v", i, err)
			return errors.New(errMsg)
		}
	}
	return nil
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
