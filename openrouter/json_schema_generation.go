package openrouter

import (
	"encoding/json"
	"fmt"
)

// GenerateSchemaFromJSONExample creates a schema definition from a JSON example
func GenerateSchemaFromJSONExample(example json.RawMessage) (*Definition, error) {
	if len(example) == 0 {
		return nil, fmt.Errorf("empty JSON example")
	}
	
	// Parse the JSON to determine its structure
	var parsed interface{}
	if err := json.Unmarshal(example, &parsed); err != nil {
		return nil, fmt.Errorf("invalid JSON example: %w", err)
	}
	
	// Generate schema from the parsed structure
	return generateSchemaFromInterface(parsed)
}

// generateSchemaFromInterface recursively builds schema from parsed JSON
func generateSchemaFromInterface(v interface{}) (*Definition, error) {
	switch val := v.(type) {
	case map[string]interface{}:
		return generateObjectSchema(val)
	case []interface{}:
		return generateArraySchema(val)
	case string:
		return &Definition{Type: String}, nil
	case float64:
		// JSON numbers are always float64
		if val == float64(int64(val)) {
			return &Definition{Type: Integer}, nil
		}
		return &Definition{Type: Number}, nil
	case bool:
		return &Definition{Type: Boolean}, nil
	case nil:
		return &Definition{Type: Null}, nil
	default:
		return nil, fmt.Errorf("unsupported type: %T", v)
	}
}

func generateObjectSchema(obj map[string]interface{}) (*Definition, error) {
	def := &Definition{
		Type:                 Object,
		Properties:           make(map[string]Definition),
		AdditionalProperties: false,
	}
	
	var required []string
	
	for key, value := range obj {
		propSchema, err := generateSchemaFromInterface(value)
		if err != nil {
			return nil, fmt.Errorf("error generating schema for property %s: %w", key, err)
		}
		
		def.Properties[key] = *propSchema
		required = append(required, key)
	}
	
	def.Required = required
	return def, nil
}

func generateArraySchema(arr []interface{}) (*Definition, error) {
	def := &Definition{Type: Array}
	
	if len(arr) > 0 {
		// Use first element to determine item schema
		itemSchema, err := generateSchemaFromInterface(arr[0])
		if err != nil {
			return nil, fmt.Errorf("error generating array item schema: %w", err)
		}
		def.Items = itemSchema
	}
	
	return def, nil
}