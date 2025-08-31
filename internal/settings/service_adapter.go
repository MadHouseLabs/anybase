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
	col := s.db.Collection("settings")
	
	filter := map[string]interface{}{
		"data": map[string]interface{}{
			"$elemMatch": map[string]interface{}{
				"type": "user",
				"user_id": userID.Hex(),
			},
		},
	}
	
	// For PostgreSQL, we need to use JSONB query
	// Since settings are stored in data field, we'll use a simpler approach
	filter = map[string]interface{}{
		"data.type": "user",
		"data.user_id": userID.Hex(),
	}
	
	var result map[string]interface{}
	if err := col.FindOne(ctx, filter, &result); err != nil {
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
	
	// Extract settings from result
	settings := &models.Settings{
		UserID:    userID,
		Theme:     "light",
		Language:  "en",
		Timezone:  "UTC",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	if data, ok := result["data"].(map[string]interface{}); ok {
		if theme, ok := data["theme"].(string); ok {
			settings.Theme = theme
		}
		if lang, ok := data["language"].(string); ok {
			settings.Language = lang
		}
		if tz, ok := data["timezone"].(string); ok {
			settings.Timezone = tz
		}
		if df, ok := data["date_format"].(string); ok {
			settings.DateFormat = df
		}
		if tf, ok := data["time_format"].(string); ok {
			settings.TimeFormat = tf
		}
		if en, ok := data["email_notifications"].(bool); ok {
			settings.EmailNotifications = en
		}
		if sa, ok := data["security_alerts"].(bool); ok {
			settings.SecurityAlerts = sa
		}
	}
	
	return settings, nil
}

// UpdateUserSettings updates user-specific settings
func (s *AdapterService) UpdateUserSettings(ctx context.Context, userID primitive.ObjectID, settings *models.Settings) error {
	col := s.db.Collection("settings")
	
	settings.UserID = userID
	settings.UpdatedAt = time.Now()
	
	filter := map[string]interface{}{
		"data.type": "user",
		"data.user_id": userID.Hex(),
	}
	
	// Convert settings to map for update
	updateData := map[string]interface{}{
		"type":                "user",
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
		"$set": map[string]interface{}{
			"data": updateData,
		},
	}
	
	result, err := col.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return fmt.Errorf("failed to update user settings: %w", err)
	}
	
	// If no document was matched, insert a new one
	if result.MatchedCount == 0 {
		// Create new settings document with ID
		updateData["created_at"] = time.Now()
		
		newDoc := map[string]interface{}{
			"_id":  primitive.NewObjectID().Hex(),
			"data": updateData,
		}
		
		if _, err := col.InsertOne(ctx, newDoc); err != nil {
			return fmt.Errorf("failed to create user settings: %w", err)
		}
	}
	
	return nil
}

// GetSystemSettings retrieves system-wide settings
func (s *AdapterService) GetSystemSettings(ctx context.Context) (*models.SystemSettings, error) {
	col := s.db.Collection("settings")
	
	// System settings should have type: "system"
	filter := map[string]interface{}{
		"data.type": "system",
	}
	
	var result map[string]interface{}
	if err := col.FindOne(ctx, filter, &result); err != nil {
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
	
	// Extract settings from result
	settings := &models.SystemSettings{
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
	}
	
	if data, ok := result["data"].(map[string]interface{}); ok {
		if cps, ok := data["connection_pool_size"].(float64); ok {
			settings.ConnectionPoolSize = int(cps)
		}
		if qt, ok := data["query_timeout"].(float64); ok {
			settings.QueryTimeout = int(qt)
		}
		if mr, ok := data["max_retries"].(float64); ok {
			settings.MaxRetries = int(mr)
		}
		if ce, ok := data["compression_enabled"].(bool); ok {
			settings.CompressionEnabled = ce
		}
		if ee, ok := data["encryption_enabled"].(bool); ok {
			settings.EncryptionEnabled = ee
		}
		if st, ok := data["session_timeout"].(float64); ok {
			settings.SessionTimeout = int(st)
		}
		if pp, ok := data["password_policy"].(string); ok {
			settings.PasswordPolicy = pp
		}
		if mfa, ok := data["mfa_required"].(bool); ok {
			settings.MFARequired = mfa
		}
		if ale, ok := data["audit_log_enabled"].(bool); ok {
			settings.AuditLogEnabled = ale
		}
		if rl, ok := data["rate_limit"].(float64); ok {
			settings.RateLimit = int(rl)
		}
		if bl, ok := data["burst_limit"].(float64); ok {
			settings.BurstLimit = int(bl)
		}
		if cors, ok := data["cors_enabled"].(bool); ok {
			settings.CORSEnabled = cors
		}
	}
	
	return settings, nil
}

// UpdateSystemSettings updates system-wide settings
func (s *AdapterService) UpdateSystemSettings(ctx context.Context, userID primitive.ObjectID, settings *models.SystemSettings) error {
	col := s.db.Collection("settings")
	
	settings.UpdatedAt = time.Now()
	settings.UpdatedBy = userID
	
	filter := map[string]interface{}{
		"data.type": "system",
	}
	
	// Convert settings to map for update
	updateData := map[string]interface{}{
		"type":                 "system",
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
		"$set": map[string]interface{}{
			"data": updateData,
		},
	}
	
	result, err := col.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return fmt.Errorf("failed to update system settings: %w", err)
	}
	
	// If no document was matched, insert a new one
	if result.MatchedCount == 0 {
		newDoc := map[string]interface{}{
			"_id":  primitive.NewObjectID().Hex(),
			"data": updateData,
		}
		
		if _, err := col.InsertOne(ctx, newDoc); err != nil {
			return fmt.Errorf("failed to create system settings: %w", err)
		}
	}
	
	return nil
}

