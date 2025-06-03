package openrouter

import (
	"encoding/json"
	"testing"

	"github.com/llmang/llmango/testhelpers"
)

type TestOutputStruct struct {
	Sentiment  string  `json:"sentiment"`
	Confidence float64 `json:"confidence"`
	Reasoning  string  `json:"reasoning"`
}

func TestUseOpenRouterJsonFormat(t *testing.T) {
	tests := []struct {
		name           string
		exampleOutput  interface{}
		schemaName     string
		wantErr        bool
		validateSchema func(t *testing.T, result json.RawMessage)
	}{
		{
			name: "simple struct",
			exampleOutput: TestOutputStruct{
				Sentiment:  "positive",
				Confidence: 0.9,
				Reasoning:  "test reasoning",
			},
			schemaName: "sentiment-analysis",
			wantErr:    false,
			validateSchema: func(t *testing.T, result json.RawMessage) {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("validateSchema panicked: %v", r)
					}
				}()
				
				var responseFormat map[string]interface{}
				err := json.Unmarshal(result, &responseFormat)
				testhelpers.RequireNoError(t, err)

				testhelpers.AssertEqual(t, "json_schema", responseFormat["type"])

				jsonSchema, ok := responseFormat["json_schema"].(map[string]interface{})
				testhelpers.AssertTrue(t, ok, "json_schema should be a map")

				testhelpers.AssertEqual(t, "sentiment_analysis", jsonSchema["name"])
				testhelpers.AssertEqual(t, true, jsonSchema["strict"])
				testhelpers.AssertNotNil(t, jsonSchema["schema"])
			},
		},
		{
			name: "nested struct",
			exampleOutput: struct {
				User struct {
					Name string `json:"name"`
					Age  int    `json:"age"`
				} `json:"user"`
				Score float64 `json:"score"`
			}{
				User:  struct{ Name string `json:"name"`; Age int `json:"age"` }{Name: "John", Age: 30},
				Score: 95.5,
			},
			schemaName: "nested-test",
			wantErr:    false,
			validateSchema: func(t *testing.T, result json.RawMessage) {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("validateSchema panicked: %v", r)
					}
				}()
				
				var responseFormat map[string]interface{}
				err := json.Unmarshal(result, &responseFormat)
				testhelpers.RequireNoError(t, err)

				jsonSchemaRaw, exists := responseFormat["json_schema"]
				if !exists {
					t.Fatalf("json_schema field not found in response")
				}
				
				jsonSchema, ok := jsonSchemaRaw.(map[string]interface{})
				if !ok {
					t.Fatalf("json_schema is not a map, got type: %T", jsonSchemaRaw)
				}
				
				schemaData := jsonSchema["schema"]
				
				var schemaBytes []byte
				
				// Handle both json.RawMessage and map[string]interface{} cases
				switch v := schemaData.(type) {
				case json.RawMessage:
					schemaBytes = v
				case map[string]interface{}:
					var marshalErr error
					schemaBytes, marshalErr = json.Marshal(v)
					testhelpers.RequireNoError(t, marshalErr)
				default:
					t.Fatalf("Unexpected schema type: %T", v)
				}

				var schema map[string]interface{}
				unmarshalErr := json.Unmarshal(schemaBytes, &schema)
				testhelpers.RequireNoError(t, unmarshalErr)

				propertiesRaw, exists := schema["properties"]
				if !exists {
					t.Fatalf("properties field not found in schema")
				}
				
				properties, ok := propertiesRaw.(map[string]interface{})
				if !ok {
					t.Fatalf("properties is not a map, got type: %T", propertiesRaw)
				}
				
				_, hasUser := properties["user"]
				_, hasScore := properties["score"]
				testhelpers.AssertTrue(t, hasUser, "schema should contain 'user' property")
				testhelpers.AssertTrue(t, hasScore, "schema should contain 'score' property")
			},
		},
		{
			name: "array field",
			exampleOutput: struct {
				Items []string `json:"items"`
				Count int      `json:"count"`
			}{
				Items: []string{"item1", "item2"},
				Count: 2,
			},
			schemaName: "array-test",
			wantErr:    false,
			validateSchema: func(t *testing.T, result json.RawMessage) {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("validateSchema panicked: %v", r)
					}
				}()
				
				var responseFormat map[string]interface{}
				err := json.Unmarshal(result, &responseFormat)
				testhelpers.RequireNoError(t, err)

				jsonSchemaRaw, exists := responseFormat["json_schema"]
				if !exists {
					t.Fatalf("json_schema field not found in response")
				}
				
				jsonSchema, ok := jsonSchemaRaw.(map[string]interface{})
				if !ok {
					t.Fatalf("json_schema is not a map, got type: %T", jsonSchemaRaw)
				}
				
				schemaData := jsonSchema["schema"]
				
				var schemaBytes []byte
				
				// Handle both json.RawMessage and map[string]interface{} cases
				switch v := schemaData.(type) {
				case json.RawMessage:
					schemaBytes = v
				case map[string]interface{}:
					var marshalErr error
					schemaBytes, marshalErr = json.Marshal(v)
					testhelpers.RequireNoError(t, marshalErr)
				default:
					t.Fatalf("Unexpected schema type: %T", v)
				}

				var schema map[string]interface{}
				unmarshalErr := json.Unmarshal(schemaBytes, &schema)
				testhelpers.RequireNoError(t, unmarshalErr)

				propertiesRaw, exists := schema["properties"]
				if !exists {
					t.Fatalf("properties field not found in schema")
				}
				
				properties, ok := propertiesRaw.(map[string]interface{})
				if !ok {
					t.Fatalf("properties is not a map, got type: %T", propertiesRaw)
				}
				
				_, hasItems := properties["items"]
				_, hasCount := properties["count"]
				testhelpers.AssertTrue(t, hasItems, "schema should contain 'items' property")
				testhelpers.AssertTrue(t, hasCount, "schema should contain 'count' property")

				// Verify items is an array type
				itemsPropRaw, exists := properties["items"]
				if !exists {
					t.Fatalf("items property not found")
				}
				
				itemsProp, ok := itemsPropRaw.(map[string]interface{})
				if !ok {
					t.Fatalf("items property is not a map, got type: %T", itemsPropRaw)
				}
				
				testhelpers.AssertEqual(t, "array", itemsProp["type"])
			},
		},
		{
			name: "special characters in schema name",
			exampleOutput: TestOutputStruct{
				Sentiment:  "positive",
				Confidence: 0.9,
				Reasoning:  "test",
			},
			schemaName: "test-schema with spaces & symbols!",
			wantErr:    false,
			validateSchema: func(t *testing.T, result json.RawMessage) {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("validateSchema panicked: %v", r)
					}
				}()
				
				var responseFormat map[string]interface{}
				err := json.Unmarshal(result, &responseFormat)
				testhelpers.RequireNoError(t, err)

				jsonSchemaRaw, exists := responseFormat["json_schema"]
				if !exists {
					t.Fatalf("json_schema field not found in response")
				}
				
				jsonSchema, ok := jsonSchemaRaw.(map[string]interface{})
				if !ok {
					t.Fatalf("json_schema is not a map, got type: %T", jsonSchemaRaw)
				}
				
				// Should clean the name to be valid
				nameRaw, exists := jsonSchema["name"]
				if !exists {
					t.Fatalf("name field not found in json_schema")
				}
				
				name, ok := nameRaw.(string)
				if !ok {
					t.Fatalf("name is not a string, got type: %T", nameRaw)
				}
				
				testhelpers.AssertNotContains(t, name, " ")
				testhelpers.AssertNotContains(t, name, "&")
				testhelpers.AssertNotContains(t, name, "!")
			},
		},
		{
			name:          "nil input",
			exampleOutput: nil,
			schemaName:    "test",
			wantErr:       true,
		},
		{
			name:          "empty schema name",
			exampleOutput: TestOutputStruct{},
			schemaName:    "",
			wantErr:       false, // Should handle empty name gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := UseOpenRouterJsonFormat(tt.exampleOutput, tt.schemaName)

			if tt.wantErr {
				testhelpers.AssertError(t, err)
				return
			}

			testhelpers.RequireNoError(t, err)
			testhelpers.AssertNotNil(t, result)

			// Verify it's valid JSON
			var temp interface{}
			err = json.Unmarshal(result, &temp)
			testhelpers.RequireNoError(t, err, "Result should be valid JSON")

			if tt.validateSchema != nil {
				tt.validateSchema(t, result)
			}
		})
	}
}

func TestPseudoStructuredResponseCleaner(t *testing.T) {
	tests := []struct {
		name     string
		response string
		expected string
	}{
		{
			name:     "clean JSON",
			response: `{"sentiment": "positive", "confidence": 0.9}`,
			expected: `{"sentiment": "positive", "confidence": 0.9}`,
		},
		{
			name:     "JSON with prefix text",
			response: `Here is the analysis: {"sentiment": "positive", "confidence": 0.9}`,
			expected: `{"sentiment": "positive", "confidence": 0.9}`,
		},
		{
			name:     "JSON with suffix text",
			response: `{"sentiment": "positive", "confidence": 0.9} This is the result.`,
			expected: `{"sentiment": "positive", "confidence": 0.9}`,
		},
		{
			name:     "JSON with both prefix and suffix",
			response: `Analysis result: {"sentiment": "positive", "confidence": 0.9} - End of analysis`,
			expected: `{"sentiment": "positive", "confidence": 0.9}`,
		},
		{
			name:     "nested JSON objects",
			response: `{"user": {"name": "John", "age": 30}, "score": 95.5}`,
			expected: `{"user": {"name": "John", "age": 30}, "score": 95.5}`,
		},
		{
			name:     "JSON array",
			response: `[{"name": "item1"}, {"name": "item2"}]`,
			expected: `[{"name": "item1"}, {"name": "item2"}]`,
		},
		{
			name:     "JSON with newlines and spaces",
			response: `Here's the result:
			{
				"sentiment": "positive",
				"confidence": 0.9
			}
			That's the analysis.`,
			expected: `{
				"sentiment": "positive",
				"confidence": 0.9
			}`,
		},
		{
			name:     "multiple JSON objects - should extract first to last",
			response: `{"first": "object"} some text {"second": "object"}`,
			expected: `{"first": "object"} some text {"second": "object"}`,
		},
		{
			name:     "no JSON found",
			response: `This is just text without JSON`,
			expected: `This is just text without JSON`,
		},
		{
			name:     "malformed JSON braces - missing closing",
			response: `{"sentiment": "positive", "confidence": 0.9`,
			expected: `{"sentiment": "positive", "confidence": 0.9`,
		},
		{
			name:     "malformed JSON braces - missing opening",
			response: `"sentiment": "positive", "confidence": 0.9}`,
			expected: `"sentiment": "positive", "confidence": 0.9}`,
		},
		{
			name:     "empty string",
			response: ``,
			expected: ``,
		},
		{
			name:     "only braces",
			response: `{}`,
			expected: `{}`,
		},
		{
			name:     "braces in wrong order",
			response: `} some text {`,
			expected: `} some text {`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PseudoStructuredResponseCleaner(tt.response)
			testhelpers.AssertEqual(t, tt.expected, result)
		})
	}
}

func TestPseudoStructuredResponseCleanerWithValidJSON(t *testing.T) {
	// Test that the cleaner preserves valid JSON structure
	validJSONTests := []string{
		`{"simple": "object"}`,
		`{"nested": {"object": {"with": "values"}}}`,
		`{"array": ["item1", "item2", "item3"]}`,
		`{"mixed": {"array": [1, 2, 3], "string": "value", "number": 42}}`,
		`[{"array": "of"}, {"objects": "here"}]`,
	}

	for _, jsonStr := range validJSONTests {
		t.Run("valid_json_"+jsonStr[:10], func(t *testing.T) {
			result := PseudoStructuredResponseCleaner(jsonStr)
			testhelpers.AssertEqual(t, jsonStr, result)

			// Verify the result is still valid JSON
			var temp interface{}
			err := json.Unmarshal([]byte(result), &temp)
			testhelpers.AssertNoError(t, err, "Cleaned result should be valid JSON")
		})
	}
}

func TestPseudoStructuredResponseCleanerEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		response string
		expected string
	}{
		{
			name:     "JSON with escaped braces in strings",
			response: `{"message": "This has \\{ escaped \\} braces", "valid": true}`,
			expected: `{"message": "This has \\{ escaped \\} braces", "valid": true}`,
		},
		{
			name:     "JSON with quotes in strings",
			response: `{"quote": "He said \"hello\" to me", "response": "ok"}`,
			expected: `{"quote": "He said \"hello\" to me", "response": "ok"}`,
		},
		{
			name:     "very long JSON",
			response: `{"data": "` + string(make([]byte, 1000)) + `"}`,
			expected: `{"data": "` + string(make([]byte, 1000)) + `"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PseudoStructuredResponseCleaner(tt.response)
			testhelpers.AssertEqual(t, tt.expected, result)
		})
	}
}