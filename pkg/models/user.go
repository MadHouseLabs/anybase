package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserType defines the type of user account
type UserType string

const (
	UserTypeRegular        UserType = "regular"
	UserTypeServiceAccount UserType = "service_account"
)

// User represents a user in the system
type User struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email             string             `bson:"email" json:"email" validate:"required,email"`
	Username          string             `bson:"username,omitempty" json:"username,omitempty" validate:"omitempty,min=3,max=50"`
	Password          string             `bson:"password" json:"-"` // Never expose password in JSON
	FirstName         string             `bson:"first_name,omitempty" json:"first_name,omitempty"`
	LastName          string             `bson:"last_name,omitempty" json:"last_name,omitempty"`
	Avatar            string             `bson:"avatar,omitempty" json:"avatar,omitempty"`
	EmailVerified     bool               `bson:"email_verified" json:"email_verified"`
	EmailVerificationToken string        `bson:"email_verification_token,omitempty" json:"-"`
	PasswordResetToken    string         `bson:"password_reset_token,omitempty" json:"-"`
	PasswordResetExpiry   *time.Time     `bson:"password_reset_expiry,omitempty" json:"-"`
	UserType          UserType           `bson:"user_type" json:"user_type"`
	Role              string             `bson:"role" json:"role"` // Only "admin" or "developer" for regular users
	Metadata          map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
	LastLogin         *time.Time         `bson:"last_login,omitempty" json:"last_login,omitempty"`
	LoginAttempts     int                `bson:"login_attempts" json:"-"`
	LockedUntil       *time.Time         `bson:"locked_until,omitempty" json:"-"`
	Active            bool               `bson:"active" json:"active"`
	CreatedAt         time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time          `bson:"updated_at" json:"updated_at"`
	DeletedAt         *time.Time         `bson:"deleted_at,omitempty" json:"-"`
}

// UserRegistration represents the user registration request
type UserRegistration struct {
	Email     string `json:"email" validate:"required,email"`
	Username  string `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

// UserLogin represents the user login request
type UserLogin struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// UserUpdate represents the user update request
type UserUpdate struct {
	Username  string                 `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
	FirstName string                 `json:"first_name,omitempty"`
	LastName  string                 `json:"last_name,omitempty"`
	Avatar    string                 `json:"avatar,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// PasswordReset represents the password reset request
type PasswordReset struct {
	Email string `json:"email" validate:"required,email"`
}

// PasswordUpdate represents the password update request
type PasswordUpdate struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// Session represents a user session
type Session struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID       primitive.ObjectID `bson:"user_id" json:"user_id"`
	Token        string             `bson:"token" json:"token"`
	RefreshToken string             `bson:"refresh_token" json:"refresh_token"`
	IPAddress    string             `bson:"ip_address" json:"ip_address"`
	UserAgent    string             `bson:"user_agent" json:"user_agent"`
	ExpiresAt    time.Time          `bson:"expires_at" json:"expires_at"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}


// AccessKey represents an API access key with direct permissions
type AccessKey struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name" validate:"required"`
	Description string             `bson:"description,omitempty" json:"description"`
	Key         string             `bson:"-" json:"key,omitempty"` // Never stored, only returned on creation
	KeyHash     string             `bson:"key_hash" json:"-"`       // Hashed version stored in DB
	Permissions []string           `bson:"permissions" json:"permissions"` // Direct permissions using pattern type:name:action
	CreatedBy   primitive.ObjectID `bson:"created_by" json:"created_by"`
	LastUsed    *time.Time         `bson:"last_used,omitempty" json:"last_used,omitempty"`
	ExpiresAt   *time.Time         `bson:"expires_at,omitempty" json:"expires_at,omitempty"`
	Active      bool               `bson:"active" json:"active"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID         primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	UserID     *primitive.ObjectID    `bson:"user_id,omitempty" json:"user_id,omitempty"`
	Action     string                 `bson:"action" json:"action"`
	Resource   string                 `bson:"resource" json:"resource"`
	ResourceID string                 `bson:"resource_id,omitempty" json:"resource_id,omitempty"`
	Details    map[string]interface{} `bson:"details,omitempty" json:"details,omitempty"`
	IPAddress  string                 `bson:"ip_address" json:"ip_address"`
	UserAgent  string                 `bson:"user_agent" json:"user_agent"`
	Status     string                 `bson:"status" json:"status"` // success, failure
	Error      string                 `bson:"error,omitempty" json:"error,omitempty"`
	CreatedAt  time.Time              `bson:"created_at" json:"created_at"`
}