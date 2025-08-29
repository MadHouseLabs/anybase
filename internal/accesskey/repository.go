package accesskey

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/madhouselabs/anybase/internal/database"
	"github.com/madhouselabs/anybase/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	List(ctx context.Context, filter bson.M, opts *options.FindOptions) ([]*models.AccessKey, error)
	Count(ctx context.Context, filter bson.M) (int64, error)
	RegenerateKey(ctx context.Context, id primitive.ObjectID) (string, error)
	UpdateLastUsed(ctx context.Context, id primitive.ObjectID) error
	ValidateKey(ctx context.Context, key string) (*models.AccessKey, error)
}

type repository struct {
	db *database.Database
}

func NewRepository(db *database.Database) Repository {
	return &repository{db: db}
}

func (r *repository) collection() *mongo.Collection {
	return r.db.Collection("access_keys")
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

	_, err = r.collection().InsertOne(ctx, ak)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return "", ErrAccessKeyAlreadyExists
		}
		return "", fmt.Errorf("failed to create access key: %w", err)
	}

	return accessKey, nil
}

func (r *repository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.AccessKey, error) {
	var ak models.AccessKey
	filter := bson.M{"_id": id}

	err := r.collection().FindOne(ctx, filter).Decode(&ak)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrAccessKeyNotFound
		}
		return nil, fmt.Errorf("failed to get access key: %w", err)
	}

	return &ak, nil
}

func (r *repository) GetByName(ctx context.Context, name string) (*models.AccessKey, error) {
	var ak models.AccessKey
	filter := bson.M{"name": name}

	err := r.collection().FindOne(ctx, filter).Decode(&ak)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrAccessKeyNotFound
		}
		return nil, fmt.Errorf("failed to get access key by name: %w", err)
	}

	return &ak, nil
}

func (r *repository) GetByKey(ctx context.Context, key string) (*models.AccessKey, error) {
	// Get all active keys and check them one by one (bcrypt doesn't allow direct lookup)
	cursor, err := r.collection().Find(ctx, bson.M{"active": true})
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
	filter := bson.M{"_id": id}
	
	if update["$set"] == nil {
		update["$set"] = bson.M{}
	}
	update["$set"].(bson.M)["updated_at"] = time.Now()

	result, err := r.collection().UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update access key: %w", err)
	}

	if result.MatchedCount == 0 {
		return ErrAccessKeyNotFound
	}

	return nil
}

func (r *repository) Delete(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	result, err := r.collection().DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete access key: %w", err)
	}

	if result.DeletedCount == 0 {
		return ErrAccessKeyNotFound
	}

	return nil
}

func (r *repository) List(ctx context.Context, filter bson.M, opts *options.FindOptions) ([]*models.AccessKey, error) {
	if filter == nil {
		filter = bson.M{}
	}

	cursor, err := r.collection().Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list access keys: %w", err)
	}
	defer cursor.Close(ctx)

	var accessKeys []*models.AccessKey
	if err := cursor.All(ctx, &accessKeys); err != nil {
		return nil, fmt.Errorf("failed to decode access keys: %w", err)
	}

	return accessKeys, nil
}

func (r *repository) Count(ctx context.Context, filter bson.M) (int64, error) {
	if filter == nil {
		filter = bson.M{}
	}

	count, err := r.collection().CountDocuments(ctx, filter)
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

	_, err := r.collection().UpdateOne(ctx, filter, update)
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