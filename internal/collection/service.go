package collection

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/karthik/anybase/internal/database"
	"github.com/karthik/anybase/internal/governance"
	"github.com/karthik/anybase/internal/validator"
	"github.com/karthik/anybase/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Service interface {
	// Collection management
	CreateCollection(ctx context.Context, userID primitive.ObjectID, collection *models.Collection) error
	GetCollection(ctx context.Context, userID primitive.ObjectID, name string) (*models.Collection, error)
	UpdateCollection(ctx context.Context, userID primitive.ObjectID, name string, updates bson.M) error
	DeleteCollection(ctx context.Context, userID primitive.ObjectID, name string) error
	ListCollections(ctx context.Context, userID primitive.ObjectID) ([]*models.Collection, error)
	ListAllCollections(ctx context.Context) ([]*models.Collection, error)

	// View management
	CreateView(ctx context.Context, userID primitive.ObjectID, view *models.View) error
	GetView(ctx context.Context, userID primitive.ObjectID, name string) (*models.View, error)
	UpdateView(ctx context.Context, userID primitive.ObjectID, name string, updates bson.M) error
	DeleteView(ctx context.Context, userID primitive.ObjectID, name string) error
	ListViews(ctx context.Context, userID primitive.ObjectID) ([]*models.View, error)
	QueryView(ctx context.Context, userID primitive.ObjectID, viewName string, opts QueryOptions) ([]bson.M, error)

	// Data operations with governance
	InsertDocument(ctx context.Context, mutation *models.DataMutation) (*models.Document, error)
	UpdateDocument(ctx context.Context, mutation *models.DataMutation) error
	DeleteDocument(ctx context.Context, mutation *models.DataMutation) error
	QueryDocuments(ctx context.Context, query *models.DataQuery) ([]models.Document, error)
	GetDocument(ctx context.Context, userID primitive.ObjectID, collection string, docID primitive.ObjectID) (*models.Document, error)
	CountDocuments(ctx context.Context, userID primitive.ObjectID, collection string, filter map[string]interface{}) (int, error)

	// Access control
	CanRead(ctx context.Context, userID primitive.ObjectID, collection string) (bool, error)
	CanWrite(ctx context.Context, userID primitive.ObjectID, collection string) (bool, error)
	CanUpdate(ctx context.Context, userID primitive.ObjectID, collection string) (bool, error)
	CanDelete(ctx context.Context, userID primitive.ObjectID, collection string) (bool, error)
	
	// Index management
	ListIndexes(ctx context.Context, collectionName string) ([]interface{}, error)
	CreateIndex(ctx context.Context, collectionName string, indexName string, keys map[string]interface{}, options map[string]interface{}) error
	DeleteIndex(ctx context.Context, collectionName string, indexName string) error
}

type QueryOptions struct {
	Limit int
	Skip  int
	Sort  bson.M
}

type service struct {
	db          *database.Database
	rbacService governance.RBACService
	validator   *validator.SchemaValidator
}

func NewService(db *database.Database, rbacService governance.RBACService) Service {
	return &service{
		db:          db,
		rbacService: rbacService,
		validator:   validator.NewSchemaValidator(),
	}
}

// Collection Management

func (s *service) CreateCollection(ctx context.Context, userID primitive.ObjectID, collection *models.Collection) error {
	// Check if user has permission to create collections
	hasPermission, err := s.rbacService.HasPermission(ctx, userID, "collections", "create")
	if err != nil {
		return fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		s.logAccess(ctx, userID, collection.Name, nil, "create", "denied", "insufficient permissions")
		return fmt.Errorf("insufficient permissions to create collection")
	}

	// Set metadata
	collection.ID = primitive.NewObjectID()
	collection.CreatedBy = userID
	collection.CreatedAt = time.Now()
	collection.UpdatedAt = time.Now()

	// Create the actual MongoDB collection
	if err := s.db.GetDatabase().CreateCollection(ctx, "data_"+collection.Name); err != nil {
		// Collection might already exist, which is okay for our metadata
		if !mongo.IsDuplicateKeyError(err) && err.Error() != "collection already exists" {
			return fmt.Errorf("failed to create MongoDB collection: %w", err)
		}
	}

	// Create indexes if specified
	if len(collection.Indexes) > 0 {
		if err := s.createIndexes(ctx, collection); err != nil {
			return fmt.Errorf("failed to create indexes: %w", err)
		}
	}

	// Store collection metadata
	if _, err := s.db.Collection("collections").InsertOne(ctx, collection); err != nil {
		return fmt.Errorf("failed to store collection metadata: %w", err)
	}

	// Permissions are now dynamically generated, no need to persist them

	s.logAccess(ctx, userID, collection.Name, nil, "create", "success", "")
	return nil
}

func (s *service) GetCollection(ctx context.Context, userID primitive.ObjectID, name string) (*models.Collection, error) {
	var collection models.Collection
	if err := s.db.Collection("collections").FindOne(ctx, bson.M{"name": name}).Decode(&collection); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("collection not found")
		}
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}

	// Check read permission
	canRead, err := s.canAccessCollection(ctx, userID, &collection, "read")
	if err != nil {
		return nil, err
	}
	if !canRead {
		s.logAccess(ctx, userID, name, nil, "read", "denied", "insufficient permissions")
		return nil, fmt.Errorf("access denied")
	}

	return &collection, nil
}

func (s *service) UpdateCollection(ctx context.Context, userID primitive.ObjectID, name string, updates bson.M) error {
	// Get collection first to check permissions
	collection, err := s.GetCollection(ctx, userID, name)
	if err != nil {
		return err
	}

	// Check if user can manage this collection
	if collection.CreatedBy != userID {
		hasPermission, err := s.rbacService.HasPermission(ctx, userID, "collections", "update")
		if err != nil {
			return err
		}
		if !hasPermission {
			s.logAccess(ctx, userID, name, nil, "update", "denied", "insufficient permissions")
			return fmt.Errorf("insufficient permissions to update collection")
		}
	}

	updates["updated_at"] = time.Now()
	if _, err := s.db.Collection("collections").UpdateOne(
		ctx,
		bson.M{"name": name},
		bson.M{"$set": updates},
	); err != nil {
		return fmt.Errorf("failed to update collection: %w", err)
	}

	s.logAccess(ctx, userID, name, nil, "update", "success", "")
	return nil
}

func (s *service) DeleteCollection(ctx context.Context, userID primitive.ObjectID, name string) error {
	// Get collection first to check permissions
	collection, err := s.GetCollection(ctx, userID, name)
	if err != nil {
		return err
	}

	// Check if user can delete this collection
	if collection.CreatedBy != userID {
		hasPermission, err := s.rbacService.HasPermission(ctx, userID, "collections", "delete")
		if err != nil {
			return err
		}
		if !hasPermission {
			s.logAccess(ctx, userID, name, nil, "delete", "denied", "insufficient permissions")
			return fmt.Errorf("insufficient permissions to delete collection")
		}
	}

	// Drop the actual MongoDB collection
	if err := s.db.GetDatabase().Collection("data_" + name).Drop(ctx); err != nil {
		return fmt.Errorf("failed to drop MongoDB collection: %w", err)
	}

	// Delete collection metadata
	if _, err := s.db.Collection("collections").DeleteOne(ctx, bson.M{"name": name}); err != nil {
		return fmt.Errorf("failed to delete collection metadata: %w", err)
	}

	s.logAccess(ctx, userID, name, nil, "delete", "success", "")
	return nil
}

func (s *service) ListCollections(ctx context.Context, userID primitive.ObjectID) ([]*models.Collection, error) {
	// Get user role
	userRole, err := s.rbacService.GetUserRole(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	// Build filter to get collections user can access
	filter := bson.M{
		"$or": []bson.M{
			{"created_by": userID},
			{"permissions.read.public": true},
			{"permissions.read.users": userID},
			{"permissions.read.roles": userRole},
		},
	}

	cursor, err := s.db.Collection("collections").Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}
	defer cursor.Close(ctx)

	var collections []*models.Collection
	if err := cursor.All(ctx, &collections); err != nil {
		return nil, fmt.Errorf("failed to decode collections: %w", err)
	}

	return collections, nil
}

func (s *service) ListAllCollections(ctx context.Context) ([]*models.Collection, error) {
	// Return all collections without filtering
	cursor, err := s.db.Collection("collections").Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}
	defer cursor.Close(ctx)

	var collections []*models.Collection
	if err := cursor.All(ctx, &collections); err != nil {
		return nil, fmt.Errorf("failed to decode collections: %w", err)
	}

	return collections, nil
}

// View Management

func (s *service) CreateView(ctx context.Context, userID primitive.ObjectID, view *models.View) error {
	// Check if user has permission to create views
	hasPermission, err := s.rbacService.HasPermission(ctx, userID, "views", "create")
	if err != nil {
		return fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return fmt.Errorf("insufficient permissions to create view")
	}

	// Verify the collection exists and user can read it
	if _, err := s.GetCollection(ctx, userID, view.Collection); err != nil {
		return fmt.Errorf("cannot create view on collection: %w", err)
	}

	view.ID = primitive.NewObjectID()
	view.CreatedBy = userID
	view.CreatedAt = time.Now()
	view.UpdatedAt = time.Now()

	if _, err := s.db.Collection("views").InsertOne(ctx, view); err != nil {
		return fmt.Errorf("failed to create view: %w", err)
	}

	// Permissions are now dynamically generated, no need to persist them

	return nil
}

func (s *service) GetView(ctx context.Context, userID primitive.ObjectID, name string) (*models.View, error) {
	var view models.View
	if err := s.db.Collection("views").FindOne(ctx, bson.M{"name": name}).Decode(&view); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("view not found")
		}
		return nil, fmt.Errorf("failed to get view: %w", err)
	}

	// Check if user can access this view
	if !s.canAccessView(ctx, userID, &view) {
		return nil, fmt.Errorf("access denied")
	}

	return &view, nil
}

func (s *service) UpdateView(ctx context.Context, userID primitive.ObjectID, name string, updates bson.M) error {
	view, err := s.GetView(ctx, userID, name)
	if err != nil {
		return err
	}

	// Only creator or admin can update
	if view.CreatedBy != userID {
		hasPermission, err := s.rbacService.HasPermission(ctx, userID, "views", "update")
		if err != nil {
			return err
		}
		if !hasPermission {
			return fmt.Errorf("insufficient permissions to update view")
		}
	}

	updates["updated_at"] = time.Now()
	if _, err := s.db.Collection("views").UpdateOne(
		ctx,
		bson.M{"name": name},
		bson.M{"$set": updates},
	); err != nil {
		return fmt.Errorf("failed to update view: %w", err)
	}

	return nil
}

func (s *service) DeleteView(ctx context.Context, userID primitive.ObjectID, name string) error {
	view, err := s.GetView(ctx, userID, name)
	if err != nil {
		return err
	}

	// Only creator or admin can delete
	if view.CreatedBy != userID {
		hasPermission, err := s.rbacService.HasPermission(ctx, userID, "views", "delete")
		if err != nil {
			return err
		}
		if !hasPermission {
			return fmt.Errorf("insufficient permissions to delete view")
		}
	}

	if _, err := s.db.Collection("views").DeleteOne(ctx, bson.M{"name": name}); err != nil {
		return fmt.Errorf("failed to delete view: %w", err)
	}

	return nil
}

func (s *service) ListViews(ctx context.Context, userID primitive.ObjectID) ([]*models.View, error) {
	userRole, err := s.rbacService.GetUserRole(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user role: %w", err)
	}

	filter := bson.M{
		"$or": []bson.M{
			{"created_by": userID},
			{"permissions.public": true},
			{"permissions.users": userID},
			{"permissions.roles": userRole},
		},
	}

	cursor, err := s.db.Collection("views").Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list views: %w", err)
	}
	defer cursor.Close(ctx)

	var views []*models.View
	if err := cursor.All(ctx, &views); err != nil {
		return nil, fmt.Errorf("failed to decode views: %w", err)
	}

	return views, nil
}

func (s *service) QueryView(ctx context.Context, userID primitive.ObjectID, viewName string, opts QueryOptions) ([]bson.M, error) {
	view, err := s.GetView(ctx, userID, viewName)
	if err != nil {
		return nil, err
	}

	// Build aggregation pipeline
	pipeline := view.Pipeline
	if pipeline == nil {
		pipeline = []bson.M{}
	}

	// Add filter if specified in view
	if view.Filter != nil {
		pipeline = append([]bson.M{{"$match": view.Filter}}, pipeline...)
	}

	// Add field projection if specified
	if len(view.Fields) > 0 {
		projection := bson.M{}
		for _, field := range view.Fields {
			projection[field] = 1
		}
		pipeline = append(pipeline, bson.M{"$project": projection})
	}

	// Add sort if specified
	if view.Sort != nil || opts.Sort != nil {
		sort := view.Sort
		if opts.Sort != nil {
			sort = opts.Sort
		}
		pipeline = append(pipeline, bson.M{"$sort": sort})
	}

	// Add pagination
	if opts.Skip > 0 {
		pipeline = append(pipeline, bson.M{"$skip": opts.Skip})
	}
	if opts.Limit > 0 {
		pipeline = append(pipeline, bson.M{"$limit": opts.Limit})
	}

	// Execute aggregation
	cursor, err := s.db.Collection("data_"+view.Collection).Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to query view: %w", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode results: %w", err)
	}

	return results, nil
}

// Data Operations

func (s *service) InsertDocument(ctx context.Context, mutation *models.DataMutation) (*models.Document, error) {
	// Check if access key has already been validated
	var collection *models.Collection
	if validated, ok := ctx.Value("access_key_validated").(bool); ok && validated {
		// Access key permissions already checked in handler, proceed
		// Get collection metadata without permission check
		var col models.Collection
		if err := s.db.Collection("collections").FindOne(ctx, bson.M{"name": mutation.Collection}).Decode(&col); err != nil {
			if err == mongo.ErrNoDocuments {
				// Collection doesn't exist in metadata, create minimal collection
				collection = &models.Collection{
					Name: mutation.Collection,
				}
			} else {
				return nil, fmt.Errorf("failed to get collection: %w", err)
			}
		} else {
			collection = &col
		}
	} else {
		// Regular user auth - check write permission
		canWrite, err := s.CanWrite(ctx, mutation.UserID, mutation.Collection)
		if err != nil {
			return nil, err
		}
		if !canWrite {
			s.logAccess(ctx, mutation.UserID, mutation.Collection, nil, "write", "denied", "insufficient permissions")
			return nil, fmt.Errorf("insufficient permissions to write to collection")
		}

		// Get collection metadata for validation
		collection, err = s.GetCollection(ctx, mutation.UserID, mutation.Collection)
		if err != nil {
			return nil, err
		}
	}

	// Validate against schema if defined
	if collection.Schema != nil {
		if err := s.validator.ValidateDocument(mutation.Data, collection.Schema); err != nil {
			return nil, fmt.Errorf("document validation failed: %w", err)
		}
	}

	// Create document
	doc := &models.Document{
		ID:        primitive.NewObjectID(),
		Data:      mutation.Data,
		CreatedBy: mutation.UserID,
		UpdatedBy: mutation.UserID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   1,
	}

	// Insert into collection
	if _, err := s.db.Collection("data_"+mutation.Collection).InsertOne(ctx, doc); err != nil {
		return nil, fmt.Errorf("failed to insert document: %w", err)
	}

	s.logAccess(ctx, mutation.UserID, mutation.Collection, &doc.ID, "write", "success", "")
	return doc, nil
}

func (s *service) UpdateDocument(ctx context.Context, mutation *models.DataMutation) error {
	// Check if access key has already been validated
	var collection *models.Collection
	if validated, ok := ctx.Value("access_key_validated").(bool); ok && validated {
		// Access key permissions already checked in handler, proceed
		// Get collection metadata without permission check
		var col models.Collection
		if err := s.db.Collection("collections").FindOne(ctx, bson.M{"name": mutation.Collection}).Decode(&col); err != nil {
			if err == mongo.ErrNoDocuments {
				// Collection doesn't exist in metadata, create minimal collection
				collection = &models.Collection{
					Name: mutation.Collection,
				}
			} else {
				return fmt.Errorf("failed to get collection: %w", err)
			}
		} else {
			collection = &col
		}
	} else {
		// Regular user auth - check update permission
		canUpdate, err := s.CanUpdate(ctx, mutation.UserID, mutation.Collection)
		if err != nil {
			return err
		}
		if !canUpdate {
			s.logAccess(ctx, mutation.UserID, mutation.Collection, &mutation.DocumentID, "update", "denied", "insufficient permissions")
			return fmt.Errorf("insufficient permissions to update document")
		}

		// Get collection metadata for validation
		collection, err = s.GetCollection(ctx, mutation.UserID, mutation.Collection)
		if err != nil {
			return err
		}
	}

	filter := bson.M{"_id": mutation.DocumentID}
	if mutation.Filter != nil {
		for k, v := range mutation.Filter {
			filter[k] = v
		}
	}

	// Check if this is an access key request (partial update) or UI request (full replacement)
	if validated, ok := ctx.Value("access_key_validated").(bool); ok && validated {
		// For partial updates, we need to validate the merged document
		if collection.Schema != nil {
			// Get existing document first
			var existingDoc bson.M
			err := s.db.Collection("data_"+mutation.Collection).FindOne(ctx, filter).Decode(&existingDoc)
			if err != nil {
				return fmt.Errorf("failed to find document for validation: %w", err)
			}
			
			// Create merged document for validation (simulate the result after update)
			mergedDoc := make(map[string]interface{})
			// Copy existing data (excluding metadata fields)
			for k, v := range existingDoc {
				if !strings.HasPrefix(k, "_") && k != "id" && k != "created_at" && k != "updated_at" && 
				   k != "created_by" && k != "updated_by" && k != "version" && k != "collection" {
					mergedDoc[k] = v
				}
			}
			// Apply updates
			for k, v := range mutation.Data {
				if !strings.HasPrefix(k, "_") && k != "id" && k != "created_at" && k != "updated_at" && 
				   k != "created_by" && k != "updated_by" && k != "version" && k != "collection" {
					mergedDoc[k] = v
				}
			}
			
			// Validate the merged document
			if err := s.validator.ValidateDocument(mergedDoc, collection.Schema); err != nil {
				return fmt.Errorf("document validation failed: %w", err)
			}
		}
		// Access key: Use $set for partial updates
		updateFields := bson.M{
			"_updated_at": time.Now(),
		}
		
		// Add user ID if available
		if mutation.UserID != primitive.NilObjectID {
			updateFields["_updated_by"] = mutation.UserID
		}
		
		// Add all the data fields from the mutation
		for k, v := range mutation.Data {
			// Skip any metadata fields that might have been accidentally included
			if !strings.HasPrefix(k, "_") && k != "id" && k != "created_at" && k != "updated_at" && 
			   k != "created_by" && k != "updated_by" && k != "version" && k != "collection" {
				updateFields[k] = v
			}
		}
		
		update := bson.M{
			"$set": updateFields,
			"$inc": bson.M{"_version": 1},
		}
		
		// Perform partial update
		result, err := s.db.Collection("data_"+mutation.Collection).UpdateOne(ctx, filter, update)
		if err != nil {
			return fmt.Errorf("failed to update document: %w", err)
		}
		
		if result.ModifiedCount == 0 {
			return fmt.Errorf("document not found or not modified")
		}
	} else {
		// UI/JWT: Full document replacement (existing behavior)
		// Get existing document to preserve metadata
		var existingDoc bson.M
		err := s.db.Collection("data_"+mutation.Collection).FindOne(ctx, bson.M{"_id": mutation.DocumentID}).Decode(&existingDoc)
		if err != nil {
			return fmt.Errorf("failed to find document: %w", err)
		}
		
		// Create replacement document with new data but preserving certain metadata
		replacementDoc := bson.M{
			"_id": mutation.DocumentID,
			"_created_by": existingDoc["_created_by"],
			"_created_at": existingDoc["_created_at"],
			"_updated_by": mutation.UserID,
			"_updated_at": time.Now(),
		}
		
		// Handle version increment safely
		if v, ok := existingDoc["_version"]; ok {
			switch version := v.(type) {
			case int32:
				replacementDoc["_version"] = version + 1
			case int64:
				replacementDoc["_version"] = int32(version) + 1
			case int:
				replacementDoc["_version"] = int32(version) + 1
			default:
				replacementDoc["_version"] = int32(2)
			}
		} else {
			replacementDoc["_version"] = int32(1)
		}
		
		// Add all the data fields from the mutation
		for k, v := range mutation.Data {
			// Skip any metadata fields that might have been accidentally included
			if !strings.HasPrefix(k, "_") && k != "id" && k != "created_at" && k != "updated_at" && 
			   k != "created_by" && k != "updated_by" && k != "version" && k != "collection" {
				replacementDoc[k] = v
			}
		}
		
		// Validate against schema if defined (for full replacement)
		if collection.Schema != nil {
			// Validate only the data fields (exclude metadata)
			dataOnly := make(map[string]interface{})
			for k, v := range replacementDoc {
				if !strings.HasPrefix(k, "_") {
					dataOnly[k] = v
				}
			}
			if err := s.validator.ValidateDocument(dataOnly, collection.Schema); err != nil {
				return fmt.Errorf("document validation failed: %w", err)
			}
		}

		// Replace the entire document
		result, err := s.db.Collection("data_"+mutation.Collection).ReplaceOne(ctx, filter, replacementDoc)
		if err != nil {
			return fmt.Errorf("failed to update document: %w", err)
		}

		if result.MatchedCount == 0 {
			return fmt.Errorf("document not found or filter didn't match")
		}
	}

	s.logAccess(ctx, mutation.UserID, mutation.Collection, &mutation.DocumentID, "update", "success", "")
	return nil
}

func (s *service) DeleteDocument(ctx context.Context, mutation *models.DataMutation) error {
	// Check if access key has already been validated
	if validated, ok := ctx.Value("access_key_validated").(bool); ok && validated {
		// Access key permissions already checked in handler, proceed
	} else {
		// Regular user auth - check delete permission
		canDelete, err := s.CanDelete(ctx, mutation.UserID, mutation.Collection)
		if err != nil {
			return err
		}
		if !canDelete {
			s.logAccess(ctx, mutation.UserID, mutation.Collection, &mutation.DocumentID, "delete", "denied", "insufficient permissions")
			return fmt.Errorf("insufficient permissions to delete document")
		}
	}

	// Get collection settings
	var collection *models.Collection
	if validated, ok := ctx.Value("access_key_validated").(bool); ok && validated {
		// For access keys, get collection without permission check
		var col models.Collection
		if err := s.db.Collection("collections").FindOne(ctx, bson.M{"name": mutation.Collection}).Decode(&col); err != nil {
			if err == mongo.ErrNoDocuments {
				// Collection doesn't exist in metadata, create minimal collection with default soft delete
				collection = &models.Collection{
					Name: mutation.Collection,
					Settings: models.CollectionSettings{
						SoftDelete: true, // Default to soft delete
					},
				}
			} else {
				return fmt.Errorf("failed to get collection: %w", err)
			}
		} else {
			collection = &col
		}
	} else {
		// Regular user auth
		var err error
		collection, err = s.GetCollection(ctx, mutation.UserID, mutation.Collection)
		if err != nil {
			return err
		}
	}

	filter := bson.M{"_id": mutation.DocumentID}
	if mutation.Filter != nil {
		for k, v := range mutation.Filter {
			filter[k] = v
		}
	}

	if collection.Settings.SoftDelete {
		// Soft delete
		update := bson.M{
			"$set": bson.M{
				"_deleted_at": time.Now(),
				"_updated_by": mutation.UserID,
				"_updated_at": time.Now(),
			},
		}
		result, err := s.db.Collection("data_"+mutation.Collection).UpdateOne(ctx, filter, update)
		if err != nil {
			return fmt.Errorf("failed to soft delete document: %w", err)
		}
		if result.MatchedCount == 0 {
			return fmt.Errorf("document not found")
		}
	} else {
		// Hard delete
		result, err := s.db.Collection("data_"+mutation.Collection).DeleteOne(ctx, filter)
		if err != nil {
			return fmt.Errorf("failed to delete document: %w", err)
		}
		if result.DeletedCount == 0 {
			return fmt.Errorf("document not found")
		}
	}

	s.logAccess(ctx, mutation.UserID, mutation.Collection, &mutation.DocumentID, "delete", "success", "")
	return nil
}

func (s *service) QueryDocuments(ctx context.Context, query *models.DataQuery) ([]models.Document, error) {
	// Check if access key has already been validated
	if validated, ok := ctx.Value("access_key_validated").(bool); ok && validated {
		// Access key permissions already checked in handler, proceed
	} else {
		// Regular user auth - check read permission
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
	filter := bson.M{"_deleted_at": nil} // Exclude soft-deleted documents
	if query.Filter != nil {
		for k, v := range query.Filter {
			filter[k] = v
		}
	}

	// Get collection metadata
	var collection *models.Collection
	if validated, ok := ctx.Value("access_key_validated").(bool); ok && validated {
		// For access keys, get collection without permission check
		var col models.Collection
		if err := s.db.Collection("collections").FindOne(ctx, bson.M{"name": query.Collection}).Decode(&col); err != nil {
			if err == mongo.ErrNoDocuments {
				// Collection doesn't exist in metadata, but might exist as MongoDB collection
				// Create a minimal collection object
				collection = &models.Collection{
					Name: query.Collection,
				}
			} else {
				return nil, fmt.Errorf("failed to get collection: %w", err)
			}
		} else {
			collection = &col
		}
	} else {
		// Regular user auth - get collection with permission check
		var err error
		collection, err = s.GetCollection(ctx, query.UserID, query.Collection)
		if err != nil {
			return nil, err
		}
	}

	// Build projection based on field permissions
	userRole := ""
	if len(query.UserRoles) > 0 {
		userRole = query.UserRoles[0]
	}
	projection := s.buildProjection(collection, userRole, query.Fields)

	// Build options
	findOptions := options.Find()
	if len(projection) > 0 {
		findOptions.SetProjection(projection)
	}
	if query.Sort != nil {
		findOptions.SetSort(query.Sort)
	}
	if query.Limit > 0 {
		findOptions.SetLimit(int64(query.Limit))
	}
	if query.Skip > 0 {
		findOptions.SetSkip(int64(query.Skip))
	}

	// Execute query
	cursor, err := s.db.Collection("data_"+query.Collection).Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to query documents: %w", err)
	}
	defer cursor.Close(ctx)

	var documents []models.Document
	if err := cursor.All(ctx, &documents); err != nil {
		return nil, fmt.Errorf("failed to decode documents: %w", err)
	}

	// Set collection name in response
	for i := range documents {
		documents[i].Collection = query.Collection
	}

	s.logAccess(ctx, query.UserID, query.Collection, nil, "read", "success", fmt.Sprintf("returned %d documents", len(documents)))
	return documents, nil
}

func (s *service) GetDocument(ctx context.Context, userID primitive.ObjectID, collection string, docID primitive.ObjectID) (*models.Document, error) {
	query := &models.DataQuery{
		Collection: collection,
		Filter:     map[string]interface{}{"_id": docID},
		UserID:     userID,
	}

	// Only get user role if we have a userID (not for access keys)
	if userID != primitive.NilObjectID {
		userRole, err := s.rbacService.GetUserRole(ctx, userID)
		if err != nil {
			return nil, err
		}
		query.UserRoles = []string{userRole}
	}

	docs, err := s.QueryDocuments(ctx, query)
	if err != nil {
		return nil, err
	}

	if len(docs) == 0 {
		return nil, fmt.Errorf("document not found")
	}

	return &docs[0], nil
}

func (s *service) CountDocuments(ctx context.Context, userID primitive.ObjectID, collection string, filter map[string]interface{}) (int, error) {
	// Check if access key has already been validated
	if validated, ok := ctx.Value("access_key_validated").(bool); ok && validated {
		// Access key permissions already checked in handler, proceed
	} else if userID != primitive.NilObjectID {
		// Regular user auth - check read permission
		canRead, err := s.CanRead(ctx, userID, collection)
		if err != nil {
			return 0, err
		}
		if !canRead {
			return 0, fmt.Errorf("insufficient permissions to read collection")
		}
	}

	// Build filter
	countFilter := bson.M{"_deleted_at": nil} // Exclude soft-deleted documents
	if filter != nil {
		for k, v := range filter {
			countFilter[k] = v
		}
	}

	// Count documents
	count, err := s.db.Collection("data_"+collection).CountDocuments(ctx, countFilter)
	if err != nil {
		return 0, fmt.Errorf("failed to count documents: %w", err)
	}

	return int(count), nil
}

// Access Control Methods

func (s *service) CanRead(ctx context.Context, userID primitive.ObjectID, collectionName string) (bool, error) {
	collection, err := s.GetCollection(ctx, userID, collectionName)
	if err != nil {
		// If collection doesn't exist or user can't access metadata, deny
		return false, nil
	}

	return s.canAccessCollection(ctx, userID, collection, "read")
}

func (s *service) CanWrite(ctx context.Context, userID primitive.ObjectID, collectionName string) (bool, error) {
	collection, err := s.GetCollection(ctx, userID, collectionName)
	if err != nil {
		return false, nil
	}

	return s.canAccessCollection(ctx, userID, collection, "write")
}

func (s *service) CanUpdate(ctx context.Context, userID primitive.ObjectID, collectionName string) (bool, error) {
	collection, err := s.GetCollection(ctx, userID, collectionName)
	if err != nil {
		return false, nil
	}

	return s.canAccessCollection(ctx, userID, collection, "update")
}

func (s *service) CanDelete(ctx context.Context, userID primitive.ObjectID, collectionName string) (bool, error) {
	collection, err := s.GetCollection(ctx, userID, collectionName)
	if err != nil {
		return false, nil
	}

	return s.canAccessCollection(ctx, userID, collection, "delete")
}

// Index Management

func (s *service) ListIndexes(ctx context.Context, collectionName string) ([]interface{}, error) {
	// Get the MongoDB collection with data_ prefix
	coll := s.db.Collection("data_" + collectionName)
	
	// List indexes
	cursor, err := coll.Indexes().List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list indexes: %w", err)
	}
	defer cursor.Close(ctx)
	
	// Decode indexes as bson.M to properly handle the structure
	var rawIndexes []bson.M
	if err := cursor.All(ctx, &rawIndexes); err != nil {
		return nil, fmt.Errorf("failed to decode indexes: %w", err)
	}
	
	// Convert to cleaner format
	var indexes []interface{}
	for _, idx := range rawIndexes {
		index := make(map[string]interface{})
		
		// Extract index name
		if name, ok := idx["name"].(string); ok {
			index["name"] = name
		}
		
		// Extract key specification
		if key, ok := idx["key"].(bson.M); ok {
			index["key"] = key
		}
		
		// Extract unique flag
		if unique, ok := idx["unique"].(bool); ok {
			index["unique"] = unique
		}
		
		// Extract sparse flag  
		if sparse, ok := idx["sparse"].(bool); ok {
			index["sparse"] = sparse
		}
		
		// Extract version
		if v, ok := idx["v"].(int32); ok {
			index["v"] = v
		}
		
		indexes = append(indexes, index)
	}
	
	return indexes, nil
}

func (s *service) CreateIndex(ctx context.Context, collectionName string, indexName string, keys map[string]interface{}, indexOptions map[string]interface{}) error {
	// Get the MongoDB collection with data_ prefix
	coll := s.db.Collection("data_" + collectionName)
	
	// Convert keys to bson.D for ordered index keys
	var indexKeys bson.D
	for field, direction := range keys {
		indexKeys = append(indexKeys, bson.E{Key: field, Value: direction})
	}
	
	// Create index model
	indexModel := mongo.IndexModel{
		Keys: indexKeys,
	}
	
	// Set index options
	indexOpts := options.Index()
	if indexName != "" {
		indexOpts = indexOpts.SetName(indexName)
	}
	
	// Handle common options from the provided options map
	if unique, ok := indexOptions["unique"].(bool); ok {
		indexOpts = indexOpts.SetUnique(unique)
	}
	if sparse, ok := indexOptions["sparse"].(bool); ok {
		indexOpts = indexOpts.SetSparse(sparse)
	}
	
	indexModel.Options = indexOpts
	
	// Create the index
	_, err := coll.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	
	return nil
}

func (s *service) DeleteIndex(ctx context.Context, collectionName string, indexName string) error {
	// Get the MongoDB collection with data_ prefix
	coll := s.db.Collection("data_" + collectionName)
	
	// Drop the index
	_, err := coll.Indexes().DropOne(ctx, indexName)
	if err != nil {
		return fmt.Errorf("failed to delete index: %w", err)
	}
	
	return nil
}

// Helper Methods

func (s *service) canAccessCollection(ctx context.Context, userID primitive.ObjectID, collection *models.Collection, operation string) (bool, error) {
	// Creator always has full access
	if collection.CreatedBy == userID {
		return true, nil
	}

	// Get the appropriate permission rule
	var rule models.PermissionRule
	switch operation {
	case "read":
		rule = collection.Permissions.Read
	case "write":
		rule = collection.Permissions.Write
	case "update":
		rule = collection.Permissions.Update
	case "delete":
		rule = collection.Permissions.Delete
	default:
		return false, fmt.Errorf("unknown operation: %s", operation)
	}

	// Check public access
	if rule.Public {
		return true, nil
	}

	// Check specific user access
	for _, allowedUser := range rule.Users {
		if allowedUser == userID {
			return true, nil
		}
	}

	// Check role-based access
	userRole, err := s.rbacService.GetUserRole(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, role := range rule.Roles {
		if role == userRole {
			return true, nil
		}
	}

	return false, nil
}

func (s *service) canAccessView(ctx context.Context, userID primitive.ObjectID, view *models.View) bool {
	// Creator always has access
	if view.CreatedBy == userID {
		return true
	}

	// Check public access
	if view.Permissions.Public {
		return true
	}

	// Check specific user access
	for _, allowedUser := range view.Permissions.Users {
		if allowedUser == userID {
			return true
		}
	}

	// Check role-based access
	userRole, _ := s.rbacService.GetUserRole(ctx, userID)
	for _, role := range view.Permissions.Roles {
		if role == userRole {
			return true
		}
	}

	return false
}

func (s *service) createIndexes(ctx context.Context, collection *models.Collection) error {
	indexModels := []mongo.IndexModel{}
	
	for _, idx := range collection.Indexes {
		indexModel := mongo.IndexModel{
			Keys: idx.Fields,
			Options: options.Index().
				SetName(idx.Name).
				SetUnique(idx.Unique).
				SetSparse(idx.Sparse),
		}
		indexModels = append(indexModels, indexModel)
	}

	if len(indexModels) > 0 {
		_, err := s.db.Collection("data_"+collection.Name).Indexes().CreateMany(ctx, indexModels)
		return err
	}

	return nil
}


func (s *service) buildProjection(collection *models.Collection, userRole string, requestedFields []string) bson.M {
	projection := bson.M{}

	// If no schema or no field permissions, return requested fields or all
	if collection.Schema == nil {
		if len(requestedFields) > 0 {
			for _, field := range requestedFields {
				projection[field] = 1
			}
		}
		return projection
	}

	// For OpenAPI schema, we don't have field-level permissions yet
	// Just return requested fields or all fields from schema
	if len(requestedFields) > 0 {
		for _, field := range requestedFields {
			projection[field] = 1
		}
	} else if collection.Schema.Properties != nil {
		// Include all properties from schema
		for propName := range collection.Schema.Properties {
			projection[propName] = 1
		}
	}

	// Always include system fields
	projection["_id"] = 1
	projection["_created_at"] = 1
	projection["_updated_at"] = 1

	return projection
}

func (s *service) logAccess(ctx context.Context, userID primitive.ObjectID, collection string, documentID *primitive.ObjectID, operation, status, error string) {
	accessLog := models.AccessLog{
		ID:         primitive.NewObjectID(),
		UserID:     userID,
		Collection: collection,
		DocumentID: documentID,
		Operation:  operation,
		Status:     status,
		Error:      error,
		CreatedAt:  time.Now(),
	}

	// Fire and forget - don't fail the operation if logging fails
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.db.Collection("access_logs").InsertOne(ctx, accessLog)
	}()
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
