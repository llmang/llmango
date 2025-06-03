package openrouter

import (
	"encoding/json"
	"fmt"
	"regexp"
)

// UseOpenRouterJsonFormatFromJSON creates a JSON schema response format for OpenRouter requests
// specifically designed for JSON examples (not Go structs). This is optimized for the dual-path
// execution system where we work with JSON data throughout the pipeline.
func UseOpenRouterJsonFormatFromJSON(jsonExample json.RawMessage, schemaName string) (json.RawMessage, error) {
	if len(jsonExample) == 0 {
		return nil, fmt.Errorf("empty JSON example provided")
	}

	// Generate schema definition directly from JSON example
	// This uses the existing GenerateSchemaFromJSONExample which is designed for this purpose
	schemaDef, err := GenerateSchemaFromJSONExample(jsonExample)
	if err != nil {
		return nil, fmt.Errorf("failed to generate schema from JSON example: %w", err)
	}

	// Convert the Definition to JSON for the schema
	schemaBytes, err := json.Marshal(schemaDef)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema definition: %w", err)
	}

	// Clean the schema name to ensure it's valid for OpenRouter (same as original)
	safeName := regexp.MustCompile(`[^a-zA-Z0-9]+`).ReplaceAllString(schemaName, "_")
	if safeName == "" {
		safeName = "generated_schema"
	}

	// Create OpenRouter-compatible response format structure
	responseFormat := map[string]interface{}{
		"type": "json_schema",
		"json_schema": map[string]interface{}{
			"name":   safeName,
			"schema": json.RawMessage(schemaBytes),
			"strict": true, // Enable strict mode for better compliance
		},
	}

	// Marshal the response format to JSON
	result, err := json.Marshal(responseFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response format: %w", err)
	}

	return result, nil
}

// ValidateJSONResponseFormat validates that a response format is properly structured
// for OpenRouter's JSON schema requirements
func ValidateJSONResponseFormat(responseFormat json.RawMessage) error {
	var format map[string]interface{}
	if err := json.Unmarshal(responseFormat, &format); err != nil {
		return fmt.Errorf("invalid response format JSON: %w", err)
	}

	// Check required fields
	if format["type"] != "json_schema" {
		return fmt.Errorf("response format type must be 'json_schema', got: %v", format["type"])
	}

	jsonSchema, ok := format["json_schema"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("json_schema field must be an object")
	}

	if _, ok := jsonSchema["name"].(string); !ok {
		return fmt.Errorf("json_schema.name must be a string")
	}

	if _, ok := jsonSchema["schema"]; !ok {
		return fmt.Errorf("json_schema.schema is required")
	}

	return nil
}