package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/madhouselabs/anybase/pkg/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PostgresCursor wraps SQL rows to implement types.Cursor
type PostgresCursor struct {
	rows *sql.Rows
}

// Next advances the cursor to the next document
func (c *PostgresCursor) Next(ctx context.Context) bool {
	return c.rows.Next()
}

// Decode decodes the current document into result
func (c *PostgresCursor) Decode(result interface{}) error {
	var id uuid.UUID
	var data json.RawMessage
	var createdBy, updatedBy sql.NullString
	var createdAt, updatedAt sql.NullTime
	var version int
	
	err := c.rows.Scan(&id, &data, &createdBy, &updatedBy, &createdAt, &updatedAt, &version)
	if err != nil {
		return err
	}
	
	// First unmarshal into a map to get all data including _id
	var dataMap map[string]interface{}
	if err := json.Unmarshal(data, &dataMap); err != nil {
		return err
	}
	
	// Unmarshal into result
	if err := json.Unmarshal(data, result); err != nil {
		return err
	}
	
	// Handle specific types
	if userPtr, ok := result.(*models.User); ok {
		// Handle User type - similar to FindOne
		// Manually set the password field since it has json:"-" tag
		if pwd, ok := dataMap["password"].(string); ok {
			userPtr.Password = pwd
		}
		
		// Get the _id from the data if it exists
		if idStr, ok := dataMap["_id"].(string); ok {
			if objID, err := primitive.ObjectIDFromHex(idStr); err == nil {
				userPtr.ID = objID
			}
		}
		
		// If we still don't have an ID, generate one (shouldn't happen with fixed Create)
		if userPtr.ID.IsZero() {
			userPtr.ID = primitive.NewObjectID()
		}
		
		// Store the actual UUID in metadata for reference
		if userPtr.Metadata == nil {
			userPtr.Metadata = make(map[string]interface{})
		}
		userPtr.Metadata["_postgres_id"] = id.String()
		
		if createdAt.Valid {
			userPtr.CreatedAt = createdAt.Time
		}
		if updatedAt.Valid {
			userPtr.UpdatedAt = updatedAt.Time
		}
	} else if docPtr, ok := result.(*models.Document); ok {
		// Handle Document type specifically
		// Get the _id from the JSONB data
		if idStr, ok := dataMap["_id"].(string); ok {
			if objID, err := primitive.ObjectIDFromHex(idStr); err == nil {
				docPtr.ID = objID
			}
		}
		
		// Handle both old (nested) and new (flat) structures
		if nestedData, ok := dataMap["data"].(map[string]interface{}); ok {
			// Old structure: data is nested in "data" field
			docPtr.Data = nestedData
			
			// Extract metadata from the old structure
			// Collection info is in the old structure but not needed for Document type
			if cb, ok := dataMap["created_by"].(string); ok {
				if objID, err := primitive.ObjectIDFromHex(cb); err == nil {
					docPtr.CreatedBy = objID
				}
			}
			if ub, ok := dataMap["updated_by"].(string); ok {
				if objID, err := primitive.ObjectIDFromHex(ub); err == nil {
					docPtr.UpdatedBy = objID
				}
			}
		} else {
			// New structure: data fields are at top level
			cleanData := make(map[string]interface{})
			for k, v := range dataMap {
				// Skip metadata fields that shouldn't be in data
				if k != "_id" && k != "collection" && k != "created_by" && 
				   k != "updated_by" && k != "created_at" && k != "updated_at" &&
				   k != "_created_at" && k != "_updated_at" && k != "_version" {
					cleanData[k] = v
				}
			}
			docPtr.Data = cleanData
		}
		
		// Set created_by and updated_by from column values
		if createdBy.Valid && createdBy.String != "" {
			if objID, err := primitive.ObjectIDFromHex(createdBy.String); err == nil {
				docPtr.CreatedBy = objID
			} else {
				// If not a valid ObjectID, generate one
				docPtr.CreatedBy = primitive.NewObjectID()
			}
		}
		
		if updatedBy.Valid && updatedBy.String != "" {
			if objID, err := primitive.ObjectIDFromHex(updatedBy.String); err == nil {
				docPtr.UpdatedBy = objID
			} else {
				// If not a valid ObjectID, generate one
				docPtr.UpdatedBy = primitive.NewObjectID()
			}
		}
		
		// Use timestamps from database columns
		if createdAt.Valid {
			docPtr.CreatedAt = createdAt.Time
		}
		if updatedAt.Valid {
			docPtr.UpdatedAt = updatedAt.Time
		}
		docPtr.Version = version
	} else if keyPtr, ok := result.(*models.AccessKey); ok {
		// Handle AccessKey type
		// Get the _id from the JSONB data
		if idStr, ok := dataMap["_id"].(string); ok {
			if objID, err := primitive.ObjectIDFromHex(idStr); err == nil {
				keyPtr.ID = objID
			}
		}
		
		// If we still don't have an ID, generate one
		if keyPtr.ID.IsZero() {
			keyPtr.ID = primitive.NewObjectID()
		}
		
		// Store the actual UUID in a temporary field for reference if needed
		// (AccessKey doesn't have a Metadata field)
		
		if createdAt.Valid {
			keyPtr.CreatedAt = createdAt.Time
		}
		if updatedAt.Valid {
			keyPtr.UpdatedAt = updatedAt.Time
		}
	} else if m, ok := result.(*map[string]interface{}); ok {
		// Add system fields if result is a map
		// Use the _id from JSONB data if it exists
		if idStr, ok := dataMap["_id"].(string); ok {
			(*m)["_id"] = idStr
		} else {
			(*m)["_id"] = id.String()
		}
		if createdAt.Valid {
			(*m)["_created_at"] = createdAt.Time
		}
		if updatedAt.Valid {
			(*m)["_updated_at"] = updatedAt.Time
		}
		(*m)["_version"] = version
	}
	
	return nil
}

// Close closes the cursor
func (c *PostgresCursor) Close(ctx context.Context) error {
	return c.rows.Close()
}

// All decodes all remaining documents into results
func (c *PostgresCursor) All(ctx context.Context, results interface{}) error {
	// This would need reflection to handle different result types
	// For now, handle the common case of []map[string]interface{}
	if docs, ok := results.(*[]map[string]interface{}); ok {
		*docs = []map[string]interface{}{}
		
		for c.rows.Next() {
			var doc map[string]interface{}
			if err := c.Decode(&doc); err != nil {
				return err
			}
			*docs = append(*docs, doc)
		}
		
		return nil
	}
	
	return fmt.Errorf("unsupported result type for All(): %T", results)
}