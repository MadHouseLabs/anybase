package collection

import (
	"context"
	"fmt"
	"time"

	"github.com/madhouselabs/anybase/internal/database/types"
	"github.com/madhouselabs/anybase/pkg/models"
)

// ListIndexes lists all indexes for a collection
func (s *AdapterService) ListIndexes(ctx context.Context, collectionName string) ([]interface{}, error) {
	// Get the data collection
	_ = s.db.Collection("data_" + collectionName)
	
	// Note: GetIndexes is not available in the interface, needs to be handled differently based on adapter
	// For now, return empty list
	return []interface{}{}, nil
}

// CreateIndex creates an index on a collection
func (s *AdapterService) CreateIndex(ctx context.Context, collectionName string, indexName string, keys map[string]interface{}, options map[string]interface{}) error {
	// Get the data collection
	dataCol := s.db.Collection("data_" + collectionName)
	
	// Convert to types.Index
	index := types.Index{
		Name: indexName,
		Keys: make(map[string]int),
	}

	// Convert keys
	for field, order := range keys {
		if orderInt, ok := order.(int); ok {
			index.Keys[field] = orderInt
		} else if orderFloat, ok := order.(float64); ok {
			index.Keys[field] = int(orderFloat)
		} else {
			index.Keys[field] = 1 // Default to ascending
		}
	}

	// Parse options
	if options != nil {
		if unique, ok := options["unique"].(bool); ok {
			index.Unique = unique
		}
		if sparse, ok := options["sparse"].(bool); ok {
			index.Sparse = sparse
		}
		if ttl, ok := options["expireAfterSeconds"].(int); ok {
			ttlDuration := time.Duration(ttl) * time.Second
			index.TTL = &ttlDuration
		}
	}

	// Create the index
	if err := dataCol.CreateIndex(ctx, index); err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	return nil
}

// DeleteIndex drops an index from a collection
func (s *AdapterService) DeleteIndex(ctx context.Context, collectionName string, indexName string) error {
	// Get the data collection
	dataCol := s.db.Collection("data_" + collectionName)
	
	// Drop the index
	if err := dataCol.DropIndex(ctx, indexName); err != nil {
		return fmt.Errorf("failed to drop index: %w", err)
	}

	return nil
}

// Helper method to create indexes for a collection
func (s *AdapterService) createIndexes(ctx context.Context, collection *models.Collection) error {
	dataCol := s.db.Collection("data_" + collection.Name)
	
	for _, index := range collection.Indexes {
		// Convert index definition to types.Index
		idx := types.Index{
			Name:   index.Name,
			Keys:   make(map[string]int),
			Unique: index.Unique,
			Sparse: index.Sparse,
		}
		
		// Convert keys
		for field, order := range index.Fields {
			if orderInt, ok := order.(int); ok {
				idx.Keys[field] = orderInt
			} else if orderFloat, ok := order.(float64); ok {
				idx.Keys[field] = int(orderFloat)
			} else {
				idx.Keys[field] = 1 // Default to ascending
			}
		}
		
		if err := dataCol.CreateIndex(ctx, idx); err != nil {
			return fmt.Errorf("failed to create index %s: %w", index.Name, err)
		}
	}
	
	return nil
}