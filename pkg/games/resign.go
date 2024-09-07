package games

import (
	"fmt"
	"net/http"

	"github.com/thinktt/yowking/pkg/db"
	"github.com/thinktt/yowking/pkg/utils"
)

func Resign(id, userID, color string) error {
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
		return utils.NewHTTPError(http.StatusBadRequest, "cannot resign, game is finished")
	}

	// check if user is playing this game
	if !game.HasPlayer(userID) {
		return utils.NewHTTPError(http.StatusBadRequest, "not your game")
	}

	// check if user is proper color
	if !game.UserIsColor(userID, color) {
		return utils.NewHTTPError(http.StatusBadRequest,
			fmt.Sprintf("you are not playing as %s", color))
	}

	// update the game to reflect the resignation
	_, err = db.ResignGame(id, color)
	if err != nil {
		return utils.NewHTTPError(http.StatusInternalServerError, "DB Error: "+err.Error())
	}

	// publish the game update
	PublishGameUpdates(game.ID)

	return nil
}
