package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/madhouselabs/anybase/internal/collection"
	"github.com/madhouselabs/anybase/pkg/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Service defines the AI service interface
type Service interface {
	// Provider management
	CreateProvider(ctx context.Context, userID primitive.ObjectID, provider *models.AIProvider) error
	GetProvider(ctx context.Context, id primitive.ObjectID) (*models.AIProvider, error)
	ListProviders(ctx context.Context) ([]*models.AIProvider, error)
	UpdateProvider(ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID, updates map[string]interface{}) error
	DeleteProvider(ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID) error
	GetProviderModels(ctx context.Context, providerID primitive.ObjectID) ([]*models.AIModel, error)
	
	// RAG configuration
	ConfigureRAG(ctx context.Context, userID primitive.ObjectID, config *models.CollectionRAGConfig) error
	GetRAGConfig(ctx context.Context, collectionName string) (*models.CollectionRAGConfig, error)
	UpdateRAGConfig(ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID, updates map[string]interface{}) error
	DeleteRAGConfig(ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID) error
	ListRAGConfigs(ctx context.Context) ([]*models.CollectionRAGConfig, error)
	
	// Embedding operations
	GenerateEmbeddings(ctx context.Context, userID primitive.ObjectID, collectionName string, documentIDs []string) (*models.EmbeddingJob, error)
	GenerateAllEmbeddings(ctx context.Context, userID primitive.ObjectID, collectionName string) (*models.EmbeddingJob, error)
	GetEmbeddingJob(ctx context.Context, jobID primitive.ObjectID) (*models.EmbeddingJob, error)
	ListEmbeddingJobs(ctx context.Context, collectionName string) ([]*models.EmbeddingJob, error)
	CancelEmbeddingJob(ctx context.Context, userID primitive.ObjectID, jobID primitive.ObjectID) error
	
	// RAG queries
	QueryRAG(ctx context.Context, userID primitive.ObjectID, query *models.RAGQuery) (*models.RAGResponse, error)
	
	// Document hooks for auto-embedding
	OnDocumentCreated(ctx context.Context, collectionName string, document *models.Document) error
	OnDocumentUpdated(ctx context.Context, collectionName string, document *models.Document) error
	OnDocumentDeleted(ctx context.Context, collectionName string, documentID primitive.ObjectID) error
	
	// Internal methods for job processor
	GetRepository() Repository
	GetCollectionService() collection.Service
	GetEmbeddingClient() EmbeddingClient
	ExtractContent(document map[string]interface{}, config *models.CollectionRAGConfig) string
}

type service struct {
	repo              Repository
	collectionService collection.Service
	embeddingClient   EmbeddingClient
	jobProcessor      *JobProcessor
}

// NewService creates a new AI service
func NewService(db interface{}, collectionService collection.Service) Service {
	s := &service{
		repo:              NewRepository(),
		collectionService: collectionService,
		embeddingClient:   NewEmbeddingClient(),
	}
	
	// Initialize job processor
	s.jobProcessor = NewJobProcessor(s)
	
	return s
}

// Provider management

func (s *service) CreateProvider(ctx context.Context, userID primitive.ObjectID, provider *models.AIProvider) error {
	provider.CreatedBy = userID
	provider.Active = true
	
	// Initialize Models to nil (temporary compatibility)
	provider.Models = nil
	provider.IsDefault = false
	
	// Validate provider configuration
	if err := s.validateProvider(provider); err != nil {
		return fmt.Errorf("invalid provider configuration: %w", err)
	}
	
	// Set default rate limits if not provided
	if provider.RateLimits.RequestsPerMinute == 0 {
		provider.RateLimits.RequestsPerMinute = 60
	}
	if provider.RateLimits.BatchSize == 0 {
		provider.RateLimits.BatchSize = 100
	}
	
	return s.repo.CreateProvider(ctx, provider)
}

func (s *service) GetProvider(ctx context.Context, id primitive.ObjectID) (*models.AIProvider, error) {
	return s.repo.GetProvider(ctx, id)
}

func (s *service) ListProviders(ctx context.Context) ([]*models.AIProvider, error) {
	return s.repo.ListProviders(ctx, true) // Only list active providers
}

func (s *service) UpdateProvider(ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID, updates map[string]interface{}) error {
	// Verify provider exists
	provider, err := s.repo.GetProvider(ctx, id)
	if err != nil {
		return err
	}
	
	// Check ownership or admin role
	// Allow update if user is the creator OR if created_by is empty (for migrated data)
	if provider.CreatedBy != userID && !provider.CreatedBy.IsZero() {
		// TODO: Check if user is admin through RBAC service
		return fmt.Errorf("unauthorized to update this provider")
	}
	
	return s.repo.UpdateProvider(ctx, id, updates)
}

func (s *service) DeleteProvider(ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID) error {
	// Verify provider exists
	provider, err := s.repo.GetProvider(ctx, id)
	if err != nil {
		return err
	}
	
	// Check ownership or admin role
	// Allow deletion if user is the creator OR if created_by is empty (for migrated data)
	if provider.CreatedBy != userID && !provider.CreatedBy.IsZero() {
		// TODO: Check if user is admin through RBAC service
		return fmt.Errorf("unauthorized to delete this provider")
	}
	
	// Check if provider is in use by any RAG configs
	configs, err := s.repo.ListRAGConfigs(ctx)
	if err != nil {
		return err
	}
	
	for _, config := range configs {
		if config.ProviderID == id {
			return fmt.Errorf("provider is in use by collection '%s'", config.CollectionName)
		}
	}
	
	return s.repo.DeleteProvider(ctx, id)
}

func (s *service) GetProviderModels(ctx context.Context, providerID primitive.ObjectID) ([]*models.AIModel, error) {
	// Get the provider details
	provider, err := s.repo.GetProvider(ctx, providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	// Check if provider has API key for providers that need it
	if (provider.Type == models.ProviderOpenAI || provider.Type == models.ProviderOllama) && provider.APIKeyHash == "" {
		// For providers that need API keys but don't have them, return default models
		switch provider.Type {
		case models.ProviderOpenAI:
			return []*models.AIModel{
				{
					ID:          "text-embedding-ada-002",
					Name:        "Text Embedding Ada 002",
					Description: "OpenAI's most capable embedding model",
					Type:        "embedding",
					Dimensions:  1536,
				},
				{
					ID:          "text-embedding-3-small",
					Name:        "Text Embedding 3 Small",
					Description: "OpenAI's smaller, faster embedding model",
					Type:        "embedding",
					Dimensions:  1536,
				},
				{
					ID:          "text-embedding-3-large",
					Name:        "Text Embedding 3 Large",
					Description: "OpenAI's largest embedding model",
					Type:        "embedding",
					Dimensions:  3072,
				},
			}, nil
		case models.ProviderOllama:
			return []*models.AIModel{
				{
					ID:          "nomic-embed-text",
					Name:        "Nomic Embed Text",
					Description: "Ollama embedding model",
					Type:        "embedding",
					Dimensions:  768,
				},
			}, nil
		}
	}

	switch provider.Type {
	case models.ProviderOpenAI:
		return s.getOpenAIModels(ctx, provider)
	case models.ProviderAnthropic:
		return s.getAnthropicModels(ctx, provider)
	case models.ProviderCohere:
		return s.getCohereModels(ctx, provider)
	case models.ProviderHuggingFace:
		return s.getHuggingFaceModels(ctx, provider)
	case models.ProviderOllama:
		return s.getOllamaModels(ctx, provider)
	default:
		return []*models.AIModel{}, nil
	}
}

func (s *service) getOpenAIModels(ctx context.Context, provider *models.AIProvider) ([]*models.AIModel, error) {
	baseURL := provider.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com"
	}

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/v1/models", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+provider.APIKeyHash) // APIKeyHash contains the actual key
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		// Return default models if API call fails
		return s.getDefaultOpenAIModels(), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Return default models for any non-200 status (401, 403, etc.)
		return s.getDefaultOpenAIModels(), nil
	}

	var response struct {
		Object string `json:"object"`
		Data   []struct {
			ID      string   `json:"id"`
			Object  string   `json:"object"`
			Created int64    `json:"created"`
			OwnedBy string   `json:"owned_by"`
			Root    string   `json:"root,omitempty"`
			Parent  string   `json:"parent,omitempty"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var aiModels []*models.AIModel
	for _, model := range response.Data {
		// Filter for embedding models
		if strings.Contains(model.ID, "embedding") || strings.Contains(model.ID, "ada") {
			aiModel := &models.AIModel{
				ID:   model.ID,
				Name: model.ID,
				Type: "embedding",
			}

			// Set known dimensions for common models
			switch model.ID {
			case "text-embedding-ada-002":
				aiModel.Dimensions = 1536
				aiModel.Name = "Text Embedding Ada 002"
				aiModel.Description = "OpenAI's most capable embedding model"
			case "text-embedding-3-small":
				aiModel.Dimensions = 1536
				aiModel.Name = "Text Embedding 3 Small"
				aiModel.Description = "OpenAI's smaller, faster embedding model"
			case "text-embedding-3-large":
				aiModel.Dimensions = 3072
				aiModel.Name = "Text Embedding 3 Large"
				aiModel.Description = "OpenAI's largest embedding model"
			}

			aiModels = append(aiModels, aiModel)
		}
	}

	return aiModels, nil
}

func (s *service) getDefaultOpenAIModels() []*models.AIModel {
	return []*models.AIModel{
		{
			ID:          "text-embedding-ada-002",
			Name:        "Text Embedding Ada 002",
			Description: "OpenAI's most capable embedding model",
			Type:        "embedding",
			Dimensions:  1536,
		},
		{
			ID:          "text-embedding-3-small",
			Name:        "Text Embedding 3 Small",
			Description: "OpenAI's smaller, faster embedding model",
			Type:        "embedding",
			Dimensions:  1536,
		},
		{
			ID:          "text-embedding-3-large",
			Name:        "Text Embedding 3 Large",
			Description: "OpenAI's largest embedding model",
			Type:        "embedding",
			Dimensions:  3072,
		},
	}
}

func (s *service) getAnthropicModels(ctx context.Context, provider *models.AIProvider) ([]*models.AIModel, error) {
	// Anthropic doesn't have embedding models, return empty list
	return []*models.AIModel{}, nil
}

func (s *service) getCohereModels(ctx context.Context, provider *models.AIProvider) ([]*models.AIModel, error) {
	// Return predefined Cohere embedding models
	return []*models.AIModel{
		{
			ID:          "embed-english-v3.0",
			Name:        "Embed English v3.0",
			Description: "Cohere's English embedding model",
			Type:        "embedding",
			Dimensions:  1024,
		},
		{
			ID:          "embed-multilingual-v3.0",
			Name:        "Embed Multilingual v3.0",
			Description: "Cohere's multilingual embedding model",
			Type:        "embedding",
			Dimensions:  1024,
		},
	}, nil
}

func (s *service) getHuggingFaceModels(ctx context.Context, provider *models.AIProvider) ([]*models.AIModel, error) {
	// Return some popular HuggingFace embedding models
	return []*models.AIModel{
		{
			ID:          "sentence-transformers/all-MiniLM-L6-v2",
			Name:        "All-MiniLM-L6-v2",
			Description: "Lightweight sentence transformer",
			Type:        "embedding",
			Dimensions:  384,
		},
		{
			ID:          "sentence-transformers/all-mpnet-base-v2",
			Name:        "All-MPNet-Base-v2",
			Description: "High-quality sentence transformer",
			Type:        "embedding",
			Dimensions:  768,
		},
	}, nil
}

func (s *service) getOllamaModels(ctx context.Context, provider *models.AIProvider) ([]*models.AIModel, error) {
	baseURL := provider.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ollama API returned status %d", resp.StatusCode)
	}

	var response struct {
		Models []struct {
			Name       string `json:"name"`
			Size       int64  `json:"size"`
			ModifiedAt string `json:"modified_at"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var aiModels []*models.AIModel
	for _, model := range response.Models {
		// Filter for embedding-capable models
		if strings.Contains(model.Name, "embed") || strings.Contains(model.Name, "nomic") {
			aiModels = append(aiModels, &models.AIModel{
				ID:          model.Name,
				Name:        model.Name,
				Description: "Ollama embedding model",
				Type:        "embedding",
				Dimensions:  768, // Default dimension, might vary
			})
		}
	}

	return aiModels, nil
}

// RAG Configuration

func (s *service) ConfigureRAG(ctx context.Context, userID primitive.ObjectID, config *models.CollectionRAGConfig) error {
	config.CreatedBy = userID
	
	// Verify collection exists
	collection, err := s.collectionService.GetCollection(ctx, userID, config.CollectionName)
	if err != nil {
		return fmt.Errorf("collection not found: %w", err)
	}
	
	// Verify provider exists and is active
	provider, err := s.repo.GetProvider(ctx, config.ProviderID)
	if err != nil {
		return fmt.Errorf("provider not found: %w", err)
	}
	if !provider.Active {
		return fmt.Errorf("provider is not active")
	}
	
	// Validate vector field exists or create it
	vectorFieldExists := false
	for _, field := range collection.VectorFields {
		if field.Name == config.VectorFieldName {
			vectorFieldExists = true
			// Verify dimensions match
			if field.Dimensions != GetEmbeddingDimensions(provider) {
				return fmt.Errorf("vector field dimensions (%d) don't match provider embedding dimensions (%d)", 
					field.Dimensions, GetEmbeddingDimensions(provider))
			}
			break
		}
	}
	
	if !vectorFieldExists {
		// Add vector field to collection
		vectorField := models.VectorField{
			Name:           config.VectorFieldName,
			Dimensions:     GetEmbeddingDimensions(provider),
			EmbeddingModel: GetEmbeddingModel(provider),
			IndexType:      "hnsw",
			Metric:         "cosine",
		}
		
		if err := s.collectionService.AddVectorField(ctx, userID, config.CollectionName, vectorField); err != nil {
			return fmt.Errorf("failed to add vector field: %w", err)
		}
	}
	
	// Set default values
	if config.FieldSeparator == "" {
		config.FieldSeparator = " "
	}
	if config.ChunkSize == 0 {
		config.ChunkSize = 512
	}
	if config.ChunkOverlap == 0 {
		config.ChunkOverlap = 50
	}
	
	// Get total document count
	docCount, err := s.collectionService.CountDocuments(ctx, userID, config.CollectionName, map[string]interface{}{})
	if err == nil {
		config.TotalDocuments = docCount
	}
	
	return s.repo.CreateRAGConfig(ctx, config)
}

func (s *service) GetRAGConfig(ctx context.Context, collectionName string) (*models.CollectionRAGConfig, error) {
	return s.repo.GetRAGConfig(ctx, collectionName)
}

func (s *service) UpdateRAGConfig(ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID, updates map[string]interface{}) error {
	// Verify config exists
	config, err := s.repo.GetRAGConfigByID(ctx, id)
	if err != nil {
		return err
	}
	
	// Check ownership or admin role
	if config.CreatedBy != userID {
		// TODO: Check if user is admin
		return fmt.Errorf("unauthorized to update this configuration")
	}
	
	// If provider is being changed, validate it
	if providerID, ok := updates["provider_id"].(primitive.ObjectID); ok {
		provider, err := s.repo.GetProvider(ctx, providerID)
		if err != nil {
			return fmt.Errorf("provider not found: %w", err)
		}
		if !provider.Active {
			return fmt.Errorf("provider is not active")
		}
	}
	
	return s.repo.UpdateRAGConfig(ctx, id, updates)
}

func (s *service) DeleteRAGConfig(ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID) error {
	// Verify config exists
	config, err := s.repo.GetRAGConfigByID(ctx, id)
	if err != nil {
		return err
	}
	
	// Check ownership or admin role
	if config.CreatedBy != userID {
		// TODO: Check if user is admin
		return fmt.Errorf("unauthorized to delete this configuration")
	}
	
	// Cancel any pending jobs
	jobs, err := s.repo.ListEmbeddingJobs(ctx, map[string]interface{}{
		"config_id": id,
		"status": "pending",
	})
	if err == nil {
		for _, job := range jobs {
			s.repo.UpdateEmbeddingJob(ctx, job.ID, map[string]interface{}{
				"status": "cancelled",
			})
		}
	}
	
	return s.repo.DeleteRAGConfig(ctx, id)
}

func (s *service) ListRAGConfigs(ctx context.Context) ([]*models.CollectionRAGConfig, error) {
	return s.repo.ListRAGConfigs(ctx)
}

// Embedding operations

func (s *service) GenerateEmbeddings(ctx context.Context, userID primitive.ObjectID, collectionName string, documentIDs []string) (*models.EmbeddingJob, error) {
	// Get RAG config
	config, err := s.repo.GetRAGConfig(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("RAG not configured for collection: %w", err)
	}
	
	if !config.Enabled {
		return nil, fmt.Errorf("RAG is disabled for this collection")
	}
	
	// Convert string IDs to ObjectIDs
	var objIDs []primitive.ObjectID
	for _, id := range documentIDs {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, fmt.Errorf("invalid document ID: %s", id)
		}
		objIDs = append(objIDs, objID)
	}
	
	// Create embedding job
	job := &models.EmbeddingJob{
		Type:           "document",
		CollectionName: collectionName,
		ProviderID:     config.ProviderID,
		ConfigID:       config.ID,
		DocumentIDs:    objIDs,
		TotalDocuments: len(objIDs),
		MaxRetries:     3,
		CreatedBy:      userID,
	}
	
	if err := s.repo.CreateEmbeddingJob(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to create embedding job: %w", err)
	}
	
	// Start processing in background
	go s.jobProcessor.ProcessJob(job.ID)
	
	return job, nil
}

func (s *service) GenerateAllEmbeddings(ctx context.Context, userID primitive.ObjectID, collectionName string) (*models.EmbeddingJob, error) {
	// Get RAG config
	config, err := s.repo.GetRAGConfig(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("RAG not configured for collection: %w", err)
	}
	
	if !config.Enabled {
		return nil, fmt.Errorf("RAG is disabled for this collection")
	}
	
	// Get document count
	docCount, err := s.collectionService.CountDocuments(ctx, userID, collectionName, map[string]interface{}{})
	if err != nil {
		return nil, fmt.Errorf("failed to count documents: %w", err)
	}
	
	// Create embedding job
	job := &models.EmbeddingJob{
		Type:           "full",
		CollectionName: collectionName,
		ProviderID:     config.ProviderID,
		ConfigID:       config.ID,
		TotalDocuments: docCount,
		MaxRetries:     3,
		CreatedBy:      userID,
	}
	
	if err := s.repo.CreateEmbeddingJob(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to create embedding job: %w", err)
	}
	
	// Start processing in background
	go s.jobProcessor.ProcessJob(job.ID)
	
	return job, nil
}

func (s *service) GetEmbeddingJob(ctx context.Context, jobID primitive.ObjectID) (*models.EmbeddingJob, error) {
	return s.repo.GetEmbeddingJob(ctx, jobID)
}

func (s *service) ListEmbeddingJobs(ctx context.Context, collectionName string) ([]*models.EmbeddingJob, error) {
	filter := map[string]interface{}{}
	if collectionName != "" {
		filter["collection_name"] = collectionName
	}
	return s.repo.ListEmbeddingJobs(ctx, filter)
}

func (s *service) CancelEmbeddingJob(ctx context.Context, userID primitive.ObjectID, jobID primitive.ObjectID) error {
	job, err := s.repo.GetEmbeddingJob(ctx, jobID)
	if err != nil {
		return err
	}
	
	// Check ownership
	if job.CreatedBy != userID {
		// TODO: Check if user is admin
		return fmt.Errorf("unauthorized to cancel this job")
	}
	
	if job.Status != "pending" && job.Status != "processing" {
		return fmt.Errorf("job cannot be cancelled (status: %s)", job.Status)
	}
	
	return s.repo.UpdateEmbeddingJob(ctx, jobID, map[string]interface{}{
		"status": "cancelled",
		"completed_at": time.Now().UTC(),
	})
}

// RAG queries

func (s *service) QueryRAG(ctx context.Context, userID primitive.ObjectID, query *models.RAGQuery) (*models.RAGResponse, error) {
	// Get RAG config
	config, err := s.repo.GetRAGConfig(ctx, query.CollectionName)
	if err != nil {
		return nil, fmt.Errorf("RAG not configured for collection: %w", err)
	}
	
	if !config.Enabled {
		return nil, fmt.Errorf("RAG is disabled for this collection")
	}
	
	// Get provider
	provider, err := s.repo.GetProvider(ctx, config.ProviderID)
	if err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}
	
	startTime := time.Now()
	
	// Generate embedding for query
	queryEmbedding, err := s.embeddingClient.GenerateEmbedding(ctx, provider, query.Query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}
	
	// Set default values
	if query.TopK == 0 {
		query.TopK = 10
	}
	
	// Perform vector search
	searchOpts := collection.VectorSearchOptions{
		VectorField:  config.VectorFieldName,
		QueryVector:  queryEmbedding,
		TopK:         query.TopK,
		Filter:       query.Filter,
		IncludeScore: true,
	}
	
	// Use collection service for vector search
	results, err := s.collectionService.VectorSearch(ctx, userID, query.CollectionName, searchOpts)
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}
	
	// Convert results to RAG documents
	var ragDocs []models.RAGDocument
	for _, result := range results {
		doc := models.RAGDocument{
			Data: result,
		}
		
		// Extract score if present
		if score, ok := result["_score"].(float32); ok {
			doc.Score = score
			delete(result, "_score") // Remove from data
		}
		
		// Extract distance if present
		if distance, ok := result["_distance"].(float32); ok {
			doc.Distance = distance
			delete(result, "_distance") // Remove from data
		}
		
		// Extract ID
		if id, ok := result["_id"].(string); ok {
			if objID, err := primitive.ObjectIDFromHex(id); err == nil {
				doc.ID = objID
			}
		}
		
		ragDocs = append(ragDocs, doc)
	}
	
	// Build response
	response := &models.RAGResponse{
		Query:      query.Query,
		Documents:  ragDocs,
		TotalFound: len(ragDocs),
		Provider:   string(provider.Type),
		Model:      GetEmbeddingModel(provider),
		SearchType: "vector",
		Latency:    time.Since(startTime).Milliseconds(),
	}
	
	if query.HybridSearch {
		response.SearchType = "hybrid"
	}
	
	return response, nil
}

// Document hooks for auto-embedding

func (s *service) OnDocumentCreated(ctx context.Context, collectionName string, document *models.Document) error {
	// Check if auto-embed is enabled
	config, err := s.repo.GetRAGConfig(ctx, collectionName)
	if err != nil || !config.Enabled || !config.AutoEmbed {
		return nil // Silent fail - RAG not configured or auto-embed disabled
	}
	
	// Create single document embedding job
	job := &models.EmbeddingJob{
		Type:           "incremental",
		CollectionName: collectionName,
		ProviderID:     config.ProviderID,
		ConfigID:       config.ID,
		DocumentIDs:    []primitive.ObjectID{document.ID},
		TotalDocuments: 1,
		MaxRetries:     3,
		CreatedBy:      document.CreatedBy,
	}
	
	if err := s.repo.CreateEmbeddingJob(ctx, job); err != nil {
		// Log error but don't fail document creation
		fmt.Printf("Failed to create embedding job for document %s: %v\n", document.ID.Hex(), err)
		return nil
	}
	
	// Process immediately for single documents
	go s.jobProcessor.ProcessJob(job.ID)
	
	return nil
}

func (s *service) OnDocumentUpdated(ctx context.Context, collectionName string, document *models.Document) error {
	// Check if auto-embed is enabled
	config, err := s.repo.GetRAGConfig(ctx, collectionName)
	if err != nil || !config.Enabled || !config.AutoEmbed {
		return nil
	}
	
	// Check if content fields were updated
	// For now, regenerate embedding on any update
	// TODO: Check if specific fields were updated
	
	job := &models.EmbeddingJob{
		Type:           "incremental",
		CollectionName: collectionName,
		ProviderID:     config.ProviderID,
		ConfigID:       config.ID,
		DocumentIDs:    []primitive.ObjectID{document.ID},
		TotalDocuments: 1,
		MaxRetries:     3,
		CreatedBy:      document.UpdatedBy,
	}
	
	if err := s.repo.CreateEmbeddingJob(ctx, job); err != nil {
		fmt.Printf("Failed to create embedding job for updated document %s: %v\n", document.ID.Hex(), err)
		return nil
	}
	
	go s.jobProcessor.ProcessJob(job.ID)
	
	return nil
}

func (s *service) OnDocumentDeleted(ctx context.Context, collectionName string, documentID primitive.ObjectID) error {
	// No action needed - vector will be removed with document
	return nil
}

// Helper methods

func (s *service) validateProvider(provider *models.AIProvider) error {
	if provider.Name == "" {
		return fmt.Errorf("provider name is required")
	}
	
	if provider.Type == "" {
		return fmt.Errorf("provider type is required")
	}
	
	// Validate based on provider type
	switch provider.Type {
	case models.ProviderOpenAI:
		if provider.APIKey == "" {
			return fmt.Errorf("API key is required for OpenAI provider")
		}
		
	case models.ProviderAnthropic:
		if provider.APIKey == "" {
			return fmt.Errorf("API key is required for Anthropic provider")
		}
		
	case models.ProviderCohere:
		if provider.APIKey == "" {
			return fmt.Errorf("API key is required for Cohere provider")
		}
		
	case models.ProviderHuggingFace:
		if provider.APIKey == "" {
			return fmt.Errorf("API key is required for HuggingFace provider")
		}
		
	case models.ProviderOllama:
		if provider.BaseURL == "" {
			provider.BaseURL = "http://localhost:11434"
		}
		
	case models.ProviderCustom:
		if provider.BaseURL == "" {
			return fmt.Errorf("base URL is required for custom provider")
		}
		
	default:
		return fmt.Errorf("unsupported provider type: %s", provider.Type)
	}
	
	return nil
}

// extractContent extracts content from document based on RAG configuration
func (s *service) extractContent(document map[string]interface{}, config *models.CollectionRAGConfig) string {
	var contents []string
	
	for _, field := range config.ContentFields {
		if value, ok := document[field]; ok {
			// Convert value to string
			var strValue string
			switch v := value.(type) {
			case string:
				strValue = v
			case int, int32, int64, float32, float64:
				strValue = fmt.Sprintf("%v", v)
			case bool:
				strValue = fmt.Sprintf("%v", v)
			default:
				// Skip complex types for now
				continue
			}
			
			// Apply preprocessing rules
			for _, rule := range config.PreprocessingRules {
				if rule.Field == field {
					strValue = s.applyPreprocessing(strValue, rule)
				}
			}
			
			if strValue != "" {
				contents = append(contents, strValue)
			}
		}
	}
	
	// Combine fields
	separator := config.FieldSeparator
	if separator == "" {
		separator = " "
	}
	
	return strings.Join(contents, separator)
}

// applyPreprocessing applies preprocessing rules to text
func (s *service) applyPreprocessing(text string, rule models.PreprocessRule) string {
	switch rule.Operation {
	case "lowercase":
		return strings.ToLower(text)
	case "strip_html":
		// Simple HTML stripping - in production, use a proper HTML parser
		// This is a placeholder
		return text
	case "truncate":
		if maxLen, ok := rule.Options["max_length"].(int); ok && len(text) > maxLen {
			return text[:maxLen]
		}
	case "regex_replace":
		// Implement regex replacement
		// This is a placeholder
	}
	return text
}

// GetRepository returns the repository
func (s *service) GetRepository() Repository {
	return s.repo
}

// GetCollectionService returns the collection service
func (s *service) GetCollectionService() collection.Service {
	return s.collectionService
}

// GetEmbeddingClient returns the embedding client
func (s *service) GetEmbeddingClient() EmbeddingClient {
	return s.embeddingClient
}

// ExtractContent extracts content from document (public wrapper)
func (s *service) ExtractContent(document map[string]interface{}, config *models.CollectionRAGConfig) string {
	return s.extractContent(document, config)
}