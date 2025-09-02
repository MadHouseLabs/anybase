package accesskey

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	types "github.com/madhouselabs/anybase/internal/database/types"
	"github.com/madhouselabs/anybase/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrAccessKeyNotFound      = errors.New("access key not found")
	ErrAccessKeyAlreadyExists = errors.New("access key already exists")
	ErrInvalidAccessKey       = errors.New("invalid access key")
	ErrExpiredAccessKey       = errors.New("access key has expired")
)

type Repository interface {
	Create(ctx context.Context, ak *models.AccessKey) (string, error) // Returns generated key
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.AccessKey, error)
	GetByName(ctx context.Context, name string) (*models.AccessKey, error)
	GetByKey(ctx context.Context, key string) (*models.AccessKey, error)
	Update(ctx context.Context, id primitive.ObjectID, update bson.M) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	List(ctx context.Context, filter interface{}, opts interface{}) ([]*models.AccessKey, error)
	Count(ctx context.Context, filter interface{}) (int64, error)
	RegenerateKey(ctx context.Context, id primitive.ObjectID) (string, error)
	UpdateLastUsed(ctx context.Context, id primitive.ObjectID) error
	ValidateKey(ctx context.Context, key string) (*models.AccessKey, error)
}

type repository struct {
	db types.DB
	collection types.Collection
}

func NewRepository(db types.DB) Repository {
	return &repository{
		db: db,
		collection: db.Collection("access_keys"),
	}
}

func (r *repository) Create(ctx context.Context, ak *models.AccessKey) (string, error) {
	// Generate access key
	accessKey := generateAccessKey()
	
	// Hash the key for storage
	hashedKey, err := hashKey(accessKey)
	if err != nil {
		return "", fmt.Errorf("failed to hash access key: %w", err)
	}

	ak.ID = primitive.NewObjectID()
	ak.KeyHash = hashedKey
	ak.Active = true
	ak.CreatedAt = time.Now()
	ak.UpdatedAt = time.Now()

	doc := map[string]interface{}{
		"_id": ak.ID,
		"name": ak.Name,
		"key_hash": ak.KeyHash,
		"description": ak.Description,
		"permissions": ak.Permissions,
		"created_by": ak.CreatedBy,
		"active": ak.Active,
		"expires_at": ak.ExpiresAt,
		"created_at": ak.CreatedAt,
		"updated_at": ak.UpdatedAt,
	}
	
	_, err = r.collection.InsertOne(ctx, doc)
	if err != nil {
		if err.Error() == "duplicate key error" {
			return "", ErrAccessKeyAlreadyExists
		}
		return "", fmt.Errorf("failed to create access key: %w", err)
	}

	return accessKey, nil
}

func (r *repository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.AccessKey, error) {
	var ak models.AccessKey
	filter := map[string]interface{}{"_id": id}

	err := r.collection.FindOne(ctx, filter, &ak)
	if err != nil {
		if err.Error() == "no documents found" {
			return nil, ErrAccessKeyNotFound
		}
		return nil, fmt.Errorf("failed to get access key: %w", err)
	}

	return &ak, nil
}

func (r *repository) GetByName(ctx context.Context, name string) (*models.AccessKey, error) {
	var ak models.AccessKey
	filter := map[string]interface{}{"name": name}

	err := r.collection.FindOne(ctx, filter, &ak)
	if err != nil {
		if err.Error() == "no documents found" {
			return nil, ErrAccessKeyNotFound
		}
		return nil, fmt.Errorf("failed to get access key by name: %w", err)
	}

	return &ak, nil
}

func (r *repository) GetByKey(ctx context.Context, key string) (*models.AccessKey, error) {
	// Get all active keys and check them one by one (bcrypt doesn't allow direct lookup)
	cursor, err := r.collection.Find(ctx, map[string]interface{}{"active": true}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to find access keys: %w", err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var ak models.AccessKey
		if err := cursor.Decode(&ak); err != nil {
			continue
		}
		
		// Check if this key matches
		if err := verifyKey(ak.KeyHash, key); err == nil {
			// Check if expired
			if ak.ExpiresAt != nil && ak.ExpiresAt.Before(time.Now()) {
				return nil, ErrExpiredAccessKey
			}

			// Update last used timestamp
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				r.UpdateLastUsed(ctx, ak.ID)
			}()

			return &ak, nil
		}
	}

	return nil, ErrAccessKeyNotFound
}

func (r *repository) ValidateKey(ctx context.Context, key string) (*models.AccessKey, error) {
	return r.GetByKey(ctx, key)
}

func (r *repository) Update(ctx context.Context, id primitive.ObjectID, update bson.M) error {
	filter := map[string]interface{}{"_id": id}
	
	updateMap := map[string]interface{}(update)
	if updateMap["$set"] == nil {
		updateMap["$set"] = map[string]interface{}{}
	}
	updateMap["$set"].(map[string]interface{})["updated_at"] = time.Now()

	result, err := r.collection.UpdateOne(ctx, filter, updateMap)
	if err != nil {
		return fmt.Errorf("failed to update access key: %w", err)
	}

	if result.MatchedCount == 0 {
		return ErrAccessKeyNotFound
	}

	return nil
}

func (r *repository) Delete(ctx context.Context, id primitive.ObjectID) error {
	filter := map[string]interface{}{"_id": id}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete access key: %w", err)
	}

	if result.DeletedCount == 0 {
		return ErrAccessKeyNotFound
	}

	return nil
}

func (r *repository) List(ctx context.Context, filter interface{}, opts interface{}) ([]*models.AccessKey, error) {
	if filter == nil {
		filter = map[string]interface{}{}
	}
	
	// Convert bson.M to map if needed
	filterMap, ok := filter.(map[string]interface{})
	if !ok {
		if bsonFilter, ok := filter.(bson.M); ok {
			filterMap = map[string]interface{}(bsonFilter)
		} else {
			filterMap = map[string]interface{}{}
		}
	}

	// Convert opts if needed
	var findOpts *types.FindOptions
	if opts != nil {
		// Check if it's already types.FindOptions
		if typedOpts, ok := opts.(*types.FindOptions); ok {
			findOpts = typedOpts
		} else {
			// Convert from MongoDB options to types.FindOptions
			findOpts = &types.FindOptions{}
			// MongoDB options don't directly map, so we'll use defaults
		}
	}
	
	cursor, err := r.collection.Find(ctx, filterMap, findOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to list access keys: %w", err)
	}
	defer cursor.Close(ctx)

	var accessKeys []*models.AccessKey
	for cursor.Next(ctx) {
		var ak models.AccessKey
		if err := cursor.Decode(&ak); err != nil {
			return nil, fmt.Errorf("failed to decode access key: %w", err)
		}
		accessKeys = append(accessKeys, &ak)
	}


	return accessKeys, nil
}

func (r *repository) Count(ctx context.Context, filter interface{}) (int64, error) {
	if filter == nil {
		filter = map[string]interface{}{}
	}
	
	// Convert bson.M to map if needed
	filterMap, ok := filter.(map[string]interface{})
	if !ok {
		if bsonFilter, ok := filter.(bson.M); ok {
			filterMap = map[string]interface{}(bsonFilter)
		} else {
			filterMap = map[string]interface{}{}
		}
	}

	count, err := r.collection.CountDocuments(ctx, filterMap)
	if err != nil {
		return 0, fmt.Errorf("failed to count access keys: %w", err)
	}

	return count, nil
}

func (r *repository) RegenerateKey(ctx context.Context, id primitive.ObjectID) (string, error) {
	// Generate new access key
	accessKey := generateAccessKey()
	
	// Hash the key for storage
	hashedKey, err := hashKey(accessKey)
	if err != nil {
		return "", fmt.Errorf("failed to hash access key: %w", err)
	}

	// Update the access key with new hash
	update := bson.M{
		"$set": bson.M{
			"key_hash":   hashedKey,
			"updated_at": time.Now(),
		},
	}

	err = r.Update(ctx, id, update)
	if err != nil {
		return "", err
	}

	return accessKey, nil
}

func (r *repository) UpdateLastUsed(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"last_used": time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// Helper functions

func generateAccessKey() string {
	// Generate a 32-byte random key
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("failed to generate random bytes: %v", err))
	}
	
	// Encode to hex and add prefix for easy identification
	return "ak_" + hex.EncodeToString(b)
}

func hashKey(key string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(key), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func verifyKey(hashedKey, key string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedKey), []byte(key))
}