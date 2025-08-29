package v1

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/madhouselabs/anybase/internal/accesskey"
	"github.com/madhouselabs/anybase/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AccessKeyHandler struct {
	repo accesskey.Repository
}

func NewAccessKeyHandler(repo accesskey.Repository) *AccessKeyHandler {
	return &AccessKeyHandler{repo: repo}
}

// CreateAccessKeyRequest represents the request to create an access key
type CreateAccessKeyRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions" binding:"required"`
	ExpiresIn   int      `json:"expires_in,omitempty"` // Hours until expiration, 0 = never expires
}

// AccessKeyResponse represents the response for access key operations
type AccessKeyResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Key         string    `json:"key,omitempty"` // Only included on creation
	Permissions []string  `json:"permissions"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	LastUsed    *time.Time `json:"last_used,omitempty"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateAccessKey creates a new access key
// @Summary Create access key
// @Description Create a new access key with specific permissions
// @Tags Access Keys
// @Accept json
// @Produce json
// @Param request body CreateAccessKeyRequest true "Access key details"
// @Success 201 {object} AccessKeyResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security Bearer
// @Router /api/v1/admin/access-keys [post]
func (h *AccessKeyHandler) CreateAccessKey(c *gin.Context) {
	var req CreateAccessKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current user ID from context
	userID, _ := c.Get("user_id")
	creatorID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	// Create access key model
	ak := &models.AccessKey{
		Name:        req.Name,
		Description: req.Description,
		Permissions: req.Permissions,
		CreatedBy:   creatorID,
	}

	// Set expiration if provided
	if req.ExpiresIn > 0 {
		expiresAt := time.Now().Add(time.Duration(req.ExpiresIn) * time.Hour)
		ak.ExpiresAt = &expiresAt
	}

	// Create the access key
	key, err := h.repo.Create(c.Request.Context(), ak)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return response with the generated key
	response := &AccessKeyResponse{
		ID:          ak.ID.Hex(),
		Name:        ak.Name,
		Description: ak.Description,
		Key:         key, // Include the key only on creation
		Permissions: ak.Permissions,
		ExpiresAt:   ak.ExpiresAt,
		Active:      ak.Active,
		CreatedAt:   ak.CreatedAt,
		UpdatedAt:   ak.UpdatedAt,
	}

	c.JSON(http.StatusCreated, response)
}

// ListAccessKeys lists all access keys for the current user
// @Summary List access keys
// @Description Get a list of all access keys created by the current user
// @Tags Access Keys
// @Produce json
// @Success 200 {array} AccessKeyResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security Bearer
// @Router /api/v1/admin/access-keys [get]
func (h *AccessKeyHandler) ListAccessKeys(c *gin.Context) {
	// Get current user ID from context
	userID, _ := c.Get("user_id")
	creatorID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	// List access keys created by this user
	filter := bson.M{"created_by": creatorID}
	opts := options.Find().SetSort(bson.M{"created_at": -1})
	
	keys, err := h.repo.List(c.Request.Context(), filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to response format
	responses := make([]*AccessKeyResponse, len(keys))
	for i, key := range keys {
		responses[i] = &AccessKeyResponse{
			ID:          key.ID.Hex(),
			Name:        key.Name,
			Description: key.Description,
			Permissions: key.Permissions,
			ExpiresAt:   key.ExpiresAt,
			LastUsed:    key.LastUsed,
			Active:      key.Active,
			CreatedAt:   key.CreatedAt,
			UpdatedAt:   key.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{"access_keys": responses})
}

// GetAccessKey gets a specific access key
// @Summary Get access key
// @Description Get details of a specific access key
// @Tags Access Keys
// @Produce json
// @Param id path string true "Access key ID"
// @Success 200 {object} AccessKeyResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security Bearer
// @Router /api/v1/admin/access-keys/{id} [get]
func (h *AccessKeyHandler) GetAccessKey(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid access key id"})
		return
	}

	key, err := h.repo.GetByID(c.Request.Context(), objID)
	if err != nil {
		if err == accesskey.ErrAccessKeyNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "access key not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Check if the current user created this key
	userID, _ := c.Get("user_id")
	creatorID, _ := primitive.ObjectIDFromHex(userID.(string))
	if key.CreatedBy != creatorID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	response := &AccessKeyResponse{
		ID:          key.ID.Hex(),
		Name:        key.Name,
		Description: key.Description,
		Permissions: key.Permissions,
		ExpiresAt:   key.ExpiresAt,
		LastUsed:    key.LastUsed,
		Active:      key.Active,
		CreatedAt:   key.CreatedAt,
		UpdatedAt:   key.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateAccessKey updates an access key
// @Summary Update access key
// @Description Update an access key's details (cannot change permissions)
// @Tags Access Keys
// @Accept json
// @Produce json
// @Param id path string true "Access key ID"
// @Param request body map[string]interface{} true "Updates"
// @Success 200 {object} AccessKeyResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security Bearer
// @Router /api/v1/admin/access-keys/{id} [put]
func (h *AccessKeyHandler) UpdateAccessKey(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid access key id"})
		return
	}

	// Get the access key first to check ownership
	key, err := h.repo.GetByID(c.Request.Context(), objID)
	if err != nil {
		if err == accesskey.ErrAccessKeyNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "access key not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Check if the current user created this key
	userID, _ := c.Get("user_id")
	creatorID, _ := primitive.ObjectIDFromHex(userID.(string))
	if key.CreatedBy != creatorID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build update document (only allow certain fields to be updated)
	updateDoc := bson.M{}
	if name, ok := updates["name"].(string); ok {
		updateDoc["name"] = name
	}
	if description, ok := updates["description"].(string); ok {
		updateDoc["description"] = description
	}
	if active, ok := updates["active"].(bool); ok {
		updateDoc["active"] = active
	}

	if len(updateDoc) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no valid updates provided"})
		return
	}

	// Update the access key
	err = h.repo.Update(c.Request.Context(), objID, bson.M{"$set": updateDoc})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "access key updated successfully"})
}

// RegenerateAccessKey regenerates an access key
// @Summary Regenerate access key
// @Description Generate a new key for an existing access key
// @Tags Access Keys
// @Produce json
// @Param id path string true "Access key ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security Bearer
// @Router /api/v1/admin/access-keys/{id}/regenerate [post]
func (h *AccessKeyHandler) RegenerateAccessKey(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid access key id"})
		return
	}

	// Get the access key first to check ownership
	key, err := h.repo.GetByID(c.Request.Context(), objID)
	if err != nil {
		if err == accesskey.ErrAccessKeyNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "access key not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Check if the current user created this key
	userID, _ := c.Get("user_id")
	creatorID, _ := primitive.ObjectIDFromHex(userID.(string))
	if key.CreatedBy != creatorID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Regenerate the key
	newKey, err := h.repo.RegenerateKey(c.Request.Context(), objID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "access key regenerated successfully",
		"key":     newKey,
	})
}

// DeleteAccessKey deletes an access key
// @Summary Delete access key
// @Description Delete an access key permanently
// @Tags Access Keys
// @Param id path string true "Access key ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security Bearer
// @Router /api/v1/admin/access-keys/{id} [delete]
func (h *AccessKeyHandler) DeleteAccessKey(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid access key id"})
		return
	}

	// Get the access key first to check ownership
	key, err := h.repo.GetByID(c.Request.Context(), objID)
	if err != nil {
		if err == accesskey.ErrAccessKeyNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "access key not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Check if the current user created this key
	userID, _ := c.Get("user_id")
	creatorID, _ := primitive.ObjectIDFromHex(userID.(string))
	if key.CreatedBy != creatorID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Delete the access key
	err = h.repo.Delete(c.Request.Context(), objID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "access key deleted successfully"})
}