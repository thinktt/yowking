package games

import (
	"fmt"
	"net/http"

	"github.com/thinktt/yowking/pkg/db"
	"github.com/thinktt/yowking/pkg/utils"
)

func Resign(id, user string) error {
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
	userColor := GetUsercolor(game, user)
	if userColor == "" {
		return utils.NewHTTPError(http.StatusBadRequest, "not your game")
	}

	// update the game to reflect the resignation
	_, err = db.ResignGame(id, userColor)
	if err != nil {
		return utils.NewHTTPError(http.StatusInternalServerError, "DB Error: "+err.Error())
	}

	// publish the game update
	PublishGameUpdates(game.ID)

	return nil
}
