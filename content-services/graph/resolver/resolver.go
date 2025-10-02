package resolver

import "go.mongodb.org/mongo-driver/mongo"

// Resolver serves as dependency injection for your app, add any dependencies you require here.
type Resolver struct {
	DB *mongo.Database
}
