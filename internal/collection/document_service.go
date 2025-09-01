package collection

import (
	"context"
	"fmt"
	"time"

	"github.com/madhouselabs/anybase/internal/database/types"
	"github.com/madhouselabs/anybase/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// InsertDocument inserts a document with governance checks
func (s *AdapterService) InsertDocument(ctx context.Context, mutation *models.DataMutation) (*models.Document, error) {
	// Check permissions
	hasPermission, err := s.rbacService.HasPermission(ctx, mutation.UserID, fmt.Sprintf("collection:%s", mutation.Collection), "write")
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		s.logAccess(ctx, mutation.UserID, mutation.Collection, nil, "insert", "denied", "insufficient permissions")
		return nil, fmt.Errorf("insufficient permissions to insert document")
	}

	// Get collection to check if it exists
	col, err := s.GetCollection(ctx, mutation.UserID, mutation.Collection)
	if err != nil {
		return nil, fmt.Errorf("collection not found: %w", err)
	}

	// Validate against schema if exists
	if col.Schema != nil {
		if err := s.ValidateAgainstSchema(ctx, mutation.Collection, mutation.Data); err != nil {
			return nil, fmt.Errorf("schema validation failed: %w", err)
		}
	}

	// Get the data collection
	dataCol := s.db.Collection("data_" + mutation.Collection)

	// Create document with metadata
	doc := &models.Document{
		ID:         primitive.NewObjectID(),
		Collection: mutation.Collection,
		Data:       mutation.Data,
		CreatedBy:  mutation.UserID,
		UpdatedBy:  mutation.UserID,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
		Version:    1,
	}

	// Insert the document
	insertDoc := bson.M{
		"_id":         doc.ID.Hex(),
		"data":        doc.Data,
		"_created_by": doc.CreatedBy.Hex(),
		"_updated_by": doc.UpdatedBy.Hex(),
		"_created_at": doc.CreatedAt,
		"_updated_at": doc.UpdatedAt,
		"_version":    doc.Version,
	}

	_, err = dataCol.InsertOne(ctx, insertDoc)
	if err != nil {
		return nil, fmt.Errorf("failed to insert document: %w", err)
	}

	s.logAccess(ctx, mutation.UserID, mutation.Collection, doc.ID, "insert", "allowed", "document inserted")
	return doc, nil
}

// UpdateDocument updates a document with governance checks
func (s *AdapterService) UpdateDocument(ctx context.Context, mutation *models.DataMutation) error {
	// Check permissions
	hasPermission, err := s.rbacService.HasPermission(ctx, mutation.UserID, fmt.Sprintf("collection:%s", mutation.Collection), "update")
	if err != nil {
		return fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		s.logAccess(ctx, mutation.UserID, mutation.Collection, mutation.DocumentID, "update", "denied", "insufficient permissions")
		return fmt.Errorf("insufficient permissions to update document")
	}

	// Get collection to check if it exists
	col, err := s.GetCollection(ctx, mutation.UserID, mutation.Collection)
	if err != nil {
		return fmt.Errorf("collection not found: %w", err)
	}

	// Validate against schema if exists
	if col.Schema != nil {
		if err := s.ValidateAgainstSchema(ctx, mutation.Collection, mutation.Data); err != nil {
			return fmt.Errorf("schema validation failed: %w", err)
		}
	}

	// Get the data collection
	dataCol := s.db.Collection("data_" + mutation.Collection)

	// Update the document
	filter := map[string]interface{}{
		"_id": mutation.DocumentID.Hex(),
		"_deleted_at": nil,
	}

	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"data":        mutation.Data,
			"_updated_by": mutation.UserID.Hex(),
			"_updated_at": time.Now().UTC(),
		},
		"$inc": map[string]interface{}{
			"_version": 1,
		},
	}

	result, err := dataCol.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("document not found or already deleted")
	}

	s.logAccess(ctx, mutation.UserID, mutation.Collection, mutation.DocumentID, "update", "allowed", "document updated")
	return nil
}

// DeleteDocument deletes a document with governance checks
func (s *AdapterService) DeleteDocument(ctx context.Context, mutation *models.DataMutation) error {
	// Check permissions
	hasPermission, err := s.rbacService.HasPermission(ctx, mutation.UserID, fmt.Sprintf("collection:%s", mutation.Collection), "delete")
	if err != nil {
		return fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		s.logAccess(ctx, mutation.UserID, mutation.Collection, mutation.DocumentID, "delete", "denied", "insufficient permissions")
		return fmt.Errorf("insufficient permissions to delete document")
	}

	// Get the data collection
	dataCol := s.db.Collection("data_" + mutation.Collection)

	// Soft delete the document
	filter := map[string]interface{}{
		"_id": mutation.DocumentID.Hex(),
	}

	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"_deleted_at": time.Now().UTC(),
			"_deleted_by": mutation.UserID.Hex(),
		},
	}

	result, err := dataCol.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("document not found")
	}

	s.logAccess(ctx, mutation.UserID, mutation.Collection, mutation.DocumentID, "delete", "allowed", "document deleted")
	return nil
}

// QueryDocuments queries documents with governance checks
func (s *AdapterService) QueryDocuments(ctx context.Context, query *models.DataQuery) ([]models.Document, error) {
	// Check permissions
	hasPermission, err := s.rbacService.HasPermission(ctx, query.UserID, fmt.Sprintf("collection:%s", query.Collection), "read")
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		s.logAccess(ctx, query.UserID, query.Collection, nil, "query", "denied", "insufficient permissions")
		return nil, fmt.Errorf("insufficient permissions to query collection")
	}

	// Get the data collection
	dataCol := s.db.Collection("data_" + query.Collection)

	// Build filter
	filter := query.Filter
	if filter == nil {
		filter = map[string]interface{}{}
	}
	
	// Always exclude soft-deleted documents
	filter["_deleted_at"] = nil

	// Build options
	opts := &types.FindOptions{
		Sort: query.Sort,
	}
	
	if query.Limit > 0 {
		limit := int64(query.Limit)
		opts.Limit = &limit
	}
	
	if query.Skip > 0 {
		skip := int64(query.Skip)
		opts.Skip = &skip
	}

	// Execute query
	cursor, err := dataCol.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to query documents: %w", err)
	}
	defer cursor.Close(ctx)

	var documents []models.Document
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}

		// Convert to Document model
		document := models.Document{
			Collection: query.Collection,
		}

		// Extract metadata fields
		if id, ok := doc["_id"].(string); ok {
			if objID, err := primitive.ObjectIDFromHex(id); err == nil {
				document.ID = objID
			}
		}

		if data, ok := doc["data"].(map[string]interface{}); ok {
			document.Data = data
		} else if data, ok := doc["data"]; ok {
			// If it's not a map, store as-is
			document.Data = map[string]interface{}{"value": data}
		}

		if createdBy, ok := doc["_created_by"].(string); ok {
			if objID, err := primitive.ObjectIDFromHex(createdBy); err == nil {
				document.CreatedBy = objID
			}
		}

		if createdAt, ok := doc["_created_at"].(time.Time); ok {
			document.CreatedAt = createdAt
		}

		if updatedAt, ok := doc["_updated_at"].(time.Time); ok {
			document.UpdatedAt = updatedAt
		}

		if version, ok := doc["_version"].(int32); ok {
			document.Version = int(version)
		} else if version, ok := doc["_version"].(int64); ok {
			document.Version = int(version)
		}

		documents = append(documents, document)
	}

	s.logAccess(ctx, query.UserID, query.Collection, nil, "query", "allowed", fmt.Sprintf("%d results", len(documents)))
	return documents, nil
}

// GetDocument retrieves a single document with governance checks
func (s *AdapterService) GetDocument(ctx context.Context, userID primitive.ObjectID, collection string, docID primitive.ObjectID) (*models.Document, error) {
	// Check permissions
	hasPermission, err := s.rbacService.HasPermission(ctx, userID, fmt.Sprintf("collection:%s", collection), "read")
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		s.logAccess(ctx, userID, collection, docID, "read", "denied", "insufficient permissions")
		return nil, fmt.Errorf("insufficient permissions to read document")
	}

	// Get the data collection
	dataCol := s.db.Collection("data_" + collection)

	// Find the document
	filter := map[string]interface{}{
		"_id": docID.Hex(),
		"_deleted_at": nil,
	}

	var doc bson.M
	err = dataCol.FindOne(ctx, filter, &doc)
	if err != nil {
		if err == types.ErrNoDocuments {
			return nil, fmt.Errorf("document not found")
		}
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	// Convert to Document model
	document := &models.Document{
		ID:         docID,
		Collection: collection,
	}

	if data, ok := doc["data"].(map[string]interface{}); ok {
		document.Data = data
	} else if data, ok := doc["data"]; ok {
		// If it's not a map, store as-is
		document.Data = map[string]interface{}{"value": data}
	}

	if createdBy, ok := doc["_created_by"].(string); ok {
		if objID, err := primitive.ObjectIDFromHex(createdBy); err == nil {
			document.CreatedBy = objID
		}
	}

	if createdAt, ok := doc["_created_at"].(time.Time); ok {
		document.CreatedAt = createdAt
	}

	if updatedAt, ok := doc["_updated_at"].(time.Time); ok {
		document.UpdatedAt = updatedAt
	}

	if version, ok := doc["_version"].(int32); ok {
		document.Version = int(version)
	} else if version, ok := doc["_version"].(int64); ok {
		document.Version = int(version)
	}

	s.logAccess(ctx, userID, collection, docID, "read", "allowed", "document retrieved")
	return document, nil
}

// CountDocuments counts documents in a collection with governance checks
func (s *AdapterService) CountDocuments(ctx context.Context, userID primitive.ObjectID, collection string, filter map[string]interface{}) (int, error) {
	// Check permissions
	hasPermission, err := s.rbacService.HasPermission(ctx, userID, fmt.Sprintf("collection:%s", collection), "read")
	if err != nil {
		return 0, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return 0, fmt.Errorf("insufficient permissions to read collection")
	}

	// Get the data collection
	dataCol := s.db.Collection("data_" + collection)

	// Always exclude soft-deleted documents
	if filter == nil {
		filter = map[string]interface{}{}
	}
	filter["_deleted_at"] = nil

	count, err := dataCol.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count documents: %w", err)
	}

	return int(count), nil
}