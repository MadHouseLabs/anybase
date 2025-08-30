package database

import (
	"context"
	
	"github.com/madhouselabs/anybase/internal/database/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// Database is a compatibility wrapper for repositories that still expect the old Database type
// TODO: Remove this once all repositories are updated to use the abstraction
type Database struct {
	adapter types.DB
}

// WrapAdapter creates a compatibility wrapper for the database adapter
func WrapAdapter(adapter types.DB) *Database {
	return &Database{
		adapter: adapter,
	}
}

// GetDatabase returns the MongoDB database (for compatibility)
func (d *Database) GetDatabase() *mongo.Database {
	return GetMongoDatabase()
}

// GetClient returns the MongoDB client (for compatibility)
func (d *Database) GetClient() *mongo.Client {
	return GetMongoClient()
}

// Collection returns a MongoDB collection (for compatibility)
func (d *Database) Collection(name string) *mongo.Collection {
	return GetMongoDatabase().Collection(name)
}

// Close closes the database connection
func (d *Database) Close(ctx context.Context) error {
	return d.adapter.Close(ctx)
}

// Ping checks if the database is reachable
func (d *Database) Ping(ctx context.Context) error {
	return d.adapter.Ping(ctx)
}

// RunInTransaction executes a function within a transaction (MongoDB specific for compatibility)
func (d *Database) RunInTransaction(ctx context.Context, fn func(sessCtx mongo.SessionContext) error) error {
	client := GetMongoClient()
	session, err := client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		err := fn(sessCtx)
		return nil, err
	}

	_, err = session.WithTransaction(ctx, callback)
	return err
}

// CreateIndexes is no longer needed as indexes are created during initialization
func (d *Database) CreateIndexes(ctx context.Context) error {
	// Indexes are created during database initialization
	return nil
}