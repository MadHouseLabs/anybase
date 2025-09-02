package ai

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/madhouselabs/anybase/internal/database"
	"github.com/madhouselabs/anybase/internal/database/types"
	"github.com/madhouselabs/anybase/pkg/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrProviderNotFound = errors.New("AI provider not found")
	ErrProviderExists   = errors.New("AI provider already exists")
	ErrConfigNotFound   = errors.New("RAG configuration not found")
	ErrConfigExists     = errors.New("RAG configuration already exists for this collection")
)

// Repository defines the interface for AI provider operations
type Repository interface {
	// Provider operations
	CreateProvider(ctx context.Context, provider *models.AIProvider) error
	GetProvider(ctx context.Context, id primitive.ObjectID) (*models.AIProvider, error)
	GetProviderByName(ctx context.Context, name string) (*models.AIProvider, error)
	ListProviders(ctx context.Context, active bool) ([]*models.AIProvider, error)
	UpdateProvider(ctx context.Context, id primitive.ObjectID, updates map[string]interface{}) error
	DeleteProvider(ctx context.Context, id primitive.ObjectID) error
	GetDefaultProvider(ctx context.Context) (*models.AIProvider, error)
	
	// RAG configuration operations
	CreateRAGConfig(ctx context.Context, config *models.CollectionRAGConfig) error
	GetRAGConfig(ctx context.Context, collectionName string) (*models.CollectionRAGConfig, error)
	GetRAGConfigByID(ctx context.Context, id primitive.ObjectID) (*models.CollectionRAGConfig, error)
	ListRAGConfigs(ctx context.Context) ([]*models.CollectionRAGConfig, error)
	UpdateRAGConfig(ctx context.Context, id primitive.ObjectID, updates map[string]interface{}) error
	DeleteRAGConfig(ctx context.Context, id primitive.ObjectID) error
	
	// Embedding job operations
	CreateEmbeddingJob(ctx context.Context, job *models.EmbeddingJob) error
	GetEmbeddingJob(ctx context.Context, id primitive.ObjectID) (*models.EmbeddingJob, error)
	ListEmbeddingJobs(ctx context.Context, filter map[string]interface{}) ([]*models.EmbeddingJob, error)
	UpdateEmbeddingJob(ctx context.Context, id primitive.ObjectID, updates map[string]interface{}) error
	GetPendingJobs(ctx context.Context, limit int) ([]*models.EmbeddingJob, error)
}

type repository struct {
	db types.DB
}

// NewRepository creates a new AI repository
func NewRepository() Repository {
	return &repository{
		db: database.GetDB(),
	}
}

// Provider operations

func (r *repository) CreateProvider(ctx context.Context, provider *models.AIProvider) error {
	// Encrypt the API key if provided (we need to be able to decrypt it for API calls)
	if provider.APIKey != "" {
		// For now, we'll store it as-is and implement proper encryption later
		// TODO: Implement AES encryption for API keys
		provider.APIKeyHash = provider.APIKey
		provider.APIKey = "" // Clear the plain text key from the struct
	}
	
	provider.ID = primitive.NewObjectID()
	provider.CreatedAt = time.Now().UTC()
	provider.UpdatedAt = time.Now().UTC()
	
	col := r.db.Collection("ai_providers")
	
	// Convert to map for storage
	providerDoc := map[string]interface{}{
		"_id":           provider.ID.Hex(),
		"name":          provider.Name,
		"type":          provider.Type,
		"api_key_hash":  provider.APIKeyHash,
		"base_url":      provider.BaseURL,
		"settings":      provider.Settings,
		"rate_limits":   provider.RateLimits,
		"active":        provider.Active,
		"created_by":    provider.CreatedBy.Hex(),
		"created_at":    provider.CreatedAt,
		"updated_at":    provider.UpdatedAt,
	}
	
	_, err := col.InsertOne(ctx, providerDoc)
	if err != nil {
		return fmt.Errorf("failed to create AI provider: %w", err)
	}
	
	return nil
}

func (r *repository) GetProvider(ctx context.Context, id primitive.ObjectID) (*models.AIProvider, error) {
	col := r.db.Collection("ai_providers")
	
	var provider models.AIProvider
	err := col.FindOne(ctx, map[string]interface{}{"_id": id.Hex()}, &provider)
	if err != nil {
		if err == types.ErrNoDocuments {
			return nil, ErrProviderNotFound
		}
		return nil, fmt.Errorf("failed to get AI provider: %w", err)
	}
	
	return &provider, nil
}

func (r *repository) GetProviderByName(ctx context.Context, name string) (*models.AIProvider, error) {
	col := r.db.Collection("ai_providers")
	
	var provider models.AIProvider
	err := col.FindOne(ctx, map[string]interface{}{"name": name}, &provider)
	if err != nil {
		if err == types.ErrNoDocuments {
			return nil, ErrProviderNotFound
		}
		return nil, fmt.Errorf("failed to get AI provider by name: %w", err)
	}
	
	return &provider, nil
}

func (r *repository) ListProviders(ctx context.Context, active bool) ([]*models.AIProvider, error) {
	col := r.db.Collection("ai_providers")
	
	filter := map[string]interface{}{}
	if active {
		filter["active"] = true
	}
	
	cursor, err := col.Find(ctx, filter, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list AI providers: %w", err)
	}
	defer cursor.Close(ctx)
	
	var providers []*models.AIProvider
	for cursor.Next(ctx) {
		var provider models.AIProvider
		if err := cursor.Decode(&provider); err != nil {
			continue
		}
		providers = append(providers, &provider)
	}
	
	return providers, nil
}

func (r *repository) UpdateProvider(ctx context.Context, id primitive.ObjectID, updates map[string]interface{}) error {
	col := r.db.Collection("ai_providers")
	
	// Handle API key update
	if apiKey, ok := updates["api_key"].(string); ok && apiKey != "" {
		// For now, store API key as-is (implement proper encryption later)
		// TODO: Implement AES encryption for API keys
		updates["api_key_hash"] = apiKey
		delete(updates, "api_key") // Remove plain text key from updates
	}
	
	updates["updated_at"] = time.Now().UTC()
	
	// If setting as default, unset others
	if isDefault, ok := updates["is_default"].(bool); ok && isDefault {
		_, err := col.UpdateMany(ctx, map[string]interface{}{"is_default": true}, map[string]interface{}{
			"$set": map[string]interface{}{"is_default": false},
		})
		if err != nil {
			return fmt.Errorf("failed to unset default providers: %w", err)
		}
	}
	
	result, err := col.UpdateOne(ctx, map[string]interface{}{"_id": id.Hex()}, map[string]interface{}{
		"$set": updates,
	})
	if err != nil {
		return fmt.Errorf("failed to update AI provider: %w", err)
	}
	
	if result.ModifiedCount == 0 {
		return ErrProviderNotFound
	}
	
	return nil
}

func (r *repository) DeleteProvider(ctx context.Context, id primitive.ObjectID) error {
	col := r.db.Collection("ai_providers")
	
	result, err := col.DeleteOne(ctx, map[string]interface{}{"_id": id.Hex()})
	if err != nil {
		return fmt.Errorf("failed to delete AI provider: %w", err)
	}
	
	if result.DeletedCount == 0 {
		return ErrProviderNotFound
	}
	
	return nil
}

func (r *repository) GetDefaultProvider(ctx context.Context) (*models.AIProvider, error) {
	col := r.db.Collection("ai_providers")
	
	var provider models.AIProvider
	err := col.FindOne(ctx, map[string]interface{}{
		"is_default": true,
		"active": true,
	}, &provider)
	if err != nil {
		if err == types.ErrNoDocuments {
			// If no default, get first active provider
			err = col.FindOne(ctx, map[string]interface{}{"active": true}, &provider)
			if err != nil {
				return nil, fmt.Errorf("no active AI providers found")
			}
		} else {
			return nil, fmt.Errorf("failed to get default AI provider: %w", err)
		}
	}
	
	return &provider, nil
}

// RAG Configuration operations

func (r *repository) CreateRAGConfig(ctx context.Context, config *models.CollectionRAGConfig) error {
	// Check if config already exists for this collection
	existing, _ := r.GetRAGConfig(ctx, config.CollectionName)
	if existing != nil {
		return ErrConfigExists
	}
	
	config.ID = primitive.NewObjectID()
	config.CreatedAt = time.Now().UTC()
	config.UpdatedAt = time.Now().UTC()
	config.ProcessingStatus = "pending"
	
	col := r.db.Collection("rag_configs")
	
	// Convert to map for storage
	configDoc := map[string]interface{}{
		"_id":                config.ID.Hex(),
		"collection_name":    config.CollectionName,
		"provider_id":        config.ProviderID.Hex(),
		"enabled":            config.Enabled,
		"auto_embed":         config.AutoEmbed,
		"content_fields":     config.ContentFields,
		"vector_field_name":  config.VectorFieldName,
		"metadata_fields":    config.MetadataFields,
		"combine_fields":     config.CombineFields,
		"field_separator":    config.FieldSeparator,
		"preprocessing_rules": config.PreprocessingRules,
		"enable_chunking":    config.EnableChunking,
		"chunk_size":         config.ChunkSize,
		"chunk_overlap":      config.ChunkOverlap,
		"processing_status":  config.ProcessingStatus,
		"documents_processed": config.DocumentsProcessed,
		"total_documents":    config.TotalDocuments,
		"created_by":         config.CreatedBy.Hex(),
		"created_at":         config.CreatedAt,
		"updated_at":         config.UpdatedAt,
	}
	
	_, err := col.InsertOne(ctx, configDoc)
	if err != nil {
		return fmt.Errorf("failed to create RAG configuration: %w", err)
	}
	
	return nil
}

func (r *repository) GetRAGConfig(ctx context.Context, collectionName string) (*models.CollectionRAGConfig, error) {
	col := r.db.Collection("rag_configs")
	
	var config models.CollectionRAGConfig
	err := col.FindOne(ctx, map[string]interface{}{"collection_name": collectionName}, &config)
	if err != nil {
		if err == types.ErrNoDocuments {
			return nil, ErrConfigNotFound
		}
		return nil, fmt.Errorf("failed to get RAG configuration: %w", err)
	}
	
	return &config, nil
}

func (r *repository) GetRAGConfigByID(ctx context.Context, id primitive.ObjectID) (*models.CollectionRAGConfig, error) {
	col := r.db.Collection("rag_configs")
	
	var config models.CollectionRAGConfig
	err := col.FindOne(ctx, map[string]interface{}{"_id": id.Hex()}, &config)
	if err != nil {
		if err == types.ErrNoDocuments {
			return nil, ErrConfigNotFound
		}
		return nil, fmt.Errorf("failed to get RAG configuration: %w", err)
	}
	
	return &config, nil
}

func (r *repository) ListRAGConfigs(ctx context.Context) ([]*models.CollectionRAGConfig, error) {
	col := r.db.Collection("rag_configs")
	
	cursor, err := col.Find(ctx, map[string]interface{}{}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list RAG configurations: %w", err)
	}
	defer cursor.Close(ctx)
	
	var configs []*models.CollectionRAGConfig
	for cursor.Next(ctx) {
		var config models.CollectionRAGConfig
		if err := cursor.Decode(&config); err != nil {
			continue
		}
		configs = append(configs, &config)
	}
	
	return configs, nil
}

func (r *repository) UpdateRAGConfig(ctx context.Context, id primitive.ObjectID, updates map[string]interface{}) error {
	col := r.db.Collection("rag_configs")
	
	updates["updated_at"] = time.Now().UTC()
	
	result, err := col.UpdateOne(ctx, map[string]interface{}{"_id": id.Hex()}, map[string]interface{}{
		"$set": updates,
	})
	if err != nil {
		return fmt.Errorf("failed to update RAG configuration: %w", err)
	}
	
	if result.ModifiedCount == 0 {
		return ErrConfigNotFound
	}
	
	return nil
}

func (r *repository) DeleteRAGConfig(ctx context.Context, id primitive.ObjectID) error {
	col := r.db.Collection("rag_configs")
	
	result, err := col.DeleteOne(ctx, map[string]interface{}{"_id": id.Hex()})
	if err != nil {
		return fmt.Errorf("failed to delete RAG configuration: %w", err)
	}
	
	if result.DeletedCount == 0 {
		return ErrConfigNotFound
	}
	
	return nil
}

// Embedding Job operations

func (r *repository) CreateEmbeddingJob(ctx context.Context, job *models.EmbeddingJob) error {
	job.ID = primitive.NewObjectID()
	job.CreatedAt = time.Now().UTC()
	job.Status = "pending"
	
	col := r.db.Collection("embedding_jobs")
	
	// Convert to map for storage
	jobDoc := map[string]interface{}{
		"_id":              job.ID.Hex(),
		"type":             job.Type,
		"collection_name":  job.CollectionName,
		"provider_id":      job.ProviderID.Hex(),
		"config_id":        job.ConfigID.Hex(),
		"document_ids":     job.DocumentIDs,
		"filter":           job.Filter,
		"status":           job.Status,
		"progress":         job.Progress,
		"total_documents":  job.TotalDocuments,
		"processed_docs":   job.ProcessedDocs,
		"failed_docs":      job.FailedDocs,
		"errors":           job.Errors,
		"retry_count":      job.RetryCount,
		"max_retries":      job.MaxRetries,
		"created_by":       job.CreatedBy.Hex(),
		"created_at":       job.CreatedAt,
	}
	
	_, err := col.InsertOne(ctx, jobDoc)
	if err != nil {
		return fmt.Errorf("failed to create embedding job: %w", err)
	}
	
	return nil
}

func (r *repository) GetEmbeddingJob(ctx context.Context, id primitive.ObjectID) (*models.EmbeddingJob, error) {
	col := r.db.Collection("embedding_jobs")
	
	var job models.EmbeddingJob
	err := col.FindOne(ctx, map[string]interface{}{"_id": id.Hex()}, &job)
	if err != nil {
		if err == types.ErrNoDocuments {
			return nil, fmt.Errorf("embedding job not found")
		}
		return nil, fmt.Errorf("failed to get embedding job: %w", err)
	}
	
	return &job, nil
}

func (r *repository) ListEmbeddingJobs(ctx context.Context, filter map[string]interface{}) ([]*models.EmbeddingJob, error) {
	col := r.db.Collection("embedding_jobs")
	
	cursor, err := col.Find(ctx, filter, &types.FindOptions{
		Sort: map[string]int{"created_at": -1},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list embedding jobs: %w", err)
	}
	defer cursor.Close(ctx)
	
	var jobs []*models.EmbeddingJob
	for cursor.Next(ctx) {
		var job models.EmbeddingJob
		if err := cursor.Decode(&job); err != nil {
			continue
		}
		jobs = append(jobs, &job)
	}
	
	return jobs, nil
}

func (r *repository) UpdateEmbeddingJob(ctx context.Context, id primitive.ObjectID, updates map[string]interface{}) error {
	col := r.db.Collection("embedding_jobs")
	
	result, err := col.UpdateOne(ctx, map[string]interface{}{"_id": id.Hex()}, map[string]interface{}{
		"$set": updates,
	})
	if err != nil {
		return fmt.Errorf("failed to update embedding job: %w", err)
	}
	
	if result.ModifiedCount == 0 {
		return fmt.Errorf("embedding job not found")
	}
	
	return nil
}

func (r *repository) GetPendingJobs(ctx context.Context, limit int) ([]*models.EmbeddingJob, error) {
	col := r.db.Collection("embedding_jobs")
	
	limitInt64 := int64(limit)
	cursor, err := col.Find(ctx, map[string]interface{}{
		"status": "pending",
	}, &types.FindOptions{
		Limit: &limitInt64,
		Sort:  map[string]int{"created_at": 1}, // FIFO
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get pending jobs: %w", err)
	}
	defer cursor.Close(ctx)
	
	var jobs []*models.EmbeddingJob
	for cursor.Next(ctx) {
		var job models.EmbeddingJob
		if err := cursor.Decode(&job); err != nil {
			continue
		}
		jobs = append(jobs, &job)
	}
	
	return jobs, nil
}