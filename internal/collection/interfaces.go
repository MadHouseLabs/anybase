package collection

import (
	"context"

	"github.com/madhouselabs/anybase/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Service defines the collection service interface
type Service interface {
	// Collection management
	CreateCollection(ctx context.Context, userID primitive.ObjectID, collection *models.Collection) error
	GetCollection(ctx context.Context, userID primitive.ObjectID, name string) (*models.Collection, error)
	ListCollections(ctx context.Context) ([]*models.Collection, error)
	UpdateCollection(ctx context.Context, userID primitive.ObjectID, name string, updates bson.M) error
	DeleteCollection(ctx context.Context, userID primitive.ObjectID, name string) error
	
	// Index management
	CreateIndex(ctx context.Context, collectionName string, indexName string, keys map[string]interface{}, options map[string]interface{}) error
	ListIndexes(ctx context.Context, collectionName string) ([]interface{}, error)
	DeleteIndex(ctx context.Context, collectionName string, indexName string) error
	
	// View management
	CreateView(ctx context.Context, userID primitive.ObjectID, view *models.View) error
	GetView(ctx context.Context, name string) (*models.View, error)
	ListViews(ctx context.Context) ([]*models.View, error)
	UpdateView(ctx context.Context, userID primitive.ObjectID, name string, updates bson.M) error
	DeleteView(ctx context.Context, userID primitive.ObjectID, name string) error
	QueryView(ctx context.Context, userID primitive.ObjectID, viewName string, opts QueryOptions) ([]bson.M, error)
	
	// Document operations
	InsertDocument(ctx context.Context, mutation *models.DataMutation) (*models.Document, error)
	QueryDocuments(ctx context.Context, query *models.DataQuery) ([]models.Document, error)
	GetDocument(ctx context.Context, userID primitive.ObjectID, collectionName string, docID primitive.ObjectID) (*models.Document, error)
	UpdateDocument(ctx context.Context, mutation *models.DataMutation) error
	DeleteDocument(ctx context.Context, mutation *models.DataMutation) error
	CountDocuments(ctx context.Context, userID primitive.ObjectID, collection string, filter map[string]interface{}) (int, error)
	
	// Schema validation
	ValidateAgainstSchema(ctx context.Context, collectionName string, document interface{}) error
	
	// Permissions
	CanRead(ctx context.Context, userID primitive.ObjectID, collection string) (bool, error)
	CanWrite(ctx context.Context, userID primitive.ObjectID, collection string) (bool, error)
	CanUpdate(ctx context.Context, userID primitive.ObjectID, collection string) (bool, error)
	CanDelete(ctx context.Context, userID primitive.ObjectID, collection string) (bool, error)
}

// QueryOptions defines options for querying documents
type QueryOptions struct {
	Filter      bson.M
	Sort        bson.M
	Limit       int
	Skip        int
	Projection  bson.M
	ExtraFilter bson.M
}

// QueryResult represents the result of a query operation
type QueryResult struct {
	Documents []bson.M `json:"documents"`
	Total     int64    `json:"total"`
	Page      int      `json:"page"`
	PageSize  int      `json:"pageSize"`
}