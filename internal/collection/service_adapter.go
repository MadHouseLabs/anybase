package collection

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/madhouselabs/anybase/internal/database/adapters/postgres"
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

	// Set metadata
	collection.ID = primitive.NewObjectID()
	collection.CreatedBy = userID
	collection.CreatedAt = time.Now()
	collection.UpdatedAt = time.Now()

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
	collectionsCol := s.db.Collection("collections")
	
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
	updates["updated_at"] = time.Now()

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

// ListCollections lists collections with governance checks
func (s *AdapterService) ListCollections(ctx context.Context, userID primitive.ObjectID) ([]*models.Collection, error) {
	// Get user role
	userRole, err := s.rbacService.GetUserRole(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user role: %w", err)
	}

	collectionsCol := s.db.Collection("collections")
	
	// Build filter based on role
	filter := map[string]interface{}{}
	
	// Non-admin users can only see collections they created or have explicit permission for
	if userRole != "admin" {
		// For now, developers can see all collections
		if userRole != "developer" {
			filter["created_by"] = userID.Hex()
		}
	}

	cursor, err := collectionsCol.Find(ctx, filter, nil)
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

// ListAllCollections lists all collections (admin only)
func (s *AdapterService) ListAllCollections(ctx context.Context) ([]*models.Collection, error) {
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
		"timestamp":   time.Now(),
	}
	
	// Best effort logging - don't fail the operation if logging fails
	logsCol.InsertOne(ctx, logEntry)
}

// Stub implementations for remaining methods - these need to be implemented
func (s *AdapterService) CreateView(ctx context.Context, userID primitive.ObjectID, view *models.View) error {
	// TODO: Implement using database adapter
	return fmt.Errorf("not implemented for database adapter yet")
}

func (s *AdapterService) GetView(ctx context.Context, userID primitive.ObjectID, name string) (*models.View, error) {
	// TODO: Implement using database adapter
	return nil, fmt.Errorf("not implemented for database adapter yet")
}

func (s *AdapterService) UpdateView(ctx context.Context, userID primitive.ObjectID, name string, updates bson.M) error {
	// TODO: Implement using database adapter
	return fmt.Errorf("not implemented for database adapter yet")
}

func (s *AdapterService) DeleteView(ctx context.Context, userID primitive.ObjectID, name string) error {
	// TODO: Implement using database adapter
	return fmt.Errorf("not implemented for database adapter yet")
}

func (s *AdapterService) ListViews(ctx context.Context, userID primitive.ObjectID) ([]*models.View, error) {
	// TODO: Implement using database adapter
	return nil, fmt.Errorf("not implemented for database adapter yet")
}

func (s *AdapterService) QueryView(ctx context.Context, userID primitive.ObjectID, viewName string, opts QueryOptions) ([]bson.M, error) {
	// TODO: Implement using database adapter
	return nil, fmt.Errorf("not implemented for database adapter yet")
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
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Convert to map for storage
	docMap := map[string]interface{}{
		"_id":        doc.ID.Hex(),
		"collection": doc.Collection,
		"data":       doc.Data,
		"created_by": mutation.UserID.Hex(),
		"created_at": doc.CreatedAt,
		"updated_at": doc.UpdatedAt,
	}

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
	updateFields["updated_at"] = time.Now()

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
	// For PostgreSQL, we'll query pg_indexes to get index information
	// This is a simplified implementation
	tableName := "data_" + collectionName
	
	// Get the underlying database connection
	if pgAdapter, ok := s.db.(*postgres.PostgresAdapter); ok {
		indexes, err := pgAdapter.GetIndexes(ctx, tableName)
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

	return nil, fmt.Errorf("database adapter does not support index listing")
}

func (s *AdapterService) CreateIndex(ctx context.Context, collectionName string, indexName string, keys map[string]interface{}, options map[string]interface{}) error {
	// For PostgreSQL, create JSONB indexes
	tableName := "data_" + collectionName
	
	// Build index fields
	var indexFields []string
	for field := range keys {
		// Create GIN index on JSONB field for efficient querying
		indexFields = append(indexFields, fmt.Sprintf("(data->'%s')", field))
	}

	if len(indexFields) == 0 {
		return fmt.Errorf("no fields specified for index")
	}

	// Check if it's a unique index
	unique := false
	if options != nil {
		if u, ok := options["unique"].(bool); ok {
			unique = u
		}
	}

	if pgAdapter, ok := s.db.(*postgres.PostgresAdapter); ok {
		return pgAdapter.CreateIndex(ctx, tableName, indexName, indexFields, unique)
	}

	return fmt.Errorf("database adapter does not support index creation")
}

func (s *AdapterService) DeleteIndex(ctx context.Context, collectionName string, indexName string) error {
	if pgAdapter, ok := s.db.(*postgres.PostgresAdapter); ok {
		return pgAdapter.DropIndex(ctx, indexName)
	}

	return fmt.Errorf("database adapter does not support index deletion")
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