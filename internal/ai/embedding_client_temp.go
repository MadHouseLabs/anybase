package ai

import (
	"github.com/madhouselabs/anybase/pkg/models"
)

// GetEmbeddingModel returns the embedding model for a provider
// This is a temporary helper function until we refactor the embedding client
func GetEmbeddingModel(provider *models.AIProvider) string {
	if provider.Models != nil && provider.Models.EmbeddingModel != "" {
		return provider.Models.EmbeddingModel
	}
	
	// Return default model based on provider type
	switch provider.Type {
	case models.ProviderOpenAI:
		return "text-embedding-ada-002"
	case models.ProviderCohere:
		return "embed-english-v3.0"
	case models.ProviderHuggingFace:
		return "sentence-transformers/all-MiniLM-L6-v2"
	case models.ProviderOllama:
		return "nomic-embed-text"
	default:
		return "default-embedding-model"
	}
}

// GetEmbeddingDimensions returns the embedding dimensions for a provider
func GetEmbeddingDimensions(provider *models.AIProvider) int {
	if provider.Models != nil && provider.Models.EmbeddingDim > 0 {
		return provider.Models.EmbeddingDim
	}
	
	// Return default dimensions based on provider type
	switch provider.Type {
	case models.ProviderOpenAI:
		return 1536
	case models.ProviderCohere:
		return 1024
	case models.ProviderHuggingFace:
		return 384
	case models.ProviderOllama:
		return 768
	default:
		return 768
	}
}