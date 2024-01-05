package db

import (
	"context"
	"fmt"
	"os"

	"github.com/thinktt/yowking/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client
var yowDatabase *mongo.Database

// UserIDsResponse represents the structure of the response
type UserIDsResponse struct {
	Count int      `json:"count"`
	IDs   []string `json:"ids"`
}

func init() {
	// Connect to MongoDB
	mongoHost := os.Getenv("MONGO_HOST")
	if mongoHost == "" {
		mongoHost = "localhost:27017"
	}
	uri := fmt.Sprintf("mongodb://%s/?maxPoolSize=20&w=majority", mongoHost)
	var err error
	clientOptions := options.Client().ApplyURI(uri)
	client, err = mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		panic(err)
	}

	yowDatabase = client.Database("yow")
}

func CreateUser(user models.User) (*mongo.UpdateResult, error) {
	usersCollection := yowDatabase.Collection("users")

	// Manually create a bson.M map from the User struct
	userMap := bson.M{
		"id":                    user.ID,
		"kingCmVersion":         user.KingCmVersion,
		"hasAcceptedDisclaimer": user.HasAcceptedDisclaimer,
	}

	filter := bson.M{"id": userMap["id"]}
	update := bson.M{"$setOnInsert": userMap}
	options := options.Update().SetUpsert(true)

	return usersCollection.UpdateOne(context.Background(), filter, update, options)
}

func GetUser(id string) (bson.M, error) {
	usersCollection := yowDatabase.Collection("users")

	filter := bson.M{"id": id}
	var result bson.M
	err := usersCollection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // No user found, return nil without error
		}
		return nil, err // An error occurred during the query
	}

	delete(result, "_id") // Remove the MongoDB _id field
	return result, nil
}

// GetAllUsers retrieves all user IDs
func GetAllUsers() (UserIDsResponse, error) {
	usersCollection := yowDatabase.Collection("users")

	// Define the projection to include only the id field
	projection := bson.M{"id": 1, "_id": 0}

	// Find all documents with the specified projection
	cursor, err := usersCollection.Find(
		context.Background(),
		bson.M{}, options.Find().SetProjection(projection),
	)
	if err != nil {
		return UserIDsResponse{}, err
	}
	defer cursor.Close(context.Background())

	var results []bson.M
	if err = cursor.All(context.Background(), &results); err != nil {
		return UserIDsResponse{}, err
	}

	// Extract IDs from results
	var ids []string
	for _, result := range results {
		if id, ok := result["id"].(string); ok {
			ids = append(ids, id)
		}
	}

	return UserIDsResponse{
		Count: len(ids),
		IDs:   ids,
	}, nil
}
