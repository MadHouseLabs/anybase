package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

// MongoCursor wraps a MongoDB cursor to implement database.Cursor
type MongoCursor struct {
	cursor *mongo.Cursor
}

// Next advances the cursor to the next document
func (c *MongoCursor) Next(ctx context.Context) bool {
	return c.cursor.Next(ctx)
}

// Decode decodes the current document into result
func (c *MongoCursor) Decode(result interface{}) error {
	return c.cursor.Decode(result)
}

// Close closes the cursor
func (c *MongoCursor) Close(ctx context.Context) error {
	return c.cursor.Close(ctx)
}

// All decodes all remaining documents into results
func (c *MongoCursor) All(ctx context.Context, results interface{}) error {
	return c.cursor.All(ctx, results)
}