export interface SchemaProperty {
  type?: string | string[];
  format?: string;
  description?: string;
  default?: any;
  enum?: any[];
  pattern?: string;
  minLength?: number;
  maxLength?: number;
  minimum?: number;
  maximum?: number;
  minItems?: number;
  maxItems?: number;
  uniqueItems?: boolean;
  items?: SchemaProperty;
  properties?: Record<string, SchemaProperty>;
  required?: string[];
  readOnly?: boolean;
  writeOnly?: boolean;
}

export interface Schema {
  type: string;
  properties?: Record<string, SchemaProperty>;
  required?: string[];
  additionalProperties?: boolean | SchemaProperty;
  description?: string;
}

export interface ValidationError {
  field: string;
  message: string;
}

export function validateDocument(document: any, schema?: Schema | null): ValidationError[] {
  const errors: ValidationError[] = [];
  
  if (!schema || !schema.properties) {
    return errors; // No schema, no validation
  }

  // Check required fields
  if (schema.required) {
    for (const field of schema.required) {
      if (!(field in document) || document[field] === undefined || document[field] === null || document[field] === '') {
        errors.push({
          field,
          message: `Field "${field}" is required`,
        });
      }
    }
  }

  // Check each property in the document
  for (const [key, value] of Object.entries(document)) {
    const property = schema.properties[key];
    
    if (!property) {
      // Check if additional properties are allowed
      if (schema.additionalProperties === false) {
        errors.push({
          field: key,
          message: `Field "${key}" is not allowed`,
        });
      }
      continue;
    }

    // Validate the property
    const propErrors = validateProperty(key, value, property);
    errors.push(...propErrors);
  }

  return errors;
}

function validateProperty(
  fieldPath: string,
  value: any,
  property: SchemaProperty
): ValidationError[] {
  const errors: ValidationError[] = [];

  // Get the expected type(s)
  const types = Array.isArray(property.type) 
    ? property.type 
    : property.type 
    ? [property.type] 
    : [];

  if (types.length === 0) return errors; // No type specified, skip validation

  // Allow null if it's in the type array
  if (value === null) {
    if (!types.includes('null')) {
      errors.push({
        field: fieldPath,
        message: `Field "${fieldPath}" cannot be null`,
      });
    }
    return errors;
  }

  // Check type
  const actualType = getJsonType(value);
  
  // Special case: 'number' type should accept both integers and floats
  let typeMatches = false;
  if (types.includes('number') && (actualType === 'number' || actualType === 'integer')) {
    typeMatches = true;
  } else if (types.includes(actualType)) {
    typeMatches = true;
  }
  
  if (!typeMatches) {
    errors.push({
      field: fieldPath,
      message: `Field "${fieldPath}" must be of type ${types.join(' or ')}, got ${actualType}`,
    });
    return errors; // Stop validation if type is wrong
  }

  // Type-specific validations
  switch (actualType) {
    case 'string':
      errors.push(...validateString(fieldPath, value, property));
      break;
    case 'number':
    case 'integer':
      errors.push(...validateNumber(fieldPath, value, property, actualType === 'integer'));
      break;
    case 'array':
      errors.push(...validateArray(fieldPath, value, property));
      break;
    case 'object':
      errors.push(...validateObject(fieldPath, value, property));
      break;
  }

  // Check enum values
  if (property.enum && property.enum.length > 0) {
    if (!property.enum.includes(value)) {
      errors.push({
        field: fieldPath,
        message: `Field "${fieldPath}" must be one of: ${property.enum.join(', ')}`,
      });
    }
  }

  return errors;
}

function validateString(
  fieldPath: string,
  value: string,
  property: SchemaProperty
): ValidationError[] {
  const errors: ValidationError[] = [];

  // Check length constraints
  if (property.minLength !== undefined && value.length < property.minLength) {
    errors.push({
      field: fieldPath,
      message: `Field "${fieldPath}" must be at least ${property.minLength} characters`,
    });
  }

  if (property.maxLength !== undefined && value.length > property.maxLength) {
    errors.push({
      field: fieldPath,
      message: `Field "${fieldPath}" must not exceed ${property.maxLength} characters`,
    });
  }

  // Check pattern
  if (property.pattern) {
    try {
      const regex = new RegExp(property.pattern);
      if (!regex.test(value)) {
        errors.push({
          field: fieldPath,
          message: `Field "${fieldPath}" does not match pattern: ${property.pattern}`,
        });
      }
    } catch (e) {
      // Invalid regex pattern
    }
  }

  // Check format
  if (property.format) {
    const formatErrors = validateFormat(fieldPath, value, property.format);
    errors.push(...formatErrors);
  }

  return errors;
}

function validateNumber(
  fieldPath: string,
  value: number,
  property: SchemaProperty,
  mustBeInteger: boolean
): ValidationError[] {
  const errors: ValidationError[] = [];

  if (mustBeInteger && !Number.isInteger(value)) {
    errors.push({
      field: fieldPath,
      message: `Field "${fieldPath}" must be an integer`,
    });
  }

  if (property.minimum !== undefined && value < property.minimum) {
    errors.push({
      field: fieldPath,
      message: `Field "${fieldPath}" must be at least ${property.minimum}`,
    });
  }

  if (property.maximum !== undefined && value > property.maximum) {
    errors.push({
      field: fieldPath,
      message: `Field "${fieldPath}" must not exceed ${property.maximum}`,
    });
  }

  return errors;
}

function validateArray(
  fieldPath: string,
  value: any[],
  property: SchemaProperty
): ValidationError[] {
  const errors: ValidationError[] = [];

  if (property.minItems !== undefined && value.length < property.minItems) {
    errors.push({
      field: fieldPath,
      message: `Field "${fieldPath}" must have at least ${property.minItems} items`,
    });
  }

  if (property.maxItems !== undefined && value.length > property.maxItems) {
    errors.push({
      field: fieldPath,
      message: `Field "${fieldPath}" must not exceed ${property.maxItems} items`,
    });
  }

  if (property.uniqueItems) {
    const uniqueValues = new Set(value.map(v => JSON.stringify(v)));
    if (uniqueValues.size !== value.length) {
      errors.push({
        field: fieldPath,
        message: `Field "${fieldPath}" must have unique items`,
      });
    }
  }

  // Validate each item
  if (property.items) {
    value.forEach((item, index) => {
      const itemErrors = validateProperty(
        `${fieldPath}[${index}]`,
        item,
        property.items!
      );
      errors.push(...itemErrors);
    });
  }

  return errors;
}

function validateObject(
  fieldPath: string,
  value: any,
  property: SchemaProperty
): ValidationError[] {
  const errors: ValidationError[] = [];

  if (property.properties) {
    // Check required fields
    if (property.required) {
      for (const field of property.required) {
        if (!(field in value)) {
          errors.push({
            field: `${fieldPath}.${field}`,
            message: `Field "${fieldPath}.${field}" is required`,
          });
        }
      }
    }

    // Validate each nested property
    for (const [key, val] of Object.entries(value)) {
      const nestedProp = property.properties[key];
      if (nestedProp) {
        const nestedErrors = validateProperty(
          `${fieldPath}.${key}`,
          val,
          nestedProp
        );
        errors.push(...nestedErrors);
      }
    }
  }

  return errors;
}

function validateFormat(
  fieldPath: string,
  value: string,
  format: string
): ValidationError[] {
  const errors: ValidationError[] = [];

  switch (format) {
    case 'email':
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      if (!emailRegex.test(value)) {
        errors.push({
          field: fieldPath,
          message: `Field "${fieldPath}" must be a valid email address`,
        });
      }
      break;

    case 'date':
      // YYYY-MM-DD format
      const dateRegex = /^\d{4}-\d{2}-\d{2}$/;
      if (!dateRegex.test(value)) {
        errors.push({
          field: fieldPath,
          message: `Field "${fieldPath}" must be a valid date (YYYY-MM-DD)`,
        });
      }
      break;

    case 'date-time':
      // ISO 8601 format
      const dateTime = new Date(value);
      if (isNaN(dateTime.getTime())) {
        errors.push({
          field: fieldPath,
          message: `Field "${fieldPath}" must be a valid date-time`,
        });
      }
      break;

    case 'uri':
    case 'url':
      try {
        new URL(value);
      } catch {
        errors.push({
          field: fieldPath,
          message: `Field "${fieldPath}" must be a valid URL`,
        });
      }
      break;

    case 'uuid':
      const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
      if (!uuidRegex.test(value)) {
        errors.push({
          field: fieldPath,
          message: `Field "${fieldPath}" must be a valid UUID`,
        });
      }
      break;

    case 'ipv4':
      const ipv4Regex = /^(\d{1,3}\.){3}\d{1,3}$/;
      if (!ipv4Regex.test(value)) {
        errors.push({
          field: fieldPath,
          message: `Field "${fieldPath}" must be a valid IPv4 address`,
        });
      }
      break;

    case 'ipv6':
      const ipv6Regex = /^([\da-f]{1,4}:){7}[\da-f]{1,4}$/i;
      if (!ipv6Regex.test(value)) {
        errors.push({
          field: fieldPath,
          message: `Field "${fieldPath}" must be a valid IPv6 address`,
        });
      }
      break;
  }

  return errors;
}

function getJsonType(value: any): string {
  if (value === null) return 'null';
  if (Array.isArray(value)) return 'array';
  
  const type = typeof value;
  if (type === 'number') {
    return Number.isInteger(value) ? 'integer' : 'number';
  }
  
  return type;
}