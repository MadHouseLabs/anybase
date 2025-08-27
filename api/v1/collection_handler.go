package v1

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/karthik/anybase/internal/collection"
	"github.com/karthik/anybase/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CollectionHandler struct {
	collectionService collection.Service
}

func NewCollectionHandler(collectionService collection.Service) *CollectionHandler {
	return &CollectionHandler{
		collectionService: collectionService,
	}
}

// CreateCollection creates a new collection
func (h *CollectionHandler) CreateCollection(c *gin.Context) {
	userID := getUserID(c)
	if userID == primitive.NilObjectID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var col models.Collection
	if err := c.ShouldBindJSON(&col); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.collectionService.CreateCollection(c.Request.Context(), userID, &col); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, col)
}

// GetCollection retrieves a collection by name
func (h *CollectionHandler) GetCollection(c *gin.Context) {
	userID := getUserID(c)
	name := c.Param("name")

	col, err := h.collectionService.GetCollection(c.Request.Context(), userID, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, col)
}

// UpdateCollection updates a collection
func (h *CollectionHandler) UpdateCollection(c *gin.Context) {
	userID := getUserID(c)
	name := c.Param("name")

	var updates bson.M
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.collectionService.UpdateCollection(c.Request.Context(), userID, name, updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "collection updated successfully"})
}

// DeleteCollection deletes a collection
func (h *CollectionHandler) DeleteCollection(c *gin.Context) {
	userID := getUserID(c)
	name := c.Param("name")

	if err := h.collectionService.DeleteCollection(c.Request.Context(), userID, name); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "collection deleted successfully"})
}

// ListCollections lists all accessible collections
func (h *CollectionHandler) ListCollections(c *gin.Context) {
	// Check if authenticated via access key
	authType, _ := c.Get("auth_type")
	if authType == "access_key" {
		// For access keys, check permissions and return all collections they have access to
		permissions, _ := c.Get("permissions")
		perms := permissions.([]string)
		
		// Check if has collection read permission
		hasPermission := false
		for _, perm := range perms {
			if perm == "collection:*:read" || perm == "*:*:*" {
				hasPermission = true
				break
			}
			// Check for specific collection permissions
			if strings.HasPrefix(perm, "collection:") && strings.HasSuffix(perm, ":read") {
				hasPermission = true
				break
			}
		}
		
		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to list collections"})
			return
		}
		
		// For now, return all collections (in production, filter by specific permissions)
		collections, err := h.collectionService.ListAllCollections(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{"collections": collections})
		return
	}
	
	// JWT auth - use userID
	userID := getUserID(c)
	if userID == primitive.NilObjectID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	collections, err := h.collectionService.ListCollections(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"collections": collections})
}

// CreateView creates a new view
func (h *CollectionHandler) CreateView(c *gin.Context) {
	userID := getUserID(c)

	var view models.View
	if err := c.ShouldBindJSON(&view); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.collectionService.CreateView(c.Request.Context(), userID, &view); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, view)
}

// GetView retrieves a view by name
func (h *CollectionHandler) GetView(c *gin.Context) {
	userID := getUserID(c)
	name := c.Param("name")

	view, err := h.collectionService.GetView(c.Request.Context(), userID, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, view)
}

// QueryView executes a view query
func (h *CollectionHandler) QueryView(c *gin.Context) {
	userID := getUserID(c)
	name := c.Param("name")

	// Parse query options
	opts := collection.QueryOptions{}
	if limit := c.Query("limit"); limit != "" {
		// Simple conversion - you might want to use strconv.Atoi
		opts.Limit = 10 // Default limit
	}
	if skip := c.Query("skip"); skip != "" {
		opts.Skip = 0 // Default skip
	}

	results, err := h.collectionService.QueryView(c.Request.Context(), userID, name, opts)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": results})
}

// ListViews lists all accessible views
func (h *CollectionHandler) ListViews(c *gin.Context) {
	userID := getUserID(c)

	views, err := h.collectionService.ListViews(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"views": views})
}

// InsertDocument inserts a new document into a collection
func (h *CollectionHandler) InsertDocument(c *gin.Context) {
	collectionName := c.Param("collection")

	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if authenticated via access key
	authType, _ := c.Get("auth_type")
	if authType == "access_key" {
		// For access keys, check permissions
		permissions, _ := c.Get("permissions")
		perms := permissions.([]string)
		
		// Check if has permission to write to this specific collection
		hasPermission := false
		requiredPerm := "collection:" + collectionName + ":write"
		for _, perm := range perms {
			if perm == requiredPerm || perm == "collection:*:write" || perm == "*:*:*" {
				hasPermission = true
				break
			}
		}
		
		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to write to collection"})
			return
		}
		
		// Use special context to bypass user checks
		ctx := context.WithValue(c.Request.Context(), "access_key_validated", true)
		
		mutation := &models.DataMutation{
			Collection: collectionName,
			Operation:  "insert",
			Data:       data,
			UserID:     primitive.NilObjectID, // No user for access keys
			UserRoles:  []string{},
		}
		
		doc, err := h.collectionService.InsertDocument(ctx, mutation)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusCreated, doc)
		return
	}

	// JWT auth - use userID
	userID := getUserID(c)
	if userID == primitive.NilObjectID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	mutation := &models.DataMutation{
		Collection: collectionName,
		Operation:  "insert",
		Data:       data,
		UserID:     userID,
		UserRoles:  getUserRoles(c),
	}

	doc, err := h.collectionService.InsertDocument(c.Request.Context(), mutation)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, doc)
}

// GetDocument retrieves a single document
func (h *CollectionHandler) GetDocument(c *gin.Context) {
	collectionName := c.Param("collection")
	docID := c.Param("id")

	objectID, err := primitive.ObjectIDFromHex(docID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document ID"})
		return
	}

	// Check if authenticated via access key
	authType, _ := c.Get("auth_type")
	if authType == "access_key" {
		// For access keys, check permissions
		permissions, _ := c.Get("permissions")
		perms := permissions.([]string)
		
		// Check if has permission to read this specific collection
		hasPermission := false
		requiredPerm := "collection:" + collectionName + ":read"
		for _, perm := range perms {
			if perm == requiredPerm || perm == "collection:*:read" || perm == "*:*:*" {
				hasPermission = true
				break
			}
		}
		
		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to read collection"})
			return
		}
		
		// Use special context to bypass user checks
		ctx := context.WithValue(c.Request.Context(), "access_key_validated", true)
		doc, err := h.collectionService.GetDocument(ctx, primitive.NilObjectID, collectionName, objectID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, doc)
		return
	}

	// JWT auth - use userID
	userID := getUserID(c)
	if userID == primitive.NilObjectID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	doc, err := h.collectionService.GetDocument(c.Request.Context(), userID, collectionName, objectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, doc)
}

// UpdateDocument updates a document in a collection
func (h *CollectionHandler) UpdateDocument(c *gin.Context) {
	collectionName := c.Param("collection")
	docID := c.Param("id")

	objectID, err := primitive.ObjectIDFromHex(docID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document ID"})
		return
	}

	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if authenticated via access key
	authType, _ := c.Get("auth_type")
	if authType == "access_key" {
		// For access keys, check permissions
		permissions, _ := c.Get("permissions")
		perms := permissions.([]string)
		
		// Check if has permission to update this specific collection
		hasPermission := false
		requiredPerm := "collection:" + collectionName + ":update"
		for _, perm := range perms {
			if perm == requiredPerm || perm == "collection:*:update" || perm == "*:*:*" {
				hasPermission = true
				break
			}
		}
		
		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to update collection"})
			return
		}
		
		// Use special context to bypass user checks
		ctx := context.WithValue(c.Request.Context(), "access_key_validated", true)
		
		mutation := &models.DataMutation{
			Collection: collectionName,
			Operation:  "update",
			DocumentID: objectID,
			Data:       data,
			UserID:     primitive.NilObjectID,
			UserRoles:  []string{},
		}
		
		if err := h.collectionService.UpdateDocument(ctx, mutation); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{"message": "document updated successfully"})
		return
	}

	// JWT auth - use userID
	userID := getUserID(c)
	if userID == primitive.NilObjectID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	mutation := &models.DataMutation{
		Collection: collectionName,
		Operation:  "update",
		DocumentID: objectID,
		Data:       data,
		UserID:     userID,
		UserRoles:  getUserRoles(c),
	}

	if err := h.collectionService.UpdateDocument(c.Request.Context(), mutation); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "document updated successfully"})
}

// DeleteDocument deletes a document from a collection
func (h *CollectionHandler) DeleteDocument(c *gin.Context) {
	collectionName := c.Param("collection")
	docID := c.Param("id")

	objectID, err := primitive.ObjectIDFromHex(docID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document ID"})
		return
	}

	// Check if authenticated via access key
	authType, _ := c.Get("auth_type")
	if authType == "access_key" {
		// For access keys, check permissions
		permissions, _ := c.Get("permissions")
		perms := permissions.([]string)
		
		// Check if has permission to delete from this specific collection
		hasPermission := false
		requiredPerm := "collection:" + collectionName + ":delete"
		for _, perm := range perms {
			if perm == requiredPerm || perm == "collection:*:delete" || perm == "*:*:*" {
				hasPermission = true
				break
			}
		}
		
		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to delete from collection"})
			return
		}
		
		// Use special context to bypass user checks
		ctx := context.WithValue(c.Request.Context(), "access_key_validated", true)
		
		mutation := &models.DataMutation{
			Collection: collectionName,
			Operation:  "delete",
			DocumentID: objectID,
			UserID:     primitive.NilObjectID,
			UserRoles:  []string{},
		}
		
		if err := h.collectionService.DeleteDocument(ctx, mutation); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{"message": "document deleted successfully"})
		return
	}

	// JWT auth - use userID
	userID := getUserID(c)
	if userID == primitive.NilObjectID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	mutation := &models.DataMutation{
		Collection: collectionName,
		Operation:  "delete",
		DocumentID: objectID,
		UserID:     userID,
		UserRoles:  getUserRoles(c),
	}

	if err := h.collectionService.DeleteDocument(c.Request.Context(), mutation); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "document deleted successfully"})
}

// QueryDocuments queries documents in a collection
func (h *CollectionHandler) QueryDocuments(c *gin.Context) {
	collectionName := c.Param("collection")
	
	// Check if authenticated via access key
	authType, _ := c.Get("auth_type")
	if authType == "access_key" {
		// For access keys, check permissions
		permissions, _ := c.Get("permissions")
		perms := permissions.([]string)
		
		// Check if has permission to read this specific collection
		hasPermission := false
		requiredPerm := "collection:" + collectionName + ":read"
		for _, perm := range perms {
			if perm == requiredPerm || perm == "collection:*:read" || perm == "*:*:*" {
				hasPermission = true
				break
			}
		}
		
		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to read collection"})
			return
		}
		
		// For access keys, we bypass the service's CanRead check by using a special context
		// The service should check for this and allow access
		ctx := context.WithValue(c.Request.Context(), "access_key_validated", true)
		
		// Query without userID for access keys
		query := &models.DataQuery{
			Collection: collectionName,
			UserID:     primitive.NilObjectID, // No user for access keys
			UserRoles:  []string{},
		}
		
		// Parse filter from query string (JSON format)
		if filterStr := c.Query("filter"); filterStr != "" {
			var filter map[string]interface{}
			if err := json.Unmarshal([]byte(filterStr), &filter); err == nil {
				query.Filter = filter
			}
		}
		
		// Parse sort from query string (JSON format)
		// Example: ?sort={"name":1,"price":-1} for ascending name, descending price
		if sortStr := c.Query("sort"); sortStr != "" {
			var sort map[string]int
			if err := json.Unmarshal([]byte(sortStr), &sort); err == nil {
				query.Sort = sort
			}
		}
		
		// Parse query parameters
		if fields := c.QueryArray("fields"); len(fields) > 0 {
			query.Fields = fields
		}
		
		// Parse pagination parameters
		page := 1
		if p := c.Query("page"); p != "" {
			if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
				page = parsed
			}
		}
		
		limit := 20 // Default page size
		if l := c.Query("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
				limit = parsed
			}
		}
		
		query.Limit = limit
		query.Skip = (page - 1) * limit
		
		docs, err := h.collectionService.QueryDocuments(ctx, query)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		// Get total count for pagination (for access keys, pass nil userID)
		totalCount, err := h.collectionService.CountDocuments(ctx, primitive.NilObjectID, collectionName, query.Filter)
		if err != nil {
			// If count fails, just return the data without total
			c.JSON(http.StatusOK, gin.H{
				"data":  docs,
				"count": len(docs),
				"page":  page,
				"limit": limit,
			})
			return
		}
		
		totalPages := (totalCount + limit - 1) / limit
		
		c.JSON(http.StatusOK, gin.H{
			"data":       docs,
			"total":      totalCount,
			"page":       page,
			"limit":      limit,
			"totalPages": totalPages,
		})
		return
	}
	
	// JWT auth - use userID as before
	userID := getUserID(c)
	if userID == primitive.NilObjectID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	query := &models.DataQuery{
		Collection: collectionName,
		UserID:     userID,
		UserRoles:  getUserRoles(c),
	}

	// Parse filter from query string (JSON format)
	if filterStr := c.Query("filter"); filterStr != "" {
		var filter map[string]interface{}
		if err := json.Unmarshal([]byte(filterStr), &filter); err == nil {
			query.Filter = filter
		}
	}
	
	// Parse sort from query string (JSON format)
	// Example: ?sort={"name":1,"price":-1} for ascending name, descending price
	if sortStr := c.Query("sort"); sortStr != "" {
		var sort map[string]int
		if err := json.Unmarshal([]byte(sortStr), &sort); err == nil {
			query.Sort = sort
		}
	}

	// Parse query parameters
	if fields := c.QueryArray("fields"); len(fields) > 0 {
		query.Fields = fields
	}

	// Parse pagination parameters
	page := 1
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	
	limit := 20 // Default page size
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	
	query.Limit = limit
	query.Skip = (page - 1) * limit

	docs, err := h.collectionService.QueryDocuments(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Get total count for pagination
	totalCount, err := h.collectionService.CountDocuments(c.Request.Context(), userID, collectionName, query.Filter)
	if err != nil {
		// If count fails, just return the data without total
		c.JSON(http.StatusOK, gin.H{
			"data":  docs,
			"count": len(docs),
			"page":  page,
			"limit": limit,
		})
		return
	}
	
	totalPages := (totalCount + limit - 1) / limit

	c.JSON(http.StatusOK, gin.H{
		"data":       docs,
		"total":      totalCount,
		"page":       page,
		"limit":      limit,
		"totalPages": totalPages,
	})
}

// Helper functions
// ListIndexes lists all indexes for a collection
func (h *CollectionHandler) ListIndexes(c *gin.Context) {
	collectionName := c.Param("name")
	
	// Check if user has access to the collection
	userID := getUserID(c)
	if userID == primitive.NilObjectID && c.GetString("auth_type") != "access_key" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	
	indexes, err := h.collectionService.ListIndexes(c.Request.Context(), collectionName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"indexes": indexes})
}

// CreateIndex creates a new index on a collection
func (h *CollectionHandler) CreateIndex(c *gin.Context) {
	collectionName := c.Param("name")
	
	// Check if user has access to the collection
	userID := getUserID(c)
	if userID == primitive.NilObjectID && c.GetString("auth_type") != "access_key" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	
	var indexRequest struct {
		Name    string                 `json:"name"`
		Keys    map[string]interface{} `json:"keys"`
		Options map[string]interface{} `json:"options"`
	}
	
	if err := c.ShouldBindJSON(&indexRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	err := h.collectionService.CreateIndex(c.Request.Context(), collectionName, indexRequest.Name, indexRequest.Keys, indexRequest.Options)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{"message": "index created successfully"})
}

// DeleteIndex deletes an index from a collection
func (h *CollectionHandler) DeleteIndex(c *gin.Context) {
	collectionName := c.Param("name")
	indexName := c.Param("index")
	
	// Check if user has access to the collection
	userID := getUserID(c)
	if userID == primitive.NilObjectID && c.GetString("auth_type") != "access_key" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	
	err := h.collectionService.DeleteIndex(c.Request.Context(), collectionName, indexName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "index deleted successfully"})
}

func getUserID(c *gin.Context) primitive.ObjectID {
	userIDStr, exists := c.Get("userID")
	if !exists {
		return primitive.NilObjectID
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		return primitive.NilObjectID
	}

	return userID
}

func getUserRoles(c *gin.Context) []string {
	roles, exists := c.Get("roles")
	if !exists {
		return []string{}
	}

	if rolesSlice, ok := roles.([]string); ok {
		return rolesSlice
	}

	return []string{}
}