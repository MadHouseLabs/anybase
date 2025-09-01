package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/madhouselabs/anybase/internal/collection"
	"github.com/madhouselabs/anybase/pkg/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// VectorHandler handles vector field operations
type VectorHandler struct {
	collectionService collection.Service
}

// NewVectorHandler creates a new vector handler
func NewVectorHandler(collectionService collection.Service) *VectorHandler {
	return &VectorHandler{
		collectionService: collectionService,
	}
}

// AddVectorField adds a vector field to a collection
func (h *VectorHandler) AddVectorField(c *gin.Context) {
	collectionName := c.Param("name")
	
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Parse request body
	var field models.VectorField
	if err := c.ShouldBindJSON(&field); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate field
	if field.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "field name is required"})
		return
	}
	if field.Dimensions <= 0 || field.Dimensions > 65536 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dimensions must be between 1 and 65536"})
		return
	}

	// Set defaults
	if field.Metric == "" {
		field.Metric = "cosine"
	}
	if field.IndexType == "" {
		field.IndexType = "ivfflat"
	}

	// Add vector field
	if err := h.collectionService.AddVectorField(c.Request.Context(), userObjID, collectionName, field); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "vector field added successfully",
		"field":   field,
	})
}

// RemoveVectorField removes a vector field from a collection
func (h *VectorHandler) RemoveVectorField(c *gin.Context) {
	collectionName := c.Param("name")
	fieldName := c.Param("field")
	
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Remove vector field
	if err := h.collectionService.RemoveVectorField(c.Request.Context(), userObjID, collectionName, fieldName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "vector field removed successfully",
	})
}

// ListVectorFields lists all vector fields for a collection
func (h *VectorHandler) ListVectorFields(c *gin.Context) {
	collectionName := c.Param("name")
	
	// List vector fields
	fields, err := h.collectionService.ListVectorFields(c.Request.Context(), collectionName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"fields": fields,
	})
}

// VectorSearch performs vector similarity search
func (h *VectorHandler) VectorSearch(c *gin.Context) {
	collectionName := c.Param("collection")
	
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Parse request body
	var opts collection.VectorSearchOptions
	if err := c.ShouldBindJSON(&opts); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate options
	if opts.VectorField == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "vector_field is required"})
		return
	}
	if len(opts.QueryVector) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query_vector is required"})
		return
	}
	if opts.TopK <= 0 {
		opts.TopK = 10 // default
	}
	if opts.Metric == "" {
		opts.Metric = "cosine" // default
	}

	// Perform vector search
	results, err := h.collectionService.VectorSearch(c.Request.Context(), userObjID, collectionName, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"count":   len(results),
	})
}

// HybridSearch performs combined text and vector search
func (h *VectorHandler) HybridSearch(c *gin.Context) {
	collectionName := c.Param("collection")
	
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Parse request body
	var opts collection.HybridSearchOptions
	if err := c.ShouldBindJSON(&opts); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate options
	if opts.TextQuery == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "text_query is required"})
		return
	}
	if opts.VectorField == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "vector_field is required"})
		return
	}
	if len(opts.QueryVector) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query_vector is required"})
		return
	}
	if opts.TopK <= 0 {
		opts.TopK = 10 // default
	}
	if opts.Alpha < 0 || opts.Alpha > 1 {
		opts.Alpha = 0.5 // default to equal weight
	}

	// Perform hybrid search
	results, err := h.collectionService.HybridSearch(c.Request.Context(), userObjID, collectionName, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"count":   len(results),
	})
}