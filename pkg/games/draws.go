package games

import (
	"fmt"
	"net/http"

	"github.com/thinktt/yowking/pkg/db"
	"github.com/thinktt/yowking/pkg/utils"
)

func OfferDraw(id, user string) error {
	// get the current game from the DB
	game, err := db.GetGame2(id)
	if err != nil {
		return utils.NewHTTPError(http.StatusInternalServerError, "DB Error: "+err.Error())
	}
	if game.ID == "" {
		return utils.NewHTTPError(http.StatusNotFound, fmt.Sprintf("no game found for id %s", id))
	}

	// check that game is still live
	if game.Winner != "pending" {
		return utils.NewHTTPError(http.StatusBadRequest, "no draws allowed, game is finished")
	}

	// check if user is playing this game
	userColor := GetUsercolor(game, user)
	if userColor == "" {
		return utils.NewHTTPError(http.StatusBadRequest, "not your game")
	}

	// check if the other player has a draw offer
	opponentColor := GetOpponentColor(userColor)
	opponentWillDraw := false
	if opponentColor == "white" {
		opponentWillDraw = game.WhiteWillDraw
	} else if opponentColor == "black" {
		opponentWillDraw = game.BlackWillDraw
	}

	if opponentWillDraw {
		// draw the game
		_, err = db.DrawGame(id, "mutual")
		if err != nil {
			return utils.NewHTTPError(http.StatusInternalServerError, "DB Error: "+err.Error())
		}
		PublishGameUpdates(game.ID)
		return nil
	}

	// update will draw state
	_, err = db.UpdateWillDraw(id, userColor, true)
	if err != nil {
		return utils.NewHTTPError(http.StatusInternalServerError, "DB Error: "+err.Error())
	}

	return nil
}

func ClearDrawOffer(id, user string) error {
	// get the current game from the DB
	game, err := db.GetGame2(id)
	if err != nil {
		return utils.NewHTTPError(http.StatusInternalServerError, "DB Error: "+err.Error())
	}
	if game.ID == "" {
		return utils.NewHTTPError(http.StatusNotFound, fmt.Sprintf("no game found for id %s", id))
	}

	// check that game is still live
	if game.Winner != "pending" {
		return utils.NewHTTPError(http.StatusBadRequest, "cannot clear draw offer, game is finished")
	}

	// check if user is playing this game
	userColor := GetUsercolor(game, user)
	if userColor == "" {
		return utils.NewHTTPError(http.StatusBadRequest, "not your game")
	}

	// clear the draw offer
	_, err = db.UpdateWillDraw(id, userColor, false)
	if err != nil {
		return utils.NewHTTPError(http.StatusInternalServerError, "DB Error: "+err.Error())
	}

	return nil
}
