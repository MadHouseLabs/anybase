package database

import (
	"github.com/madhouselabs/anybase/internal/database/adapters/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetMongoDatabase returns the underlying MongoDB database if using MongoDB adapter
// This is a temporary helper for migration purposes
func GetMongoDatabase() *mongo.Database {
	if db == nil {
		panic("database not initialized")
	}
	
	if mongoAdapter, ok := db.(*mongodb.MongoAdapter); ok {
		return mongoAdapter.GetDatabase()
	}
	
	panic("current database is not MongoDB")
}

// GetMongoClient returns the underlying MongoDB client if using MongoDB adapter
// This is a temporary helper for migration purposes
func GetMongoClient() *mongo.Client {
	if db == nil {
		panic("database not initialized")
	}
	
	if mongoAdapter, ok := db.(*mongodb.MongoAdapter); ok {
		return mongoAdapter.GetClient()
	}
	
	panic("current database is not MongoDB")
}

// IsUsingMongoDB returns true if the current database is MongoDB
func IsUsingMongoDB() bool {
	if db == nil {
		return false
	}
	
	_, ok := db.(*mongodb.MongoAdapter)
	return ok
}

// IsUsingPostgreSQL returns true if the current database is PostgreSQL
func IsUsingPostgreSQL() bool {
	if db == nil {
		return false
	}
	
	return db.Type() == "postgres"
}