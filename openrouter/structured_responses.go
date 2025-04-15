package openrouter

import (
	"encoding/json"
	"regexp"
	"strings"
)

func PseudoStructuredResponseCleaner(response string) string {
	// Find the first opening brace
	firstBrace := strings.Index(response, "{")
	if firstBrace == -1 {
		// No JSON object found
		return response
	}

	// Find the last closing brace
	lastBrace := strings.LastIndex(response, "}")
	if lastBrace == -1 || lastBrace < firstBrace {
		// No valid closing brace found
		return response
	}

	// Extract only the content between the first { and last }
	return response[firstBrace : lastBrace+1]
}

// UseOpenRouterJsonFormat creates a JSON schema response format for OpenRouter requests
// based on the provided example response object
func UseOpenRouterJsonFormat(exampleOutput any, schemaName string) (json.RawMessage, error) {
	// Generate schema definition for the example output type
	schemaDef, err := GenerateSchemaForType(exampleOutput)
	if err != nil {
		return nil, err
	}

	// Convert the Definition to JSON for the schema
	schemaBytes, err := json.Marshal(schemaDef)
	if err != nil {
		return nil, err
	}

	// Clean the schema name to ensure it's valid
	safeName := regexp.MustCompile(`[^a-zA-Z0-9]+`).ReplaceAllString(schemaName, "_")

	// Create a response format structure with the JSON schema
	responseFormat := map[string]interface{}{
		"type": "json_schema",
		"json_schema": map[string]interface{}{
			"name":   safeName,
			"schema": json.RawMessage(schemaBytes),
			"strict": true,
		},
	}

	// Marshal the response format to JSON
	return json.Marshal(responseFormat)
}
