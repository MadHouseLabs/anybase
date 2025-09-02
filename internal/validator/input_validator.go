package validator

import (
	"fmt"
	"regexp"
	"strings"
)

// InputValidator provides input validation methods
type InputValidator struct{}

// NewInputValidator creates a new input validator
func NewInputValidator() *InputValidator {
	return &InputValidator{}
}

var (
	// Collection name must start with letter, can contain letters, numbers, underscores
	// Min 3 chars, max 100 chars, no SQL keywords
	collectionNameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]{2,99}$`)
	
	// Reserved SQL keywords that cannot be used as collection names
	sqlKeywords = map[string]bool{
		"select": true, "insert": true, "update": true, "delete": true,
		"drop": true, "create": true, "alter": true, "table": true,
		"database": true, "schema": true, "index": true, "view": true,
		"trigger": true, "procedure": true, "function": true, "grant": true,
		"revoke": true, "union": true, "join": true, "where": true,
		"order": true, "group": true, "having": true, "limit": true,
		"offset": true, "from": true, "into": true, "values": true,
		"set": true, "begin": true, "commit": true, "rollback": true,
		"transaction": true, "primary": true, "foreign": true, "key": true,
		"references": true, "constraint": true, "unique": true, "default": true,
		"null": true, "not": true, "and": true, "or": true,
		"in": true, "exists": true, "between": true, "like": true,
		"as": true, "on": true, "using": true, "with": true,
	}

	// Reserved collection name prefixes
	reservedPrefixes = []string{
		"system_", "sys_", "_", "pg_", "information_", "performance_",
	}
)

// ValidateCollectionName validates a collection name for security and naming conventions
func (v *InputValidator) ValidateCollectionName(name string) error {
	if name == "" {
		return fmt.Errorf("collection name cannot be empty")
	}

	if len(name) < 3 {
		return fmt.Errorf("collection name must be at least 3 characters long")
	}

	if len(name) > 100 {
		return fmt.Errorf("collection name must not exceed 100 characters")
	}

	// Check against regex pattern
	if !collectionNameRegex.MatchString(name) {
		return fmt.Errorf("collection name must start with a letter and contain only letters, numbers, and underscores")
	}

	// Check for SQL keywords (case-insensitive)
	lowerName := strings.ToLower(name)
	if sqlKeywords[lowerName] {
		return fmt.Errorf("collection name '%s' is a reserved SQL keyword", name)
	}

	// Check for reserved prefixes
	for _, prefix := range reservedPrefixes {
		if strings.HasPrefix(lowerName, strings.ToLower(prefix)) {
			return fmt.Errorf("collection name cannot start with reserved prefix '%s'", prefix)
		}
	}

	// Check for potentially dangerous patterns
	if strings.Contains(lowerName, "--") {
		return fmt.Errorf("collection name cannot contain SQL comment sequences")
	}

	if strings.Contains(lowerName, "/*") || strings.Contains(lowerName, "*/") {
		return fmt.Errorf("collection name cannot contain SQL comment blocks")
	}

	if strings.Contains(name, ";") {
		return fmt.Errorf("collection name cannot contain semicolons")
	}

	if strings.Contains(name, "'") || strings.Contains(name, "\"") || strings.Contains(name, "`") {
		return fmt.Errorf("collection name cannot contain quotes")
	}

	return nil
}

// ValidateFieldName validates a field name within a collection
func (v *InputValidator) ValidateFieldName(name string) error {
	if name == "" {
		return fmt.Errorf("field name cannot be empty")
	}

	if len(name) > 100 {
		return fmt.Errorf("field name must not exceed 100 characters")
	}

	// Field names can start with letter or underscore (for internal fields)
	fieldNameRegex := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	if !fieldNameRegex.MatchString(name) {
		return fmt.Errorf("field name must start with a letter or underscore and contain only letters, numbers, and underscores")
	}

	// Check for SQL injection patterns
	lowerName := strings.ToLower(name)
	if strings.Contains(lowerName, "--") || strings.Contains(lowerName, "/*") || strings.Contains(lowerName, "*/") {
		return fmt.Errorf("field name contains invalid characters")
	}

	if strings.Contains(name, ";") || strings.Contains(name, "'") || strings.Contains(name, "\"") {
		return fmt.Errorf("field name contains invalid characters")
	}

	return nil
}

// SanitizeIdentifier ensures an identifier is safe for use in SQL queries
// This should be used as a last resort after validation
func (v *InputValidator) SanitizeIdentifier(identifier string) string {
	// Remove any non-alphanumeric characters except underscores
	sanitized := regexp.MustCompile(`[^a-zA-Z0-9_]`).ReplaceAllString(identifier, "")
	
	// Ensure it starts with a letter
	if len(sanitized) > 0 && !regexp.MustCompile(`^[a-zA-Z]`).MatchString(sanitized) {
		sanitized = "col_" + sanitized
	}
	
	// Truncate if too long
	if len(sanitized) > 100 {
		sanitized = sanitized[:100]
	}
	
	// If empty after sanitization, use a default
	if sanitized == "" {
		sanitized = "unnamed_collection"
	}
	
	return sanitized
}

// ValidateQueryParameter validates parameters used in queries to prevent injection
func (v *InputValidator) ValidateQueryParameter(param interface{}) error {
	switch v := param.(type) {
	case string:
		// Check for SQL injection patterns in string parameters
		if strings.Contains(v, "--") || strings.Contains(v, "/*") || strings.Contains(v, "*/") {
			return fmt.Errorf("parameter contains SQL comment sequences")
		}
		// Note: We don't check for quotes here as they might be legitimate in data
		// The database driver should handle parameterized queries properly
	case []byte:
		// Binary data is generally safe if properly parameterized
		if len(v) > 10*1024*1024 { // 10MB limit for binary data
			return fmt.Errorf("binary parameter exceeds size limit")
		}
	}
	return nil
}