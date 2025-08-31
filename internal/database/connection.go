package database

import (
	"context"
	"fmt"

	"github.com/madhouselabs/anybase/internal/config"
	"github.com/madhouselabs/anybase/internal/database/adapters/mongodb"
	"github.com/madhouselabs/anybase/internal/database/adapters/postgres"
	"github.com/madhouselabs/anybase/internal/database/types"
)

// Global database instance
var db types.DB

// Initialize creates and connects to the appropriate database based on configuration
func Initialize(cfg *config.DatabaseConfig) error {
	ctx := context.Background()
	
	var adapter types.DB
	
	fmt.Printf("Initializing database with type: %s\n", cfg.Type)
	
	switch cfg.Type {
	case "", "mongodb":
		// Default to MongoDB for backward compatibility
		fmt.Println("Using MongoDB adapter")
		adapter = mongodb.NewMongoAdapter(cfg)
		
	case "postgres", "postgresql":
		fmt.Println("Using PostgreSQL adapter")
		adapter = postgres.NewPostgresAdapter(cfg)
		
	default:
		return fmt.Errorf("unsupported database type: %s", cfg.Type)
	}
	
	// Connect to the database
	if err := adapter.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to %s: %w", cfg.Type, err)
	}
	
	// Create standard indexes
	if err := createStandardIndexes(ctx, adapter); err != nil {
		// Log warning but don't fail initialization
		fmt.Printf("Warning: Failed to create indexes: %v\n", err)
	}
	
	db = adapter
	return nil
}

// GetDB returns the current database adapter
func GetDB() types.DB {
	if db == nil {
		panic("database not initialized")
	}
	return db
}

// createStandardIndexes creates the standard system indexes
func createStandardIndexes(ctx context.Context, adapter types.DB) error {
	// Users collection indexes
	usersCol := adapter.Collection("users")
	userIndexes := []types.Index{
		{
			Name:   "email_unique",
			Keys:   map[string]int{"email": 1},
			Unique: true,
		},
		{
			Name: "created_at_desc",
			Keys: map[string]int{"created_at": -1},
		},
	}
	
	for _, idx := range userIndexes {
		if err := usersCol.CreateIndex(ctx, idx); err != nil {
			// Continue even if index creation fails (might already exist)
			fmt.Printf("Warning: Failed to create index %s on users: %v\n", idx.Name, err)
		}
	}
	
	// Access keys collection indexes
	accessKeysCol := adapter.Collection("access_keys")
	accessKeyIndexes := []types.Index{
		{
			Name:   "name_unique",
			Keys:   map[string]int{"name": 1},
			Unique: true,
		},
		{
			Name:   "key_hash_unique",
			Keys:   map[string]int{"key_hash": 1},
			Unique: true,
		},
		{
			Name: "owner_id",
			Keys: map[string]int{"owner_id": 1},
		},
	}
	
	for _, idx := range accessKeyIndexes {
		if err := accessKeysCol.CreateIndex(ctx, idx); err != nil {
			fmt.Printf("Warning: Failed to create index %s on access_keys: %v\n", idx.Name, err)
		}
	}
	
	// Collections collection indexes
	collectionsCol := adapter.Collection("collections")
	collectionIndexes := []types.Index{
		{
			Name:   "name_unique",
			Keys:   map[string]int{"name": 1},
			Unique: true,
		},
		{
			Name: "created_by",
			Keys: map[string]int{"created_by": 1},
		},
	}
	
	for _, idx := range collectionIndexes {
		if err := collectionsCol.CreateIndex(ctx, idx); err != nil {
			fmt.Printf("Warning: Failed to create index %s on collections: %v\n", idx.Name, err)
		}
	}
	
	// Views collection indexes
	viewsCol := adapter.Collection("views")
	viewIndexes := []types.Index{
		{
			Name:   "name_unique",
			Keys:   map[string]int{"name": 1},
			Unique: true,
		},
		{
			Name: "collection",
			Keys: map[string]int{"collection": 1},
		},
		{
			Name: "created_by",
			Keys: map[string]int{"created_by": 1},
		},
	}
	
	for _, idx := range viewIndexes {
		if err := viewsCol.CreateIndex(ctx, idx); err != nil {
			fmt.Printf("Warning: Failed to create index %s on views: %v\n", idx.Name, err)
		}
	}
	
	// Audit logs collection indexes
	auditLogsCol := adapter.Collection("audit_logs")
	auditIndexes := []types.Index{
		{
			Name: "user_created",
			Keys: map[string]int{
				"user_id":    1,
				"created_at": -1,
			},
		},
		{
			Name: "action_created",
			Keys: map[string]int{
				"action":     1,
				"created_at": -1,
			},
		},
	}
	
	for _, idx := range auditIndexes {
		if err := auditLogsCol.CreateIndex(ctx, idx); err != nil {
			fmt.Printf("Warning: Failed to create index %s on audit_logs: %v\n", idx.Name, err)
		}
	}
	
	// Sessions collection indexes with TTL
	sessionsCol := adapter.Collection("sessions")
	ttl := SessionTTL
	sessionIndexes := []types.Index{
		{
			Name:   "token_unique",
			Keys:   map[string]int{"token": 1},
			Unique: true,
		},
		{
			Name: "user_id",
			Keys: map[string]int{"user_id": 1},
		},
		{
			Name: "expires_at_ttl",
			Keys: map[string]int{"expires_at": 1},
			TTL:  &ttl,
		},
	}
	
	for _, idx := range sessionIndexes {
		if err := sessionsCol.CreateIndex(ctx, idx); err != nil {
			fmt.Printf("Warning: Failed to create index %s on sessions: %v\n", idx.Name, err)
		}
	}
	
	return nil
}