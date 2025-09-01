package collection

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/madhouselabs/anybase/internal/database/adapters/postgres"
	"github.com/madhouselabs/anybase/internal/database/types"
	"github.com/madhouselabs/anybase/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AddVectorField adds a vector field to a collection
func (s *AdapterService) AddVectorField(ctx context.Context, userID primitive.ObjectID, collectionName string, field models.VectorField) error {
	// Check permissions
	hasPermission, err := s.rbacService.HasPermission(ctx, userID, fmt.Sprintf("collection:%s", collectionName), "update")
	if err != nil {
		return fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return fmt.Errorf("insufficient permissions to add vector field")
	}

	// Get the collection
	collection, err := s.GetCollection(ctx, userID, collectionName)
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}

	// Check if field already exists
	for _, vf := range collection.VectorFields {
		if vf.Name == field.Name {
			return fmt.Errorf("vector field '%s' already exists", field.Name)
		}
	}

	// Add field to collection
	collection.VectorFields = append(collection.VectorFields, field)
	collection.UpdatedAt = time.Now().UTC()

	// Update collection metadata
	collectionsCol := s.db.Collection("collections")
	filter := map[string]interface{}{
		"name": collectionName,
	}
	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"vector_fields": collection.VectorFields,
			"updated_at": collection.UpdatedAt,
		},
	}

	_, err = collectionsCol.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update collection: %w", err)
	}

	// Create vector column in data table
	dataTableName := "data_" + collectionName
	dataCol := s.db.Collection(dataTableName)
	
	// Check if this is a PostgreSQL collection and create the vector column
	if pgCol, ok := dataCol.(*postgres.PostgresCollection); ok {
		vectorOps := pgCol.GetVectorOperations()
		if err := vectorOps.CreateVectorColumn(ctx, field); err != nil {
			// Log but don't fail - column might already exist
			fmt.Printf("Warning: Failed to create vector column: %v\n", err)
		}
	}

	s.logAccess(ctx, userID, collectionName, nil, "add_vector_field", "success", field.Name)
	return nil
}

// RemoveVectorField removes a vector field from a collection
func (s *AdapterService) RemoveVectorField(ctx context.Context, userID primitive.ObjectID, collectionName string, fieldName string) error {
	// Check permissions
	hasPermission, err := s.rbacService.HasPermission(ctx, userID, fmt.Sprintf("collection:%s", collectionName), "update")
	if err != nil {
		return fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return fmt.Errorf("insufficient permissions to remove vector field")
	}

	// Get the collection
	collection, err := s.GetCollection(ctx, userID, collectionName)
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}

	// Remove field from collection
	var newFields []models.VectorField
	found := false
	for _, vf := range collection.VectorFields {
		if vf.Name != fieldName {
			newFields = append(newFields, vf)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("vector field '%s' not found", fieldName)
	}

	collection.VectorFields = newFields
	collection.UpdatedAt = time.Now().UTC()

	// Update collection metadata
	collectionsCol := s.db.Collection("collections")
	filter := map[string]interface{}{
		"name": collectionName,
	}
	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"vector_fields": collection.VectorFields,
			"updated_at": collection.UpdatedAt,
		},
	}

	_, err = collectionsCol.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update collection: %w", err)
	}

	// Drop vector column from data table
	dataTableName := "data_" + collectionName
	dataCol := s.db.Collection(dataTableName)
	
	// Check if this is a PostgreSQL collection and drop the vector column
	if pgCol, ok := dataCol.(*postgres.PostgresCollection); ok {
		vectorOps := pgCol.GetVectorOperations()
		if err := vectorOps.DropVectorColumn(ctx, fieldName); err != nil {
			// Log but don't fail - column might not exist
			fmt.Printf("Warning: Failed to drop vector column: %v\n", err)
		}
	}

	s.logAccess(ctx, userID, collectionName, nil, "remove_vector_field", "success", fieldName)
	return nil
}

// ListVectorFields lists all vector fields for a collection
func (s *AdapterService) ListVectorFields(ctx context.Context, collectionName string) ([]models.VectorField, error) {
	collectionsCol := s.db.Collection("collections")
	
	filter := map[string]interface{}{
		"name": collectionName,
		"_deleted_at": nil,
	}

	var collection models.Collection
	err := collectionsCol.FindOne(ctx, filter, &collection)
	if err != nil {
		if err == types.ErrNoDocuments {
			return nil, fmt.Errorf("collection not found")
		}
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}

	return collection.VectorFields, nil
}

// VectorSearch performs vector similarity search
func (s *AdapterService) VectorSearch(ctx context.Context, userID primitive.ObjectID, collectionName string, opts VectorSearchOptions) ([]bson.M, error) {
	// Check read permissions
	hasPermission, err := s.rbacService.HasPermission(ctx, userID, fmt.Sprintf("collection:%s", collectionName), "read")
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, fmt.Errorf("insufficient permissions to search collection")
	}

	// Validate vector field exists
	fields, err := s.ListVectorFields(ctx, collectionName)
	if err != nil {
		return nil, err
	}

	found := false
	for _, f := range fields {
		if f.Name == opts.VectorField {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("vector field '%s' not found", opts.VectorField)
	}

	// Get the PostgreSQL collection
	dataCol := s.db.Collection("data_" + collectionName)
	pgCol, ok := dataCol.(*postgres.PostgresCollection)
	if !ok {
		return nil, fmt.Errorf("vector search requires PostgreSQL adapter")
	}

	// Perform vector search using PostgreSQL operations
	vectorOps := pgCol.GetVectorOperations()
	rows, err := vectorOps.VectorSearch(ctx, opts.VectorField, opts.QueryVector, opts.TopK, opts.Metric)
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}
	defer rows.Close()

	// Parse results
	var results []bson.M
	for rows.Next() {
		var id string
		var data interface{}
		var distance float32
		
		if err := rows.Scan(&id, &data, &distance); err != nil {
			return nil, fmt.Errorf("failed to scan result: %w", err)
		}
		
		// Parse the JSONB data
		doc := bson.M{"_id": id}
		if jsonData, ok := data.([]byte); ok {
			if err := json.Unmarshal(jsonData, &doc); err == nil {
				doc["_id"] = id
				doc["_distance"] = distance
			}
		}
		
		results = append(results, doc)
	}

	s.logAccess(ctx, userID, collectionName, nil, "vector_search", "success", opts.VectorField)
	return results, nil
}

// HybridSearch performs combined text and vector search
func (s *AdapterService) HybridSearch(ctx context.Context, userID primitive.ObjectID, collectionName string, opts HybridSearchOptions) ([]bson.M, error) {
	// Check read permissions
	hasPermission, err := s.rbacService.HasPermission(ctx, userID, fmt.Sprintf("collection:%s", collectionName), "read")
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, fmt.Errorf("insufficient permissions to search collection")
	}

	// Validate vector field exists
	fields, err := s.ListVectorFields(ctx, collectionName)
	if err != nil {
		return nil, err
	}

	found := false
	for _, f := range fields {
		if f.Name == opts.VectorField {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("vector field '%s' not found", opts.VectorField)
	}

	// Get the PostgreSQL collection
	dataCol := s.db.Collection("data_" + collectionName)
	pgCol, ok := dataCol.(*postgres.PostgresCollection)
	if !ok {
		return nil, fmt.Errorf("hybrid search requires PostgreSQL adapter")
	}

	// Perform hybrid search using PostgreSQL operations
	vectorOps := pgCol.GetVectorOperations()
	rows, err := vectorOps.HybridSearch(ctx, opts.VectorField, opts.QueryVector, opts.TextQuery, opts.TopK, opts.Alpha)
	if err != nil {
		return nil, fmt.Errorf("hybrid search failed: %w", err)
	}
	defer rows.Close()

	// Parse results
	var results []bson.M
	for rows.Next() {
		var id string
		var data interface{}
		var score float32
		
		if err := rows.Scan(&id, &data, &score); err != nil {
			return nil, fmt.Errorf("failed to scan result: %w", err)
		}
		
		// Parse the JSONB data
		doc := bson.M{"_id": id}
		if jsonData, ok := data.([]byte); ok {
			if err := json.Unmarshal(jsonData, &doc); err == nil {
				doc["_id"] = id
				doc["_score"] = score
			}
		}
		
		results = append(results, doc)
	}

	s.logAccess(ctx, userID, collectionName, nil, "hybrid_search", "success", fmt.Sprintf("text+%s", opts.VectorField))
	return results, nil
}