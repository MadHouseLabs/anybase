package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ProviderType defines the type of AI provider
type ProviderType string

const (
	ProviderOpenAI      ProviderType = "openai"
	ProviderAnthropic   ProviderType = "anthropic"
	ProviderCohere      ProviderType = "cohere"
	ProviderHuggingFace ProviderType = "huggingface"
	ProviderOllama      ProviderType = "ollama"
	ProviderCustom      ProviderType = "custom"
)

// AIModel represents a model available from an AI provider
type AIModel struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Type        string                 `json:"type"` // "embedding", "completion", "chat", etc.
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	Dimensions  int                    `json:"dimensions,omitempty"` // For embedding models
	Pricing     *ModelPricing          `json:"pricing,omitempty"`
	Capabilities []string              `json:"capabilities,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ModelPricing represents pricing information for a model
type ModelPricing struct {
	InputPer1K  float64 `json:"input_per_1k,omitempty"`  // Cost per 1K input tokens
	OutputPer1K float64 `json:"output_per_1k,omitempty"` // Cost per 1K output tokens
	Currency    string  `json:"currency,omitempty"`
}

// AIProvider represents an AI service provider configuration
type AIProvider struct {
	ID          primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	Name        string                 `bson:"name" json:"name" validate:"required,min=3,max=100"`
	Type        ProviderType           `bson:"type" json:"type" validate:"required"`
	APIKey      string                 `bson:"-" json:"api_key,omitempty"` // Only for input, never stored
	APIKeyHash  string                 `bson:"api_key_hash" json:"-"`      // Stored but never exposed
	BaseURL     string                 `bson:"base_url,omitempty" json:"base_url,omitempty"`
	Settings    map[string]interface{} `bson:"settings,omitempty" json:"settings,omitempty"`
	RateLimits  RateLimitConfig        `bson:"rate_limits" json:"rate_limits"`
	Active      bool                   `bson:"active" json:"active"`
	CreatedBy   primitive.ObjectID     `bson:"created_by" json:"created_by"`
	CreatedAt   time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time              `bson:"updated_at" json:"updated_at"`
	
	// Temporary compatibility fields - will be removed
	Models      *ProviderModels        `bson:"-" json:"-"`
	IsDefault   bool                   `bson:"-" json:"-"`
}

// ProviderModels - temporary struct for compatibility
type ProviderModels struct {
	EmbeddingModel string `json:"embedding_model"`
	EmbeddingDim   int    `json:"embedding_dim"`
}

// RateLimitConfig defines rate limiting for API calls
type RateLimitConfig struct {
	RequestsPerMinute int `bson:"requests_per_minute" json:"requests_per_minute"`
	TokensPerMinute   int `bson:"tokens_per_minute,omitempty" json:"tokens_per_minute,omitempty"`
	BatchSize         int `bson:"batch_size" json:"batch_size"`
}

// CollectionRAGConfig represents RAG configuration for a collection
type CollectionRAGConfig struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CollectionName    string             `bson:"collection_name" json:"collection_name" validate:"required"`
	ProviderID        primitive.ObjectID `bson:"provider_id" json:"provider_id" validate:"required"`
	EmbeddingModel    string             `bson:"embedding_model" json:"embedding_model" validate:"required"`
	EmbeddingDim      int                `bson:"embedding_dim" json:"embedding_dim" validate:"required,min=1"`
	Enabled           bool               `bson:"enabled" json:"enabled"`
	AutoEmbed         bool               `bson:"auto_embed" json:"auto_embed"`
	
	// Field configuration
	ContentFields     []string           `bson:"content_fields" json:"content_fields" validate:"required,min=1"`
	VectorFieldName   string             `bson:"vector_field_name" json:"vector_field_name" validate:"required"`
	MetadataFields    []string           `bson:"metadata_fields,omitempty" json:"metadata_fields,omitempty"`
	
	// Processing configuration
	CombineFields     bool               `bson:"combine_fields" json:"combine_fields"`
	FieldSeparator    string             `bson:"field_separator" json:"field_separator"`
	PreprocessingRules []PreprocessRule  `bson:"preprocessing_rules,omitempty" json:"preprocessing_rules,omitempty"`
	
	// Chunking configuration (for large field values)
	EnableChunking    bool               `bson:"enable_chunking" json:"enable_chunking"`
	ChunkSize         int                `bson:"chunk_size,omitempty" json:"chunk_size,omitempty"`
	ChunkOverlap      int                `bson:"chunk_overlap,omitempty" json:"chunk_overlap,omitempty"`
	
	// Processing status
	LastProcessed     *time.Time         `bson:"last_processed,omitempty" json:"last_processed,omitempty"`
	ProcessingStatus  string             `bson:"processing_status" json:"processing_status"`
	DocumentsProcessed int               `bson:"documents_processed" json:"documents_processed"`
	TotalDocuments    int                `bson:"total_documents" json:"total_documents"`
	
	CreatedBy         primitive.ObjectID `bson:"created_by" json:"created_by"`
	CreatedAt         time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time          `bson:"updated_at" json:"updated_at"`
}

// PreprocessRule defines how to preprocess field values before embedding
type PreprocessRule struct {
	Field     string `bson:"field" json:"field"`
	Operation string `bson:"operation" json:"operation"` // "lowercase", "strip_html", "truncate", "regex_replace"
	Options   map[string]interface{} `bson:"options,omitempty" json:"options,omitempty"`
}

// EmbeddingJob represents a background job for generating embeddings
type EmbeddingJob struct {
	ID             primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	Type           string                 `bson:"type" json:"type"` // "full", "incremental", "document"
	CollectionName string                 `bson:"collection_name" json:"collection_name"`
	ProviderID     primitive.ObjectID     `bson:"provider_id" json:"provider_id"`
	ConfigID       primitive.ObjectID     `bson:"config_id" json:"config_id"`
	DocumentIDs    []primitive.ObjectID   `bson:"document_ids,omitempty" json:"document_ids,omitempty"`
	Filter         map[string]interface{} `bson:"filter,omitempty" json:"filter,omitempty"`
	
	// Job status
	Status         string                 `bson:"status" json:"status"` // "pending", "processing", "completed", "failed", "cancelled"
	Progress       int                    `bson:"progress" json:"progress"`
	TotalDocuments int                    `bson:"total_documents" json:"total_documents"`
	ProcessedDocs  int                    `bson:"processed_docs" json:"processed_docs"`
	FailedDocs     int                    `bson:"failed_docs" json:"failed_docs"`
	
	// Timing
	StartedAt      *time.Time             `bson:"started_at,omitempty" json:"started_at,omitempty"`
	CompletedAt    *time.Time             `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
	EstimatedTime  int                    `bson:"estimated_time_seconds,omitempty" json:"estimated_time_seconds,omitempty"`
	
	// Error handling
	Errors         []JobError             `bson:"errors,omitempty" json:"errors,omitempty"`
	RetryCount     int                    `bson:"retry_count" json:"retry_count"`
	MaxRetries     int                    `bson:"max_retries" json:"max_retries"`
	
	CreatedBy      primitive.ObjectID     `bson:"created_by" json:"created_by"`
	CreatedAt      time.Time              `bson:"created_at" json:"created_at"`
}

// JobError represents an error during job processing
type JobError struct {
	DocumentID primitive.ObjectID `bson:"document_id,omitempty" json:"document_id,omitempty"`
	Error      string             `bson:"error" json:"error"`
	Timestamp  time.Time          `bson:"timestamp" json:"timestamp"`
}

// RAGQuery represents a RAG query request on collection documents
type RAGQuery struct {
	Query           string                 `json:"query" validate:"required"`
	CollectionName  string                 `json:"collection_name" validate:"required"`
	
	// Search parameters
	TopK            int                    `json:"top_k,omitempty"`
	MinScore        float32                `json:"min_score,omitempty"`
	MaxDistance     float32                `json:"max_distance,omitempty"`
	
	// Filtering
	Filter          map[string]interface{} `json:"filter,omitempty"`
	FieldsToReturn  []string               `json:"fields_to_return,omitempty"`
	
	// Search options
	IncludeVector   bool                   `json:"include_vector,omitempty"`
	IncludeScore    bool                   `json:"include_score,omitempty"`
	IncludeDistance bool                   `json:"include_distance,omitempty"`
	
	// Hybrid search
	HybridSearch    bool                   `json:"hybrid_search,omitempty"`
	HybridAlpha     float32                `json:"hybrid_alpha,omitempty"` // 0 = pure vector, 1 = pure keyword
	KeywordFields   []string               `json:"keyword_fields,omitempty"`
	
	// Reranking
	Rerank          bool                   `json:"rerank,omitempty"`
	RerankModel     string                 `json:"rerank_model,omitempty"`
}

// RAGResponse represents the response from a RAG query
type RAGResponse struct {
	Query      string                 `json:"query"`
	Documents  []RAGDocument          `json:"documents"`
	TotalFound int                    `json:"total_found"`
	Provider   string                 `json:"provider"`
	Model      string                 `json:"model"`
	SearchType string                 `json:"search_type"` // "vector", "hybrid", "keyword"
	Latency    int64                  `json:"latency_ms"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// RAGDocument represents a document in RAG results
type RAGDocument struct {
	ID         primitive.ObjectID     `json:"id"`
	Data       map[string]interface{} `json:"data"`
	Score      float32                `json:"score,omitempty"`
	Distance   float32                `json:"distance,omitempty"`
	Relevance  float32                `json:"relevance,omitempty"`
	Highlights map[string][]string    `json:"highlights,omitempty"`
}