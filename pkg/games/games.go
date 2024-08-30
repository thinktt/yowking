package games

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/akamensky/base58"
	"github.com/notnil/chess"
	"github.com/thinktt/yowking/pkg/db"
	"github.com/thinktt/yowking/pkg/events"
	"github.com/thinktt/yowking/pkg/models"
	"github.com/thinktt/yowking/pkg/utils"
)

func PublishGameUpdates(gameID string) error {
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
	events.Pub.PublishMessage(game.ID, string(jsonData))

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

// GetTurnColor given a move index returns the color that move belongs to
func GetTurnColor(moveIndex int) string {
	if moveIndex%2 == 0 {
		return "white"
	}
	return "black"
}

// AddMove validates and adds a move to the game, and then triggers an event
// message that moves were added, it returns cutsom HTTPErrors so errors can
// play nicely with an http routers
func AddMove(id, user string, moveData models.MoveData2) error {
	game, err := db.GetGame2(id)
	if err != nil {
		err = utils.NewHTTPError(
			http.StatusInternalServerError, "DB Error: "+err.Error())
		return err
	}

	if game.ID == "" {
		err = utils.NewHTTPError(
			http.StatusNotFound,
			fmt.Sprintf("no game found for id %s", id))
		return err
	}

	userColor := ""
	if user == game.WhitePlayer.ID {
		userColor = "white"
	} else if user == game.BlackPlayer.ID {
		userColor = "black"
	}

	if userColor == "" {
		err = utils.NewHTTPError(http.StatusBadRequest, "not your game")
		return err
	}

	moves := strings.Fields(game.Moves)

	if len(moves) != moveData.Index {
		err = utils.NewHTTPError(http.StatusBadRequest,
			fmt.Sprintf("invalid move index, next move index is %d", len(moves)))
		return err
	}

	turnColor := GetTurnColor(moveData.Index)
	if turnColor != userColor {
		err = utils.NewHTTPError(http.StatusBadRequest, "not your turn")
		return err
	}

	moves = append(moves, moveData.Move)

	if _, err := CheckMoves(moves); err != nil {
		err = utils.NewHTTPError(http.StatusBadRequest, err.Error())
		return err
	}

	if _, err := db.CreateMove(id, moveData.Move); err != nil {
		err = utils.NewHTTPError(http.StatusInternalServerError, "DB Error: "+err.Error())
		return err
	}

	PublishGameUpdates(game.ID)

	return nil
}
