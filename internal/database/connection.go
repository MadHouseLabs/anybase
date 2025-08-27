package database

import (
	"context"
	"fmt"
	"time"

	"github.com/karthik/anybase/internal/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Database struct {
	client   *mongo.Client
	database *mongo.Database
	config   *config.DatabaseConfig
}

var db *Database

// Initialize creates a new database connection
func Initialize(cfg *config.DatabaseConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().
		ApplyURI(cfg.URI).
		SetMaxPoolSize(cfg.MaxPoolSize).
		SetMinPoolSize(cfg.MinPoolSize).
		SetMaxConnIdleTime(cfg.MaxIdleTime).
		SetHeartbeatInterval(cfg.HeartbeatInterval).
		SetServerSelectionTimeout(cfg.ServerSelectionTimeout)

	if cfg.ReplicaSet != "" {
		clientOptions.SetReplicaSet(cfg.ReplicaSet)
	}

	// For AWS DocumentDB, disable retryable writes if needed
	if !cfg.RetryWrites {
		clientOptions.SetRetryWrites(false)
	}

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Ping the database to verify connection
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	db = &Database{
		client:   client,
		database: client.Database(cfg.Database),
		config:   cfg,
	}

	return nil
}

// GetDB returns the database instance
func GetDB() *Database {
	if db == nil {
		panic("database not initialized")
	}
	return db
}

// GetClient returns the MongoDB client
func (d *Database) GetClient() *mongo.Client {
	return d.client
}

// GetDatabase returns the database
func (d *Database) GetDatabase() *mongo.Database {
	return d.database
}

// Collection returns a collection from the database
func (d *Database) Collection(name string) *mongo.Collection {
	return d.database.Collection(name)
}

// Close closes the database connection
func (d *Database) Close(ctx context.Context) error {
	if d.client != nil {
		return d.client.Disconnect(ctx)
	}
	return nil
}

// Ping checks if the database is reachable
func (d *Database) Ping(ctx context.Context) error {
	return d.client.Ping(ctx, readpref.Primary())
}

// RunInTransaction executes a function within a transaction
func (d *Database) RunInTransaction(ctx context.Context, fn func(sessCtx mongo.SessionContext) error) error {
	session, err := d.client.StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		err := fn(sessCtx)
		return nil, err
	}

	_, err = session.WithTransaction(ctx, callback)
	return err
}

// CreateIndexes creates indexes for collections
func (d *Database) CreateIndexes(ctx context.Context) error {
	// Create indexes for users collection
	userIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{
				"email": 1,
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: map[string]interface{}{
				"username": 1,
			},
			Options: options.Index().SetUnique(true).SetSparse(true),
		},
		{
			Keys: map[string]interface{}{
				"created_at": -1,
			},
		},
	}

	if _, err := d.Collection("users").Indexes().CreateMany(ctx, userIndexes); err != nil {
		return fmt.Errorf("failed to create user indexes: %w", err)
	}

	// Create indexes for sessions collection
	sessionIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{
				"token": 1,
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: map[string]interface{}{
				"user_id": 1,
			},
		},
		{
			Keys: map[string]interface{}{
				"expires_at": 1,
			},
			Options: options.Index().SetExpireAfterSeconds(0),
		},
	}

	if _, err := d.Collection("sessions").Indexes().CreateMany(ctx, sessionIndexes); err != nil {
		return fmt.Errorf("failed to create session indexes: %w", err)
	}

	// Create indexes for audit_logs collection
	auditIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{
				"user_id": 1,
				"created_at": -1,
			},
		},
		{
			Keys: map[string]interface{}{
				"action": 1,
				"created_at": -1,
			},
		},
	}

	if _, err := d.Collection("audit_logs").Indexes().CreateMany(ctx, auditIndexes); err != nil {
		return fmt.Errorf("failed to create audit log indexes: %w", err)
	}

	return nil
}