package governance

import (
	"context"
	"fmt"
	"strings"

	"github.com/karthik/anybase/internal/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// SystemRoles defines the fixed system roles
var SystemRoles = map[string][]string{
	"admin": {
		"*:*:*", // Full access to everything
	},
	"developer": {
		"collection:*:*", // Full access to all collections
		"view:*:*",       // Full access to all views
		"api:*:read",     // Read access to API resources
	},
}

type RBACService interface {
	// User role assignment (simplified - just strings now)
	GetUserRole(ctx context.Context, userID primitive.ObjectID) (string, error)
	SetUserRole(ctx context.Context, userID primitive.ObjectID, role string) error
	
	// Authorization checks
	HasPermission(ctx context.Context, userID primitive.ObjectID, resource, action string) (bool, error)
	HasRole(ctx context.Context, userID primitive.ObjectID, role string) (bool, error)
	GetEffectivePermissions(ctx context.Context, userID primitive.ObjectID) ([]string, error)
	
	// Available permissions (for UI)
	GetAvailablePermissions(ctx context.Context) ([]string, error)
}

type rbacService struct {
	db *database.Database
}

func NewRBACService(db *database.Database) RBACService {
	return &rbacService{db: db}
}

func (s *rbacService) usersCollection() *mongo.Collection {
	return s.db.Collection("users")
}

// GetUserRole gets the role of a user
func (s *rbacService) GetUserRole(ctx context.Context, userID primitive.ObjectID) (string, error) {
	var user struct {
		Role string `bson:"role"`
	}

	err := s.usersCollection().FindOne(
		ctx,
		bson.M{"_id": userID},
	).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", fmt.Errorf("user not found")
		}
		return "", fmt.Errorf("failed to get user role: %w", err)
	}

	// Default to developer if no role set
	if user.Role == "" {
		return "developer", nil
	}

	return user.Role, nil
}

// SetUserRole sets the role of a user (admin or developer only)
func (s *rbacService) SetUserRole(ctx context.Context, userID primitive.ObjectID, role string) error {
	// Validate role
	if role != "admin" && role != "developer" {
		return fmt.Errorf("invalid role: must be 'admin' or 'developer'")
	}

	update := bson.M{
		"$set": bson.M{"role": role},
	}

	result, err := s.usersCollection().UpdateOne(ctx, bson.M{"_id": userID}, update)
	if err != nil {
		return fmt.Errorf("failed to set user role: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// HasRole checks if a user has a specific role
func (s *rbacService) HasRole(ctx context.Context, userID primitive.ObjectID, role string) (bool, error) {
	userRole, err := s.GetUserRole(ctx, userID)
	if err != nil {
		return false, err
	}

	return userRole == role, nil
}

// GetEffectivePermissions gets all permissions for a user based on their role
func (s *rbacService) GetEffectivePermissions(ctx context.Context, userID primitive.ObjectID) ([]string, error) {
	role, err := s.GetUserRole(ctx, userID)
	if err != nil {
		return nil, err
	}

	permissions, ok := SystemRoles[role]
	if !ok {
		// Default to no permissions if role not found
		return []string{}, nil
	}

	return permissions, nil
}

// HasPermission checks if a user has a specific permission
func (s *rbacService) HasPermission(ctx context.Context, userID primitive.ObjectID, resource, action string) (bool, error) {
	// Construct the permission string
	var permissionName string
	if strings.Count(resource, ":") == 1 {
		// Already in type:name format, just append action
		permissionName = fmt.Sprintf("%s:%s", resource, action)
	} else {
		// Legacy format
		permissionName = fmt.Sprintf("%s:%s", resource, action)
	}
	
	// Get user's effective permissions
	permissions, err := s.GetEffectivePermissions(ctx, userID)
	if err != nil {
		return false, err
	}

	// Check if user has the specific permission
	for _, p := range permissions {
		// Exact match
		if p == permissionName {
			return true, nil
		}
		
		// Super admin wildcard
		if p == "*:*:*" {
			return true, nil
		}
		
		// Pattern matching for dynamic permissions
		if strings.Contains(p, "*") {
			parts := strings.Split(p, ":")
			targetParts := strings.Split(permissionName, ":")
			
			// Handle both 2-part and 3-part permissions
			if len(parts) == len(targetParts) {
				match := true
				for i, part := range parts {
					if part != "*" && part != targetParts[i] {
						match = false
						break
					}
				}
				if match {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

// GetAvailablePermissions returns all available permissions based on existing collections and views
func (s *rbacService) GetAvailablePermissions(ctx context.Context) ([]string, error) {
	permissions := []string{
		// Wildcard patterns
		"*:*:*", // Super admin - full access
		"collection:*:*", // All collections, all actions
		"collection:*:read",
		"collection:*:write",
		"collection:*:update",
		"collection:*:delete",
		"view:*:*", // All views, all actions
		"view:*:read",
		"view:*:execute",
		"api:*:*", // All API resources
		"system:*:*", // System resources
	}

	// Get all collections for specific permissions
	collectionsCursor, err := s.db.Collection("collections").Find(ctx, bson.M{})
	if err == nil {
		defer collectionsCursor.Close(ctx)
		var collections []struct{ Name string `bson:"name"` }
		if err := collectionsCursor.All(ctx, &collections); err == nil {
			for _, col := range collections {
				for _, action := range []string{"read", "write", "update", "delete"} {
					permissions = append(permissions, fmt.Sprintf("collection:%s:%s", col.Name, action))
				}
			}
		}
	}

	// Get all views for specific permissions
	viewsCursor, err := s.db.Collection("views").Find(ctx, bson.M{})
	if err == nil {
		defer viewsCursor.Close(ctx)
		var views []struct{ Name string `bson:"name"` }
		if err := viewsCursor.All(ctx, &views); err == nil {
			for _, view := range views {
				for _, action := range []string{"read", "execute"} {
					permissions = append(permissions, fmt.Sprintf("view:%s:%s", view.Name, action))
				}
			}
		}
	}

	return permissions, nil
}