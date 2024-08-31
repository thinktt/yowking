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

// GetAllLiveGameIDs find sall the games in the database that have game.Status
// started and returns a list of all of their IDs
func GetAllLiveGameIDs() ([]string, error) {
	gamesCollection := yowDatabase.Collection("games2")

	// Define the filter to find games with status "started"
	filter := bson.M{"status": "started"}

	// Define a projection to return only the "id" field
	projection := bson.M{"id": 1}

	// Find all matching games
	cursor, err := gamesCollection.Find(context.Background(), filter, options.Find().SetProjection(projection))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	// Collect all game IDs
	var gameIDs []string
	for cursor.Next(context.Background()) {
		var result struct {
			ID string `bson:"id"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		gameIDs = append(gameIDs, result.ID)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return gameIDs, nil
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
	// game.MoveList = nil

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

func UpdateGame(gameID, move, status, winner string) (*mongo.UpdateResult, error) {
	gamesCollection := yowDatabase.Collection("games2")

	filter := bson.M{"id": gameID}
	update := bson.M{
		"$push": bson.M{
			"moveList": move,
		},
		"$set": bson.M{
			"lastMoveAt": time.Now().UnixMilli(),
			"status":     status,
			"winner":     winner,
		},
	}

	return gamesCollection.UpdateOne(context.Background(), filter, update)
}
