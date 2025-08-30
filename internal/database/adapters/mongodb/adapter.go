package mongodb

import (
	"context"
	"fmt"

	"github.com/madhouselabs/anybase/internal/config"
	"github.com/madhouselabs/anybase/internal/database/types"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/bson"
)

// MongoAdapter implements the types.DB interface for MongoDB
type MongoAdapter struct {
	client   *mongo.Client
	database *mongo.Database
	config   *config.DatabaseConfig
}

// NewMongoAdapter creates a new MongoDB adapter
func NewMongoAdapter(cfg *config.DatabaseConfig) *MongoAdapter {
	return &MongoAdapter{
		config: cfg,
	}
}

// Connect establishes connection to MongoDB
func (m *MongoAdapter) Connect(ctx context.Context) error {
	clientOptions := options.Client().
		ApplyURI(m.config.URI).
		SetMaxPoolSize(m.config.MaxPoolSize).
		SetMinPoolSize(m.config.MinPoolSize).
		SetMaxConnIdleTime(m.config.MaxIdleTime).
		SetHeartbeatInterval(m.config.HeartbeatInterval).
		SetServerSelectionTimeout(m.config.ServerSelectionTimeout)

	if m.config.ReplicaSet != "" {
		clientOptions.SetReplicaSet(m.config.ReplicaSet)
	}

	// For AWS DocumentDB, disable retryable writes if needed
	if !m.config.RetryWrites {
		clientOptions.SetRetryWrites(false)
	}

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database to verify connection
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	m.client = client
	m.database = client.Database(m.config.Database)
	
	return nil
}

// Close closes the database connection
func (m *MongoAdapter) Close(ctx context.Context) error {
	if m.client != nil {
		return m.client.Disconnect(ctx)
	}
	return nil
}

// Ping checks if the database is reachable
func (m *MongoAdapter) Ping(ctx context.Context) error {
	return m.client.Ping(ctx, readpref.Primary())
}

// Collection returns a collection wrapper
func (m *MongoAdapter) Collection(name string) types.Collection {
	return &MongoCollection{
		collection: m.database.Collection(name),
		dbType:     "mongodb",
	}
}

// CreateCollection creates a new collection
func (m *MongoAdapter) CreateCollection(ctx context.Context, name string, schema interface{}) error {
	// MongoDB creates collections automatically, but we can set validation if schema is provided
	if schema != nil {
		// Convert schema to MongoDB validation rules if needed
		cmd := bson.D{
			{Key: "create", Value: name},
		}
		
		// Add validation if schema is provided as bson.M
		if validation, ok := schema.(bson.M); ok {
			cmd = append(cmd, bson.E{Key: "validator", Value: validation})
		}
		
		err := m.database.RunCommand(ctx, cmd).Err()
		if err != nil && !mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("failed to create collection: %w", err)
		}
	}
	
	return nil
}

// DropCollection drops a collection
func (m *MongoAdapter) DropCollection(ctx context.Context, name string) error {
	return m.database.Collection(name).Drop(ctx)
}

// ListCollections lists all collections
func (m *MongoAdapter) ListCollections(ctx context.Context) ([]string, error) {
	names, err := m.database.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}
	return names, nil
}

// BeginTransaction starts a new transaction
func (m *MongoAdapter) BeginTransaction(ctx context.Context) (types.Transaction, error) {
	session, err := m.client.StartSession()
	if err != nil {
		return nil, fmt.Errorf("failed to start session: %w", err)
	}
	
	if err := session.StartTransaction(); err != nil {
		session.EndSession(ctx)
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	
	return &MongoTransaction{
		session: session,
		adapter: m,
		ctx:     mongo.NewSessionContext(ctx, session),
	}, nil
}

// RunInTransaction executes a function within a transaction
func (m *MongoAdapter) RunInTransaction(ctx context.Context, fn func(ctx context.Context, tx types.Transaction) error) error {
	session, err := m.client.StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		tx := &MongoTransaction{
			session: session,
			adapter: m,
			ctx:     sessCtx,
		}
		err := fn(sessCtx, tx)
		return nil, err
	}

	_, err = session.WithTransaction(ctx, callback)
	return err
}

// Type returns the database type
func (m *MongoAdapter) Type() string {
	return "mongodb"
}

// GetClient returns the underlying MongoDB client (for migration purposes)
func (m *MongoAdapter) GetClient() *mongo.Client {
	return m.client
}

// GetDatabase returns the underlying MongoDB database (for migration purposes)
func (m *MongoAdapter) GetDatabase() *mongo.Database {
	return m.database
}

// CreateIndexes creates standard indexes for the system collections
func (m *MongoAdapter) CreateIndexes(ctx context.Context) error {
	// Create indexes for users collection
	userIndexes := []mongo.IndexModel{
		{
			Keys: bson.M{"email": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.M{"username": 1},
			Options: options.Index().SetUnique(true).SetSparse(true),
		},
		{
			Keys: bson.M{"created_at": -1},
		},
	}

	if _, err := m.database.Collection("users").Indexes().CreateMany(ctx, userIndexes); err != nil {
		return fmt.Errorf("failed to create user indexes: %w", err)
	}

	// Create indexes for sessions collection
	sessionIndexes := []mongo.IndexModel{
		{
			Keys: bson.M{"token": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.M{"user_id": 1},
		},
		{
			Keys: bson.M{"expires_at": 1},
			Options: options.Index().SetExpireAfterSeconds(0),
		},
	}

	if _, err := m.database.Collection("sessions").Indexes().CreateMany(ctx, sessionIndexes); err != nil {
		return fmt.Errorf("failed to create session indexes: %w", err)
	}

	// Create indexes for access_keys collection
	accessKeyIndexes := []mongo.IndexModel{
		{
			Keys: bson.M{"name": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.M{"key_hash": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.M{"owner_id": 1},
		},
	}

	if _, err := m.database.Collection("access_keys").Indexes().CreateMany(ctx, accessKeyIndexes); err != nil {
		return fmt.Errorf("failed to create access key indexes: %w", err)
	}

	// Create indexes for collections collection
	collectionIndexes := []mongo.IndexModel{
		{
			Keys: bson.M{"name": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.M{"created_by": 1},
		},
	}

	if _, err := m.database.Collection("collections").Indexes().CreateMany(ctx, collectionIndexes); err != nil {
		return fmt.Errorf("failed to create collection indexes: %w", err)
	}

	// Create indexes for views collection
	viewIndexes := []mongo.IndexModel{
		{
			Keys: bson.M{"name": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.M{"collection": 1},
		},
		{
			Keys: bson.M{"created_by": 1},
		},
	}

	if _, err := m.database.Collection("views").Indexes().CreateMany(ctx, viewIndexes); err != nil {
		return fmt.Errorf("failed to create view indexes: %w", err)
	}

	// Create indexes for audit_logs collection
	auditIndexes := []mongo.IndexModel{
		{
			Keys: bson.M{
				"user_id": 1,
				"created_at": -1,
			},
		},
		{
			Keys: bson.M{
				"action": 1,
				"created_at": -1,
			},
		},
	}

	if _, err := m.database.Collection("audit_logs").Indexes().CreateMany(ctx, auditIndexes); err != nil {
		return fmt.Errorf("failed to create audit log indexes: %w", err)
	}

	return nil
}