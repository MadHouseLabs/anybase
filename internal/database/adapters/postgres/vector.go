package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/madhouselabs/anybase/pkg/models"
)

// VectorOperations handles vector-specific database operations
type VectorOperations struct {
	db        *sql.DB
	tableName string
}

// NewVectorOperations creates a new vector operations handler
func NewVectorOperations(db *sql.DB, tableName string) *VectorOperations {
	return &VectorOperations{
		db:        db,
		tableName: tableName,
	}
}

// CreateVectorColumn adds a vector column to the table
func (v *VectorOperations) CreateVectorColumn(ctx context.Context, field models.VectorField) error {
	// Create column name with vec_ prefix to avoid conflicts
	columnName := fmt.Sprintf("vec_%s", field.Name)
	
	// Add vector column
	query := fmt.Sprintf(`
		ALTER TABLE %s 
		ADD COLUMN IF NOT EXISTS %s vector(%d)
	`, v.tableName, columnName, field.Dimensions)
	
	if _, err := v.db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("failed to create vector column: %w", err)
	}
	
	// Create index if specified
	if field.IndexType != "" {
		return v.CreateVectorIndex(ctx, field)
	}
	
	return nil
}

// CreateVectorIndex creates an index on a vector column
func (v *VectorOperations) CreateVectorIndex(ctx context.Context, field models.VectorField) error {
	columnName := fmt.Sprintf("vec_%s", field.Name)
	indexName := fmt.Sprintf("idx_%s_%s", v.tableName, columnName)
	
	// Determine operator class based on metric
	var opClass string
	switch field.Metric {
	case "cosine":
		opClass = "vector_cosine_ops"
	case "l2":
		opClass = "vector_l2_ops"
	case "inner_product", "ip":
		opClass = "vector_ip_ops"
	default:
		opClass = "vector_cosine_ops" // default to cosine
	}
	
	var query string
	switch field.IndexType {
	case "ivfflat":
		// Set list size (default 100 if not specified)
		listSize := field.ListSize
		if listSize == 0 {
			listSize = 100
		}
		query = fmt.Sprintf(`
			CREATE INDEX IF NOT EXISTS %s 
			ON %s 
			USING ivfflat (%s %s) 
			WITH (lists = %d)
		`, indexName, v.tableName, columnName, opClass, listSize)
		
	case "hnsw":
		// Set HNSW parameters (defaults if not specified)
		m := field.M
		if m == 0 {
			m = 16
		}
		efConstruct := field.EfConstruct
		if efConstruct == 0 {
			efConstruct = 64
		}
		query = fmt.Sprintf(`
			CREATE INDEX IF NOT EXISTS %s 
			ON %s 
			USING hnsw (%s %s) 
			WITH (m = %d, ef_construction = %d)
		`, indexName, v.tableName, columnName, opClass, m, efConstruct)
		
	default:
		// No index or unknown type
		return nil
	}
	
	if _, err := v.db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("failed to create vector index: %w", err)
	}
	
	return nil
}

// DropVectorColumn removes a vector column from the table
func (v *VectorOperations) DropVectorColumn(ctx context.Context, fieldName string) error {
	columnName := fmt.Sprintf("vec_%s", fieldName)
	
	query := fmt.Sprintf(`
		ALTER TABLE %s 
		DROP COLUMN IF EXISTS %s
	`, v.tableName, columnName)
	
	if _, err := v.db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("failed to drop vector column: %w", err)
	}
	
	return nil
}

// VectorSearch performs similarity search on a vector field
func (v *VectorOperations) VectorSearch(ctx context.Context, fieldName string, queryVector []float32, limit int, metric string) (*sql.Rows, error) {
	columnName := fmt.Sprintf("vec_%s", fieldName)
	
	// Convert vector to PostgreSQL array format
	vectorStr := v.vectorToString(queryVector)
	
	// Determine distance operator based on metric
	var operator string
	switch metric {
	case "cosine":
		operator = "<=>"
	case "l2":
		operator = "<->"
	case "inner_product", "ip":
		operator = "<#>"
	default:
		operator = "<=>" // default to cosine
	}
	
	// Build query
	query := fmt.Sprintf(`
		SELECT _id, data, %s %s '%s'::vector as distance
		FROM %s
		WHERE %s IS NOT NULL
		  AND _deleted_at IS NULL
		ORDER BY %s %s '%s'::vector
		LIMIT $1
	`, columnName, operator, vectorStr, v.tableName, columnName, columnName, operator, vectorStr)
	
	rows, err := v.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}
	
	return rows, nil
}

// HybridSearch performs combined text and vector search
func (v *VectorOperations) HybridSearch(ctx context.Context, fieldName string, queryVector []float32, textQuery string, limit int, alpha float32) (*sql.Rows, error) {
	columnName := fmt.Sprintf("vec_%s", fieldName)
	vectorStr := v.vectorToString(queryVector)
	
	// Alpha controls the weight between text search (0) and vector search (1)
	// Combined score = (1-alpha) * text_score + alpha * (1 - vector_distance)
	query := fmt.Sprintf(`
		WITH text_search AS (
			SELECT _id, data, 
				   ts_rank_cd(to_tsvector('english', data::text), plainto_tsquery('english', $2)) as text_score
			FROM %s
			WHERE _deleted_at IS NULL
			  AND to_tsvector('english', data::text) @@ plainto_tsquery('english', $2)
		),
		vector_search AS (
			SELECT _id, data,
				   1 - (%s <=> '%s'::vector) as vector_score
			FROM %s
			WHERE %s IS NOT NULL
			  AND _deleted_at IS NULL
		),
		combined AS (
			SELECT 
				COALESCE(t._id, v._id) as _id,
				COALESCE(t.data, v.data) as data,
				COALESCE(t.text_score, 0) * (1 - $3) + COALESCE(v.vector_score, 0) * $3 as combined_score
			FROM text_search t
			FULL OUTER JOIN vector_search v ON t._id = v._id
		)
		SELECT _id, data, combined_score
		FROM combined
		ORDER BY combined_score DESC
		LIMIT $1
	`, v.tableName, columnName, vectorStr, v.tableName, columnName)
	
	rows, err := v.db.QueryContext(ctx, query, limit, textQuery, alpha)
	if err != nil {
		return nil, fmt.Errorf("hybrid search failed: %w", err)
	}
	
	return rows, nil
}

// vectorToString converts a float32 slice to PostgreSQL vector string format
func (v *VectorOperations) vectorToString(vector []float32) string {
	strVals := make([]string, len(vector))
	for i, val := range vector {
		strVals[i] = fmt.Sprintf("%f", val)
	}
	return "[" + strings.Join(strVals, ",") + "]"
}

// GetVectorColumns returns all vector columns for the table
func (v *VectorOperations) GetVectorColumns(ctx context.Context) ([]string, error) {
	query := `
		SELECT column_name 
		FROM information_schema.columns 
		WHERE table_name = $1 
		  AND data_type = 'USER-DEFINED'
		  AND udt_name = 'vector'
		  AND column_name LIKE 'vec_%'
	`
	
	rows, err := v.db.QueryContext(ctx, query, v.tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get vector columns: %w", err)
	}
	defer rows.Close()
	
	var columns []string
	for rows.Next() {
		var colName string
		if err := rows.Scan(&colName); err != nil {
			return nil, err
		}
		// Remove vec_ prefix
		columns = append(columns, strings.TrimPrefix(colName, "vec_"))
	}
	
	return columns, nil
}