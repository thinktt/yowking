package games

import (
	"fmt"
	"net/http"

	"github.com/thinktt/yowking/pkg/db"
	"github.com/thinktt/yowking/pkg/utils"
)

// OfferDraw NEED TO CHECK THIS. Things could go wrong if cmp is playing both sides
func OfferDraw(id, user string) error {
	// get the current game from the DB
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

	//check that game is still live
	if game.Winner != "pending" {
		err = utils.NewHTTPError(
			http.StatusBadRequest, "no moves allowed, game is finished")
		return err
	}
	// check if user is playing this game
	userColor := GetUsercolor(game, user)
	if userColor == "" {
		return utils.NewHTTPError(http.StatusBadRequest, "not your game")
	}

	// check if it is this user's turn
	turnColor := GetTurnColor(len(game.MoveList))
	if turnColor != userColor {
		err = utils.NewHTTPError(http.StatusBadRequest, "not your turn")
		return err
	}

	// check if the other player has a draw offer
	opponentColor := GetOpponentColor(turnColor)
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
	_, err = db.UpdateWillDraw(id, turnColor, true)
	if err != nil {
		return utils.NewHTTPError(http.StatusInternalServerError, "DB Error: "+err.Error())
	}

	fmt.Println(userColor, " offered a draw")

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
