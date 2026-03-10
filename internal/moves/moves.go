package moves

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/thinktt/yowking/internal/books"
	"github.com/thinktt/yowking/internal/engine"
	"github.com/thinktt/yowking/pkg/models"
	"github.com/thinktt/yowking/pkg/personalities"
)

// HandleMoveReq resolves a move request via book lookup first, then engine fallback.
func HandleMoveReq(moveReq models.MoveReq) (models.MoveData, error) {
	logContext := logrus.WithFields(logrus.Fields{
		"gameId": moveReq.GameId,
		"moveNo": len(moveReq.Moves),
	})

	cmp, ok := personalities.CmpMap[moveReq.CmpName]
	if !ok {
		errMsg := fmt.Sprintf("%s is not a valid personality", moveReq.CmpName)
		logContext.Error(errMsg)
		return models.MoveData{Err: &errMsg}, nil
	}
	logContext.Println("playing as", cmp.Name, "using book", cmp.Book)

	if moveReq.ShouldSkipBook {
		logContext.Println("shouldSkipBook is set, skipping book check")
	} else {
		bookMove, err := books.GetMove(moveReq.Moves, cmp.Book)
		if err == nil {
			bookMove.GameId = moveReq.GameId
			logContext.Println("book move found:", bookMove.CoordinateMove)
			return bookMove, nil
		}
		logContext.Println("no book move found, sending move to engine")
	}

	settings := moveReq
	settings.CmpVals = cmp.Vals
	if moveReq.ClockTime == 0 {
		settings.ClockTime = personalities.GetClockTime(cmp)
		logContext.Println("using calibrated clock time:", settings.ClockTime)
	} else {
		logContext.Println("using manual clock time:", settings.ClockTime)
	}

	moveData, err := engine.GetMove(settings)
	if err != nil {
		logContext.Error("There was ane error getting the move: ", err)
		return models.MoveData{}, err
	}

	if moveData.Err != nil {
		logContext.Error(*moveData.Err)
		return moveData, nil
	}

	moveData.WillAcceptDraw = personalities.GetDrawEval(moveData.Eval, settings)
	moveData.Type = "engine"
	moveData.GameId = moveReq.GameId

	logContext.Println("move received from engine:", moveData.CoordinateMove)
	return moveData, nil
}
