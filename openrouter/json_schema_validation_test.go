package openrouter

import (
	"encoding/json"
	"testing"
)

func TestValidateJSONAgainstSchema(t *testing.T) {
	// Define a test schema
	schema := &Definition{
		Type: Object,
		Properties: map[string]Definition{
			"name": {Type: String},
			"age":  {Type: Integer},
			"active": {Type: Boolean},
			"tags": {
				Type:  Array,
				Items: &Definition{Type: String},
			},
			"metadata": {
				Type: Object,
				Properties: map[string]Definition{
					"created": {Type: String},
					"version": {Type: Integer},
				},
				Required: []string{"created", "version"},
				AdditionalProperties: false,
			},
		},
		Required: []string{"name", "age"},
		AdditionalProperties: false,
	}

	tests := []struct {
		name     string
		data     json.RawMessage
		hasError bool
		errorMsg string
	}{
		{
			name:     "valid complete object",
			data:     json.RawMessage(`{"name": "John", "age": 30, "active": true, "tags": ["tag1", "tag2"], "metadata": {"created": "2023-01-01", "version": 1}}`),
			hasError: false,
		},
		{
			name:     "valid minimal object",
			data:     json.RawMessage(`{"name": "John", "age": 30}`),
			hasError: false,
		},
		{
			name:     "missing required field",
			data:     json.RawMessage(`{"name": "John"}`),
			hasError: true,
			errorMsg: "missing required field age",
		},
		{
			name:     "wrong type for field",
			data:     json.RawMessage(`{"name": "John", "age": "thirty"}`),
			hasError: true,
			errorMsg: "expected integer",
		},
		{
			name:     "unexpected additional property",
			data:     json.RawMessage(`{"name": "John", "age": 30, "unexpected": "value"}`),
			hasError: true,
			errorMsg: "unexpected property unexpected",
		},
		{
			name:     "invalid array item type",
			data:     json.RawMessage(`{"name": "John", "age": 30, "tags": ["tag1", 123]}`),
			hasError: true,
			errorMsg: "expected string",
		},
		{
			name:     "invalid nested object",
			data:     json.RawMessage(`{"name": "John", "age": 30, "metadata": {"created": "2023-01-01"}}`),
			hasError: true,
			errorMsg: "missing required field version",
		},
		{
			name:     "invalid JSON",
			data:     json.RawMessage(`{invalid json`),
			hasError: true,
			errorMsg: "invalid JSON",
		},
		{
			name:     "float as integer (should fail)",
			data:     json.RawMessage(`{"name": "John", "age": 30.5}`),
			hasError: true,
			errorMsg: "expected integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateJSONAgainstSchema(tt.data, schema)
			
			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got: %s", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateValueAgainstSchema(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		schema   *Definition
		hasError bool
	}{
		{
			name:     "valid string",
			value:    "test",
			schema:   &Definition{Type: String},
			hasError: false,
		},
		{
			name:     "invalid string",
			value:    123,
			schema:   &Definition{Type: String},
			hasError: true,
		},
		{
			name:     "valid integer",
			value:    float64(42),
			schema:   &Definition{Type: Integer},
			hasError: false,
		},
		{
			name:     "invalid integer (float)",
			value:    float64(42.5),
			schema:   &Definition{Type: Integer},
			hasError: true,
		},
		{
			name:     "valid number",
			value:    float64(42.5),
			schema:   &Definition{Type: Number},
			hasError: false,
		},
		{
			name:     "invalid number",
			value:    "not a number",
			schema:   &Definition{Type: Number},
			hasError: true,
		},
		{
			name:     "valid boolean",
			value:    true,
			schema:   &Definition{Type: Boolean},
			hasError: false,
		},
		{
			name:     "invalid boolean",
			value:    "true",
			schema:   &Definition{Type: Boolean},
			hasError: true,
		},
		{
			name:     "valid null",
			value:    nil,
			schema:   &Definition{Type: Null},
			hasError: false,
		},
		{
			name:     "invalid null",
			value:    "not null",
			schema:   &Definition{Type: Null},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateValueAgainstSchema(tt.value, tt.schema, "test")
			
			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateArrayAgainstSchema(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		schema   *Definition
		hasError bool
	}{
		{
			name:  "valid string array",
			value: []interface{}{"a", "b", "c"},
			schema: &Definition{
				Type:  Array,
				Items: &Definition{Type: String},
			},
			hasError: false,
		},
		{
			name:  "invalid array item type",
			value: []interface{}{"a", 123, "c"},
			schema: &Definition{
				Type:  Array,
				Items: &Definition{Type: String},
			},
			hasError: true,
		},
		{
			name:  "empty array",
			value: []interface{}{},
			schema: &Definition{
				Type:  Array,
				Items: &Definition{Type: String},
			},
			hasError: false,
		},
		{
			name:     "not an array",
			value:    "not an array",
			schema:   &Definition{Type: Array},
			hasError: true,
		},
		{
			name:     "array without items schema",
			value:    []interface{}{"a", "b"},
			schema:   &Definition{Type: Array},
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateArrayAgainstSchema(tt.value, tt.schema, "test")
			
			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())))
}