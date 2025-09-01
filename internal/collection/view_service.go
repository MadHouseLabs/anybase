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

// CreateView creates a new view with governance checks
func (s *AdapterService) CreateView(ctx context.Context, userID primitive.ObjectID, view *models.View) error {
	// Check if user has permission on the underlying collection
	hasPermission, err := s.rbacService.HasPermission(ctx, userID, fmt.Sprintf("collection:%s", view.Collection), "read")
	if err != nil {
		return fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		s.logAccess(ctx, userID, view.Name, nil, "create", "denied", "insufficient permissions on collection")
		return fmt.Errorf("insufficient permissions on underlying collection")
	}

	// Check if view already exists
	viewsCol := s.db.Collection("_views")
	existingCount, err := viewsCol.CountDocuments(ctx, map[string]interface{}{"name": view.Name})
	if err != nil {
		return fmt.Errorf("failed to check existing view: %w", err)
	}
	if existingCount > 0 {
		return fmt.Errorf("view '%s' already exists", view.Name)
	}

	// Set metadata
	view.ID = primitive.NewObjectID()
	view.CreatedBy = userID
	view.CreatedAt = time.Now().UTC()
	view.UpdatedAt = time.Now().UTC()

	// Store view metadata
	viewDoc := map[string]interface{}{
		"_id":         view.ID.Hex(),
		"name":        view.Name,
		"collection":  view.Collection,
		"pipeline":    view.Pipeline,
		"fields":      view.Fields,
		"filter":      view.Filter,
		"permissions": view.Permissions,
		"metadata":    view.Metadata,
		"created_by":  userID.Hex(),
		"created_at":  view.CreatedAt,
		"updated_at":  view.UpdatedAt,
	}

	if _, err := viewsCol.InsertOne(ctx, viewDoc); err != nil {
		return fmt.Errorf("failed to create view: %w", err)
	}

	s.logAccess(ctx, userID, view.Name, nil, "create", "allowed", "view created")
	return nil
}

// GetView retrieves a view
func (s *AdapterService) GetView(ctx context.Context, name string) (*models.View, error) {
	viewsCol := s.db.Collection("_views")
	
	var view models.View
	filter := map[string]interface{}{
		"name": name,
		"$or": []interface{}{
			map[string]interface{}{"_deleted_at": nil},
			map[string]interface{}{"_deleted_at": map[string]interface{}{"$exists": false}},
		},
	}
	
	err := viewsCol.FindOne(ctx, filter, &view)
	if err != nil {
		if err == types.ErrNoDocuments {
			return nil, fmt.Errorf("view not found")
		}
		return nil, fmt.Errorf("failed to get view: %w", err)
	}

	return &view, nil
}

// UpdateView updates a view with governance checks
func (s *AdapterService) UpdateView(ctx context.Context, userID primitive.ObjectID, name string, updates bson.M) error {
	// Get the view first to check collection permissions
	view, err := s.GetView(ctx, name)
	if err != nil {
		return err
	}

	// Check if user has permission on the underlying collection
	hasPermission, err := s.rbacService.HasPermission(ctx, userID, fmt.Sprintf("collection:%s", view.Collection), "update")
	if err != nil {
		return fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		s.logAccess(ctx, userID, name, nil, "update", "denied", "insufficient permissions")
		return fmt.Errorf("insufficient permissions to update view")
	}

	viewsCol := s.db.Collection("_views")
	
	// Add updated_at timestamp
	if updateSet, ok := updates["$set"].(bson.M); ok {
		updateSet["updated_at"] = time.Now().UTC()
		updateSet["updated_by"] = userID.Hex()
	} else {
		updates["$set"] = bson.M{
			"updated_at": time.Now().UTC(),
			"updated_by": userID.Hex(),
		}
	}

	filter := map[string]interface{}{
		"name": name,
		"$or": []interface{}{
			map[string]interface{}{"_deleted_at": nil},
			map[string]interface{}{"_deleted_at": map[string]interface{}{"$exists": false}},
		},
	}
	
	_, err = viewsCol.UpdateOne(ctx, filter, updates)
	if err != nil {
		return fmt.Errorf("failed to update view: %w", err)
	}

	s.logAccess(ctx, userID, name, nil, "update", "allowed", "view updated")
	return nil
}

// DeleteView deletes a view with governance checks
func (s *AdapterService) DeleteView(ctx context.Context, userID primitive.ObjectID, name string) error {
	// Get the view first to check collection permissions
	view, err := s.GetView(ctx, name)
	if err != nil {
		return err
	}

	// Check if user has permission on the underlying collection
	hasPermission, err := s.rbacService.HasPermission(ctx, userID, fmt.Sprintf("collection:%s", view.Collection), "delete")
	if err != nil {
		return fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		s.logAccess(ctx, userID, name, nil, "delete", "denied", "insufficient permissions")
		return fmt.Errorf("insufficient permissions to delete view")
	}

	viewsCol := s.db.Collection("_views")
	
	// Soft delete by setting _deleted_at
	filter := map[string]interface{}{"name": name}
	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"_deleted_at": time.Now().UTC(),
			"_deleted_by": userID.Hex(),
		},
	}
	
	_, err = viewsCol.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to delete view: %w", err)
	}

	s.logAccess(ctx, userID, name, nil, "delete", "allowed", "view deleted")
	return nil
}

// ListViews lists all views
func (s *AdapterService) ListViews(ctx context.Context) ([]*models.View, error) {
	viewsCol := s.db.Collection("_views")
	
	filter := map[string]interface{}{
		"$or": []interface{}{
			map[string]interface{}{"_deleted_at": nil},
			map[string]interface{}{"_deleted_at": map[string]interface{}{"$exists": false}},
		},
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

// QueryView queries a view with governance checks
func (s *AdapterService) QueryView(ctx context.Context, userID primitive.ObjectID, viewName string, opts QueryOptions) ([]bson.M, error) {
	// Get the view
	view, err := s.GetView(ctx, viewName)
	if err != nil {
		return nil, err
	}

	// Check if user has permission on the underlying collection
	hasPermission, err := s.rbacService.HasPermission(ctx, userID, fmt.Sprintf("collection:%s", view.Collection), "read")
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		s.logAccess(ctx, userID, viewName, nil, "query", "denied", "insufficient permissions")
		return nil, fmt.Errorf("insufficient permissions to query view")
	}

	// Get the underlying collection
	dataCol := s.db.Collection("data_" + view.Collection)

	// Build the aggregation pipeline
	pipeline := []interface{}{}

	// Add view filter if specified
	if view.Filter != nil && len(view.Filter) > 0 {
		pipeline = append(pipeline, map[string]interface{}{"$match": view.Filter})
	}

	// Add query filter if specified
	if opts.Filter != nil && len(opts.Filter) > 0 {
		pipeline = append(pipeline, map[string]interface{}{"$match": opts.Filter})
	}

	// Add view pipeline if specified
	if view.Pipeline != nil && len(view.Pipeline) > 0 {
		for _, stage := range view.Pipeline {
			pipeline = append(pipeline, stage)
		}
	}

	// Add field projection if specified
	if view.Fields != nil && len(view.Fields) > 0 {
		projection := map[string]interface{}{}
		for _, field := range view.Fields {
			projection[field] = 1
		}
		if len(projection) > 0 {
			pipeline = append(pipeline, map[string]interface{}{"$project": projection})
		}
	}

	// Add sort if specified
	if opts.Sort != nil && len(opts.Sort) > 0 {
		pipeline = append(pipeline, map[string]interface{}{"$sort": opts.Sort})
	}

	// Add limit if specified
	if opts.Limit > 0 {
		pipeline = append(pipeline, map[string]interface{}{"$limit": opts.Limit})
	}

	// Convert pipeline to []map[string]interface{}
	var mapPipeline []map[string]interface{}
	for _, stage := range pipeline {
		if s, ok := stage.(map[string]interface{}); ok {
			mapPipeline = append(mapPipeline, s)
		} else if s, ok := stage.(bson.M); ok {
			mapPipeline = append(mapPipeline, map[string]interface{}(s))
		}
	}
	
	// Execute the aggregation
	cursor, err := dataCol.Aggregate(ctx, mapPipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to query view: %w", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		results = append(results, doc)
	}

	s.logAccess(ctx, userID, viewName, nil, "query", "allowed", fmt.Sprintf("%d results", len(results)))
	return results, nil
}