package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Settings struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID     primitive.ObjectID `bson:"user_id" json:"user_id"`
	Theme      string             `bson:"theme" json:"theme"`                             // light, dark, system
	Language   string             `bson:"language" json:"language"`                       // en, es, fr, de, zh
	Timezone   string             `bson:"timezone" json:"timezone"`                       // UTC, EST, CST, MST, PST
	DateFormat string             `bson:"date_format" json:"date_format"`                 // MM/DD/YYYY, DD/MM/YYYY, YYYY-MM-DD
	TimeFormat string             `bson:"time_format" json:"time_format"`                 // 12, 24
	
	// Notifications
	EmailNotifications     bool `bson:"email_notifications" json:"email_notifications"`
	SecurityAlerts         bool `bson:"security_alerts" json:"security_alerts"`
	
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}

type SystemSettings struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	
	// Database Settings (read-only for display)
	ConnectionPoolSize int  `bson:"connection_pool_size" json:"connection_pool_size"`
	QueryTimeout      int  `bson:"query_timeout" json:"query_timeout"` // seconds
	MaxRetries        int  `bson:"max_retries" json:"max_retries"`
	CompressionEnabled bool `bson:"compression_enabled" json:"compression_enabled"`
	EncryptionEnabled  bool `bson:"encryption_enabled" json:"encryption_enabled"`
	
	// Security Settings (admin only)
	SessionTimeout    int    `bson:"session_timeout" json:"session_timeout"` // hours
	PasswordPolicy    string `bson:"password_policy" json:"password_policy"` // basic, moderate, strong
	MFARequired       bool   `bson:"mfa_required" json:"mfa_required"`
	AuditLogEnabled   bool   `bson:"audit_log_enabled" json:"audit_log_enabled"`
	
	// API Settings (admin only)
	RateLimit         int  `bson:"rate_limit" json:"rate_limit"`         // requests per minute
	BurstLimit        int  `bson:"burst_limit" json:"burst_limit"`
	CORSEnabled       bool `bson:"cors_enabled" json:"cors_enabled"`
	
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
	UpdatedBy  primitive.ObjectID `bson:"updated_by" json:"updated_by"`
}