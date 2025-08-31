package settings

import (
	"context"
	"fmt"
	"time"

	"github.com/madhouselabs/anybase/internal/database/types"
	"github.com/madhouselabs/anybase/pkg/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AdapterService implements the settings service using the database adapter
type AdapterService struct {
	db types.DB
}

// NewAdapterService creates a new settings service using the database adapter
func NewAdapterService(db types.DB) Service {
	return &AdapterService{
		db: db,
	}
}

// GetUserSettings retrieves user-specific settings
func (s *AdapterService) GetUserSettings(ctx context.Context, userID primitive.ObjectID) (*models.Settings, error) {
	col := s.db.Collection("user_settings")
	
	filter := map[string]interface{}{
		"user_id": userID.Hex(),
	}
	
	var settings models.Settings
	if err := col.FindOne(ctx, filter, &settings); err != nil {
		if err == types.ErrNoDocuments {
			// Return default settings if none exist
			return &models.Settings{
				UserID:    userID,
				Theme:     "light",
				Language:  "en",
				Timezone:  "UTC",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		}
		return nil, fmt.Errorf("failed to get user settings: %w", err)
	}
	
	return &settings, nil
}

// UpdateUserSettings updates user-specific settings
func (s *AdapterService) UpdateUserSettings(ctx context.Context, userID primitive.ObjectID, settings *models.Settings) error {
	col := s.db.Collection("user_settings")
	
	settings.UserID = userID
	settings.UpdatedAt = time.Now()
	
	filter := map[string]interface{}{
		"user_id": userID.Hex(),
	}
	
	// Convert settings to map for update
	updateData := map[string]interface{}{
		"user_id":             userID.Hex(),
		"theme":               settings.Theme,
		"language":            settings.Language,
		"timezone":            settings.Timezone,
		"date_format":         settings.DateFormat,
		"time_format":         settings.TimeFormat,
		"email_notifications": settings.EmailNotifications,
		"security_alerts":     settings.SecurityAlerts,
		"updated_at":          settings.UpdatedAt,
	}
	
	// Try to update first
	updateDoc := map[string]interface{}{
		"$set": updateData,
	}
	
	result, err := col.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return fmt.Errorf("failed to update user settings: %w", err)
	}
	
	// If no document was matched, insert a new one
	if result.MatchedCount == 0 {
		// Create new settings document with ID
		updateData["_id"] = primitive.NewObjectID().Hex()
		updateData["created_at"] = time.Now()
		
		if _, err := col.InsertOne(ctx, updateData); err != nil {
			return fmt.Errorf("failed to create user settings: %w", err)
		}
	}
	
	return nil
}

// GetSystemSettings retrieves system-wide settings
func (s *AdapterService) GetSystemSettings(ctx context.Context) (*models.SystemSettings, error) {
	col := s.db.Collection("system_settings")
	
	// System settings should have a fixed ID
	filter := map[string]interface{}{
		"_id": "system",
	}
	
	var settings models.SystemSettings
	if err := col.FindOne(ctx, filter, &settings); err != nil {
		if err == types.ErrNoDocuments {
			// Return default system settings if none exist
			return &models.SystemSettings{
				ID:                primitive.NewObjectID(),
				ConnectionPoolSize: 100,
				QueryTimeout:      30,
				MaxRetries:        3,
				SessionTimeout:    24,
				PasswordPolicy:    "moderate",
				RateLimit:         100,
				BurstLimit:        200,
				CORSEnabled:       true,
				UpdatedAt:         time.Now(),
			}, nil
		}
		return nil, fmt.Errorf("failed to get system settings: %w", err)
	}
	
	return &settings, nil
}

// UpdateSystemSettings updates system-wide settings
func (s *AdapterService) UpdateSystemSettings(ctx context.Context, userID primitive.ObjectID, settings *models.SystemSettings) error {
	col := s.db.Collection("system_settings")
	
	settings.UpdatedAt = time.Now()
	settings.UpdatedBy = userID
	
	filter := map[string]interface{}{
		"_id": "system",
	}
	
	// Convert settings to map for update
	updateData := map[string]interface{}{
		"connection_pool_size": settings.ConnectionPoolSize,
		"query_timeout":        settings.QueryTimeout,
		"max_retries":          settings.MaxRetries,
		"compression_enabled":  settings.CompressionEnabled,
		"encryption_enabled":   settings.EncryptionEnabled,
		"session_timeout":      settings.SessionTimeout,
		"password_policy":      settings.PasswordPolicy,
		"mfa_required":         settings.MFARequired,
		"audit_log_enabled":    settings.AuditLogEnabled,
		"rate_limit":           settings.RateLimit,
		"burst_limit":          settings.BurstLimit,
		"cors_enabled":         settings.CORSEnabled,
		"updated_at":           settings.UpdatedAt,
		"updated_by":           userID.Hex(),
	}
	
	// Try to update first
	updateDoc := map[string]interface{}{
		"$set": updateData,
	}
	
	result, err := col.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return fmt.Errorf("failed to update system settings: %w", err)
	}
	
	// If no document was matched, insert a new one
	if result.MatchedCount == 0 {
		// Add _id for new document
		updateData["_id"] = "system"
		
		if _, err := col.InsertOne(ctx, updateData); err != nil {
			return fmt.Errorf("failed to create system settings: %w", err)
		}
	}
	
	return nil
}

