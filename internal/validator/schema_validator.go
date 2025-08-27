package validator

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/karthik/anybase/pkg/models"
)

// SchemaValidator validates documents against OpenAPI schema
type SchemaValidator struct{}

// NewSchemaValidator creates a new schema validator
func NewSchemaValidator() *SchemaValidator {
	return &SchemaValidator{}
}

// ValidateDocument validates a document against a collection schema
func (v *SchemaValidator) ValidateDocument(doc map[string]interface{}, schema *models.CollectionSchema) error {
	if schema == nil {
		return nil // No schema, no validation
	}

	// Check required fields
	for _, required := range schema.Required {
		if _, ok := doc[required]; !ok {
			return fmt.Errorf("required field '%s' is missing", required)
		}
	}

	// Validate each property
	for key, value := range doc {
		// Skip internal fields
		if strings.HasPrefix(key, "_") {
			continue
		}

		prop, ok := schema.Properties[key]
		if !ok {
			// Check if additional properties are allowed
			if schema.AdditionalProperties == false {
				return fmt.Errorf("property '%s' is not allowed", key)
			}
			continue
		}

		if err := v.validateValue(value, prop, key); err != nil {
			return err
		}
	}

	return nil
}

// validateValue validates a value against a schema property
func (v *SchemaValidator) validateValue(value interface{}, prop *models.SchemaProperty, path string) error {
	if value == nil {
		// Check if null is allowed
		if prop.Type == "null" {
			return nil
		}
		if arr, ok := prop.Type.([]interface{}); ok {
			for _, t := range arr {
				if t == "null" {
					return nil
				}
			}
		}
		return fmt.Errorf("field '%s' cannot be null", path)
	}

	// Get the type(s)
	types := v.getTypes(prop.Type)
	
	// Try to validate against any of the allowed types
	var lastErr error
	for _, t := range types {
		if err := v.validateType(value, t, prop, path); err == nil {
			return nil // Valid for this type
		} else {
			lastErr = err
		}
	}
	
	if lastErr != nil {
		return lastErr
	}

	return fmt.Errorf("field '%s' does not match any allowed type", path)
}

// getTypes extracts types from type definition
func (v *SchemaValidator) getTypes(typeValue interface{}) []string {
	switch t := typeValue.(type) {
	case string:
		return []string{t}
	case []interface{}:
		types := []string{}
		for _, tt := range t {
			if str, ok := tt.(string); ok {
				types = append(types, str)
			}
		}
		return types
	default:
		return []string{}
	}
}

// validateType validates value against a specific type
func (v *SchemaValidator) validateType(value interface{}, dataType string, prop *models.SchemaProperty, path string) error {
	switch dataType {
	case "string":
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("field '%s' must be a string", path)
		}
		return v.validateString(str, prop, path)
		
	case "number", "integer":
		var num float64
		switch n := value.(type) {
		case float64:
			num = n
		case int:
			num = float64(n)
		case int64:
			num = float64(n)
		case int32:
			num = float64(n)
		default:
			return fmt.Errorf("field '%s' must be a number", path)
		}
		
		if dataType == "integer" {
			if num != float64(int64(num)) {
				return fmt.Errorf("field '%s' must be an integer", path)
			}
		}
		
		return v.validateNumber(num, prop, path)
		
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("field '%s' must be a boolean", path)
		}
		
	case "array":
		arr, ok := value.([]interface{})
		if !ok {
			// Try to convert from other slice types
			if jsonBytes, err := json.Marshal(value); err == nil {
				if err := json.Unmarshal(jsonBytes, &arr); err != nil {
					return fmt.Errorf("field '%s' must be an array", path)
				}
			} else {
				return fmt.Errorf("field '%s' must be an array", path)
			}
		}
		return v.validateArray(arr, prop, path)
		
	case "object":
		obj, ok := value.(map[string]interface{})
		if !ok {
			return fmt.Errorf("field '%s' must be an object", path)
		}
		return v.validateObject(obj, prop, path)
	}
	
	return nil
}

// validateString validates string constraints
func (v *SchemaValidator) validateString(str string, prop *models.SchemaProperty, path string) error {
	// Check enum
	if len(prop.Enum) > 0 {
		found := false
		for _, e := range prop.Enum {
			if str == e {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("field '%s' must be one of: %v", path, prop.Enum)
		}
	}

	// Check pattern
	if prop.Pattern != "" {
		matched, err := regexp.MatchString(prop.Pattern, str)
		if err != nil {
			return fmt.Errorf("invalid pattern for field '%s': %v", path, err)
		}
		if !matched {
			return fmt.Errorf("field '%s' does not match pattern: %s", path, prop.Pattern)
		}
	}

	// Check length constraints
	if prop.MinLength != nil && len(str) < *prop.MinLength {
		return fmt.Errorf("field '%s' is too short (minimum length: %d)", path, *prop.MinLength)
	}
	if prop.MaxLength != nil && len(str) > *prop.MaxLength {
		return fmt.Errorf("field '%s' is too long (maximum length: %d)", path, *prop.MaxLength)
	}

	// Check format
	if prop.Format != "" {
		if err := v.validateFormat(str, prop.Format, path); err != nil {
			return err
		}
	}

	return nil
}

// validateNumber validates number constraints
func (v *SchemaValidator) validateNumber(num float64, prop *models.SchemaProperty, path string) error {
	// Check enum
	if len(prop.Enum) > 0 {
		found := false
		for _, e := range prop.Enum {
			if enumNum, ok := e.(float64); ok && num == enumNum {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("field '%s' must be one of: %v", path, prop.Enum)
		}
	}

	// Check range constraints
	if prop.Minimum != nil && num < *prop.Minimum {
		return fmt.Errorf("field '%s' is too small (minimum: %f)", path, *prop.Minimum)
	}
	if prop.Maximum != nil && num > *prop.Maximum {
		return fmt.Errorf("field '%s' is too large (maximum: %f)", path, *prop.Maximum)
	}

	return nil
}

// validateArray validates array constraints
func (v *SchemaValidator) validateArray(arr []interface{}, prop *models.SchemaProperty, path string) error {
	// Check length constraints
	if prop.MinItems != nil && len(arr) < *prop.MinItems {
		return fmt.Errorf("field '%s' has too few items (minimum: %d)", path, *prop.MinItems)
	}
	if prop.MaxItems != nil && len(arr) > *prop.MaxItems {
		return fmt.Errorf("field '%s' has too many items (maximum: %d)", path, *prop.MaxItems)
	}

	// Check unique items
	if prop.UniqueItems {
		seen := make(map[string]bool)
		for _, item := range arr {
			key := fmt.Sprintf("%v", item)
			if seen[key] {
				return fmt.Errorf("field '%s' must have unique items", path)
			}
			seen[key] = true
		}
	}

	// Validate each item
	if prop.Items != nil {
		for i, item := range arr {
			itemPath := fmt.Sprintf("%s[%d]", path, i)
			if err := v.validateValue(item, prop.Items, itemPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// validateObject validates object constraints
func (v *SchemaValidator) validateObject(obj map[string]interface{}, prop *models.SchemaProperty, path string) error {
	// Check required fields
	for _, required := range prop.Required {
		if _, ok := obj[required]; !ok {
			return fmt.Errorf("required field '%s.%s' is missing", path, required)
		}
	}

	// Validate properties
	if prop.Properties != nil {
		for key, value := range obj {
			if subProp, ok := prop.Properties[key]; ok {
				subPath := fmt.Sprintf("%s.%s", path, key)
				if err := v.validateValue(value, subProp, subPath); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// validateFormat validates string format
func (v *SchemaValidator) validateFormat(str string, format string, path string) error {
	switch format {
	case "date-time":
		if _, err := time.Parse(time.RFC3339, str); err != nil {
			return fmt.Errorf("field '%s' must be a valid date-time (RFC3339 format)", path)
		}
	case "date":
		if _, err := time.Parse("2006-01-02", str); err != nil {
			return fmt.Errorf("field '%s' must be a valid date (YYYY-MM-DD format)", path)
		}
	case "email":
		emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
		if matched, _ := regexp.MatchString(emailRegex, str); !matched {
			return fmt.Errorf("field '%s' must be a valid email address", path)
		}
	case "uri", "url":
		urlRegex := `^(https?|ftp)://[^\s/$.?#].[^\s]*$`
		if matched, _ := regexp.MatchString(urlRegex, str); !matched {
			return fmt.Errorf("field '%s' must be a valid URL", path)
		}
	case "uuid":
		uuidRegex := `^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`
		if matched, _ := regexp.MatchString(uuidRegex, strings.ToLower(str)); !matched {
			return fmt.Errorf("field '%s' must be a valid UUID", path)
		}
	}

	return nil
}