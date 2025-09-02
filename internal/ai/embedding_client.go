package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/madhouselabs/anybase/pkg/models"
)

// EmbeddingClient handles embedding generation from various providers
type EmbeddingClient interface {
	GenerateEmbedding(ctx context.Context, provider *models.AIProvider, text string) ([]float32, error)
	GenerateBatchEmbeddings(ctx context.Context, provider *models.AIProvider, texts []string) ([][]float32, error)
}

type embeddingClient struct {
	httpClient *http.Client
}

// NewEmbeddingClient creates a new embedding client
func NewEmbeddingClient() EmbeddingClient {
	return &embeddingClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *embeddingClient) GenerateEmbedding(ctx context.Context, provider *models.AIProvider, text string) ([]float32, error) {
	embeddings, err := c.GenerateBatchEmbeddings(ctx, provider, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embedding generated")
	}
	return embeddings[0], nil
}

func (c *embeddingClient) GenerateBatchEmbeddings(ctx context.Context, provider *models.AIProvider, texts []string) ([][]float32, error) {
	switch provider.Type {
	case models.ProviderOpenAI:
		return c.generateOpenAIEmbeddings(ctx, provider, texts)
	case models.ProviderCohere:
		return c.generateCohereEmbeddings(ctx, provider, texts)
	case models.ProviderHuggingFace:
		return c.generateHuggingFaceEmbeddings(ctx, provider, texts)
	case models.ProviderOllama:
		return c.generateOllamaEmbeddings(ctx, provider, texts)
	case models.ProviderCustom:
		return c.generateCustomEmbeddings(ctx, provider, texts)
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", provider.Type)
	}
}

// OpenAI implementation
func (c *embeddingClient) generateOpenAIEmbeddings(ctx context.Context, provider *models.AIProvider, texts []string) ([][]float32, error) {
	url := "https://api.openai.com/v1/embeddings"
	if provider.BaseURL != "" {
		url = provider.BaseURL + "/v1/embeddings"
	}
	
	requestBody := map[string]interface{}{
		"input": texts,
		"model": GetEmbeddingModel(provider),
	}
	
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+provider.APIKey)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}
	
	var response struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
		Usage struct {
			PromptTokens int `json:"prompt_tokens"`
			TotalTokens  int `json:"total_tokens"`
		} `json:"usage"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	embeddings := make([][]float32, len(response.Data))
	for i, data := range response.Data {
		embeddings[i] = data.Embedding
	}
	
	return embeddings, nil
}

// Cohere implementation
func (c *embeddingClient) generateCohereEmbeddings(ctx context.Context, provider *models.AIProvider, texts []string) ([][]float32, error) {
	url := "https://api.cohere.ai/v1/embed"
	if provider.BaseURL != "" {
		url = provider.BaseURL + "/v1/embed"
	}
	
	requestBody := map[string]interface{}{
		"texts": texts,
		"model": GetEmbeddingModel(provider),
		"input_type": "search_document",
	}
	
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+provider.APIKey)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Cohere API error (status %d): %s", resp.StatusCode, string(body))
	}
	
	var response struct {
		Embeddings [][]float32 `json:"embeddings"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return response.Embeddings, nil
}

// HuggingFace implementation
func (c *embeddingClient) generateHuggingFaceEmbeddings(ctx context.Context, provider *models.AIProvider, texts []string) ([][]float32, error) {
	url := fmt.Sprintf("https://api-inference.huggingface.co/models/%s", GetEmbeddingModel(provider))
	if provider.BaseURL != "" {
		url = provider.BaseURL
	}
	
	// HuggingFace API typically handles one text at a time for embeddings
	embeddings := make([][]float32, len(texts))
	
	for i, text := range texts {
		requestBody := map[string]interface{}{
			"inputs": text,
		}
		
		jsonBody, err := json.Marshal(requestBody)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+provider.APIKey)
		
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to send request: %w", err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("HuggingFace API error (status %d): %s", resp.StatusCode, string(body))
		}
		
		var embedding []float32
		if err := json.NewDecoder(resp.Body).Decode(&embedding); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		
		embeddings[i] = embedding
	}
	
	return embeddings, nil
}

// Ollama implementation
func (c *embeddingClient) generateOllamaEmbeddings(ctx context.Context, provider *models.AIProvider, texts []string) ([][]float32, error) {
	url := provider.BaseURL + "/api/embeddings"
	
	embeddings := make([][]float32, len(texts))
	
	for i, text := range texts {
		requestBody := map[string]interface{}{
			"model": GetEmbeddingModel(provider),
			"prompt": text,
		}
		
		jsonBody, err := json.Marshal(requestBody)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		
		req.Header.Set("Content-Type", "application/json")
		
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to send request: %w", err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("Ollama API error (status %d): %s", resp.StatusCode, string(body))
		}
		
		var response struct {
			Embedding []float32 `json:"embedding"`
		}
		
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		
		embeddings[i] = response.Embedding
	}
	
	return embeddings, nil
}

// Custom provider implementation
func (c *embeddingClient) generateCustomEmbeddings(ctx context.Context, provider *models.AIProvider, texts []string) ([][]float32, error) {
	// Custom providers should follow a standard API format
	// This is a generic implementation that can be customized
	
	url := provider.BaseURL + "/embeddings"
	
	requestBody := map[string]interface{}{
		"texts": texts,
		"model": GetEmbeddingModel(provider),
	}
	
	// Add any custom settings
	if provider.Settings != nil {
		for k, v := range provider.Settings {
			requestBody[k] = v
		}
	}
	
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	if provider.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+provider.APIKey)
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Custom API error (status %d): %s", resp.StatusCode, string(body))
	}
	
	var response struct {
		Embeddings [][]float32 `json:"embeddings"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return response.Embeddings, nil
}