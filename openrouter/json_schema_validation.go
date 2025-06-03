package openrouter

import (
	"encoding/json"
	"fmt"
)

// ValidateJSONAgainstSchema validates JSON data against a schema definition
func ValidateJSONAgainstSchema(data json.RawMessage, schema *Definition) error {
	var parsed interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	
	return validateValueAgainstSchema(parsed, schema, "root")
}

func validateValueAgainstSchema(value interface{}, schema *Definition, path string) error {
	switch schema.Type {
	case Object:
		return validateObjectAgainstSchema(value, schema, path)
	case Array:
		return validateArrayAgainstSchema(value, schema, path)
	case String:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string at %s, got %T", path, value)
		}
	case Number:
		if _, ok := value.(float64); !ok {
			return fmt.Errorf("expected number at %s, got %T", path, value)
		}
	case Integer:
		if num, ok := value.(float64); !ok || num != float64(int64(num)) {
			return fmt.Errorf("expected integer at %s, got %T", path, value)
		}
	case Boolean:
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean at %s, got %T", path, value)
		}
	case Null:
		if value != nil {
			return fmt.Errorf("expected null at %s, got %T", path, value)
		}
	}
	
	return nil
}

func validateObjectAgainstSchema(value interface{}, schema *Definition, path string) error {
	obj, ok := value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected object at %s, got %T", path, value)
	}
	
	// Check required fields
	for _, required := range schema.Required {
		if _, exists := obj[required]; !exists {
			return fmt.Errorf("missing required field %s at %s", required, path)
		}
	}
	
	// Validate each property
	for key, val := range obj {
		if propSchema, exists := schema.Properties[key]; exists {
			if err := validateValueAgainstSchema(val, &propSchema, fmt.Sprintf("%s.%s", path, key)); err != nil {
				return err
			}
		} else if schema.AdditionalProperties == false {
			return fmt.Errorf("unexpected property %s at %s", key, path)
		}
	}
	
	return nil
}

func validateArrayAgainstSchema(value interface{}, schema *Definition, path string) error {
	arr, ok := value.([]interface{})
	if !ok {
		return fmt.Errorf("expected array at %s, got %T", path, value)
	}
	
	if schema.Items != nil {
		for i, item := range arr {
			if err := validateValueAgainstSchema(item, schema.Items, fmt.Sprintf("%s[%d]", path, i)); err != nil {
				return err
			}
		}
	}
	
	return nil
}