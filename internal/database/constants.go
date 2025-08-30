package database

import "time"

const (
	// SessionTTL is the default TTL for session documents
	SessionTTL = 0 * time.Second // 0 means use expires_at field value
	
	// DefaultTimeout for database operations
	DefaultTimeout = 30 * time.Second
	
	// MaxRetries for database operations
	MaxRetries = 3
	
	// RetryDelay between retries
	RetryDelay = 1 * time.Second
)