package settings

import (
	"context"
	"time"

	"github.com/karthik/anybase/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Service interface {
	GetUserSettings(ctx context.Context, userID primitive.ObjectID) (*models.Settings, error)
	UpdateUserSettings(ctx context.Context, userID primitive.ObjectID, settings *models.Settings) error
	GetSystemSettings(ctx context.Context) (*models.SystemSettings, error)
	UpdateSystemSettings(ctx context.Context, userID primitive.ObjectID, settings *models.SystemSettings) error
}

type service struct {
	db *mongo.Database
}

func NewService(db *mongo.Database) Service {
	return &service{db: db}
}

func (s *service) GetUserSettings(ctx context.Context, userID primitive.ObjectID) (*models.Settings, error) {
	var settings models.Settings
	
	err := s.db.Collection("user_settings").FindOne(ctx, bson.M{"user_id": userID}).Decode(&settings)
	if err == mongo.ErrNoDocuments {
		// Create default settings for user
		settings = models.Settings{
			ID:                 primitive.NewObjectID(),
			UserID:             userID,
			Theme:              "system",
			Language:           "en",
			Timezone:           "UTC",
			DateFormat:         "MM/DD/YYYY",
			TimeFormat:         "12",
			EmailNotifications: true,
			SecurityAlerts:     true,
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}
		
		_, err = s.db.Collection("user_settings").InsertOne(ctx, settings)
		if err != nil {
			return nil, err
		}
		return &settings, nil
	}
	
	if err != nil {
		return nil, err
	}
	
	return &settings, nil
}

func (s *service) UpdateUserSettings(ctx context.Context, userID primitive.ObjectID, settings *models.Settings) error {
	settings.UserID = userID
	settings.UpdatedAt = time.Now()
	
	filter := bson.M{"user_id": userID}
	update := bson.M{"$set": settings}
	opts := options.Update().SetUpsert(true)
	
	_, err := s.db.Collection("user_settings").UpdateOne(ctx, filter, update, opts)
	return err
}

func (s *service) GetSystemSettings(ctx context.Context) (*models.SystemSettings, error) {
	var settings models.SystemSettings
	
	// Get the single system settings document
	err := s.db.Collection("system_settings").FindOne(ctx, bson.M{}).Decode(&settings)
	if err == mongo.ErrNoDocuments {
		// Create default system settings
		settings = models.SystemSettings{
			ID:                 primitive.NewObjectID(),
			ConnectionPoolSize: 100,
			QueryTimeout:       30,
			MaxRetries:         3,
			CompressionEnabled: true,
			EncryptionEnabled:  true,
			SessionTimeout:     24,
			PasswordPolicy:     "strong",
			MFARequired:        false,
			AuditLogEnabled:    true,
			RateLimit:          1000,
			BurstLimit:         50,
			CORSEnabled:        true,
			UpdatedAt:          time.Now(),
		}
		
		_, err = s.db.Collection("system_settings").InsertOne(ctx, settings)
		if err != nil {
			return nil, err
		}
		return &settings, nil
	}
	
	if err != nil {
		return nil, err
	}
	
	return &settings, nil
}

func (s *service) UpdateSystemSettings(ctx context.Context, userID primitive.ObjectID, settings *models.SystemSettings) error {
	settings.UpdatedAt = time.Now()
	settings.UpdatedBy = userID
	
	// Update the single system settings document
	filter := bson.M{}
	update := bson.M{"$set": settings}
	opts := options.Update().SetUpsert(true)
	
	_, err := s.db.Collection("system_settings").UpdateOne(ctx, filter, update, opts)
	return err
}