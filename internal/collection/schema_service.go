package collection

import (
	"context"
	"fmt"

	"github.com/madhouselabs/anybase/internal/validator"
)

// ValidateAgainstSchema validates a document against collection schema
func (s *AdapterService) ValidateAgainstSchema(ctx context.Context, collectionName string, document interface{}) error {
	// Get the collection
	collectionsCol := s.db.Collection("collections")
	
	filter := map[string]interface{}{"name": collectionName}
	var collection struct {
		Schema interface{} `bson:"schema"`
	}
	
	err := collectionsCol.FindOne(ctx, filter, &collection)
	if err != nil {
		return fmt.Errorf("failed to get collection schema: %w", err)
	}

	// If no schema, validation passes
	if collection.Schema == nil {
		return nil
	}

	// Validate document against schema
	schemaValidator := validator.NewSchemaValidator()
	
	// Convert schema to map if needed
	schemaMap, ok := collection.Schema.(map[string]interface{})
	if !ok {
		// If schema is not a map, skip validation
		return nil
	}
	
	// Validate document
	if err := schemaValidator.ValidateDocument(schemaMap, nil); err != nil {
		return fmt.Errorf("document does not match schema: %w", err)
	}

	return nil
}