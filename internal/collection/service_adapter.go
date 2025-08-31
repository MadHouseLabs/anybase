package collection

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/madhouselabs/anybase/internal/database/types"
	"github.com/madhouselabs/anybase/internal/governance"
	"github.com/madhouselabs/anybase/internal/validator"
	"github.com/madhouselabs/anybase/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AdapterService is a new implementation that uses the database adapter
type AdapterService struct {
	db          types.DB
	rbacService governance.RBACService
	validator   *validator.SchemaValidator
}

// NewAdapterService creates a new collection service using the database adapter
func NewAdapterService(db types.DB, rbacService governance.RBACService) Service {
	return &AdapterService{
		db:          db,
		rbacService: rbacService,
		validator:   validator.NewSchemaValidator(),
	}
}

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
	if hasPermission, err := s.rbacService.HasPermission(ctx, userID, fmt.Sprintf("collection:%s:read", name), ""); err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	} else if !hasPermission {
		// Check if user has general collection read permission
		if hasPermission, err := s.rbacService.HasPermission(ctx, userID, "collection:*:read", ""); err != nil {
			return nil, fmt.Errorf("failed to check permissions: %w", err)
		} else if !hasPermission {
			s.logAccess(ctx, userID, name, nil, "read", "denied", "insufficient permissions")
			return nil, fmt.Errorf("insufficient permissions to read collection")
		}
	}

	collectionsCol := s.db.Collection("collections")
	
	var collection models.Collection
	filter := map[string]interface{}{
		"name": name,
	}
	
	if err := collectionsCol.FindOne(ctx, filter, &collection); err != nil {
		if err == types.ErrNoDocuments {
			return nil, fmt.Errorf("collection not found")
		}
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}

	// Get collection stats
	dataCol := s.db.Collection("data_" + name)
	docCount, _ := dataCol.CountDocuments(ctx, map[string]interface{}{})
	collection.DocumentCount = int64(docCount)

	s.logAccess(ctx, userID, name, nil, "read", "allowed", "collection retrieved")
	return &collection, nil
}

// UpdateCollection updates a collection with governance checks
func (s *AdapterService) UpdateCollection(ctx context.Context, userID primitive.ObjectID, name string, updates bson.M) error {
	// Check permissions
	if hasPermission, err := s.rbacService.HasPermission(ctx, userID, fmt.Sprintf("collection:%s:update", name), ""); err != nil {
		return fmt.Errorf("failed to check permissions: %w", err)
	} else if !hasPermission {
		// Check if user has general collection update permission
		if hasPermission, err := s.rbacService.HasPermission(ctx, userID, "collection:*:update", ""); err != nil {
			return fmt.Errorf("failed to check permissions: %w", err)
		} else if !hasPermission {
			s.logAccess(ctx, userID, name, nil, "update", "denied", "insufficient permissions")
			return fmt.Errorf("insufficient permissions to update collection")
		}
	}

	// Add updated timestamp
	if updates == nil {
		updates = bson.M{}
	}
	updates["updated_at"] = time.Now().UTC()

	collectionsCol := s.db.Collection("collections")
	
	filter := map[string]interface{}{
		"name": name,
	}
	
	updateDoc := map[string]interface{}{
		"$set": updates,
	}
	
	result, err := collectionsCol.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return fmt.Errorf("failed to update collection: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("collection not found")
	}

	s.logAccess(ctx, userID, name, nil, "update", "allowed", "collection updated")
	return nil
}

// DeleteCollection deletes a collection with governance checks
func (s *AdapterService) DeleteCollection(ctx context.Context, userID primitive.ObjectID, name string) error {
	// Check permissions
	if hasPermission, err := s.rbacService.HasPermission(ctx, userID, fmt.Sprintf("collection:%s:delete", name), ""); err != nil {
		return fmt.Errorf("failed to check permissions: %w", err)
	} else if !hasPermission {
		// Check if user has general collection delete permission
		if hasPermission, err := s.rbacService.HasPermission(ctx, userID, "collection:*:delete", ""); err != nil {
			return fmt.Errorf("failed to check permissions: %w", err)
		} else if !hasPermission {
			s.logAccess(ctx, userID, name, nil, "delete", "denied", "insufficient permissions")
			return fmt.Errorf("insufficient permissions to delete collection")
		}
	}

	// Drop the data collection
	if err := s.db.DropCollection(ctx, "data_"+name); err != nil {
		// Log but don't fail if collection doesn't exist
		fmt.Printf("Warning: failed to drop data collection: %v\n", err)
	}

	// Delete collection metadata
	collectionsCol := s.db.Collection("collections")
	
	filter := map[string]interface{}{
		"name": name,
	}
	
	result, err := collectionsCol.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete collection metadata: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("collection not found")
	}

	s.logAccess(ctx, userID, name, nil, "delete", "allowed", "collection deleted")
	return nil
}

// ListCollections lists all collections in the system
func (s *AdapterService) ListCollections(ctx context.Context) ([]*models.Collection, error) {
	collectionsCol := s.db.Collection("collections")
	
	cursor, err := collectionsCol.Find(ctx, map[string]interface{}{}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list all collections: %w", err)
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

// Helper method to create indexes
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

// Helper method to log access
func (s *AdapterService) logAccess(ctx context.Context, userID primitive.ObjectID, resource string, resourceID interface{}, action, result, reason string) {
	logsCol := s.db.Collection("access_logs")
	
	logEntry := map[string]interface{}{
		"user_id":     userID.Hex(),
		"resource":    resource,
		"resource_id": resourceID,
		"action":      action,
		"result":      result,
		"reason":      reason,
		"timestamp":   time.Now().UTC(),
	}
	
	// Best effort logging - don't fail the operation if logging fails
	logsCol.InsertOne(ctx, logEntry)
}

// CreateView creates a new view (saved query)
func (s *AdapterService) CreateView(ctx context.Context, userID primitive.ObjectID, view *models.View) error {
	// Check permissions
	if hasPermission, err := s.rbacService.HasPermission(ctx, userID, "view:*:create", ""); err != nil {
		return fmt.Errorf("failed to check permissions: %w", err)
	} else if !hasPermission {
		s.logAccess(ctx, userID, view.Name, nil, "create", "denied", "insufficient permissions")
		return fmt.Errorf("insufficient permissions to create view")
	}

	// Check if view already exists
	viewsCol := s.db.Collection("_views")
	existingCount, err := viewsCol.CountDocuments(ctx, map[string]interface{}{
		"data.name": view.Name,
	})
	if err != nil {
		return fmt.Errorf("failed to check existing view: %w", err)
	}
	if existingCount > 0 {
		return fmt.Errorf("view '%s' already exists", view.Name)
	}

	// Verify the source collection exists
	collectionsCol := s.db.Collection("collections")
	collCount, err := collectionsCol.CountDocuments(ctx, map[string]interface{}{
		"name": view.Collection,
	})
	if err != nil {
		return fmt.Errorf("failed to check collection: %w", err)
	}
	if collCount == 0 {
		return fmt.Errorf("collection '%s' does not exist", view.Collection)
	}

	// Set metadata
	view.ID = primitive.NewObjectID()
	view.CreatedBy = userID
	view.CreatedAt = time.Now().UTC()
	view.UpdatedAt = time.Now().UTC()

	// Convert view to JSONB format for PostgreSQL storage
	viewDoc := map[string]interface{}{
		"_id": view.ID.Hex(),
		"data": map[string]interface{}{
			"name":        view.Name,
			"description": view.Description,
			"collection":  view.Collection,
			"pipeline":    view.Pipeline,
			"fields":      view.Fields,
			"filter":      view.Filter,
			"permissions": view.Permissions,
			"metadata":    view.Metadata,
			"created_by":  userID.Hex(),
			"created_at":  view.CreatedAt,
			"updated_at":  view.UpdatedAt,
		},
	}

	if _, err := viewsCol.InsertOne(ctx, viewDoc); err != nil {
		return fmt.Errorf("failed to store view: %w", err)
	}

	s.logAccess(ctx, userID, view.Name, nil, "create", "allowed", "view created")
	return nil
}

func (s *AdapterService) GetView(ctx context.Context, name string) (*models.View, error) {
	viewsCol := s.db.Collection("_views")
	
	filter := map[string]interface{}{
		"name": name,
		"_deleted_at": nil,
	}
	
	var result map[string]interface{}
	if err := viewsCol.FindOne(ctx, filter, &result); err != nil {
		if err == types.ErrNoDocuments {
			return nil, fmt.Errorf("view not found")
		}
		return nil, fmt.Errorf("failed to get view: %w", err)
	}

	// Extract view data from JSONB structure
	// PostgreSQL adapter may return data either nested or flat
	var data map[string]interface{}
	if nested, ok := result["data"].(map[string]interface{}); ok {
		data = nested
	} else {
		// If data is not nested, use the result directly
		// This happens when the PostgreSQL adapter flattens the JSONB
		data = result
	}

	view := &models.View{
		Name:        getString(data, "name"),
		Description: getString(data, "description"),
		Collection:  getString(data, "collection"),
	}

	// Parse ID
	if idStr := getString(result, "_id"); idStr != "" {
		if oid, err := primitive.ObjectIDFromHex(idStr); err == nil {
			view.ID = oid
		}
	}

	// Parse created_by
	if createdByStr := getString(data, "created_by"); createdByStr != "" {
		if oid, err := primitive.ObjectIDFromHex(createdByStr); err == nil {
			view.CreatedBy = oid
		}
	}

	// Parse complex fields
	if pipeline, ok := data["pipeline"].([]interface{}); ok {
		view.Pipeline = make([]bson.M, len(pipeline))
		for i, p := range pipeline {
			if pm, ok := p.(map[string]interface{}); ok {
				view.Pipeline[i] = bson.M(pm)
			}
		}
	}

	if fields, ok := data["fields"].([]interface{}); ok {
		view.Fields = make([]string, len(fields))
		for i, f := range fields {
			view.Fields[i] = fmt.Sprint(f)
		}
	}

	if filter, ok := data["filter"].(map[string]interface{}); ok {
		view.Filter = bson.M(filter)
	}

	// Parse timestamps
	if createdAt, ok := data["created_at"].(time.Time); ok {
		view.CreatedAt = createdAt
	}
	if updatedAt, ok := data["updated_at"].(time.Time); ok {
		view.UpdatedAt = updatedAt
	}

	return view, nil
}

// Helper function to safely get string from map
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		return fmt.Sprint(v)
	}
	return ""
}

func (s *AdapterService) UpdateView(ctx context.Context, userID primitive.ObjectID, name string, updates bson.M) error {
	// Check permissions
	if hasPermission, err := s.rbacService.HasPermission(ctx, userID, fmt.Sprintf("view:%s:update", name), ""); err != nil {
		return fmt.Errorf("failed to check permissions: %w", err)
	} else if !hasPermission {
		// Check if user has general view update permission
		if hasPermission, err := s.rbacService.HasPermission(ctx, userID, "view:*:update", ""); err != nil {
			return fmt.Errorf("failed to check permissions: %w", err)
		} else if !hasPermission {
			s.logAccess(ctx, userID, name, nil, "update", "denied", "insufficient permissions")
			return fmt.Errorf("insufficient permissions to update view")
		}
	}

	// Get existing view to verify it exists
	existingView, err := s.GetView(ctx, name)
	if err != nil {
		return fmt.Errorf("view not found: %w", err)
	}

	// Prepare update document
	updateDoc := map[string]interface{}{
		"$set": map[string]interface{}{},
	}
	setFields := updateDoc["$set"].(map[string]interface{})

	// Process updates - only update provided fields
	if desc, ok := updates["description"]; ok {
		setFields["data.description"] = desc
	}
	if pipeline, ok := updates["pipeline"]; ok {
		setFields["data.pipeline"] = pipeline
	}
	if fields, ok := updates["fields"]; ok {
		setFields["data.fields"] = fields
	}
	if filter, ok := updates["filter"]; ok {
		setFields["data.filter"] = filter
	}
	if permissions, ok := updates["permissions"]; ok {
		setFields["data.permissions"] = permissions
	}
	if metadata, ok := updates["metadata"]; ok {
		setFields["data.metadata"] = metadata
	}

	// Always update timestamp
	setFields["data.updated_at"] = time.Now().UTC()

	viewsCol := s.db.Collection("_views")
	filter := map[string]interface{}{
		"name": name,
		"_deleted_at": nil,
	}

	result, err := viewsCol.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return fmt.Errorf("failed to update view: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("view not found")
	}

	s.logAccess(ctx, userID, name, existingView.ID, "update", "allowed", "view updated")
	return nil
}

func (s *AdapterService) DeleteView(ctx context.Context, userID primitive.ObjectID, name string) error {
	// Check permissions
	if hasPermission, err := s.rbacService.HasPermission(ctx, userID, fmt.Sprintf("view:%s:delete", name), ""); err != nil {
		return fmt.Errorf("failed to check permissions: %w", err)
	} else if !hasPermission {
		// Check if user has general view delete permission
		if hasPermission, err := s.rbacService.HasPermission(ctx, userID, "view:*:delete", ""); err != nil {
			return fmt.Errorf("failed to check permissions: %w", err)
		} else if !hasPermission {
			s.logAccess(ctx, userID, name, nil, "delete", "denied", "insufficient permissions")
			return fmt.Errorf("insufficient permissions to delete view")
		}
	}

	// Get existing view to verify it exists
	existingView, err := s.GetView(ctx, name)
	if err != nil {
		return fmt.Errorf("view not found: %w", err)
	}

	viewsCol := s.db.Collection("_views")
	
	// Soft delete by setting _deleted_at timestamp
	filter := map[string]interface{}{
		"name": name,
		"_deleted_at": nil,
	}
	updateDoc := map[string]interface{}{
		"$set": map[string]interface{}{
			"_deleted_at": time.Now().UTC(),
		},
	}

	result, err := viewsCol.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return fmt.Errorf("failed to delete view: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("view not found")
	}

	s.logAccess(ctx, userID, name, existingView.ID, "delete", "allowed", "view deleted")
	return nil
}

func (s *AdapterService) ListViews(ctx context.Context) ([]*models.View, error) {
	viewsCol := s.db.Collection("_views")
	
	filter := map[string]interface{}{
		"_deleted_at": nil,
	}
	
	cursor, err := viewsCol.Find(ctx, filter, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list views: %w", err)
	}
	defer cursor.Close(ctx)
	

	var views []*models.View
	for cursor.Next(ctx) {
		var view models.View
		if err := cursor.Decode(&view); err != nil {
			continue
		}
		views = append(views, &view)
	}
	return views, nil
}

func (s *AdapterService) QueryView(ctx context.Context, userID primitive.ObjectID, viewName string, opts QueryOptions) ([]bson.M, error) {
	// Get the view definition
	view, err := s.GetView(ctx, viewName)
	if err != nil {
		return nil, fmt.Errorf("failed to get view: %w", err)
	}

	// Check permissions if userID is set
	if userID != primitive.NilObjectID {
		if hasPermission, err := s.rbacService.HasPermission(ctx, userID, fmt.Sprintf("view:%s:execute", viewName), ""); err != nil {
			return nil, fmt.Errorf("failed to check permissions: %w", err)
		} else if !hasPermission {
			// Check if user has general view execute permission
			if hasPermission, err := s.rbacService.HasPermission(ctx, userID, "view:*:execute", ""); err != nil {
				return nil, fmt.Errorf("failed to check permissions: %w", err)
			} else if !hasPermission {
				s.logAccess(ctx, userID, viewName, nil, "execute", "denied", "insufficient permissions")
				return nil, fmt.Errorf("insufficient permissions to execute view")
			}
		}
	}

	// Build the query from the view's filter
	filter := map[string]interface{}{}
	if view.Filter != nil {
		for k, v := range view.Filter {
			filter[k] = v
		}
	}

	// Merge with additional filters from opts
	if opts.Filter != nil {
		for k, v := range opts.Filter {
			filter[k] = v
		}
	}

	// Get the collection
	dataCol := s.db.Collection("data_" + view.Collection)

	// Create find options
	findOpts := &types.FindOptions{}
	
	// Apply limit and skip from opts (for pagination)
	if opts.Limit > 0 {
		limit := int64(opts.Limit)
		findOpts.Limit = &limit
	}
	if opts.Skip > 0 {
		skip := int64(opts.Skip)
		findOpts.Skip = &skip
	}

	// Execute the query
	cursor, err := dataCol.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to query view: %w", err)
	}
	defer cursor.Close(ctx)

	// Decode results
	var results []bson.M
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		results = append(results, doc)
	}

	if userID != primitive.NilObjectID {
		s.logAccess(ctx, userID, viewName, nil, "execute", "allowed", fmt.Sprintf("queried %d documents", len(results)))
	}

	return results, nil
}

func (s *AdapterService) InsertDocument(ctx context.Context, mutation *models.DataMutation) (*models.Document, error) {
	// Check permissions
	if mutation.UserID != primitive.NilObjectID {
		canWrite, err := s.CanWrite(ctx, mutation.UserID, mutation.Collection)
		if err != nil {
			return nil, err
		}
		if !canWrite {
			s.logAccess(ctx, mutation.UserID, mutation.Collection, nil, "write", "denied", "insufficient permissions")
			return nil, fmt.Errorf("insufficient permissions to write to collection")
		}
	}

	// Validate document against schema if validation is enabled
	if s.validator != nil {
		if err := s.ValidateAgainstSchema(ctx, mutation.Collection, mutation.Data); err != nil {
			return nil, fmt.Errorf("document validation failed: %w", err)
		}
	}

	// Create document
	doc := &models.Document{
		ID:         primitive.NewObjectID(),
		Collection: mutation.Collection,
		Data:       mutation.Data,
		CreatedBy:  mutation.UserID,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	// For PostgreSQL adapter, pass data fields at top level
	// The adapter will handle extracting metadata properly
	docMap := make(map[string]interface{})
	for k, v := range doc.Data {
		docMap[k] = v
	}
	docMap["_id"] = doc.ID.Hex()
	docMap["created_by"] = mutation.UserID.Hex()
	
	// Insert into collection
	dataCol := s.db.Collection("data_" + mutation.Collection)
	_, err := dataCol.InsertOne(ctx, docMap)
	if err != nil {
		return nil, fmt.Errorf("failed to insert document: %w", err)
	}

	s.logAccess(ctx, mutation.UserID, mutation.Collection, &doc.ID, "write", "allowed", "document inserted")
	return doc, nil
}

func (s *AdapterService) UpdateDocument(ctx context.Context, mutation *models.DataMutation) error {
	// Check permissions
	if mutation.UserID != primitive.NilObjectID {
		canUpdate, err := s.CanUpdate(ctx, mutation.UserID, mutation.Collection)
		if err != nil {
			return err
		}
		if !canUpdate {
			s.logAccess(ctx, mutation.UserID, mutation.Collection, &mutation.DocumentID, "update", "denied", "insufficient permissions")
			return fmt.Errorf("insufficient permissions to update collection")
		}
	}

	// Validate document against schema if validation is enabled
	if s.validator != nil {
		if err := s.ValidateAgainstSchema(ctx, mutation.Collection, mutation.Data); err != nil {
			return fmt.Errorf("document validation failed: %w", err)
		}
	}

	// Build update
	filter := map[string]interface{}{
		"_id": mutation.DocumentID.Hex(),
	}

	// For PostgreSQL adapter, don't wrap data in "data" field
	// Pass fields directly for the new structure
	updateFields := make(map[string]interface{})
	for k, v := range mutation.Data {
		// Skip metadata fields
		if !strings.HasPrefix(k, "_") && k != "id" && k != "created_at" && k != "updated_at" && 
		   k != "created_by" && k != "updated_by" && k != "version" && k != "collection" {
			updateFields[k] = v
		}
	}
	updateFields["updated_by"] = mutation.UserID.Hex()
	updateFields["updated_at"] = time.Now().UTC()

	update := map[string]interface{}{
		"$set": updateFields,
	}

	// Update document
	dataCol := s.db.Collection("data_" + mutation.Collection)
	result, err := dataCol.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("document not found")
	}

	s.logAccess(ctx, mutation.UserID, mutation.Collection, &mutation.DocumentID, "update", "allowed", "document updated")
	return nil
}

func (s *AdapterService) DeleteDocument(ctx context.Context, mutation *models.DataMutation) error {
	// Check permissions
	if mutation.UserID != primitive.NilObjectID {
		canDelete, err := s.CanDelete(ctx, mutation.UserID, mutation.Collection)
		if err != nil {
			return err
		}
		if !canDelete {
			s.logAccess(ctx, mutation.UserID, mutation.Collection, &mutation.DocumentID, "delete", "denied", "insufficient permissions")
			return fmt.Errorf("insufficient permissions to delete from collection")
		}
	}

	// Soft delete by setting deleted_at
	filter := map[string]interface{}{
		"_id": mutation.DocumentID.Hex(),
	}

	// Use DeleteOne which handles soft delete in PostgreSQL adapter
	dataCol := s.db.Collection("data_" + mutation.Collection)
	result, err := dataCol.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("document not found")
	}

	s.logAccess(ctx, mutation.UserID, mutation.Collection, &mutation.DocumentID, "delete", "allowed", "document deleted")
	return nil
}

func (s *AdapterService) QueryDocuments(ctx context.Context, query *models.DataQuery) ([]models.Document, error) {
	// Check permissions if userID is set
	if query.UserID != primitive.NilObjectID {
		canRead, err := s.CanRead(ctx, query.UserID, query.Collection)
		if err != nil {
			return nil, err
		}
		if !canRead {
			s.logAccess(ctx, query.UserID, query.Collection, nil, "read", "denied", "insufficient permissions")
			return nil, fmt.Errorf("insufficient permissions to read collection")
		}
	}

	// Build filter
	filter := map[string]interface{}{"deleted_at": nil}
	if query.Filter != nil {
		for k, v := range query.Filter {
			filter[k] = v
		}
	}

	// Get collection
	dataCol := s.db.Collection("data_" + query.Collection)
	
	// Create find options
	findOpts := &types.FindOptions{}
	if query.Limit > 0 {
		limit := int64(query.Limit)
		findOpts.Limit = &limit
	}
	if query.Skip > 0 {
		skip := int64(query.Skip)
		findOpts.Skip = &skip
	}
	if len(query.Sort) > 0 {
		findOpts.Sort = query.Sort
	}

	// Find documents
	cursor, err := dataCol.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to query documents: %w", err)
	}
	defer cursor.Close(ctx)

	// Decode results
	var documents []models.Document
	for cursor.Next(ctx) {
		var doc models.Document
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		documents = append(documents, doc)
	}

	if query.UserID != primitive.NilObjectID {
		s.logAccess(ctx, query.UserID, query.Collection, nil, "read", "allowed", fmt.Sprintf("queried %d documents", len(documents)))
	}

	return documents, nil
}

func (s *AdapterService) GetDocument(ctx context.Context, userID primitive.ObjectID, collection string, docID primitive.ObjectID) (*models.Document, error) {
	// Check permissions
	if userID != primitive.NilObjectID {
		canRead, err := s.CanRead(ctx, userID, collection)
		if err != nil {
			return nil, err
		}
		if !canRead {
			s.logAccess(ctx, userID, collection, &docID, "read", "denied", "insufficient permissions")
			return nil, fmt.Errorf("insufficient permissions to read collection")
		}
	}

	// Get document
	dataCol := s.db.Collection("data_" + collection)
	filter := map[string]interface{}{
		"_id":        docID.Hex(),
		"deleted_at": nil,
	}

	var doc models.Document
	if err := dataCol.FindOne(ctx, filter, &doc); err != nil {
		if err == types.ErrNoDocuments {
			return nil, fmt.Errorf("document not found")
		}
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	s.logAccess(ctx, userID, collection, &docID, "read", "allowed", "document retrieved")
	return &doc, nil
}

func (s *AdapterService) ValidateAgainstSchema(ctx context.Context, collectionName string, document interface{}) error {
	// Get collection metadata to retrieve schema
	col := s.db.Collection("collections")
	filter := map[string]interface{}{
		"name":       collectionName,
		"deleted_at": nil,
	}

	var collection models.Collection
	if err := col.FindOne(ctx, filter, &collection); err != nil {
		if err == types.ErrNoDocuments {
			// No schema defined, allow any document
			return nil
		}
		return fmt.Errorf("failed to get collection schema: %w", err)
	}

	// If no schema is defined, validation passes
	if collection.Schema == nil {
		return nil
	}

	// Use the validator if available
	if s.validator != nil {
		// Convert document to map if needed
		docMap, ok := document.(map[string]interface{})
		if !ok {
			// Try to convert using reflection or JSON marshaling
			return nil // Skip validation for non-map documents
		}
		return s.validator.ValidateDocument(docMap, collection.Schema)
	}

	return nil
}

func (s *AdapterService) ListIndexes(ctx context.Context, collectionName string) ([]interface{}, error) {
	// Get the data collection
	dataCol := s.db.Collection("data_" + collectionName)
	
	// List indexes using the collection's ListIndexes method
	indexes, err := dataCol.ListIndexes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list indexes: %w", err)
	}
	
	// Convert to interface slice
	var result []interface{}
	for _, idx := range indexes {
		result = append(result, idx)
	}
	return result, nil
}

func (s *AdapterService) CreateIndex(ctx context.Context, collectionName string, indexName string, keys map[string]interface{}, options map[string]interface{}) error {
	// Get the data collection
	dataCol := s.db.Collection("data_" + collectionName)
	
	// Build types.Index structure
	idx := types.Index{
		Name:   indexName,
		Keys:   make(map[string]int),
		Unique: false,
		Sparse: false,
	}
	
	// Convert keys to proper format
	for field, order := range keys {
		if orderInt, ok := order.(int); ok {
			idx.Keys[field] = orderInt
		} else if orderFloat, ok := order.(float64); ok {
			idx.Keys[field] = int(orderFloat)
		} else {
			idx.Keys[field] = 1 // Default to ascending
		}
	}
	
	// Check options
	if options != nil {
		if u, ok := options["unique"].(bool); ok {
			idx.Unique = u
		}
		if s, ok := options["sparse"].(bool); ok {
			idx.Sparse = s
		}
	}

	// Create the index using the collection's CreateIndex method
	if err := dataCol.CreateIndex(ctx, idx); err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	return nil
}

func (s *AdapterService) DeleteIndex(ctx context.Context, collectionName string, indexName string) error {
	// Get the data collection
	dataCol := s.db.Collection("data_" + collectionName)
	
	// Drop the index using the collection's DropIndex method
	if err := dataCol.DropIndex(ctx, indexName); err != nil {
		return fmt.Errorf("failed to drop index: %w", err)
	}

	return nil
}

// CountDocuments counts documents in a collection with permissions check
func (s *AdapterService) CountDocuments(ctx context.Context, userID primitive.ObjectID, collection string, filter map[string]interface{}) (int, error) {
	// Check read permission
	hasPermission, err := s.CanRead(ctx, userID, collection)
	if err != nil {
		return 0, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return 0, fmt.Errorf("access denied")
	}

	// Count documents
	col := s.db.Collection("data_" + collection)
	count, err := col.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count documents: %w", err)
	}

	return int(count), nil
}

// Access control methods
func (s *AdapterService) CanRead(ctx context.Context, userID primitive.ObjectID, collection string) (bool, error) {
	return s.rbacService.HasPermission(ctx, userID, fmt.Sprintf("collection:%s", collection), "read")
}

func (s *AdapterService) CanWrite(ctx context.Context, userID primitive.ObjectID, collection string) (bool, error) {
	return s.rbacService.HasPermission(ctx, userID, fmt.Sprintf("collection:%s", collection), "write")
}

func (s *AdapterService) CanUpdate(ctx context.Context, userID primitive.ObjectID, collection string) (bool, error) {
	return s.rbacService.HasPermission(ctx, userID, fmt.Sprintf("collection:%s", collection), "update")
}

func (s *AdapterService) CanDelete(ctx context.Context, userID primitive.ObjectID, collection string) (bool, error) {
	return s.rbacService.HasPermission(ctx, userID, fmt.Sprintf("collection:%s", collection), "delete")
}