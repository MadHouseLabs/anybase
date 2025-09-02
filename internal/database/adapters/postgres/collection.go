package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/madhouselabs/anybase/internal/database/types"
	"github.com/madhouselabs/anybase/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PostgresCollection wraps PostgreSQL table operations to implement types.Collection
type PostgresCollection struct {
	db        *sql.DB
	name      string
	tableName string
}

// InsertOne inserts a single document
func (c *PostgresCollection) InsertOne(ctx context.Context, document map[string]interface{}) (types.ID, error) {
	// Generate ID if not provided
	var id uuid.UUID
	var mongoID string
	
	if docID, ok := document["_id"]; ok {
		// Store the MongoDB-style ID (hex string) for application compatibility
		if strID, ok := docID.(string); ok {
			mongoID = strID
			// Try to parse as UUID for PostgreSQL _id column
			parsedID, err := uuid.Parse(strID)
			if err == nil {
				id = parsedID
			} else {
				// Not a UUID, generate a new one for PostgreSQL
				id = uuid.New()
			}
		} else {
			id = uuid.New()
			mongoID = id.String()
		}
	} else {
		id = uuid.New()
		// Generate a MongoDB-style ObjectID hex string
		mongoID = primitive.NewObjectID().Hex()
	}
	
	// Extract metadata fields - store as strings for MongoDB ObjectID compatibility
	var createdBy, updatedBy *string
	if cb, ok := document["created_by"]; ok {
		if strID, ok := cb.(string); ok {
			createdBy = &strID
		}
	}
	
	if ub, ok := document["updated_by"]; ok {
		if strID, ok := ub.(string); ok {
			updatedBy = &strID
		}
	}
	
	// Create a clean data object with only actual data fields
	dataOnly := make(map[string]interface{})
	for k, v := range document {
		// Extract only the data field if it exists, otherwise exclude metadata
		if k == "data" {
			// If there's a nested data field, use it
			if dataMap, ok := v.(map[string]interface{}); ok {
				dataOnly = dataMap
			}
		} else if k != "_id" && k != "collection" && k != "created_by" && 
		          k != "updated_by" && k != "created_at" && k != "updated_at" && 
		          k != "_created_at" && k != "_updated_at" && k != "_version" && 
		          k != "_deleted_at" {
			// Include non-metadata fields
			dataOnly[k] = v
		}
	}
	
	// If dataOnly is still empty but we have document data, use the whole document minus metadata
	if len(dataOnly) == 0 && len(document) > 0 {
		for k, v := range document {
			if k != "_id" && k != "collection" && k != "created_by" && 
			   k != "updated_by" && k != "created_at" && k != "updated_at" && 
			   k != "_created_at" && k != "_updated_at" && k != "_version" && 
			   k != "_deleted_at" {
				dataOnly[k] = v
			}
		}
	}
	
	// Add MongoDB ID to data for compatibility
	dataOnly["_id"] = mongoID
	
	// Convert data to JSON
	data, err := json.Marshal(dataOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal document: %w", err)
	}
	
	query := fmt.Sprintf(`
		INSERT INTO %s (_id, data, _created_by, _updated_by, _created_at, _updated_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING _id
	`, c.tableName)
	
	var returnedID uuid.UUID
	err = c.db.QueryRowContext(ctx, query, id, data, createdBy, updatedBy).Scan(&returnedID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert document: %w", err)
	}
	
	// Return the MongoDB-style ID for compatibility
	objID, _ := primitive.ObjectIDFromHex(mongoID)
	return types.FromObjectID(objID), nil
}

// InsertMany inserts multiple documents
func (c *PostgresCollection) InsertMany(ctx context.Context, documents []map[string]interface{}) ([]types.ID, error) {
	if len(documents) == 0 {
		return []types.ID{}, nil
	}
	
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	ids := make([]types.ID, len(documents))
	
	for i, doc := range documents {
		// Generate ID if not provided
		var id uuid.UUID
		var mongoID string
		
		if docID, ok := doc["_id"]; ok {
			// Store the MongoDB-style ID (hex string) for application compatibility
			if strID, ok := docID.(string); ok {
				mongoID = strID
				// Try to parse as UUID for PostgreSQL _id column
				parsedID, err := uuid.Parse(strID)
				if err == nil {
					id = parsedID
				} else {
					// Not a UUID, generate a new one for PostgreSQL
					id = uuid.New()
				}
			} else {
				id = uuid.New()
				mongoID = id.String()
			}
			// Keep the _id in the document for compatibility
			// delete(doc, "_id") -- Don't delete, keep for MongoDB compatibility
		} else {
			id = uuid.New()
			mongoID = id.String()
			// Add the _id to the document
			doc["_id"] = mongoID
		}
		
		data, err := json.Marshal(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal document %d: %w", i, err)
		}
		
		query := fmt.Sprintf(`
			INSERT INTO %s (_id, data, _created_at, _updated_at)
			VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`, c.tableName)
		
		if _, err := tx.ExecContext(ctx, query, id, data); err != nil {
			return nil, fmt.Errorf("failed to insert document %d: %w", i, err)
		}
		
		ids[i] = types.FromUUID(id)
	}
	
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return ids, nil
}

// FindOne finds a single document
func (c *PostgresCollection) FindOne(ctx context.Context, filter map[string]interface{}, result interface{}) error {
	where, args := c.buildWhereClause(filter)
	
	query := fmt.Sprintf(`
		SELECT _id, data, _created_by, _updated_by, _created_at, _updated_at, _version
		FROM %s
		WHERE _deleted_at IS NULL %s
		LIMIT 1
	`, c.tableName, where)
	
	var id uuid.UUID
	var data json.RawMessage
	var createdBy, updatedBy sql.NullString
	var createdAt, updatedAt sql.NullTime
	var version int
	
	err := c.db.QueryRowContext(ctx, query, args...).Scan(&id, &data, &createdBy, &updatedBy, &createdAt, &updatedAt, &version)
	if err == sql.ErrNoRows {
		return types.ErrNoDocuments
	}
	if err != nil {
		return err
	}
	
	// First unmarshal into a map to get all data including _id
	var dataMap map[string]interface{}
	if err := json.Unmarshal(data, &dataMap); err != nil {
		return err
	}
	
	// Now unmarshal into the target result
	if err := json.Unmarshal(data, result); err != nil {
		return err
	}
	
	// Handle special cases for known models
	if userPtr, ok := result.(*models.User); ok {
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
			
			// Extract metadata from the old structure if present
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
				// If not a valid ObjectID, use a zero value
				docPtr.CreatedBy = primitive.NilObjectID
			}
		}
		
		if updatedBy.Valid && updatedBy.String != "" {
			if objID, err := primitive.ObjectIDFromHex(updatedBy.String); err == nil {
				docPtr.UpdatedBy = objID
			} else {
				// If not a valid ObjectID, use a zero value
				docPtr.UpdatedBy = primitive.NilObjectID
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
	} else if providerPtr, ok := result.(*models.AIProvider); ok {
		// Handle AIProvider type
		// Get the _id from the JSONB data
		if idStr, ok := dataMap["_id"].(string); ok {
			if objID, err := primitive.ObjectIDFromHex(idStr); err == nil {
				providerPtr.ID = objID
			}
		}
		
		// If we still don't have an ID, generate one
		if providerPtr.ID.IsZero() {
			providerPtr.ID = primitive.NewObjectID()
		}
		
		// Set created_by from JSONB data
		if cbStr, ok := dataMap["created_by"].(string); ok {
			if objID, err := primitive.ObjectIDFromHex(cbStr); err == nil {
				providerPtr.CreatedBy = objID
			}
		}
		
		// Manually set APIKeyHash since it has json:"-" tag
		if keyHash, ok := dataMap["api_key_hash"].(string); ok {
			providerPtr.APIKeyHash = keyHash
		}
		
		// Use timestamps from database columns
		if createdAt.Valid {
			providerPtr.CreatedAt = createdAt.Time
		}
		if updatedAt.Valid {
			providerPtr.UpdatedAt = updatedAt.Time
		}
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
		
		// Manually set KeyHash since it has json:"-" tag
		if keyHash, ok := dataMap["key_hash"].(string); ok {
			keyPtr.KeyHash = keyHash
		}
		
		// Store the actual UUID in a temporary field for reference if needed
		// (AccessKey doesn't have a Metadata field)
		
		if createdAt.Valid {
			keyPtr.CreatedAt = createdAt.Time
		}
		if updatedAt.Valid {
			keyPtr.UpdatedAt = updatedAt.Time
		}
	} else if colPtr, ok := result.(*models.Collection); ok {
		// Handle Collection type
		// Get the _id from the JSONB data
		if idStr, ok := dataMap["_id"].(string); ok {
			if objID, err := primitive.ObjectIDFromHex(idStr); err == nil {
				colPtr.ID = objID
			}
		}
		
		// If we still don't have an ID, generate one
		if colPtr.ID.IsZero() {
			colPtr.ID = primitive.NewObjectID()
		}
		
		// Set created_by from JSONB data
		if cbStr, ok := dataMap["created_by"].(string); ok {
			if objID, err := primitive.ObjectIDFromHex(cbStr); err == nil {
				colPtr.CreatedBy = objID
			}
		}
		
		// Use timestamps from database columns
		if createdAt.Valid {
			colPtr.CreatedAt = createdAt.Time
		}
		if updatedAt.Valid {
			colPtr.UpdatedAt = updatedAt.Time
		}
	}
	
	return nil
}

// Find finds multiple documents
func (c *PostgresCollection) Find(ctx context.Context, filter map[string]interface{}, opts *types.FindOptions) (types.Cursor, error) {
	where, args := c.buildWhereClause(filter)
	
	query := fmt.Sprintf(`
		SELECT _id, data, _created_by, _updated_by, _created_at, _updated_at, _version
		FROM %s
		WHERE _deleted_at IS NULL %s
	`, c.tableName, where)
	
	// Add sorting
	if opts != nil && opts.Sort != nil {
		orderBy := c.buildOrderBy(opts.Sort)
		if orderBy != "" {
			query += " ORDER BY " + orderBy
		}
	}
	
	// Add limit and offset
	if opts != nil {
		if opts.Limit != nil {
			query += fmt.Sprintf(" LIMIT %d", *opts.Limit)
		}
		if opts.Skip != nil {
			query += fmt.Sprintf(" OFFSET %d", *opts.Skip)
		}
	}
	
	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	
	return &PostgresCursor{rows: rows}, nil
}

// UpdateOne updates a single document
func (c *PostgresCollection) UpdateOne(ctx context.Context, filter map[string]interface{}, update map[string]interface{}) (*types.UpdateResult, error) {
	// Handle MongoDB update operators
	var setOps map[string]interface{}
	hasSet := false
	
	// Check if this is a MongoDB-style update with operators
	for key := range update {
		if strings.HasPrefix(key, "$") {
			// This is an update operator
			if key == "$set" {
				// Handle both map[string]interface{} and primitive.M/bson.M types
				switch v := update["$set"].(type) {
				case map[string]interface{}:
					setOps = v
					hasSet = true
				case bson.M:
					// Convert bson.M (same as primitive.M) to map[string]interface{}
					setOps = make(map[string]interface{})
					for k, val := range v {
						setOps[k] = val
					}
					hasSet = true
				}
			}
			// Could handle other operators like $inc, $push, etc. here
		}
	}
	
	// If no operators found, treat entire update as $set
	if !hasSet {
		setOps = update
	}
	
	where, args := c.buildWhereClause(filter)
	
	// Build the update query with deep merge support
	// Pass len(args)+1 as the starting index for update parameters
	updateClause, updateArgs := c.buildJSONBUpdateClause(setOps, len(args)+1)
	args = append(args, updateArgs...)
	
	var query string
	if updateClause != "" {
		query = fmt.Sprintf(`
			UPDATE %s
			SET %s,
			    _updated_at = CURRENT_TIMESTAMP
			WHERE _deleted_at IS NULL %s
		`, c.tableName, updateClause, where)
	} else {
		// Only updating timestamp if no data changes
		query = fmt.Sprintf(`
			UPDATE %s
			SET _updated_at = CURRENT_TIMESTAMP
			WHERE _deleted_at IS NULL %s
		`, c.tableName, where)
	}
	
	result, err := c.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	
	affected, _ := result.RowsAffected()
	
	return &types.UpdateResult{
		MatchedCount:  affected,
		ModifiedCount: affected,
	}, nil
}

// ReplaceOne replaces a single document entirely
func (c *PostgresCollection) ReplaceOne(ctx context.Context, filter map[string]interface{}, replacement map[string]interface{}) (*types.UpdateResult, error) {
	// Extract metadata fields from replacement
	var createdBy, updatedBy *string
	if cb, ok := replacement["_created_by"]; ok {
		if strID, ok := cb.(string); ok {
			createdBy = &strID
		}
		delete(replacement, "_created_by")
	}
	if ub, ok := replacement["_updated_by"]; ok {
		if strID, ok := ub.(string); ok {
			updatedBy = &strID
		}
		delete(replacement, "_updated_by")
	}
	
	// Ensure _id is in the replacement data
	if id, ok := replacement["_id"]; ok {
		// Make sure it's in the data
		if _, isString := id.(string); !isString {
			if objID, ok := id.(primitive.ObjectID); ok {
				replacement["_id"] = objID.Hex()
			}
		}
	}
	
	// Remove other metadata that shouldn't be in JSONB
	delete(replacement, "_created_at")
	delete(replacement, "_updated_at")
	delete(replacement, "_version")
	delete(replacement, "collection")
	delete(replacement, "created_at")
	delete(replacement, "updated_at")
	delete(replacement, "created_by")
	delete(replacement, "updated_by")
	
	// Convert replacement to JSON
	data, err := json.Marshal(replacement)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal replacement: %w", err)
	}
	
	where, args := c.buildWhereClause(filter)
	
	// Build update query
	setClauses := []string{
		fmt.Sprintf("data = $%d::jsonb", len(args)+1),
	}
	args = append(args, string(data))
	
	// Add metadata column updates if present
	if createdBy != nil {
		setClauses = append(setClauses, fmt.Sprintf("_created_by = $%d", len(args)+1))
		args = append(args, createdBy)
	}
	if updatedBy != nil {
		setClauses = append(setClauses, fmt.Sprintf("_updated_by = $%d", len(args)+1))
		args = append(args, updatedBy)
	}
	
	query := fmt.Sprintf(`
		UPDATE %s
		SET %s,
		    _updated_at = CURRENT_TIMESTAMP,
		    _version = _version + 1
		WHERE _deleted_at IS NULL %s
	`, c.tableName, strings.Join(setClauses, ", "), where)
	
	result, err := c.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	
	affected, _ := result.RowsAffected()
	
	return &types.UpdateResult{
		MatchedCount:  affected,
		ModifiedCount: affected,
	}, nil
}

// UpdateMany updates multiple documents
func (c *PostgresCollection) UpdateMany(ctx context.Context, filter map[string]interface{}, update map[string]interface{}) (*types.UpdateResult, error) {
	// Handle $set operations
	setOps, hasSet := update["$set"].(map[string]interface{})
	if !hasSet {
		setOps = update
	}
	
	where, args := c.buildWhereClause(filter)
	
	// Build the update query with deep merge support
	// Pass len(args)+1 as the starting index for update parameters
	updateClause, updateArgs := c.buildJSONBUpdateClause(setOps, len(args)+1)
	args = append(args, updateArgs...)
	
	var query string
	if updateClause != "" {
		query = fmt.Sprintf(`
			UPDATE %s
			SET %s,
			    _updated_at = CURRENT_TIMESTAMP
			WHERE _deleted_at IS NULL %s
		`, c.tableName, updateClause, where)
	} else {
		// Only updating timestamp if no data changes
		query = fmt.Sprintf(`
			UPDATE %s
			SET _updated_at = CURRENT_TIMESTAMP
			WHERE _deleted_at IS NULL %s
		`, c.tableName, where)
	}
	
	result, err := c.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	
	affected, _ := result.RowsAffected()
	
	return &types.UpdateResult{
		MatchedCount:  affected,
		ModifiedCount: affected,
	}, nil
}

// DeleteOne deletes a single document
func (c *PostgresCollection) DeleteOne(ctx context.Context, filter map[string]interface{}) (*types.DeleteResult, error) {
	where, args := c.buildWhereClause(filter)
	
	// Soft delete by default
	query := fmt.Sprintf(`
		UPDATE %s
		SET _deleted_at = CURRENT_TIMESTAMP
		WHERE _deleted_at IS NULL %s
	`, c.tableName, where)
	
	result, err := c.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	
	affected, _ := result.RowsAffected()
	
	return &types.DeleteResult{
		DeletedCount: affected,
	}, nil
}

// DeleteMany deletes multiple documents
func (c *PostgresCollection) DeleteMany(ctx context.Context, filter map[string]interface{}) (*types.DeleteResult, error) {
	where, args := c.buildWhereClause(filter)
	
	query := fmt.Sprintf(`
		UPDATE %s
		SET _deleted_at = CURRENT_TIMESTAMP
		WHERE _deleted_at IS NULL %s
	`, c.tableName, where)
	
	result, err := c.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	
	affected, _ := result.RowsAffected()
	
	return &types.DeleteResult{
		DeletedCount: affected,
	}, nil
}

// CountDocuments counts documents matching the filter
func (c *PostgresCollection) CountDocuments(ctx context.Context, filter map[string]interface{}) (int64, error) {
	where, args := c.buildWhereClause(filter)
	
	query := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s
		WHERE _deleted_at IS NULL %s
	`, c.tableName, where)
	
	var count int64
	err := c.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

// CreateIndex creates an index on the collection
func (c *PostgresCollection) CreateIndex(ctx context.Context, index types.Index) error {
	// Check if there are any keys to index
	if len(index.Keys) == 0 {
		return fmt.Errorf("no fields specified for index")
	}
	
	// Build index creation query for JSONB fields
	// We need to preserve field order for compound indexes
	// Convert map to sorted slice for consistent ordering
	type fieldOrder struct {
		field     string
		direction int
	}
	
	var fields []fieldOrder
	for field, direction := range index.Keys {
		fields = append(fields, fieldOrder{field, direction})
	}
	
	// Sort fields by name for consistent ordering
	// This ensures compound indexes are created with predictable field order
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].field < fields[j].field
	})
	
	var indexParts []string
	for _, f := range fields {
		// For JSONB fields, use the ->> operator to extract text
		jsonPath := fmt.Sprintf("(data->>'%s')", f.field)
		
		order := ""
		if f.direction < 0 {
			order = " DESC"
		}
		
		indexParts = append(indexParts, fmt.Sprintf("%s%s", jsonPath, order))
	}
	
	indexName := index.Name
	if indexName == "" {
		indexName = fmt.Sprintf("idx_%s_%d", c.tableName, len(indexParts))
	}
	
	// Use B-tree for unique indexes (GIN doesn't support unique)
	// Use B-tree for regular indexes too for better performance on equality/range queries
	indexType := "USING btree"
	
	// For full-text search or contains queries, we might want GIN
	// But for now, stick with btree for all JSONB field indexes
	
	query := fmt.Sprintf("CREATE ")
	if index.Unique {
		query += "UNIQUE "
	}
	query += fmt.Sprintf("INDEX IF NOT EXISTS %s ON %s %s (%s)",
		indexName, c.tableName, indexType, strings.Join(indexParts, ", "))
	
	_, err := c.db.ExecContext(ctx, query)
	return err
}

// DropIndex drops an index from the collection
func (c *PostgresCollection) DropIndex(ctx context.Context, name string) error {
	query := fmt.Sprintf("DROP INDEX IF EXISTS %s", name)
	_, err := c.db.ExecContext(ctx, query)
	return err
}

// ListIndexes lists all indexes on the collection
func (c *PostgresCollection) ListIndexes(ctx context.Context) ([]types.Index, error) {
	query := `
		SELECT indexname, indexdef
		FROM pg_indexes
		WHERE tablename = $1
	`
	
	rows, err := c.db.QueryContext(ctx, query, c.tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var indexes []types.Index
	for rows.Next() {
		var name, def string
		if err := rows.Scan(&name, &def); err != nil {
			continue
		}
		
		// Initialize the index with an empty Keys map
		keys := make(map[string]int)
		
		// Parse the index definition to extract field information
		// Example formats:
		// CREATE INDEX customer_idx ON public.data_orders USING gin (((data -> 'customer_id'::text)))
		// CREATE UNIQUE INDEX data_orders_pkey ON public.data_orders USING btree (_id)
		// CREATE INDEX status_total_idx ON public.data_orders USING btree (((data -> 'status'::text)), ((data -> 'total'::text)))
		
		if strings.Contains(def, "data ->") || strings.Contains(def, "data->>") {
			// JSONB field index - extract field names from 'fieldname'::text patterns
			re := regexp.MustCompile(`'([^']+)'::text`)
			matches := re.FindAllStringSubmatch(def, -1)
			for _, match := range matches {
				if len(match) > 1 {
					fieldName := match[1]
					// Check if DESC is specified for this field
					if strings.Contains(def, "'" + fieldName + "'::text)) DESC") ||
					   strings.Contains(def, "'" + fieldName + "'::text) DESC") {
						keys[fieldName] = -1
					} else {
						keys[fieldName] = 1
					}
				}
			}
		} else {
			// Regular column index - extract from USING btree/gin (column_name)
			// Find the content between the parentheses after USING
			if idx := strings.Index(def, "USING"); idx >= 0 {
				afterUsing := def[idx+5:]
				if parenStart := strings.Index(afterUsing, "("); parenStart >= 0 {
					parenEnd := strings.LastIndex(afterUsing, ")")
					if parenEnd > parenStart {
						fieldsStr := afterUsing[parenStart+1:parenEnd]
						// Split by comma for compound indexes
						fields := strings.Split(fieldsStr, ",")
						for _, field := range fields {
							field = strings.TrimSpace(field)
							// Remove any extra parentheses
							field = strings.Trim(field, "()")
							if field != "" {
								// Check for DESC
								if strings.HasSuffix(field, " DESC") {
									fieldName := strings.TrimSuffix(field, " DESC")
									keys[fieldName] = -1
								} else {
									// Remove ASC if present
									fieldName := strings.TrimSuffix(field, " ASC")
									keys[fieldName] = 1
								}
							}
						}
					}
				}
			}
		}
		
		// If we couldn't parse any keys, try to infer from the index name
		if len(keys) == 0 {
			if strings.Contains(name, "_pkey") {
				keys["_id"] = 1
			} else if strings.Contains(name, "created_at") {
				keys["_created_at"] = 1
			} else if strings.Contains(name, "updated_at") {
				keys["_updated_at"] = 1
			} else if strings.Contains(name, "deleted_at") {
				keys["_deleted_at"] = 1
			} else if strings.Contains(name, "_data") {
				keys["data"] = 1
			} else {
				// Default: use the index name as a hint
				keys["unknown"] = 1
			}
		}
		
		idx := types.Index{
			Name:   name,
			Unique: strings.Contains(def, "UNIQUE"),
			Keys:   keys,
			Sparse: false,
			TTL:    nil,
		}
		
		indexes = append(indexes, idx)
	}
	
	return indexes, nil
}

// Aggregate performs aggregation on the collection
func (c *PostgresCollection) Aggregate(ctx context.Context, pipeline []map[string]interface{}) (types.Cursor, error) {
	// For now, return an error as aggregation is complex to implement
	// This would require translating MongoDB aggregation pipeline to SQL
	return nil, fmt.Errorf("aggregation not yet implemented for PostgreSQL")
}

// buildJSONBUpdateClause builds an UPDATE clause that handles nested JSONB fields
// It starts with the current argIndex from the WHERE clause
func (c *PostgresCollection) buildJSONBUpdateClause(updates map[string]interface{}, startArgIndex int) (string, []interface{}) {
	var setClauses []string
	var args []interface{}
	argIndex := startArgIndex
	
	// Extract metadata fields that should be updated in columns
	var updatedBy *string
	if ub, ok := updates["updated_by"]; ok {
		if strID, ok := ub.(string); ok {
			updatedBy = &strID
		}
		delete(updates, "updated_by")
	}
	
	// Add updated_by column update if present
	if updatedBy != nil {
		setClauses = append(setClauses, fmt.Sprintf("_updated_by = $%d", argIndex))
		args = append(args, updatedBy)
		argIndex++
	}
	
	// Remove other metadata fields that shouldn't be in data
	delete(updates, "created_by")
	delete(updates, "created_at")
	delete(updates, "updated_at")
	delete(updates, "_created_at")
	delete(updates, "_updated_at")
	delete(updates, "_version")
	delete(updates, "collection")
	
	// Separate nested and top-level updates for data field
	topLevel := make(map[string]interface{})
	nested := make(map[string]map[string]interface{})
	
	for key, value := range updates {
		// Check if this is a nested path (e.g., "settings.versioning")
		if parts := strings.Split(key, "."); len(parts) > 1 {
			// Handle nested field
			topKey := parts[0]
			if nested[topKey] == nil {
				nested[topKey] = make(map[string]interface{})
			}
			// Build nested path recursively
			current := nested[topKey]
			for i := 1; i < len(parts)-1; i++ {
				if current[parts[i]] == nil {
					current[parts[i]] = make(map[string]interface{})
				}
				current = current[parts[i]].(map[string]interface{})
			}
			current[parts[len(parts)-1]] = value
		} else {
			// Top-level field
			topLevel[key] = value
		}
	}
	
	// Handle nested updates using jsonb_set
	if len(nested) > 0 {
		dataUpdate := "data"
		for topKey, nestedMap := range nested {
			// For each top-level key with nested updates, use jsonb_set
			nestedJSON, _ := json.Marshal(nestedMap)
			dataUpdate = fmt.Sprintf(
				"jsonb_set(%s, '{%s}', COALESCE(%s->'%s', '{}')::jsonb || $%d::jsonb)",
				dataUpdate, topKey, dataUpdate, topKey, argIndex,
			)
			args = append(args, string(nestedJSON))
			argIndex++
		}
		
		// If we have top-level updates too, merge them
		if len(topLevel) > 0 {
			topJSON, _ := json.Marshal(topLevel)
			dataUpdate = fmt.Sprintf("%s || $%d::jsonb", dataUpdate, argIndex)
			args = append(args, string(topJSON))
		}
		
		setClauses = append(setClauses, fmt.Sprintf("data = %s", dataUpdate))
	} else if len(topLevel) > 0 {
		// Only top-level updates - use simple merge
		topJSON, _ := json.Marshal(topLevel)
		setClauses = append(setClauses, fmt.Sprintf("data = data || $%d::jsonb", argIndex))
		args = append(args, string(topJSON))
		argIndex++ // Increment for consistency, even though not used after this
	}
	
	// If no setClauses were added but we have updates, ensure we at least update the data
	if len(setClauses) == 0 && len(updates) > 0 {
		// This shouldn't happen normally, but handle it
		updateJSON, _ := json.Marshal(updates)
		setClauses = append(setClauses, fmt.Sprintf("data = data || $%d::jsonb", argIndex))
		args = append(args, string(updateJSON))
	}
	
	return strings.Join(setClauses, ", "), args
}

// buildWhereClause builds a WHERE clause from a filter
func (c *PostgresCollection) buildWhereClause(filter map[string]interface{}) (string, []interface{}) {
	if len(filter) == 0 {
		return "", nil
	}
	
	var conditions []string
	var args []interface{}
	argIndex := 1
	
	for key, value := range filter {
		// Handle special fields
		if key == "_id" {
			// For MongoDB compatibility, check the _id in JSONB data first
			// This allows queries with MongoDB ObjectID hex strings
			conditions = append(conditions, fmt.Sprintf("(data->>'_id' = $%d)", argIndex))
			// Keep as string for JSONB comparison
			if strID, ok := value.(string); ok {
				args = append(args, strID)
			} else if objID, ok := value.(primitive.ObjectID); ok {
				args = append(args, objID.Hex())
			} else {
				args = append(args, fmt.Sprintf("%v", value))
			}
			argIndex++
		} else {
			// For JSONB fields
			// Handle nil values specially - check for NULL or missing field
			if value == nil {
				conditions = append(conditions, fmt.Sprintf("(data->>'%s' IS NULL OR NOT (data ? '%s'))", key, key))
				// No argument to add for NULL check
			} else {
				// Check if value is a simple type (string, number, bool)
				switch v := value.(type) {
				case string, int, int64, float64, bool:
					conditions = append(conditions, fmt.Sprintf("data->>'%s' = $%d", key, argIndex))
					args = append(args, v)
					argIndex++
				default:
					jsonValue, _ := json.Marshal(value)
					conditions = append(conditions, fmt.Sprintf("data @> jsonb_build_object('%s', $%d::jsonb)", key, argIndex))
					args = append(args, jsonValue)
					argIndex++
				}
			}
		}
	}
	
	if len(conditions) > 0 {
		return " AND " + strings.Join(conditions, " AND "), args
	}
	
	return "", nil
}

// buildOrderBy builds an ORDER BY clause
func (c *PostgresCollection) buildOrderBy(sort map[string]int) string {
	if len(sort) == 0 {
		return ""
	}
	
	var parts []string
	for field, direction := range sort {
		order := "ASC"
		if direction < 0 {
			order = "DESC"
		}
		
		// Handle system fields
		if strings.HasPrefix(field, "_") {
			parts = append(parts, fmt.Sprintf("%s %s", field, order))
		} else {
			// JSONB fields
			parts = append(parts, fmt.Sprintf("data->>'%s' %s", field, order))
		}
	}
	
	return strings.Join(parts, ", ")
}

// GetVectorOperations returns a VectorOperations instance for this collection
func (c *PostgresCollection) GetVectorOperations() *VectorOperations {
	return NewVectorOperations(c.db, c.tableName)
}