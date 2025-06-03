package openrouter

import (
	"encoding/json"
	"testing"
)

func TestUseOpenRouterJsonFormatFromJSON(t *testing.T) {
	tests := []struct {
		name        string
		jsonExample string
		schemaName  string
		wantError   bool
		description string
	}{
		{
			name:        "Simple Object",
			jsonExample: `{"sentiment": "positive", "confidence": 0.95}`,
			schemaName:  "sentiment_analysis",
			wantError:   false,
			description: "Basic object with string and number fields",
		},
		{
			name:        "Complex Object with Array",
			jsonExample: `{"summary": "text", "key_points": ["point1", "point2"], "word_count": 150}`,
			schemaName:  "text_summary",
			wantError:   false,
			description: "Object with string, array, and number fields",
		},
		{
			name:        "Nested Object",
			jsonExample: `{"result": {"status": "success", "data": {"value": 42}}, "timestamp": "2023-01-01"}`,
			schemaName:  "nested_response",
			wantError:   false,
			description: "Nested objects to test recursion",
		},
		{
			name:        "Array of Objects",
			jsonExample: `[{"id": 1, "name": "item1"}, {"id": 2, "name": "item2"}]`,
			schemaName:  "item_list",
			wantError:   false,
			description: "Array containing objects",
		},
		{
			name:        "Mixed Types",
			jsonExample: `{"string_field": "text", "number_field": 123.45, "boolean_field": true, "null_field": null}`,
			schemaName:  "mixed_types",
			wantError:   false,
			description: "All basic JSON types",
		},
		{
			name:        "Empty JSON",
			jsonExample: ``,
			schemaName:  "empty",
			wantError:   true,
			description: "Should fail on empty input",
		},
		{
			name:        "Invalid JSON",
			jsonExample: `{"invalid": json}`,
			schemaName:  "invalid",
			wantError:   true,
			description: "Should fail on malformed JSON",
		},
		{
			name:        "Special Characters in Schema Name",
			jsonExample: `{"test": "value"}`,
			schemaName:  "test-schema with spaces & symbols!",
			wantError:   false,
			description: "Schema name should be sanitized",
		},
		{
			name:        "Empty Schema Name",
			jsonExample: `{"test": "value"}`,
			schemaName:  "",
			wantError:   false,
			description: "Should handle empty schema name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := UseOpenRouterJsonFormatFromJSON(json.RawMessage(tt.jsonExample), tt.schemaName)
			
			if tt.wantError {
				if err == nil {
					t.Errorf("UseOpenRouterJsonFormatFromJSON() expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("UseOpenRouterJsonFormatFromJSON() unexpected error: %v", err)
				return
			}
			
			// Validate the result is valid JSON
			var responseFormat map[string]interface{}
			if err := json.Unmarshal(result, &responseFormat); err != nil {
				t.Errorf("Result is not valid JSON: %v", err)
				return
			}
			
			// Validate OpenRouter response format structure
			if err := ValidateJSONResponseFormat(result); err != nil {
				t.Errorf("Invalid OpenRouter response format: %v", err)
				return
			}
			
			// Check required fields
			if responseFormat["type"] != "json_schema" {
				t.Errorf("Expected type 'json_schema', got %v", responseFormat["type"])
			}
			
			jsonSchema, ok := responseFormat["json_schema"].(map[string]interface{})
			if !ok {
				t.Errorf("json_schema field is not an object")
				return
			}
			
			// Check schema name is present and sanitized
			name, ok := jsonSchema["name"].(string)
			if !ok {
				t.Errorf("json_schema.name is not a string")
				return
			}
			
			if name == "" {
				t.Errorf("json_schema.name should not be empty")
			}
			
			// Check schema field exists
			if _, ok := jsonSchema["schema"]; !ok {
				t.Errorf("json_schema.schema field is missing")
			}
			
			// Check strict mode is enabled
			if strict, ok := jsonSchema["strict"].(bool); !ok || !strict {
				t.Errorf("json_schema.strict should be true")
			}
		})
	}
}

func TestValidateJSONResponseFormat(t *testing.T) {
	tests := []struct {
		name          string
		responseFormat string
		wantError     bool
		description   string
	}{
		{
			name: "Valid Response Format",
			responseFormat: `{
				"type": "json_schema",
				"json_schema": {
					"name": "test_schema",
					"schema": {"type": "object"},
					"strict": true
				}
			}`,
			wantError:   false,
			description: "Valid OpenRouter response format",
		},
		{
			name:          "Invalid JSON",
			responseFormat: `{invalid json}`,
			wantError:     true,
			description:   "Malformed JSON should fail",
		},
		{
			name: "Wrong Type",
			responseFormat: `{
				"type": "text",
				"json_schema": {
					"name": "test_schema",
					"schema": {"type": "object"}
				}
			}`,
			wantError:   true,
			description: "Type must be 'json_schema'",
		},
		{
			name: "Missing json_schema Field",
			responseFormat: `{
				"type": "json_schema"
			}`,
			wantError:   true,
			description: "json_schema field is required",
		},
		{
			name: "Missing Name",
			responseFormat: `{
				"type": "json_schema",
				"json_schema": {
					"schema": {"type": "object"}
				}
			}`,
			wantError:   true,
			description: "json_schema.name is required",
		},
		{
			name: "Missing Schema",
			responseFormat: `{
				"type": "json_schema",
				"json_schema": {
					"name": "test_schema"
				}
			}`,
			wantError:   true,
			description: "json_schema.schema is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateJSONResponseFormat(json.RawMessage(tt.responseFormat))
			
			if tt.wantError {
				if err == nil {
					t.Errorf("ValidateJSONResponseFormat() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("ValidateJSONResponseFormat() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestSchemaNameSanitization(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple_name", "simple_name"},
		{"name with spaces", "name_with_spaces"},
		{"name-with-dashes", "name_with_dashes"},
		{"name@with#symbols!", "name_with_symbols_"},
		{"123numeric", "123numeric"},
		{"", "generated_schema"},
		{"   ", "_"},
		{"CamelCase", "CamelCase"},
		{"___multiple___underscores___", "_multiple_underscores_"},
		{"@#$%", "_"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := UseOpenRouterJsonFormatFromJSON(
				json.RawMessage(`{"test": "value"}`),
				tt.input,
			)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			var responseFormat map[string]interface{}
			if err := json.Unmarshal(result, &responseFormat); err != nil {
				t.Errorf("Result is not valid JSON: %v", err)
				return
			}

			jsonSchema := responseFormat["json_schema"].(map[string]interface{})
			actualName := jsonSchema["name"].(string)

			if actualName != tt.expected {
				t.Errorf("Expected sanitized name '%s', got '%s'", tt.expected, actualName)
			}
		})
	}
}

// Benchmark tests to ensure performance is acceptable
func BenchmarkUseOpenRouterJsonFormatFromJSON(b *testing.B) {
	jsonExample := json.RawMessage(`{
		"sentiment": "positive",
		"confidence": 0.95,
		"reasoning": "Contains positive language",
		"details": {
			"keywords": ["love", "great", "excellent"],
			"score": 8.5,
			"metadata": {
				"processed_at": "2023-01-01T00:00:00Z",
				"version": "1.0"
			}
		}
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := UseOpenRouterJsonFormatFromJSON(jsonExample, "sentiment_analysis")
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}