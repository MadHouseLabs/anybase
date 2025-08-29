package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/madhouselabs/anybase/internal/auth"
	"github.com/madhouselabs/anybase/internal/settings"
	"github.com/madhouselabs/anybase/pkg/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SettingsHandler struct {
	settingsService settings.Service
}

func NewSettingsHandler(settingsService settings.Service) *SettingsHandler {
	return &SettingsHandler{
		settingsService: settingsService,
	}
}

// GetUserSettings retrieves user-specific settings
func (h *SettingsHandler) GetUserSettings(c *gin.Context) {
	userClaims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	claims := userClaims.(*auth.Claims)
	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	settings, err := h.settingsService.GetUserSettings(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user settings"})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// UpdateUserSettings updates user-specific settings
func (h *SettingsHandler) UpdateUserSettings(c *gin.Context) {
	userClaims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	claims := userClaims.(*auth.Claims)
	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var settings models.Settings
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err = h.settingsService.UpdateUserSettings(c.Request.Context(), userID, &settings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Settings updated successfully"})
}

// GetSystemSettings retrieves system-wide settings
func (h *SettingsHandler) GetSystemSettings(c *gin.Context) {
	settings, err := h.settingsService.GetSystemSettings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get system settings"})
		return
	}

	// Remove sensitive fields for non-admin users
	userClaims, exists := c.Get("user")
	if exists {
		claims := userClaims.(*auth.Claims)
		isAdmin := false
		for _, role := range claims.Roles {
			if role == "admin" {
				isAdmin = true
				break
			}
		}
		if !isAdmin {
			// Return read-only view for non-admins
			c.JSON(http.StatusOK, gin.H{
				"connection_pool_size": settings.ConnectionPoolSize,
				"query_timeout":        settings.QueryTimeout,
				"max_retries":          settings.MaxRetries,
				"compression_enabled":  settings.CompressionEnabled,
				"encryption_enabled":   settings.EncryptionEnabled,
				"cors_enabled":         settings.CORSEnabled,
				"rate_limit":           settings.RateLimit,
			})
			return
		}
	}

	c.JSON(http.StatusOK, settings)
}

// UpdateSystemSettings updates system-wide settings (admin only)
func (h *SettingsHandler) UpdateSystemSettings(c *gin.Context) {
	userClaims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	claims := userClaims.(*auth.Claims)
	
	// Check if user is admin
	isAdmin := false
	for _, role := range claims.Roles {
		if role == "admin" {
			isAdmin = true
			break
		}
	}
	if !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only administrators can update system settings"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var settings models.SystemSettings
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err = h.settingsService.UpdateSystemSettings(c.Request.Context(), userID, &settings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update system settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "System settings updated successfully"})
}