package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/madhouselabs/anybase/internal/database/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoCollection wraps a MongoDB collection to implement types.Collection
type MongoCollection struct {
	collection *mongo.Collection
	dbType     string
}

// InsertOne inserts a single document
func (c *MongoCollection) InsertOne(ctx context.Context, document map[string]interface{}) (types.ID, error) {
	// Ensure _id field exists
	if _, ok := document["_id"]; !ok {
		document["_id"] = primitive.NewObjectID()
	}
	
	result, err := c.collection.InsertOne(ctx, document)
	if err != nil {
		return nil, fmt.Errorf("failed to insert document: %w", err)
	}
	
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		return types.FromObjectID(oid), nil
	}
	
	return nil, fmt.Errorf("unexpected ID type: %T", result.InsertedID)
}

// InsertMany inserts multiple documents
func (c *MongoCollection) InsertMany(ctx context.Context, documents []map[string]interface{}) ([]types.ID, error) {
	// Convert to []interface{} and ensure _id fields
	docs := make([]interface{}, len(documents))
	for i, doc := range documents {
		if _, ok := doc["_id"]; !ok {
			doc["_id"] = primitive.NewObjectID()
		}
		docs[i] = doc
	}
	
	result, err := c.collection.InsertMany(ctx, docs)
	if err != nil {
		return nil, fmt.Errorf("failed to insert documents: %w", err)
	}
	
	ids := make([]types.ID, len(result.InsertedIDs))
	for i, id := range result.InsertedIDs {
		if oid, ok := id.(primitive.ObjectID); ok {
			ids[i] = types.FromObjectID(oid)
		} else {
			return nil, fmt.Errorf("unexpected ID type at index %d: %T", i, id)
		}
	}
	
	return ids, nil
}

// FindOne finds a single document
func (c *MongoCollection) FindOne(ctx context.Context, filter map[string]interface{}, result interface{}) error {
	err := c.collection.FindOne(ctx, filter).Decode(result)
	if err == mongo.ErrNoDocuments {
		return types.ErrNoDocuments
	}
	return err
}

// Find finds multiple documents
func (c *MongoCollection) Find(ctx context.Context, filter map[string]interface{}, opts *types.FindOptions) (types.Cursor, error) {
	findOpts := options.Find()
	
	if opts != nil {
		if opts.Limit != nil {
			findOpts.SetLimit(*opts.Limit)
		}
		if opts.Skip != nil {
			findOpts.SetSkip(*opts.Skip)
		}
		if opts.Sort != nil {
			findOpts.SetSort(opts.Sort)
		}
		if opts.Projection != nil {
			findOpts.SetProjection(opts.Projection)
		}
	}
	
	cursor, err := c.collection.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, err
	}
	
	return &MongoCursor{cursor: cursor}, nil
}

// UpdateOne updates a single document
func (c *MongoCollection) UpdateOne(ctx context.Context, filter map[string]interface{}, update map[string]interface{}) (*types.UpdateResult, error) {
	// Ensure update has proper MongoDB update operators
	updateDoc := c.prepareUpdate(update)
	
	result, err := c.collection.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return nil, err
	}
	
	dbResult := &types.UpdateResult{
		MatchedCount:  result.MatchedCount,
		ModifiedCount: result.ModifiedCount,
		UpsertedCount: result.UpsertedCount,
	}
	
	if result.UpsertedID != nil {
		if oid, ok := result.UpsertedID.(primitive.ObjectID); ok {
			dbResult.UpsertedID = types.FromObjectID(oid)
		}
	}
	
	return dbResult, nil
}

// UpdateMany updates multiple documents
func (c *MongoCollection) UpdateMany(ctx context.Context, filter map[string]interface{}, update map[string]interface{}) (*types.UpdateResult, error) {
	// Ensure update has proper MongoDB update operators
	updateDoc := c.prepareUpdate(update)
	
	result, err := c.collection.UpdateMany(ctx, filter, updateDoc)
	if err != nil {
		return nil, err
	}
	
	dbResult := &types.UpdateResult{
		MatchedCount:  result.MatchedCount,
		ModifiedCount: result.ModifiedCount,
		UpsertedCount: result.UpsertedCount,
	}
	
	if result.UpsertedID != nil {
		if oid, ok := result.UpsertedID.(primitive.ObjectID); ok {
			dbResult.UpsertedID = types.FromObjectID(oid)
		}
	}
	
	return dbResult, nil
}

// DeleteOne deletes a single document
func (c *MongoCollection) DeleteOne(ctx context.Context, filter map[string]interface{}) (*types.DeleteResult, error) {
	result, err := c.collection.DeleteOne(ctx, filter)
	if err != nil {
		return nil, err
	}
	
	return &types.DeleteResult{
		DeletedCount: result.DeletedCount,
	}, nil
}

// DeleteMany deletes multiple documents
func (c *MongoCollection) DeleteMany(ctx context.Context, filter map[string]interface{}) (*types.DeleteResult, error) {
	result, err := c.collection.DeleteMany(ctx, filter)
	if err != nil {
		return nil, err
	}
	
	return &types.DeleteResult{
		DeletedCount: result.DeletedCount,
	}, nil
}

// CountDocuments counts documents matching the filter
func (c *MongoCollection) CountDocuments(ctx context.Context, filter map[string]interface{}) (int64, error) {
	return c.collection.CountDocuments(ctx, filter)
}

// CreateIndex creates an index on the collection
func (c *MongoCollection) CreateIndex(ctx context.Context, index types.Index) error {
	indexModel := mongo.IndexModel{
		Keys: index.Keys,
	}
	
	opts := options.Index()
	if index.Name != "" {
		opts.SetName(index.Name)
	}
	if index.Unique {
		opts.SetUnique(true)
	}
	if index.Sparse {
		opts.SetSparse(true)
	}
	if index.TTL != nil {
		opts.SetExpireAfterSeconds(int32(index.TTL.Seconds()))
	}
	
	indexModel.Options = opts
	
	_, err := c.collection.Indexes().CreateOne(ctx, indexModel)
	return err
}

// DropIndex drops an index from the collection
func (c *MongoCollection) DropIndex(ctx context.Context, name string) error {
	_, err := c.collection.Indexes().DropOne(ctx, name)
	return err
}

// ListIndexes lists all indexes on the collection
func (c *MongoCollection) ListIndexes(ctx context.Context) ([]types.Index, error) {
	cursor, err := c.collection.Indexes().List(ctx)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var indexes []types.Index
	for cursor.Next(ctx) {
		var idx bson.M
		if err := cursor.Decode(&idx); err != nil {
			continue
		}
		
		dbIdx := types.Index{}
		
		if name, ok := idx["name"].(string); ok {
			dbIdx.Name = name
		}
		
		if keys, ok := idx["key"].(bson.M); ok {
			dbIdx.Keys = make(map[string]int)
			for k, v := range keys {
				if val, ok := v.(int32); ok {
					dbIdx.Keys[k] = int(val)
				} else if val, ok := v.(int64); ok {
					dbIdx.Keys[k] = int(val)
				} else if val, ok := v.(int); ok {
					dbIdx.Keys[k] = val
				}
			}
		}
		
		if unique, ok := idx["unique"].(bool); ok {
			dbIdx.Unique = unique
		}
		
		if sparse, ok := idx["sparse"].(bool); ok {
			dbIdx.Sparse = sparse
		}
		
		if expireAfterSeconds, ok := idx["expireAfterSeconds"].(int32); ok {
			ttl := time.Duration(expireAfterSeconds) * time.Second
			dbIdx.TTL = &ttl
		}
		
		indexes = append(indexes, dbIdx)
	}
	
	return indexes, nil
}

// Aggregate performs aggregation on the collection
func (c *MongoCollection) Aggregate(ctx context.Context, pipeline []map[string]interface{}) (types.Cursor, error) {
	// Convert pipeline to []interface{}
	pipe := make([]interface{}, len(pipeline))
	for i, stage := range pipeline {
		pipe[i] = stage
	}
	
	cursor, err := c.collection.Aggregate(ctx, pipe)
	if err != nil {
		return nil, err
	}
	
	return &MongoCursor{cursor: cursor}, nil
}

// prepareUpdate ensures the update document has proper MongoDB update operators
func (c *MongoCollection) prepareUpdate(update map[string]interface{}) interface{} {
	// Check if update already has MongoDB operators
	for key := range update {
		if len(key) > 0 && key[0] == '$' {
			// Already has operators, return as-is
			return update
		}
	}
	
	// Wrap in $set operator if no operators present
	return bson.M{"$set": update}
}

// GetUnderlying returns the underlying MongoDB collection (for migration purposes)
func (c *MongoCollection) GetUnderlying() *mongo.Collection {
	return c.collection
}