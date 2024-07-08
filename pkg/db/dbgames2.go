package db

import (
	"context"
	"strings"
	"time"

	"github.com/thinktt/yowking/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CreateGame creates or updates a game entry in the database
func CreateGame2(game models.Game2) (*mongo.UpdateResult, error) {
	gamesCollection := yowDatabase.Collection("games2")

	filter := bson.M{"id": game.ID}
	update := bson.M{"$setOnInsert": game}
	upsert := true
	opts := options.UpdateOptions{Upsert: &upsert}

	return gamesCollection.UpdateOne(context.Background(), filter, update, &opts)
}

// GetGame retrieves a game by its ID
func GetGame2(id string) (models.Game2, error) {
	gamesCollection := yowDatabase.Collection("games2")

	filter := bson.M{"id": id}
	// var result bson.M
	var game models.Game2
	err := gamesCollection.FindOne(context.Background(), filter).Decode(&game)

	// No game found, return nil without error
	if err != nil && err == mongo.ErrNoDocuments {
		return models.Game2{}, nil
	}

	// return any other errors
	if err != nil {
		return models.Game2{}, err
	}

	// // Remove the MongoDB _id field
	// delete(result, "_id")

	game.Moves = strings.Join(game.MoveList, " ")
	game.MoveList = nil

	return game, nil
}

func DeleteGame2(id string) (*mongo.DeleteResult, error) {
	gamesCollection := yowDatabase.Collection("games2")

	filter := bson.M{"id": id}
	return gamesCollection.DeleteOne(context.Background(), filter)
}

func CreateMove(gameID string, move string) (*mongo.UpdateResult, error) {
	gamesCollection := yowDatabase.Collection("games2")

	filter := bson.M{"id": gameID}
	update := bson.M{
		"$push": bson.M{
			"moveList": move,
		},
		"$set": bson.M{
			"lastMoveAt": time.Now().UnixMilli(),
		},
	}

	return gamesCollection.UpdateOne(context.Background(), filter, update)
}
