package collection

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/madhouselabs/anybase/internal/database/types"
	"github.com/madhouselabs/anybase/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateCollection creates a new collection with governance checks
func (s *AdapterService) CreateCollection(ctx context.Context, userID primitive.ObjectID, collection *models.Collection) error {
	// Get user role
	userRole, err := s.rbacService.GetUserRole(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user role: %w", err)
	}

	// Admins and developers can create collections
	if userRole != "admin" && userRole != "developer" {
		// For other users, check if they have permission to create collections
		hasPermission, err := s.rbacService.HasPermission(ctx, userID, "collections", "create")
		if err != nil {
			return fmt.Errorf("failed to check permissions: %w", err)
		}
		if !hasPermission {
			s.logAccess(ctx, userID, collection.Name, nil, "create", "denied", "insufficient permissions")
			return fmt.Errorf("insufficient permissions to create collection")
		}
	}

	// Check if collection already exists
	collectionsCol := s.db.Collection("collections")
	existingCount, err := collectionsCol.CountDocuments(ctx, map[string]interface{}{"name": collection.Name})
	if err != nil {
		return fmt.Errorf("failed to check existing collection: %w", err)
	}
	if existingCount > 0 {
		return fmt.Errorf("collection '%s' already exists", collection.Name)
	}

	// Set metadata
	collection.ID = primitive.NewObjectID()
	collection.CreatedBy = userID
	collection.CreatedAt = time.Now().UTC()
	collection.UpdatedAt = time.Now().UTC()

	// Create the actual collection in the database
	dataCollectionName := "data_" + collection.Name
	if err := s.db.CreateCollection(ctx, dataCollectionName, collection.Schema); err != nil {
		// Collection might already exist
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("failed to create collection: %w", err)
		}
	}

	// Create indexes if specified
	if len(collection.Indexes) > 0 {
		if err := s.createIndexes(ctx, collection); err != nil {
			return fmt.Errorf("failed to create indexes: %w", err)
		}
	}

	// Store collection metadata
	// collectionsCol already defined above
	
	// Convert collection to map for storage
	collectionDoc := map[string]interface{}{
		"_id":         collection.ID.Hex(),
		"name":        collection.Name,
		"description": collection.Description,
		"schema":      collection.Schema,
		"indexes":     collection.Indexes,
		"settings":    collection.Settings,
		"created_by":  userID.Hex(),
		"created_at":  collection.CreatedAt,
		"updated_at":  collection.UpdatedAt,
	}

	if _, err := collectionsCol.InsertOne(ctx, collectionDoc); err != nil {
		// Try to rollback the data collection creation
		s.db.DropCollection(ctx, dataCollectionName)
		return fmt.Errorf("failed to store collection metadata: %w", err)
	}

	s.logAccess(ctx, userID, collection.Name, nil, "create", "allowed", "collection created")
	return nil
}

// GetCollection retrieves a collection with governance checks
func (s *AdapterService) GetCollection(ctx context.Context, userID primitive.ObjectID, name string) (*models.Collection, error) {
	// Check permissions
	hasPermission, err := s.rbacService.HasPermission(ctx, userID, fmt.Sprintf("collection:%s", name), "read")
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		s.logAccess(ctx, userID, name, nil, "read", "denied", "insufficient permissions")
		return nil, fmt.Errorf("insufficient permissions to read collection")
	}

	collectionsCol := s.db.Collection("collections")
	
	var collection models.Collection
	filter := map[string]interface{}{"name": name}
	
	err = collectionsCol.FindOne(ctx, filter, &collection)
	if err != nil {
		if err == types.ErrNoDocuments {
			return nil, fmt.Errorf("collection not found")
		}
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}

	s.logAccess(ctx, userID, name, nil, "read", "allowed", "collection retrieved")
	return &collection, nil
}

// UpdateCollection updates a collection with governance checks
func (s *AdapterService) UpdateCollection(ctx context.Context, userID primitive.ObjectID, name string, updates bson.M) error {
	// Check permissions
	hasPermission, err := s.rbacService.HasPermission(ctx, userID, fmt.Sprintf("collection:%s", name), "update")
	if err != nil {
		return fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		s.logAccess(ctx, userID, name, nil, "update", "denied", "insufficient permissions")
		return fmt.Errorf("insufficient permissions to update collection")
	}

	collectionsCol := s.db.Collection("collections")
	
	// Add updated_at timestamp
	if updateSet, ok := updates["$set"].(bson.M); ok {
		updateSet["updated_at"] = time.Now().UTC()
	} else {
		updates["$set"] = bson.M{"updated_at": time.Now().UTC()}
	}

	filter := map[string]interface{}{"name": name}
	_, err = collectionsCol.UpdateOne(ctx, filter, updates)
	if err != nil {
		return fmt.Errorf("failed to update collection: %w", err)
	}

	s.logAccess(ctx, userID, name, nil, "update", "allowed", "collection updated")
	return nil
}

// DeleteCollection deletes a collection with governance checks
func (s *AdapterService) DeleteCollection(ctx context.Context, userID primitive.ObjectID, name string) error {
	// Check permissions
	hasPermission, err := s.rbacService.HasPermission(ctx, userID, fmt.Sprintf("collection:%s", name), "delete")
	if err != nil {
		return fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		s.logAccess(ctx, userID, name, nil, "delete", "denied", "insufficient permissions")
		return fmt.Errorf("insufficient permissions to delete collection")
	}

	// Drop the actual data collection
	dataCollectionName := "data_" + name
	if err := s.db.DropCollection(ctx, dataCollectionName); err != nil {
		// Collection might not exist, which is okay
		if !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "does not exist") {
			return fmt.Errorf("failed to drop collection: %w", err)
		}
	}

	// Delete collection metadata
	collectionsCol := s.db.Collection("collections")
	filter := map[string]interface{}{"name": name}
	_, err = collectionsCol.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete collection metadata: %w", err)
	}

	s.logAccess(ctx, userID, name, nil, "delete", "allowed", "collection deleted")
	return nil
}

// ListCollections lists all collections
func (s *AdapterService) ListCollections(ctx context.Context) ([]*models.Collection, error) {
	collectionsCol := s.db.Collection("collections")
	
	cursor, err := collectionsCol.Find(ctx, map[string]interface{}{}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}
	defer cursor.Close(ctx)

	var collections []*models.Collection
	for cursor.Next(ctx) {
		var col models.Collection
		if err := cursor.Decode(&col); err != nil {
			continue
		}

		// Get document count for each collection
		dataCol := s.db.Collection("data_" + col.Name)
		docCount, _ := dataCol.CountDocuments(ctx, map[string]interface{}{})
		col.DocumentCount = int64(docCount)

		collections = append(collections, &col)
	}

	return collections, nil
}