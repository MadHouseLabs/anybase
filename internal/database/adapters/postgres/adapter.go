package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/madhouselabs/anybase/internal/config"
	"github.com/madhouselabs/anybase/internal/database/types"
	_ "github.com/lib/pq"
)

// PostgresAdapter implements the types.DB interface for PostgreSQL
type PostgresAdapter struct {
	db       *sql.DB
	config   *config.DatabaseConfig
	database string
}

// NewPostgresAdapter creates a new PostgreSQL adapter
func NewPostgresAdapter(cfg *config.DatabaseConfig) *PostgresAdapter {
	return &PostgresAdapter{
		config:   cfg,
		database: cfg.Database,
	}
}

// Connect establishes connection to PostgreSQL
func (p *PostgresAdapter) Connect(ctx context.Context) error {
	// Parse connection string
	connStr := p.buildConnectionString()
	
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}
	
	// Configure connection pool
	db.SetMaxOpenConns(int(p.config.MaxPoolSize))
	db.SetMaxIdleConns(int(p.config.MinPoolSize))
	db.SetConnMaxIdleTime(p.config.MaxIdleTime)
	
	// Test connection
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}
	
	p.db = db
	
	// Initialize schema
	if err := p.initializeSchema(ctx); err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}
	
	return nil
}

// buildConnectionString builds PostgreSQL connection string
func (p *PostgresAdapter) buildConnectionString() string {
	// If URI is provided, use it directly
	if strings.HasPrefix(p.config.URI, "postgres://") || strings.HasPrefix(p.config.URI, "postgresql://") {
		return p.config.URI
	}
	
	// Build connection string from parts
	parts := []string{}
	
	// Parse URI for host, port, user, password
	if p.config.URI != "" {
		// Simple parsing for localhost connections
		if strings.Contains(p.config.URI, "localhost") {
			parts = append(parts, "host=localhost")
			parts = append(parts, "port=5432")
			parts = append(parts, "user=karthik") // Use current user
		}
	} else {
		parts = append(parts, "host=localhost")
		parts = append(parts, "port=5432")
	}
	
	if p.config.Database != "" {
		parts = append(parts, fmt.Sprintf("dbname=%s", p.config.Database))
	}
	
	if p.config.SSLMode != "" {
		parts = append(parts, fmt.Sprintf("sslmode=%s", p.config.SSLMode))
	} else {
		parts = append(parts, "sslmode=disable")
	}
	
	if p.config.ConnectTimeout > 0 {
		parts = append(parts, fmt.Sprintf("connect_timeout=%d", int(p.config.ConnectTimeout.Seconds())))
	}
	
	return strings.Join(parts, " ")
}

// initializeSchema creates the necessary tables and indexes
func (p *PostgresAdapter) initializeSchema(ctx context.Context) error {
	// Create collections metadata table
	_, err := p.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS _collections (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) UNIQUE NOT NULL,
			schema JSONB,
			settings JSONB,
			permissions JSONB,
			metadata JSONB,
			created_by UUID,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create collections table: %w", err)
	}
	
	// Create views metadata table
	_, err = p.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS _views (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) UNIQUE NOT NULL,
			collection VARCHAR(255) NOT NULL,
			pipeline JSONB,
			fields JSONB,
			filter JSONB,
			permissions JSONB,
			metadata JSONB,
			created_by UUID,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create views table: %w", err)
	}
	
	// Create standard system collections
	systemCollections := []string{
		"users",
		"sessions",
		"access_keys",
		"audit_logs",
		"settings",
		"collections",  // Add collections table for metadata storage
		"ai_providers",
		"rag_configs",
		"embedding_jobs",
	}
	
	for _, collection := range systemCollections {
		if err := p.ensureCollectionTable(ctx, collection); err != nil {
			return fmt.Errorf("failed to create %s table: %w", collection, err)
		}
	}
	
	return nil
}

// ensureCollectionTable creates a table for a collection if it doesn't exist
func (p *PostgresAdapter) ensureCollectionTable(ctx context.Context, name string) error {
	tableName := p.sanitizeTableName(name)
	
	// Create the collection table with JSONB data field
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			data JSONB NOT NULL DEFAULT '{}',
			_created_by TEXT,
			_updated_by TEXT,
			_created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			_updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			_version INTEGER DEFAULT 1,
			_deleted_at TIMESTAMP
		)
	`, tableName)
	
	if _, err := p.db.ExecContext(ctx, query); err != nil {
		return err
	}
	
	// Create indexes for system fields
	indexes := []string{
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_created_at ON %s(_created_at)", tableName, tableName),
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_updated_at ON %s(_updated_at)", tableName, tableName),
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_deleted_at ON %s(_deleted_at)", tableName, tableName),
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_data ON %s USING GIN(data)", tableName, tableName),
	}
	
	for _, idx := range indexes {
		if _, err := p.db.ExecContext(ctx, idx); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}
	
	// Create update trigger for updated_at
	triggerQuery := fmt.Sprintf(`
		CREATE OR REPLACE FUNCTION update_%s_updated_at()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW._updated_at = CURRENT_TIMESTAMP;
			NEW._version = OLD._version + 1;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;
		
		DROP TRIGGER IF EXISTS %s_updated_at ON %s;
		
		CREATE TRIGGER %s_updated_at
			BEFORE UPDATE ON %s
			FOR EACH ROW
			EXECUTE FUNCTION update_%s_updated_at();
	`, tableName, tableName, tableName, tableName, tableName, tableName)
	
	if _, err := p.db.ExecContext(ctx, triggerQuery); err != nil {
		// Trigger creation might fail if it already exists, which is okay
		// Silent fail to avoid noisy logs
	}
	
	return nil
}

// sanitizeTableName ensures table name is safe for PostgreSQL
func (p *PostgresAdapter) sanitizeTableName(name string) string {
	// Replace non-alphanumeric characters with underscores
	// PostgreSQL table names should be lowercase
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, ".", "_")
	return name
}

// GetIndexes returns the list of indexes for a table
func (p *PostgresAdapter) GetIndexes(ctx context.Context, tableName string) ([]types.Index, error) {
	sanitizedName := p.sanitizeTableName(tableName)
	
	query := `
		SELECT indexname, indexdef 
		FROM pg_indexes 
		WHERE tablename = $1 AND schemaname = 'public'
	`
	
	rows, err := p.db.QueryContext(ctx, query, sanitizedName)
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
		
		// Initialize the Keys map
		keys := make(map[string]int)
		
		// Parse the index definition to extract field information
		// Example formats:
		// CREATE INDEX customer_idx ON public.data_orders USING gin (((data -> 'customer_id'::text)))
		// CREATE UNIQUE INDEX data_orders_pkey ON public.data_orders USING btree (_id)
		// CREATE INDEX status_total_idx ON public.data_orders USING btree (((data -> 'status'::text)), ((data -> 'total'::text)))
		
		if strings.Contains(def, "data ->") || strings.Contains(def, "data->>") {
			// JSONB field index - extract field names using simple string parsing
			// Look for patterns like 'fieldname'::text
			start := 0
			for {
				idx := strings.Index(def[start:], "'")
				if idx == -1 {
					break
				}
				start = start + idx + 1
				end := strings.Index(def[start:], "'")
				if end == -1 {
					break
				}
				fieldName := def[start : start+end]
				// Skip if it's not followed by ::text (to avoid false positives)
				if strings.HasPrefix(def[start+end:], "'::text") {
					keys[fieldName] = 1
				}
				start = start + end + 1
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
			}
		}
		
		idx := types.Index{
			Name:   name,
			Keys:   keys,
			Unique: strings.Contains(def, "UNIQUE"),
			Sparse: false,
			TTL:    nil,
		}
		indexes = append(indexes, idx)
	}
	
	return indexes, nil
}

// CreateIndex creates an index on a table
func (p *PostgresAdapter) CreateIndex(ctx context.Context, tableName, indexName string, fields []string, unique bool) error {
	sanitizedTable := p.sanitizeTableName(tableName)
	sanitizedIndex := p.sanitizeTableName(indexName)
	
	indexType := ""
	if unique {
		indexType = "UNIQUE "
	}
	
	// For JSONB fields, use GIN index for better performance
	if len(fields) == 1 && strings.Contains(fields[0], "data->") {
		query := fmt.Sprintf(
			"CREATE %sINDEX IF NOT EXISTS %s ON %s USING GIN (%s)",
			indexType, sanitizedIndex, sanitizedTable, fields[0],
		)
		_, err := p.db.ExecContext(ctx, query)
		return err
	}
	
	// Regular index
	query := fmt.Sprintf(
		"CREATE %sINDEX IF NOT EXISTS %s ON %s (%s)",
		indexType, sanitizedIndex, sanitizedTable, strings.Join(fields, ", "),
	)
	
	_, err := p.db.ExecContext(ctx, query)
	return err
}

// DropIndex drops an index
func (p *PostgresAdapter) DropIndex(ctx context.Context, indexName string) error {
	sanitizedIndex := p.sanitizeTableName(indexName)
	query := fmt.Sprintf("DROP INDEX IF EXISTS %s", sanitizedIndex)
	_, err := p.db.ExecContext(ctx, query)
	return err
}

// Close closes the database connection
func (p *PostgresAdapter) Close(ctx context.Context) error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

// Ping checks if the database is reachable
func (p *PostgresAdapter) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

// Collection returns a collection wrapper
func (p *PostgresAdapter) Collection(name string) types.Collection {
	return &PostgresCollection{
		db:         p.db,
		name:       name,
		tableName:  p.sanitizeTableName(name),
	}
}

// CreateCollection creates a new collection (table)
func (p *PostgresAdapter) CreateCollection(ctx context.Context, name string, schema interface{}) error {
	// Create the table
	if err := p.ensureCollectionTable(ctx, name); err != nil {
		return err
	}
	
	// Convert schema to JSON if it's not nil
	var schemaJSON []byte
	if schema != nil {
		var err error
		schemaJSON, err = json.Marshal(schema)
		if err != nil {
			return fmt.Errorf("failed to marshal schema: %w", err)
		}
	}
	
	// Store collection metadata
	_, err := p.db.ExecContext(ctx, `
		INSERT INTO _collections (name, schema, created_at, updated_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (name) DO UPDATE SET
			schema = EXCLUDED.schema,
			updated_at = CURRENT_TIMESTAMP
	`, name, schemaJSON)
	
	return err
}

// DropCollection drops a collection (table)
func (p *PostgresAdapter) DropCollection(ctx context.Context, name string) error {
	tableName := p.sanitizeTableName(name)
	
	// Drop the table
	_, err := p.db.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", tableName))
	if err != nil {
		return err
	}
	
	// Remove from metadata
	_, err = p.db.ExecContext(ctx, "DELETE FROM _collections WHERE name = $1", name)
	return err
}

// ListCollections lists all collections
func (p *PostgresAdapter) ListCollections(ctx context.Context) ([]string, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_name NOT LIKE '\_%'
		ORDER BY table_name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var collections []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		collections = append(collections, name)
	}
	
	return collections, nil
}

// BeginTransaction starts a new transaction
func (p *PostgresAdapter) BeginTransaction(ctx context.Context) (types.Transaction, error) {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	
	return &PostgresTransaction{
		tx:      tx,
		adapter: p,
		ctx:     ctx,
	}, nil
}

// RunInTransaction executes a function within a transaction
func (p *PostgresAdapter) RunInTransaction(ctx context.Context, fn func(ctx context.Context, tx types.Transaction) error) error {
	tx, err := p.BeginTransaction(ctx)
	if err != nil {
		return err
	}
	
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(ctx)
			panic(r)
		}
	}()
	
	if err := fn(ctx, tx); err != nil {
		tx.Rollback(ctx)
		return err
	}
	
	return tx.Commit(ctx)
}

// Type returns the database type
func (p *PostgresAdapter) Type() string {
	return "postgres"
}

// GetDB returns the underlying database connection (for migration purposes)
func (p *PostgresAdapter) GetDB() *sql.DB {
	return p.db
}