package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/madhouselabs/anybase/internal/ai"
	"github.com/madhouselabs/anybase/pkg/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AIHandler handles AI provider and RAG-related requests
type AIHandler struct {
	aiService ai.Service
}

// NewAIHandler creates a new AI handler
func NewAIHandler(aiService ai.Service) *AIHandler {
	return &AIHandler{
		aiService: aiService,
	}
}

// Provider Management Endpoints

// CreateProvider creates a new AI provider
func (h *AIHandler) CreateProvider(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	
	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	
	var provider models.AIProvider
	if err := c.ShouldBindJSON(&provider); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := h.aiService.CreateProvider(c.Request.Context(), userObjID, &provider); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// Clear sensitive data before returning
	provider.APIKey = ""
	provider.APIKeyHash = ""
	
	c.JSON(http.StatusCreated, provider)
}

// ListProviders lists all AI providers
func (h *AIHandler) ListProviders(c *gin.Context) {
	providers, err := h.aiService.ListProviders(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"providers": providers})
}

// GetProvider gets a specific AI provider
func (h *AIHandler) GetProvider(c *gin.Context) {
	providerID := c.Param("id")
	
	objID, err := primitive.ObjectIDFromHex(providerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid provider ID"})
		return
	}
	
	provider, err := h.aiService.GetProvider(c.Request.Context(), objID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
		return
	}
	
	// Clear sensitive data
	provider.APIKeyHash = ""
	
	c.JSON(http.StatusOK, provider)
}

// UpdateProvider updates an AI provider
func (h *AIHandler) UpdateProvider(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	
	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	
	providerID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(providerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid provider ID"})
		return
	}
	
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := h.aiService.UpdateProvider(c.Request.Context(), userObjID, objID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "provider updated successfully"})
}

// DeleteProvider deletes an AI provider
func (h *AIHandler) DeleteProvider(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	
	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	
	providerID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(providerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid provider ID"})
		return
	}
	
	if err := h.aiService.DeleteProvider(c.Request.Context(), userObjID, objID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "provider deleted successfully"})
}

// RAG Configuration Endpoints

// ConfigureRAG configures RAG for a collection
func (h *AIHandler) ConfigureRAG(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	
	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	
	collectionName := c.Param("collection")
	
	var config models.CollectionRAGConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	config.CollectionName = collectionName
	
	if err := h.aiService.ConfigureRAG(c.Request.Context(), userObjID, &config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, config)
}

// GetRAGConfig gets RAG configuration for a collection
func (h *AIHandler) GetRAGConfig(c *gin.Context) {
	collectionName := c.Param("collection")
	
	config, err := h.aiService.GetRAGConfig(c.Request.Context(), collectionName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "RAG configuration not found"})
		return
	}
	
	c.JSON(http.StatusOK, config)
}

// UpdateRAGConfig updates RAG configuration
func (h *AIHandler) UpdateRAGConfig(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	
	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	
	collectionName := c.Param("collection")
	
	// Get existing config to find its ID
	config, err := h.aiService.GetRAGConfig(c.Request.Context(), collectionName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "RAG configuration not found"})
		return
	}
	
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := h.aiService.UpdateRAGConfig(c.Request.Context(), userObjID, config.ID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "RAG configuration updated successfully"})
}

// DeleteRAGConfig deletes RAG configuration
func (h *AIHandler) DeleteRAGConfig(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	
	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	
	collectionName := c.Param("collection")
	
	// Get existing config to find its ID
	config, err := h.aiService.GetRAGConfig(c.Request.Context(), collectionName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "RAG configuration not found"})
		return
	}
	
	if err := h.aiService.DeleteRAGConfig(c.Request.Context(), userObjID, config.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "RAG configuration deleted successfully"})
}

// ListRAGConfigs lists all RAG configurations
func (h *AIHandler) ListRAGConfigs(c *gin.Context) {
	configs, err := h.aiService.ListRAGConfigs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"configs": configs})
}

// Embedding Job Endpoints

// GenerateEmbeddings creates an embedding job for specific documents
func (h *AIHandler) GenerateEmbeddings(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	
	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	
	collectionName := c.Param("collection")
	
	var request struct {
		DocumentIDs []string `json:"document_ids"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	job, err := h.aiService.GenerateEmbeddings(c.Request.Context(), userObjID, collectionName, request.DocumentIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, job)
}

// GenerateAllEmbeddings creates an embedding job for all documents in a collection
func (h *AIHandler) GenerateAllEmbeddings(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	
	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	
	collectionName := c.Param("collection")
	
	job, err := h.aiService.GenerateAllEmbeddings(c.Request.Context(), userObjID, collectionName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, job)
}

// GetEmbeddingJob gets the status of an embedding job
func (h *AIHandler) GetEmbeddingJob(c *gin.Context) {
	jobID := c.Param("jobId")
	
	objID, err := primitive.ObjectIDFromHex(jobID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job ID"})
		return
	}
	
	job, err := h.aiService.GetEmbeddingJob(c.Request.Context(), objID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}
	
	c.JSON(http.StatusOK, job)
}

// ListEmbeddingJobs lists embedding jobs for a collection
func (h *AIHandler) ListEmbeddingJobs(c *gin.Context) {
	collectionName := c.Query("collection")
	
	jobs, err := h.aiService.ListEmbeddingJobs(c.Request.Context(), collectionName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"jobs": jobs})
}

// CancelEmbeddingJob cancels an embedding job
func (h *AIHandler) CancelEmbeddingJob(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	
	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	
	jobID := c.Param("jobId")
	objID, err := primitive.ObjectIDFromHex(jobID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job ID"})
		return
	}
	
	if err := h.aiService.CancelEmbeddingJob(c.Request.Context(), userObjID, objID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "job cancelled successfully"})
}

// RAG Query Endpoints

// QueryRAG performs a RAG query on a collection
func (h *AIHandler) QueryRAG(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	
	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	
	collectionName := c.Param("collection")
	
	var query models.RAGQuery
	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	query.CollectionName = collectionName
	
	response, err := h.aiService.QueryRAG(c.Request.Context(), userObjID, &query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, response)
}

// GetProviderModels fetches available models from a specific AI provider
func (h *AIHandler) GetProviderModels(c *gin.Context) {
	providerID := c.Param("id")
	
	objID, err := primitive.ObjectIDFromHex(providerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid provider ID"})
		return
	}
	
	models, err := h.aiService.GetProviderModels(c.Request.Context(), objID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"models": models})
}