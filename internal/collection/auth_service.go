package collection

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CanRead checks if a user can read from a collection
func (s *AdapterService) CanRead(ctx context.Context, userID primitive.ObjectID, collection string) (bool, error) {
	return s.rbacService.HasPermission(ctx, userID, fmt.Sprintf("collection:%s", collection), "read")
}

// CanWrite checks if a user can write to a collection
func (s *AdapterService) CanWrite(ctx context.Context, userID primitive.ObjectID, collection string) (bool, error) {
	return s.rbacService.HasPermission(ctx, userID, fmt.Sprintf("collection:%s", collection), "write")
}

// CanUpdate checks if a user can update a collection
func (s *AdapterService) CanUpdate(ctx context.Context, userID primitive.ObjectID, collection string) (bool, error) {
	return s.rbacService.HasPermission(ctx, userID, fmt.Sprintf("collection:%s", collection), "update")
}

// CanDelete checks if a user can delete from a collection
func (s *AdapterService) CanDelete(ctx context.Context, userID primitive.ObjectID, collection string) (bool, error) {
	return s.rbacService.HasPermission(ctx, userID, fmt.Sprintf("collection:%s", collection), "delete")
}

// logAccess logs access attempts for auditing
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
	
	// Best effort logging - don't fail operation if logging fails
	logsCol.InsertOne(ctx, logEntry)
}