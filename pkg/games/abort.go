package games

import (
	"fmt"
	"net/http"

	"github.com/thinktt/yowking/pkg/db"
	"github.com/thinktt/yowking/pkg/utils"
)

func Abort(id, user string) error {
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
		return utils.NewHTTPError(http.StatusBadRequest, "cannot abort, game is finished")
	}

	// check if user is playing this game
	userColor := GetUsercolor(game, user)
	if userColor == "" {
		return utils.NewHTTPError(http.StatusBadRequest, "not your game")
	}

	// check if the game has more than two moves
	if len(game.MoveList) > 2 {
		return utils.NewHTTPError(http.StatusBadRequest, "cannot abort, too many moves")
	}

	// delete the game
	_, err = db.DeleteGame2(id)
	if err != nil {
		return utils.NewHTTPError(http.StatusInternalServerError, "DB Error: "+err.Error())
	}

	return nil
}
