package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Collection represents a data collection in the system
type Collection struct {
	ID          primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	Name        string                 `bson:"name" json:"name" validate:"required,min=3,max=100"`
	Description string                 `bson:"description,omitempty" json:"description,omitempty"`
	Schema      *CollectionSchema      `bson:"schema,omitempty" json:"schema,omitempty"`
	Indexes     []CollectionIndex      `bson:"indexes,omitempty" json:"indexes,omitempty"`
	Permissions CollectionPermissions  `bson:"permissions" json:"permissions"`
	Settings    CollectionSettings     `bson:"settings" json:"settings"`
	Metadata    map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
	CreatedBy   primitive.ObjectID     `bson:"created_by" json:"created_by"`
	CreatedAt   time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time              `bson:"updated_at" json:"updated_at"`
}

// CollectionSchema defines the structure and validation rules using OpenAPI 3.0 format
type CollectionSchema struct {
	Type        string                            `bson:"type" json:"type"` // Should be "object"
	Properties  map[string]*SchemaProperty        `bson:"properties,omitempty" json:"properties,omitempty"`
	Required    []string                          `bson:"required,omitempty" json:"required,omitempty"`
	AdditionalProperties interface{}              `bson:"additionalProperties,omitempty" json:"additionalProperties,omitempty"`
	Description string                            `bson:"description,omitempty" json:"description,omitempty"`
}

// SchemaProperty represents a property in OpenAPI schema format
type SchemaProperty struct {
	Type        interface{}                       `bson:"type" json:"type"` // string, number, integer, boolean, object, array, null or array of types
	Format      string                            `bson:"format,omitempty" json:"format,omitempty"` // date-time, date, email, uri, etc.
	Description string                            `bson:"description,omitempty" json:"description,omitempty"`
	Default     interface{}                       `bson:"default,omitempty" json:"default,omitempty"`
	Enum        []interface{}                     `bson:"enum,omitempty" json:"enum,omitempty"`
	Pattern     string                            `bson:"pattern,omitempty" json:"pattern,omitempty"`
	MinLength   *int                              `bson:"minLength,omitempty" json:"minLength,omitempty"`
	MaxLength   *int                              `bson:"maxLength,omitempty" json:"maxLength,omitempty"`
	Minimum     *float64                          `bson:"minimum,omitempty" json:"minimum,omitempty"`
	Maximum     *float64                          `bson:"maximum,omitempty" json:"maximum,omitempty"`
	MinItems    *int                              `bson:"minItems,omitempty" json:"minItems,omitempty"`
	MaxItems    *int                              `bson:"maxItems,omitempty" json:"maxItems,omitempty"`
	UniqueItems bool                              `bson:"uniqueItems,omitempty" json:"uniqueItems,omitempty"`
	Items       *SchemaProperty                   `bson:"items,omitempty" json:"items,omitempty"` // For arrays
	Properties  map[string]*SchemaProperty        `bson:"properties,omitempty" json:"properties,omitempty"` // For nested objects
	Required    []string                          `bson:"required,omitempty" json:"required,omitempty"` // For nested objects
	ReadOnly    bool                              `bson:"readOnly,omitempty" json:"readOnly,omitempty"`
	WriteOnly   bool                              `bson:"writeOnly,omitempty" json:"writeOnly,omitempty"`
	Example     interface{}                       `bson:"example,omitempty" json:"example,omitempty"`
}

// CollectionIndex represents an index on the collection
type CollectionIndex struct {
	Name   string                 `bson:"name" json:"name"`
	Fields map[string]interface{} `bson:"fields" json:"fields"`
	Unique bool                   `bson:"unique" json:"unique"`
	Sparse bool                   `bson:"sparse" json:"sparse"`
}

// CollectionPermissions defines collection-level permissions
type CollectionPermissions struct {
	Read   PermissionRule `bson:"read" json:"read"`
	Write  PermissionRule `bson:"write" json:"write"`
	Update PermissionRule `bson:"update" json:"update"`
	Delete PermissionRule `bson:"delete" json:"delete"`
}

// PermissionRule defines who can perform an action
type PermissionRule struct {
	Roles      []string               `bson:"roles" json:"roles"`           // Allowed roles
	Users      []primitive.ObjectID   `bson:"users" json:"users"`           // Specific users
	Conditions map[string]interface{} `bson:"conditions" json:"conditions"` // Additional conditions
	Public     bool                   `bson:"public" json:"public"`         // Public access
}

// CollectionSettings contains collection configuration
type CollectionSettings struct {
	Versioning    bool `bson:"versioning" json:"versioning"`       // Enable document versioning
	SoftDelete    bool `bson:"soft_delete" json:"soft_delete"`     // Enable soft deletes
	Auditing      bool `bson:"auditing" json:"auditing"`           // Enable audit logging
	Encryption    bool `bson:"encryption" json:"encryption"`       // Enable field encryption
	MaxDocuments  int  `bson:"max_documents" json:"max_documents"` // Max documents (0 = unlimited)
	MaxSizeBytes  int  `bson:"max_size_bytes" json:"max_size_bytes"` // Max size in bytes
}

// View represents a filtered/transformed view of a collection
type View struct {
	ID          primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	Name        string                 `bson:"name" json:"name" validate:"required,min=3,max=100"`
	Description string                 `bson:"description,omitempty" json:"description,omitempty"`
	Collection  string                 `bson:"collection" json:"collection"`
	Pipeline    []bson.M               `bson:"pipeline" json:"pipeline"`           // MongoDB aggregation pipeline
	Fields      []string               `bson:"fields,omitempty" json:"fields"`     // Projected fields
	Filter      bson.M                 `bson:"filter,omitempty" json:"filter"`     // Query filter
	Sort        bson.M                 `bson:"sort,omitempty" json:"sort"`         // Sort order
	Permissions ViewPermissions        `bson:"permissions" json:"permissions"`
	Metadata    map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
	CreatedBy   primitive.ObjectID     `bson:"created_by" json:"created_by"`
	CreatedAt   time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time              `bson:"updated_at" json:"updated_at"`
}

// ViewPermissions defines who can access a view
type ViewPermissions struct {
	Roles  []string             `bson:"roles" json:"roles"`
	Users  []primitive.ObjectID `bson:"users" json:"users"`
	Public bool                 `bson:"public" json:"public"`
}

// Document represents a generic document in a collection
type Document struct {
	ID         primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	Collection string                 `bson:"-" json:"collection,omitempty"`
	Data       map[string]interface{} `bson:",inline" json:"data"`
	CreatedBy  primitive.ObjectID     `bson:"_created_by" json:"created_by"`
	UpdatedBy  primitive.ObjectID     `bson:"_updated_by" json:"updated_by"`
	CreatedAt  time.Time              `bson:"_created_at" json:"created_at"`
	UpdatedAt  time.Time              `bson:"_updated_at" json:"updated_at"`
	Version    int                    `bson:"_version" json:"version"`
	DeletedAt  *time.Time             `bson:"_deleted_at,omitempty" json:"deleted_at,omitempty"`
}

// DataQuery represents a query with governance rules
type DataQuery struct {
	Collection string                 `json:"collection"`
	Filter     map[string]interface{} `json:"filter,omitempty"`
	Fields     []string               `json:"fields,omitempty"`
	Sort       map[string]int         `json:"sort,omitempty"`
	Limit      int                    `json:"limit,omitempty"`
	Skip       int                    `json:"skip,omitempty"`
	UserID     primitive.ObjectID     `json:"-"`
	UserRoles  []string               `json:"-"`
}

// DataMutation represents a data change with governance
type DataMutation struct {
	Collection string                 `json:"collection"`
	Operation  string                 `json:"operation"` // insert, update, delete
	DocumentID primitive.ObjectID     `json:"document_id,omitempty"`
	Data       map[string]interface{} `json:"data"`
	Filter     map[string]interface{} `json:"filter,omitempty"`
	UserID     primitive.ObjectID     `json:"-"`
	UserRoles  []string               `json:"-"`
}

// AccessLog represents an access audit log
type AccessLog struct {
	ID         primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	UserID     primitive.ObjectID     `bson:"user_id" json:"user_id"`
	Collection string                 `bson:"collection" json:"collection"`
	DocumentID *primitive.ObjectID    `bson:"document_id,omitempty" json:"document_id,omitempty"`
	Operation  string                 `bson:"operation" json:"operation"` // read, write, update, delete
	Query      map[string]interface{} `bson:"query,omitempty" json:"query,omitempty"`
	Changes    map[string]interface{} `bson:"changes,omitempty" json:"changes,omitempty"`
	IPAddress  string                 `bson:"ip_address" json:"ip_address"`
	UserAgent  string                 `bson:"user_agent" json:"user_agent"`
	Status     string                 `bson:"status" json:"status"` // success, denied, error
	Error      string                 `bson:"error,omitempty" json:"error,omitempty"`
	Duration   int64                  `bson:"duration_ms" json:"duration_ms"`
	CreatedAt  time.Time              `bson:"created_at" json:"created_at"`
}