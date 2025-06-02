package openrouter

import (
	"encoding/json"
	"regexp"
	"strings"
)

func PseudoStructuredResponseCleaner(response string) string {
	// Find the first opening brace or bracket
	firstBrace := strings.Index(response, "{")
	firstBracket := strings.Index(response, "[")
	
	var firstChar int = -1
	var isArray bool
	
	// Determine which comes first: { or [
	if firstBrace != -1 && firstBracket != -1 {
		if firstBrace < firstBracket {
			firstChar = firstBrace
			isArray = false
		} else {
			firstChar = firstBracket
			isArray = true
		}
	} else if firstBrace != -1 {
		firstChar = firstBrace
		isArray = false
	} else if firstBracket != -1 {
		firstChar = firstBracket
		isArray = true
	} else {
		// No JSON found
		return response
	}

	// Find the corresponding closing character
	var lastChar int
	if isArray {
		lastChar = strings.LastIndex(response, "]")
		if lastChar == -1 || lastChar < firstChar {
			// No valid closing bracket found
			return response
		}
	} else {
		lastChar = strings.LastIndex(response, "}")
		if lastChar == -1 || lastChar < firstChar {
			// No valid closing brace found
			return response
		}
	}

	// Extract only the content between the first and last characters
	return response[firstChar : lastChar+1]
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
