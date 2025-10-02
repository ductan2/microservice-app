package db

import (
	"context"
	"time"

	"content-services/internal/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// NewMongoClient creates a new Mongo client with sane defaults.
func NewMongoClient(ctx context.Context) (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(config.GetMongoURI())
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}
	return client, nil
}

// GetDatabase returns a handle to the configured Mongo database.
func GetDatabase(client *mongo.Client) *mongo.Database {
	return client.Database(config.GetMongoDBName())
}
