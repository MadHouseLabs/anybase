package user

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	types "github.com/madhouselabs/anybase/internal/database/types"
	"github.com/madhouselabs/anybase/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type Repository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, id primitive.ObjectID, update *models.UserUpdate) error
	UpdateRaw(ctx context.Context, id primitive.ObjectID, update interface{}) error
	UpdatePassword(ctx context.Context, id primitive.ObjectID, hashedPassword string) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	SoftDelete(ctx context.Context, id primitive.ObjectID) error
	List(ctx context.Context, filter interface{}, opts interface{}) ([]*models.User, error)
	Count(ctx context.Context, filter interface{}) (int64, error)
	UpdateLoginAttempts(ctx context.Context, id primitive.ObjectID, attempts int) error
	UpdateLockedUntil(ctx context.Context, id primitive.ObjectID, until *time.Time) error
	UpdateLastLogin(ctx context.Context, id primitive.ObjectID) error
	VerifyEmail(ctx context.Context, id primitive.ObjectID) error
	SetPasswordResetToken(ctx context.Context, id primitive.ObjectID, token string, expiry time.Time) error
	GetByPasswordResetToken(ctx context.Context, token string) (*models.User, error)
}

type repository struct {
	db types.DB
	collection types.Collection
}

func NewRepository(db types.DB) Repository {
	return &repository{
		db: db,
		collection: db.Collection("users"),
	}
}

// Helper function to create ID filter that works with both MongoDB and PostgreSQL
func (r *repository) idFilter(id primitive.ObjectID) map[string]interface{} {
	// For PostgreSQL, we need to use hex string
	// For MongoDB, this will still work as it can handle hex strings
	return map[string]interface{}{
		"_id": id.Hex(),
	}
}

func (r *repository) Create(ctx context.Context, user *models.User) error {
	// Always generate an ObjectID for application-level compatibility
	user.ID = primitive.NewObjectID()
	
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.Active = true
	user.EmailVerified = false
	user.LoginAttempts = 0
	user.UserType = models.UserTypeRegular

	if user.Role == "" {
		user.Role = "developer"
	}

	// Convert to generic document
	doc := map[string]interface{}{
		"_id": user.ID.Hex(),  // Always store as hex string for cross-database compatibility
		"email": user.Email,
		"password": user.Password,
		"first_name": user.FirstName,
		"last_name": user.LastName,
		"role": user.Role,
		"user_type": user.UserType,
		"active": user.Active,
		"email_verified": user.EmailVerified,
		"login_attempts": user.LoginAttempts,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	}

	insertedID, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		// Check for duplicate key error
		if strings.Contains(err.Error(), "duplicate key") {
			return ErrUserAlreadyExists
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	// For PostgreSQL, we need to store the actual UUID somehow
	// Store it in the user's metadata for reference
	if r.db.Type() == "postgres" && insertedID != nil {
		if user.Metadata == nil {
			user.Metadata = make(map[string]interface{})
		}
		user.Metadata["_postgres_id"] = insertedID.String()
	}

	return nil
}

func (r *repository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	filter := r.idFilter(id)
	filter["deleted_at"] = nil

	var user models.User
	if err := r.collection.FindOne(ctx, filter, &user); err != nil {
		if err.Error() == "no documents found" {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *repository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	filter := map[string]interface{}{
		"email": email,
		"deleted_at": nil,
	}

	var user models.User
	if err := r.collection.FindOne(ctx, filter, &user); err != nil {
		if err == types.ErrNoDocuments || err.Error() == "no documents found" {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}


func (r *repository) Update(ctx context.Context, id primitive.ObjectID, update *models.UserUpdate) error {
	filter := map[string]interface{}{
		"_id": id,
		"deleted_at": nil,
	}

	updateDoc := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if update.FirstName != "" {
		updateDoc["first_name"] = update.FirstName
	}
	if update.LastName != "" {
		updateDoc["last_name"] = update.LastName
	}
	if update.Avatar != "" {
		updateDoc["avatar"] = update.Avatar
	}
	if update.Metadata != nil {
		updateDoc["metadata"] = update.Metadata
	}

	result, err := r.collection.UpdateOne(ctx, filter, map[string]interface{}{"$set": updateDoc})
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if result.ModifiedCount == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *repository) UpdateRaw(ctx context.Context, id primitive.ObjectID, update interface{}) error {
	// For PostgreSQL, we need to handle the ID differently
	var filter map[string]interface{}
	
	if r.db.Type() == "postgres" {
		// Try to get the user first to get the PostgreSQL ID
		user, err := r.GetByID(ctx, id)
		if err != nil {
			// If we can't find by ObjectID, this might be a new user
			// Try to find by any criteria available
			filter = map[string]interface{}{
				"deleted_at": nil,
			}
		} else if user.Metadata != nil && user.Metadata["_postgres_id"] != nil {
			// Use the PostgreSQL ID if available
			filter = map[string]interface{}{
				"_id": user.Metadata["_postgres_id"],
				"deleted_at": nil,
			}
		} else {
			filter = map[string]interface{}{
				"deleted_at": nil,
			}
		}
	} else {
		filter = r.idFilter(id)
		filter["deleted_at"] = nil
	}
	
	// Convert update to map if it's bson.M
	var updateMap map[string]interface{}
	if update == nil {
		updateMap = map[string]interface{}{}
	} else if m, ok := update.(map[string]interface{}); ok {
		updateMap = m
	} else if bsonM, ok := update.(bson.M); ok {
		updateMap = map[string]interface{}(bsonM)
	} else {
		updateMap = map[string]interface{}{}
	}
	
	updateMap["updated_at"] = time.Now()
	
	updateDoc := map[string]interface{}{"$set": updateMap}
	
	result, err := r.collection.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	
	if result.ModifiedCount == 0 {
		return ErrUserNotFound
	}
	
	return nil
}

func (r *repository) UpdatePassword(ctx context.Context, id primitive.ObjectID, hashedPassword string) error {
	filter := map[string]interface{}{
		"_id": id,
		"deleted_at": nil,
	}
	
	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"password":              hashedPassword,
			"password_reset_token":  "",
			"password_reset_expiry": nil,
			"updated_at":            time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	if result.ModifiedCount == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *repository) Delete(ctx context.Context, id primitive.ObjectID) error {
	filter := r.idFilter(id)
	
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.DeletedCount == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *repository) SoftDelete(ctx context.Context, id primitive.ObjectID) error {
	now := time.Now()
	filter := r.idFilter(id)
	filter["deleted_at"] = nil
	
	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"deleted_at": now,
			"updated_at": now,
			"active":     false,
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to soft delete user: %w", err)
	}

	if result.ModifiedCount == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *repository) List(ctx context.Context, filter interface{}, opts interface{}) ([]*models.User, error) {
	if filter == nil {
		filter = map[string]interface{}{}
	}
	
	// Ensure we add deleted_at filter
	filterMap, ok := filter.(map[string]interface{})
	if !ok {
		// If it's a bson.M, convert it
		if bsonFilter, ok := filter.(bson.M); ok {
			filterMap = map[string]interface{}(bsonFilter)
		} else {
			filterMap = map[string]interface{}{}
		}
	}
	filterMap["deleted_at"] = nil

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
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer cursor.Close(ctx)

	var users []*models.User
	for cursor.Next(ctx) {
		var user models.User
		if err := cursor.Decode(&user); err != nil {
			return nil, fmt.Errorf("failed to decode user: %w", err)
		}
		users = append(users, &user)
	}


	return users, nil
}

func (r *repository) Count(ctx context.Context, filter interface{}) (int64, error) {
	if filter == nil {
		filter = map[string]interface{}{}
	}
	
	// Ensure we add deleted_at filter
	filterMap, ok := filter.(map[string]interface{})
	if !ok {
		// If it's a bson.M, convert it
		if bsonFilter, ok := filter.(bson.M); ok {
			filterMap = map[string]interface{}(bsonFilter)
		} else {
			filterMap = map[string]interface{}{}
		}
	}
	filterMap["deleted_at"] = nil

	count, err := r.collection.CountDocuments(ctx, filterMap)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

func (r *repository) UpdateLoginAttempts(ctx context.Context, id primitive.ObjectID, attempts int) error {
	filter := r.idFilter(id)
	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"login_attempts": attempts,
			"updated_at":     time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *repository) UpdateLockedUntil(ctx context.Context, id primitive.ObjectID, until *time.Time) error {
	filter := r.idFilter(id)
	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"locked_until": until,
			"updated_at":   time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *repository) UpdateLastLogin(ctx context.Context, id primitive.ObjectID) error {
	now := time.Now()
	filter := r.idFilter(id)
	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"last_login":     now,
			"login_attempts": 0,
			"locked_until":   nil,
			"updated_at":     now,
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *repository) VerifyEmail(ctx context.Context, id primitive.ObjectID) error {
	filter := r.idFilter(id)
	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"email_verified":           true,
			"email_verification_token": "",
			"updated_at":               time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to verify email: %w", err)
	}

	if result.ModifiedCount == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *repository) SetPasswordResetToken(ctx context.Context, id primitive.ObjectID, token string, expiry time.Time) error {
	filter := r.idFilter(id)
	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"password_reset_token":  token,
			"password_reset_expiry": expiry,
			"updated_at":            time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to set password reset token: %w", err)
	}

	if result.ModifiedCount == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *repository) GetByPasswordResetToken(ctx context.Context, token string) (*models.User, error) {
	filter := map[string]interface{}{
		"password_reset_token": token,
		"password_reset_expiry": map[string]interface{}{"$gt": time.Now()},
		"deleted_at":           nil,
	}

	var user models.User
	if err := r.collection.FindOne(ctx, filter, &user); err != nil {
		if err.Error() == "no documents found" {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by reset token: %w", err)
	}

	return &user, nil
}