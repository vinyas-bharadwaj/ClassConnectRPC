package mongodb

import (
	"ClassConnectRPC/pkg/utils"
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateMongoClient() (*mongo.Client, error) {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return nil, utils.ErrorHandler(err, "Unable to connect to database")
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Unable to ping the database")
	}

	log.Println("Connected to mongodb successfully")
	return client, nil
}
