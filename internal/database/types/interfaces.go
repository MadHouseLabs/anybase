package types

import (
	"context"
	"time"
)

// DB represents the main database connection interface
type DB interface {
	// Connection management
	Connect(ctx context.Context) error
	Close(ctx context.Context) error
	Ping(ctx context.Context) error
	
	// Collection/Table operations (collections map to tables with JSONB data field)
	Collection(name string) Collection
	CreateCollection(ctx context.Context, name string, schema interface{}) error
	DropCollection(ctx context.Context, name string) error
	ListCollections(ctx context.Context) ([]string, error)
	
	// Transaction support
	BeginTransaction(ctx context.Context) (Transaction, error)
	RunInTransaction(ctx context.Context, fn func(ctx context.Context, tx Transaction) error) error
	
	// Database info
	Type() string // Returns "mongodb" or "postgres"
}

// Collection represents a collection/table interface
// For PostgreSQL, this maps to a table with structure: (id, data JSONB, created_at, updated_at)
type Collection interface {
	// CRUD operations using JSONB queries
	InsertOne(ctx context.Context, document map[string]interface{}) (ID, error)
	InsertMany(ctx context.Context, documents []map[string]interface{}) ([]ID, error)
	FindOne(ctx context.Context, filter map[string]interface{}, result interface{}) error
	Find(ctx context.Context, filter map[string]interface{}, options *FindOptions) (Cursor, error)
	UpdateOne(ctx context.Context, filter map[string]interface{}, update map[string]interface{}) (*UpdateResult, error)
	UpdateMany(ctx context.Context, filter map[string]interface{}, update map[string]interface{}) (*UpdateResult, error)
	DeleteOne(ctx context.Context, filter map[string]interface{}) (*DeleteResult, error)
	DeleteMany(ctx context.Context, filter map[string]interface{}) (*DeleteResult, error)
	CountDocuments(ctx context.Context, filter map[string]interface{}) (int64, error)
	
	// Index operations (creates GIN indexes on JSONB paths for PostgreSQL)
	CreateIndex(ctx context.Context, index Index) error
	DropIndex(ctx context.Context, name string) error
	ListIndexes(ctx context.Context) ([]Index, error)
	
	// Aggregation (translates to JSONB operations in PostgreSQL)
	Aggregate(ctx context.Context, pipeline []map[string]interface{}) (Cursor, error)
}

// Transaction represents a database transaction
type Transaction interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	Collection(name string) Collection
	Context() context.Context
}

// Cursor represents a result cursor
type Cursor interface {
	Next(ctx context.Context) bool
	Decode(result interface{}) error
	Close(ctx context.Context) error
	All(ctx context.Context, results interface{}) error
}

// FindOptions represents query options
type FindOptions struct {
	Limit      *int64
	Skip       *int64
	Sort       map[string]int         // field -> 1 (asc) or -1 (desc)
	Projection map[string]int         // field -> 1 (include) or 0 (exclude)
}

// UpdateResult represents the result of an update operation
type UpdateResult struct {
	MatchedCount  int64
	ModifiedCount int64
	UpsertedCount int64
	UpsertedID    ID
}

// DeleteResult represents the result of a delete operation
type DeleteResult struct {
	DeletedCount int64
}

// Index represents a database index
// For PostgreSQL, this creates GIN indexes on JSONB paths
type Index struct {
	Name   string         `json:"name"`
	Keys   map[string]int `json:"keys"` // For PostgreSQL: creates GIN index on data->'field'
	Unique bool           `json:"unique"`
	Sparse bool           `json:"sparse"`
	TTL    *time.Duration `json:"ttl"` // For PostgreSQL: handled via trigger
}

// ID represents a generic database ID
type ID interface {
	String() string
	Bytes() []byte
	IsZero() bool
	MarshalJSON() ([]byte, error)
	UnmarshalJSON([]byte) error
	Equals(other ID) bool
}

// QueryOperators that work consistently across MongoDB and PostgreSQL JSONB
const (
	OpEqual        = "$eq"
	OpNotEqual     = "$ne"
	OpGreater      = "$gt"
	OpGreaterEqual = "$gte"
	OpLess         = "$lt"
	OpLessEqual    = "$lte"
	OpIn           = "$in"
	OpNotIn        = "$nin"
	OpExists       = "$exists"
	OpRegex        = "$regex"
	OpAnd          = "$and"
	OpOr           = "$or"
	OpNot          = "$not"
)

// UpdateOperators that work consistently across both databases
const (
	OpSet       = "$set"
	OpUnset     = "$unset"
	OpInc       = "$inc"
	OpPush      = "$push"
	OpPull      = "$pull"
	OpAddToSet  = "$addToSet"
)

// Helper functions for building queries that work with both databases
type QueryHelper interface {
	// Builds a filter that works with both MongoDB and PostgreSQL JSONB
	BuildFilter(conditions map[string]interface{}) map[string]interface{}
	
	// Builds an update that works with both databases
	BuildUpdate(operations map[string]interface{}) map[string]interface{}
	
	// Converts between MongoDB aggregation pipeline and PostgreSQL JSONB operations
	BuildAggregation(pipeline []map[string]interface{}) []map[string]interface{}
}