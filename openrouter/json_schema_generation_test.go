package openrouter

import (
	"encoding/json"
	"testing"
)

func TestGenerateSchemaFromJSONExample(t *testing.T) {
	tests := []struct {
		name     string
		example  json.RawMessage
		expected *Definition
		hasError bool
	}{
		{
			name:    "simple object",
			example: json.RawMessage(`{"name": "test", "age": 25}`),
			expected: &Definition{
				Type:                 Object,
				AdditionalProperties: false,
				Properties: map[string]Definition{
					"name": {Type: String},
					"age":  {Type: Integer},
				},
				Required: []string{"name", "age"},
			},
			hasError: false,
		},
		{
			name:    "object with nested object",
			example: json.RawMessage(`{"user": {"name": "test"}, "count": 42}`),
			expected: &Definition{
				Type:                 Object,
				AdditionalProperties: false,
				Properties: map[string]Definition{
					"user": {
						Type:                 Object,
						AdditionalProperties: false,
						Properties: map[string]Definition{
							"name": {Type: String},
						},
						Required: []string{"name"},
					},
					"count": {Type: Integer},
				},
				Required: []string{"user", "count"},
			},
			hasError: false,
		},
		{
			name:    "object with array",
			example: json.RawMessage(`{"tags": ["tag1", "tag2"], "active": true}`),
			expected: &Definition{
				Type:                 Object,
				AdditionalProperties: false,
				Properties: map[string]Definition{
					"tags": {
						Type: Array,
						Items: &Definition{Type: String},
					},
					"active": {Type: Boolean},
				},
				Required: []string{"tags", "active"},
			},
			hasError: false,
		},
		{
			name:    "object with float",
			example: json.RawMessage(`{"price": 19.99, "discount": 0.1}`),
			expected: &Definition{
				Type:                 Object,
				AdditionalProperties: false,
				Properties: map[string]Definition{
					"price":    {Type: Number},
					"discount": {Type: Number},
				},
				Required: []string{"price", "discount"},
			},
			hasError: false,
		},
		{
			name:     "empty JSON",
			example:  json.RawMessage(``),
			expected: nil,
			hasError: true,
		},
		{
			name:     "invalid JSON",
			example:  json.RawMessage(`{invalid json`),
			expected: nil,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateSchemaFromJSONExample(tt.example)
			
			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if !compareDefinitions(result, tt.expected) {
				t.Errorf("Schema mismatch.\nExpected: %+v\nGot: %+v", tt.expected, result)
			}
		})
	}
}

func TestGenerateSchemaFromInterface(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected *Definition
		hasError bool
	}{
		{
			name:     "string",
			input:    "test",
			expected: &Definition{Type: String},
			hasError: false,
		},
		{
			name:     "integer (from float64)",
			input:    float64(42),
			expected: &Definition{Type: Integer},
			hasError: false,
		},
		{
			name:     "float",
			input:    float64(42.5),
			expected: &Definition{Type: Number},
			hasError: false,
		},
		{
			name:     "boolean",
			input:    true,
			expected: &Definition{Type: Boolean},
			hasError: false,
		},
		{
			name:     "null",
			input:    nil,
			expected: &Definition{Type: Null},
			hasError: false,
		},
		{
			name:     "empty array",
			input:    []interface{}{},
			expected: &Definition{Type: Array},
			hasError: false,
		},
		{
			name:  "string array",
			input: []interface{}{"a", "b"},
			expected: &Definition{
				Type:  Array,
				Items: &Definition{Type: String},
			},
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := generateSchemaFromInterface(tt.input)
			
			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if !compareDefinitions(result, tt.expected) {
				t.Errorf("Schema mismatch.\nExpected: %+v\nGot: %+v", tt.expected, result)
			}
		})
	}
}

// Helper function to compare Definition structs
func compareDefinitions(a, b *Definition) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	
	if a.Type != b.Type {
		return false
	}
	
	if a.Description != b.Description {
		return false
	}
	
	if !compareStringSlices(a.Required, b.Required) {
		return false
	}
	
	if !compareStringSlices(a.Enum, b.Enum) {
		return false
	}
	
	if a.AdditionalProperties != b.AdditionalProperties {
		return false
	}
	
	// Compare Items
	if !compareDefinitions(a.Items, b.Items) {
		return false
	}
	
	// Compare Properties
	if len(a.Properties) != len(b.Properties) {
		return false
	}
	
	for key, propA := range a.Properties {
		propB, exists := b.Properties[key]
		if !exists {
			return false
		}
		if !compareDefinitions(&propA, &propB) {
			return false
		}
	}
	
	return true
}

// Helper function to compare string slices (order doesn't matter for Required fields)
func compareStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	
	// Create maps to check existence
	mapA := make(map[string]bool)
	mapB := make(map[string]bool)
	
	for _, s := range a {
		mapA[s] = true
	}
	for _, s := range b {
		mapB[s] = true
	}
	
	for key := range mapA {
		if !mapB[key] {
			return false
		}
	}
	
	return true
}