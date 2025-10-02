package config

import (
	"os"

	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
	}
}

func GetPort() string {
	if p := os.Getenv("PORT"); p != "" {
		return p
	}
	return "8004"
}

// GetMongoURI returns the MongoDB connection string.
// Example: mongodb://user:pass@localhost:27017
func GetMongoURI() string {
	if v := os.Getenv("MONGO_URI"); v != "" {
		return v
	}
	return "mongodb://localhost:27017"
}

// GetMongoDBName returns the Mongo database name to use.
func GetMongoDBName() string {
	if v := os.Getenv("MONGO_DB"); v != "" {
		return v
	}
	return "content"
}

// GetGraphQLPlaygroundEnabled toggles GraphQL Playground exposure.
func GetGraphQLPlaygroundEnabled() bool {
	if v := os.Getenv("GRAPHQL_PLAYGROUND"); v != "" {
		return v == "1" || v == "true" || v == "TRUE"
	}
	return true
}
