package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/madhouselabs/anybase/internal/database"
	"github.com/madhouselabs/anybase/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	Update(ctx context.Context, id primitive.ObjectID, update *models.UserUpdate) error
	UpdateRaw(ctx context.Context, id primitive.ObjectID, update bson.M) error
	UpdatePassword(ctx context.Context, id primitive.ObjectID, hashedPassword string) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	SoftDelete(ctx context.Context, id primitive.ObjectID) error
	List(ctx context.Context, filter bson.M, opts *options.FindOptions) ([]*models.User, error)
	Count(ctx context.Context, filter bson.M) (int64, error)
	UpdateLoginAttempts(ctx context.Context, id primitive.ObjectID, attempts int) error
	UpdateLockedUntil(ctx context.Context, id primitive.ObjectID, until *time.Time) error
	UpdateLastLogin(ctx context.Context, id primitive.ObjectID) error
	VerifyEmail(ctx context.Context, id primitive.ObjectID) error
	SetPasswordResetToken(ctx context.Context, id primitive.ObjectID, token string, expiry time.Time) error
	GetByPasswordResetToken(ctx context.Context, token string) (*models.User, error)
}

type repository struct {
	db *database.Database
}

func NewRepository(db *database.Database) Repository {
	return &repository{db: db}
}

func (r *repository) collection() *mongo.Collection {
	return r.db.Collection("users")
}

func (r *repository) Create(ctx context.Context, user *models.User) error {
	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.Active = true
	user.EmailVerified = false
	user.LoginAttempts = 0
	user.UserType = models.UserTypeRegular // Set as regular user

	if user.Role == "" {
		user.Role = "developer" // Default role for new users
	}

	_, err := r.collection().InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrUserAlreadyExists
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *repository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	var user models.User
	filter := bson.M{"_id": id, "deleted_at": nil}

	err := r.collection().FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *repository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	filter := bson.M{"email": email, "deleted_at": nil}

	err := r.collection().FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

func (r *repository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	filter := bson.M{"username": username, "deleted_at": nil}

	err := r.collection().FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &user, nil
}

func (r *repository) Update(ctx context.Context, id primitive.ObjectID, update *models.UserUpdate) error {
	updateDoc := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	if update.Username != "" {
		updateDoc["$set"].(bson.M)["username"] = update.Username
	}
	if update.FirstName != "" {
		updateDoc["$set"].(bson.M)["first_name"] = update.FirstName
	}
	if update.LastName != "" {
		updateDoc["$set"].(bson.M)["last_name"] = update.LastName
	}
	if update.Avatar != "" {
		updateDoc["$set"].(bson.M)["avatar"] = update.Avatar
	}
	if update.Metadata != nil {
		updateDoc["$set"].(bson.M)["metadata"] = update.Metadata
	}

	filter := bson.M{"_id": id, "deleted_at": nil}
	result, err := r.collection().UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if result.MatchedCount == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *repository) UpdateRaw(ctx context.Context, id primitive.ObjectID, update bson.M) error {
	filter := bson.M{"_id": id, "deleted_at": nil}
	
	// Add updated_at to the update
	if update == nil {
		update = bson.M{}
	}
	update["updated_at"] = time.Now()
	
	updateDoc := bson.M{"$set": update}
	
	result, err := r.collection().UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	
	if result.MatchedCount == 0 {
		return ErrUserNotFound
	}
	
	return nil
}

func (r *repository) UpdatePassword(ctx context.Context, id primitive.ObjectID, hashedPassword string) error {
	filter := bson.M{"_id": id, "deleted_at": nil}
	update := bson.M{
		"$set": bson.M{
			"password":              hashedPassword,
			"password_reset_token":  "",
			"password_reset_expiry": nil,
			"updated_at":            time.Now(),
		},
	}

	result, err := r.collection().UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	if result.MatchedCount == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *repository) Delete(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	result, err := r.collection().DeleteOne(ctx, filter)
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
	filter := bson.M{"_id": id, "deleted_at": nil}
	update := bson.M{
		"$set": bson.M{
			"deleted_at": now,
			"updated_at": now,
			"active":     false,
		},
	}

	result, err := r.collection().UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to soft delete user: %w", err)
	}

	if result.MatchedCount == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *repository) List(ctx context.Context, filter bson.M, opts *options.FindOptions) ([]*models.User, error) {
	if filter == nil {
		filter = bson.M{}
	}
	filter["deleted_at"] = nil

	cursor, err := r.collection().Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer cursor.Close(ctx)

	var users []*models.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, fmt.Errorf("failed to decode users: %w", err)
	}

	return users, nil
}

func (r *repository) Count(ctx context.Context, filter bson.M) (int64, error) {
	if filter == nil {
		filter = bson.M{}
	}
	filter["deleted_at"] = nil

	count, err := r.collection().CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

func (r *repository) UpdateLoginAttempts(ctx context.Context, id primitive.ObjectID, attempts int) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"login_attempts": attempts,
			"updated_at":     time.Now(),
		},
	}

	_, err := r.collection().UpdateOne(ctx, filter, update)
	return err
}

func (r *repository) UpdateLockedUntil(ctx context.Context, id primitive.ObjectID, until *time.Time) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"locked_until": until,
			"updated_at":   time.Now(),
		},
	}

	_, err := r.collection().UpdateOne(ctx, filter, update)
	return err
}

func (r *repository) UpdateLastLogin(ctx context.Context, id primitive.ObjectID) error {
	now := time.Now()
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"last_login":     now,
			"login_attempts": 0,
			"locked_until":   nil,
			"updated_at":     now,
		},
	}

	_, err := r.collection().UpdateOne(ctx, filter, update)
	return err
}

func (r *repository) VerifyEmail(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"email_verified":           true,
			"email_verification_token": "",
			"updated_at":               time.Now(),
		},
	}

	result, err := r.collection().UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to verify email: %w", err)
	}

	if result.MatchedCount == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *repository) SetPasswordResetToken(ctx context.Context, id primitive.ObjectID, token string, expiry time.Time) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"password_reset_token":  token,
			"password_reset_expiry": expiry,
			"updated_at":            time.Now(),
		},
	}

	result, err := r.collection().UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to set password reset token: %w", err)
	}

	if result.MatchedCount == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *repository) GetByPasswordResetToken(ctx context.Context, token string) (*models.User, error) {
	var user models.User
	filter := bson.M{
		"password_reset_token": token,
		"password_reset_expiry": bson.M{"$gt": time.Now()},
		"deleted_at":           nil,
	}

	err := r.collection().FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by reset token: %w", err)
	}

	return &user, nil
}