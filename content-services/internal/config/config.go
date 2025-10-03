package config

import (
	"os"
	"strconv"
	"time"

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

func GetS3Endpoint() string {
	return os.Getenv("S3_ENDPOINT")
}

func GetS3Region() string {
	return getenv("S3_REGION", "us-east-1")
}

func GetS3Bucket() string {
	return getenv("S3_BUCKET", "content-media")
}

func GetS3AccessKeyID() string {
	return os.Getenv("S3_ACCESS_KEY_ID")
}

func GetS3SecretAccessKey() string {
	return os.Getenv("S3_SECRET_ACCESS_KEY")
}

func GetS3UsePathStyle() bool {
	if v := os.Getenv("S3_USE_PATH_STYLE"); v != "" {
		parsed, err := strconv.ParseBool(v)
		if err == nil {
			return parsed
		}
	}
	return false
}

func GetS3PresignTTL() time.Duration {
	if v := os.Getenv("S3_PRESIGN_TTL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return 15 * time.Minute
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
