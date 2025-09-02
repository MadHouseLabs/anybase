package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/madhouselabs/anybase/pkg/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// JobProcessor handles background processing of embedding jobs
type JobProcessor struct {
	service Service
	running bool
}

// NewJobProcessor creates a new job processor
func NewJobProcessor(s Service) *JobProcessor {
	return &JobProcessor{
		service: s,
		running: false,
	}
}

// Start begins processing jobs
func (p *JobProcessor) Start() {
	if p.running {
		return
	}
	
	p.running = true
	go p.processLoop()
}

// Stop stops the job processor
func (p *JobProcessor) Stop() {
	p.running = false
}

// processLoop continuously processes pending jobs
func (p *JobProcessor) processLoop() {
	for p.running {
		ctx := context.Background()
		
		// Get pending jobs
		jobs, err := p.service.GetRepository().GetPendingJobs(ctx, 5)
		if err != nil {
			fmt.Printf("Error getting pending jobs: %v\n", err)
			time.Sleep(10 * time.Second)
			continue
		}
		
		if len(jobs) == 0 {
			// No jobs, wait before checking again
			time.Sleep(5 * time.Second)
			continue
		}
		
		// Process each job
		for _, job := range jobs {
			if !p.running {
				break
			}
			p.ProcessJob(job.ID)
		}
	}
}

// ProcessJob processes a single embedding job
func (p *JobProcessor) ProcessJob(jobID primitive.ObjectID) {
	ctx := context.Background()
	
	// Get job details
	job, err := p.service.GetRepository().GetEmbeddingJob(ctx, jobID)
	if err != nil {
		fmt.Printf("Error getting job %s: %v\n", jobID.Hex(), err)
		return
	}
	
	// Check if job is still pending
	if job.Status != "pending" {
		return
	}
	
	// Update job status to processing
	now := time.Now().UTC()
	err = p.service.GetRepository().UpdateEmbeddingJob(ctx, jobID, map[string]interface{}{
		"status": "processing",
		"started_at": now,
	})
	if err != nil {
		fmt.Printf("Error updating job status: %v\n", err)
		return
	}
	
	// Get RAG configuration
	config, err := p.service.GetRepository().GetRAGConfigByID(ctx, job.ConfigID)
	if err != nil {
		p.failJob(ctx, job, fmt.Sprintf("Failed to get RAG config: %v", err))
		return
	}
	
	// Get provider
	provider, err := p.service.GetRepository().GetProvider(ctx, job.ProviderID)
	if err != nil {
		p.failJob(ctx, job, fmt.Sprintf("Failed to get provider: %v", err))
		return
	}
	
	// Process based on job type
	switch job.Type {
	case "full":
		p.processFullCollection(ctx, job, config, provider)
	case "incremental", "document":
		p.processDocuments(ctx, job, config, provider)
	default:
		p.failJob(ctx, job, fmt.Sprintf("Unknown job type: %s", job.Type))
	}
}

// processFullCollection processes all documents in a collection
func (p *JobProcessor) processFullCollection(ctx context.Context, job *models.EmbeddingJob, config *models.CollectionRAGConfig, provider *models.AIProvider) {
	// Get all documents from collection
	query := &models.DataQuery{
		Collection: job.CollectionName,
		Limit:      100, // Process in batches
	}
	
	processedCount := 0
	failedCount := 0
	skip := 0
	
	for {
		query.Skip = skip
		
		// Get batch of documents
		documents, err := p.service.GetCollectionService().QueryDocuments(ctx, query)
		if err != nil {
			p.failJob(ctx, job, fmt.Sprintf("Failed to query documents: %v", err))
			return
		}
		
		if len(documents) == 0 {
			break // No more documents
		}
		
		// Process each document
		for _, doc := range documents {
			if !p.running {
				p.pauseJob(ctx, job, processedCount, failedCount)
				return
			}
			
			// Check for cancellation
			currentJob, _ := p.service.GetRepository().GetEmbeddingJob(ctx, job.ID)
			if currentJob != nil && currentJob.Status == "cancelled" {
				return
			}
			
			// Generate embedding for document
			err := p.processDocument(ctx, &doc, config, provider)
			if err != nil {
				failedCount++
				// Log error
				p.addJobError(ctx, job.ID, doc.ID, err.Error())
			} else {
				processedCount++
			}
			
			// Update progress
			progress := int(float64(processedCount+failedCount) / float64(job.TotalDocuments) * 100)
			p.service.GetRepository().UpdateEmbeddingJob(ctx, job.ID, map[string]interface{}{
				"progress": progress,
				"processed_docs": processedCount,
				"failed_docs": failedCount,
			})
			
			// Rate limiting
			time.Sleep(time.Duration(60000/provider.RateLimits.RequestsPerMinute) * time.Millisecond)
		}
		
		skip += len(documents)
	}
	
	// Mark job as completed
	completedAt := time.Now().UTC()
	status := "completed"
	if failedCount > 0 {
		status = "completed_with_errors"
	}
	
	p.service.GetRepository().UpdateEmbeddingJob(ctx, job.ID, map[string]interface{}{
		"status": status,
		"progress": 100,
		"processed_docs": processedCount,
		"failed_docs": failedCount,
		"completed_at": completedAt,
	})
	
	// Update RAG config stats
	p.service.GetRepository().UpdateRAGConfig(ctx, config.ID, map[string]interface{}{
		"last_processed": completedAt,
		"documents_processed": processedCount,
		"processing_status": status,
	})
}

// processDocuments processes specific documents
func (p *JobProcessor) processDocuments(ctx context.Context, job *models.EmbeddingJob, config *models.CollectionRAGConfig, provider *models.AIProvider) {
	processedCount := 0
	failedCount := 0
	
	for _, docID := range job.DocumentIDs {
		if !p.running {
			p.pauseJob(ctx, job, processedCount, failedCount)
			return
		}
		
		// Get document
		doc, err := p.service.GetCollectionService().GetDocument(ctx, primitive.NilObjectID, job.CollectionName, docID)
		if err != nil {
			failedCount++
			p.addJobError(ctx, job.ID, docID, fmt.Sprintf("Failed to get document: %v", err))
			continue
		}
		
		// Generate embedding
		err = p.processDocument(ctx, doc, config, provider)
		if err != nil {
			failedCount++
			p.addJobError(ctx, job.ID, docID, err.Error())
		} else {
			processedCount++
		}
		
		// Update progress
		progress := int(float64(processedCount+failedCount) / float64(len(job.DocumentIDs)) * 100)
		p.service.GetRepository().UpdateEmbeddingJob(ctx, job.ID, map[string]interface{}{
			"progress": progress,
			"processed_docs": processedCount,
			"failed_docs": failedCount,
		})
		
		// Rate limiting
		time.Sleep(time.Duration(60000/provider.RateLimits.RequestsPerMinute) * time.Millisecond)
	}
	
	// Mark job as completed
	completedAt := time.Now().UTC()
	status := "completed"
	if failedCount > 0 {
		status = "completed_with_errors"
	}
	
	p.service.GetRepository().UpdateEmbeddingJob(ctx, job.ID, map[string]interface{}{
		"status": status,
		"progress": 100,
		"processed_docs": processedCount,
		"failed_docs": failedCount,
		"completed_at": completedAt,
	})
}

// processDocument generates and stores embedding for a single document
func (p *JobProcessor) processDocument(ctx context.Context, doc *models.Document, config *models.CollectionRAGConfig, provider *models.AIProvider) error {
	// Extract content from document
	content := p.service.ExtractContent(doc.Data, config)
	if content == "" {
		return fmt.Errorf("no content to embed")
	}
	
	// Check if chunking is needed
	if config.EnableChunking && len(content) > config.ChunkSize {
		// TODO: Implement chunking
		// For now, truncate
		if len(content) > config.ChunkSize {
			content = content[:config.ChunkSize]
		}
	}
	
	// Generate embedding
	embedding, err := p.service.GetEmbeddingClient().GenerateEmbedding(ctx, provider, content)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}
	
	// Update document with embedding
	mutation := &models.DataMutation{
		Collection: doc.Collection,
		Operation:  "update",
		DocumentID: doc.ID,
		Data: map[string]interface{}{
			config.VectorFieldName: embedding,
		},
	}
	
	err = p.service.GetCollectionService().UpdateDocument(ctx, mutation)
	if err != nil {
		return fmt.Errorf("failed to update document with embedding: %w", err)
	}
	
	return nil
}

// failJob marks a job as failed
func (p *JobProcessor) failJob(ctx context.Context, job *models.EmbeddingJob, errorMsg string) {
	p.service.GetRepository().UpdateEmbeddingJob(ctx, job.ID, map[string]interface{}{
		"status": "failed",
		"completed_at": time.Now().UTC(),
		"errors": []models.JobError{{
			Error:     errorMsg,
			Timestamp: time.Now().UTC(),
		}},
	})
}

// pauseJob pauses a job
func (p *JobProcessor) pauseJob(ctx context.Context, job *models.EmbeddingJob, processed, failed int) {
	p.service.GetRepository().UpdateEmbeddingJob(ctx, job.ID, map[string]interface{}{
		"status": "paused",
		"processed_docs": processed,
		"failed_docs": failed,
	})
}

// addJobError adds an error to a job
func (p *JobProcessor) addJobError(ctx context.Context, jobID, docID primitive.ObjectID, errorMsg string) {
	job, err := p.service.GetRepository().GetEmbeddingJob(ctx, jobID)
	if err != nil {
		return
	}
	
	errors := job.Errors
	errors = append(errors, models.JobError{
		DocumentID: docID,
		Error:      errorMsg,
		Timestamp:  time.Now().UTC(),
	})
	
	// Keep only last 100 errors
	if len(errors) > 100 {
		errors = errors[len(errors)-100:]
	}
	
	p.service.GetRepository().UpdateEmbeddingJob(ctx, jobID, map[string]interface{}{
		"errors": errors,
	})
}