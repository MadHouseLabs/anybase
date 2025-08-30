package mongodb

import (
	"context"

	"github.com/madhouselabs/anybase/internal/database/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// MongoTransaction wraps a MongoDB session to implement types.Transaction
type MongoTransaction struct {
	session mongo.Session
	adapter *MongoAdapter
	ctx     context.Context
}

// Commit commits the transaction
func (t *MongoTransaction) Commit(ctx context.Context) error {
	return t.session.CommitTransaction(ctx)
}

// Rollback rolls back the transaction
func (t *MongoTransaction) Rollback(ctx context.Context) error {
	return t.session.AbortTransaction(ctx)
}

// Collection returns a collection that uses this transaction
func (t *MongoTransaction) Collection(name string) types.Collection {
	// Use the session context for transactional operations
	return &MongoCollection{
		collection: t.adapter.database.Collection(name),
		dbType:     "mongodb",
	}
}

// Context returns the transaction context
func (t *MongoTransaction) Context() context.Context {
	return t.ctx
}