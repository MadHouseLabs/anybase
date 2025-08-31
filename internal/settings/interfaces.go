package settings

import (
	"context"
	
	"github.com/madhouselabs/anybase/pkg/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Service defines the settings service interface
type Service interface {
	GetSystemSettings(ctx context.Context) (*models.SystemSettings, error)
	UpdateSystemSettings(ctx context.Context, userID primitive.ObjectID, settings *models.SystemSettings) error
	GetUserSettings(ctx context.Context, userID primitive.ObjectID) (*models.Settings, error)
	UpdateUserSettings(ctx context.Context, userID primitive.ObjectID, settings *models.Settings) error
}