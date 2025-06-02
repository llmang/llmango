package openrouter

import (
	"encoding/json"
	"testing"

	"github.com/llmang/llmango/testhelpers"
)

func TestGenerateUniversalSystemPrompt(t *testing.T) {
	tests := []struct {
		name              string
		schema            map[string]interface{}
		inputExample      json.RawMessage
		outputExample     json.RawMessage
		expectedContains  []string
		expectedNotContains []string
	}{
		{
			name: "sentiment analysis schema",
			schema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"sentiment": map[string]interface{}{
						"type": "string",
						"enum": []string{"positive", "negative", "neutral"},
					},
					"confidence": map[string]interface{}{
						"type":    "number",
						"minimum": 0,
						"maximum": 1,
					},
					"reasoning": map[string]interface{}{
						"type": "string",
					},
				},
				"required": []string{"sentiment", "confidence", "reasoning"},
			},
			inputExample:  json.RawMessage(`{"text": "I love this product!"}`),
			outputExample: json.RawMessage(`{"sentiment": "positive", "confidence": 0.9, "reasoning": "Contains positive language"}`),
			expectedContains: []string{
				"valid JSON",
				"sentiment",
				"confidence", 
				"reasoning",
				"positive",
				"negative",
				"neutral",
				"required",
				"JSON only",
				"I love this product!",
				"Contains positive language",
				"CRITICAL INSTRUCTIONS",
				"no explanations",
				"no markdown",
			},
			expectedNotContains: []string{
				"function",
				"tool",
				"API",
			},
		},
		{
			name: "entity extraction schema",
			schema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"entities": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"text": map[string]interface{}{"type": "string"},
								"type": map[string]interface{}{"type": "string"},
								"start": map[string]interface{}{"type": "integer"},
								"end": map[string]interface{}{"type": "integer"},
							},
							"required": []string{"text", "type", "start", "end"},
						},
					},
				},
				"required": []string{"entities"},
			},
			inputExample:  json.RawMessage(`{"text": "John works at Apple", "entity_types": ["person", "organization"]}`),
			outputExample: json.RawMessage(`{"entities": [{"text": "John", "type": "person", "start": 0, "end": 4}]}`),
			expectedContains: []string{
				"entities",
				"array",
				"text",
				"type",
				"start", 
				"end",
				"JSON only",
				"John works at Apple",
				"person",
				"organization",
				"integer",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("GenerateUniversalSystemPrompt panicked: %v", r)
				}
			}()

			result := GenerateUniversalSystemPrompt(tt.schema, tt.inputExample, tt.outputExample)

			// Basic validation
			testhelpers.AssertNotEmpty(t, result, "System prompt should not be empty")

			// Check for required content
			for _, expectedText := range tt.expectedContains {
				testhelpers.AssertContains(t, result, expectedText, "System prompt should contain: %s", expectedText)
			}

			// Check for content that should not be present
			for _, unexpectedText := range tt.expectedNotContains {
				testhelpers.AssertNotContains(t, result, unexpectedText, "System prompt should not contain: %s", unexpectedText)
			}

			// Verify examples are included
			testhelpers.AssertContains(t, result, string(tt.inputExample), "System prompt should contain input example")
			testhelpers.AssertContains(t, result, string(tt.outputExample), "System prompt should contain output example")

			// Verify structure
			testhelpers.AssertContains(t, result, "SCHEMA REQUIREMENTS:", "Should have schema requirements section")
			testhelpers.AssertContains(t, result, "INPUT EXAMPLE:", "Should have input example section")
			testhelpers.AssertContains(t, result, "EXPECTED OUTPUT EXAMPLE:", "Should have output example section")
			testhelpers.AssertContains(t, result, "CRITICAL INSTRUCTIONS:", "Should have critical instructions section")
		})
	}
}

func TestMergeSystemPrompts(t *testing.T) {
	tests := []struct {
		name                string
		existingSystemPrompt string
		universalPrompt     string
		expectedContains    []string
		expectedOrder       []string // To verify the order of content
	}{
		{
			name:                "merge with existing system prompt",
			existingSystemPrompt: "You are a helpful AI assistant. Always be polite and professional.",
			universalPrompt:     "You must respond with valid JSON only. Follow the schema exactly.",
			expectedContains: []string{
				"You are a helpful AI assistant",
				"Always be polite and professional",
				"You must respond with valid JSON only",
				"Follow the schema exactly",
			},
			expectedOrder: []string{
				"You are a helpful AI assistant", // Original should come first
				"You must respond with valid JSON only", // Universal should be appended
			},
		},
		{
			name:                "empty existing system prompt",
			existingSystemPrompt: "",
			universalPrompt:     "You must respond with valid JSON only.",
			expectedContains: []string{
				"You must respond with valid JSON only",
			},
		},
		{
			name:                "existing prompt with newlines",
			existingSystemPrompt: "You are an expert analyst.\nProvide detailed insights.\n\nBe thorough in your analysis.",
			universalPrompt:     "CRITICAL: Respond with JSON only.",
			expectedContains: []string{
				"You are an expert analyst",
				"Provide detailed insights",
				"Be thorough in your analysis",
				"CRITICAL: Respond with JSON only",
			},
		},
		{
			name:                "very long existing prompt",
			existingSystemPrompt: "You are a specialized AI assistant with expertise in multiple domains. " +
				"Your role is to provide accurate, helpful, and contextually appropriate responses. " +
				"Always maintain a professional tone and ensure your answers are well-structured.",
			universalPrompt: "OVERRIDE: You must now respond with valid JSON that matches the provided schema.",
			expectedContains: []string{
				"specialized AI assistant",
				"multiple domains",
				"professional tone",
				"OVERRIDE: You must now respond with valid JSON",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("MergeSystemPrompts panicked: %v", r)
				}
			}()

			result := MergeSystemPrompts(tt.existingSystemPrompt, tt.universalPrompt)

			// Basic validation
			testhelpers.AssertNotEmpty(t, result, "Merged prompt should not be empty")

			// Check for required content
			for _, expectedText := range tt.expectedContains {
				testhelpers.AssertContains(t, result, expectedText, "Merged prompt should contain: %s", expectedText)
			}

			// Check order if specified
			if len(tt.expectedOrder) > 1 {
				firstPos := -1
				secondPos := -1
				for i, text := range tt.expectedOrder {
					pos := len(result) // Default to end if not found
					if idx := findStringIndex(result, text); idx != -1 {
						pos = idx
					}
					
					if i == 0 {
						firstPos = pos
					} else if i == 1 {
						secondPos = pos
					}
				}
				
				if firstPos >= 0 && secondPos >= 0 && firstPos >= secondPos {
					t.Errorf("Expected '%s' to come before '%s' in merged prompt", tt.expectedOrder[0], tt.expectedOrder[1])
				}
			}
		})
	}
}

func TestCreateUniversalCompatibilityPrompt(t *testing.T) {
	tests := []struct {
		name                string
		existingSystemPrompt string
		schema              map[string]interface{}
		inputExample        json.RawMessage
		outputExample       json.RawMessage
		expectedContains    []string
	}{
		{
			name:                "complete integration test",
			existingSystemPrompt: "You are an expert data analyst.",
			schema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"category": map[string]interface{}{
						"type": "string",
						"enum": []string{"urgent", "normal", "low"},
					},
				},
				"required": []string{"category"},
			},
			inputExample:  json.RawMessage(`{"ticket": "Server is down!"}`),
			outputExample: json.RawMessage(`{"category": "urgent"}`),
			expectedContains: []string{
				"You are an expert data analyst", // Original prompt preserved
				"category",                       // Schema content
				"urgent",
				"normal", 
				"low",
				"Server is down!",                // Input example
				"JSON only",                      // Universal instructions
				"CRITICAL INSTRUCTIONS",
			},
		},
		{
			name:                "no existing prompt",
			existingSystemPrompt: "",
			schema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"result": map[string]interface{}{"type": "boolean"},
				},
			},
			inputExample:  json.RawMessage(`{"question": "Is this valid?"}`),
			outputExample: json.RawMessage(`{"result": true}`),
			expectedContains: []string{
				"result",
				"boolean",
				"Is this valid?",
				"JSON only",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("CreateUniversalCompatibilityPrompt panicked: %v", r)
				}
			}()

			result := CreateUniversalCompatibilityPrompt(tt.existingSystemPrompt, tt.schema, tt.inputExample, tt.outputExample)

			// Basic validation
			testhelpers.AssertNotEmpty(t, result, "Final prompt should not be empty")

			// Check for required content
			for _, expectedText := range tt.expectedContains {
				testhelpers.AssertContains(t, result, expectedText, "Final prompt should contain: %s", expectedText)
			}

			// Verify examples are included
			testhelpers.AssertContains(t, result, string(tt.inputExample), "Final prompt should contain input example")
			testhelpers.AssertContains(t, result, string(tt.outputExample), "Final prompt should contain output example")
		})
	}
}

func TestFormatSchemaForPrompt(t *testing.T) {
	tests := []struct {
		name             string
		schema           map[string]interface{}
		expectedContains []string
		expectedFormat   []string // Expected formatting patterns
	}{
		{
			name: "simple object schema",
			schema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type": "string",
					},
					"age": map[string]interface{}{
						"type":    "integer",
						"minimum": 0,
					},
				},
				"required": []string{"name"},
			},
			expectedContains: []string{
				"name",
				"string",
				"REQUIRED",
				"age",
				"integer",
				"minimum: 0",
			},
			expectedFormat: []string{
				"- name (REQUIRED):",
				"- age:",
			},
		},
		{
			name: "array schema",
			schema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"items": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
				},
			},
			expectedContains: []string{
				"items",
				"array",
				"string",
			},
		},
		{
			name: "enum schema",
			schema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"status": map[string]interface{}{
						"type": "string",
						"enum": []string{"active", "inactive", "pending"},
					},
				},
			},
			expectedContains: []string{
				"status",
				"active",
				"inactive", 
				"pending",
				"Must be one of:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("FormatSchemaForPrompt panicked: %v", r)
				}
			}()

			result := FormatSchemaForPrompt(tt.schema)

			// Basic validation
			testhelpers.AssertNotEmpty(t, result, "Formatted schema should not be empty")

			// Check for required content
			for _, expectedText := range tt.expectedContains {
				testhelpers.AssertContains(t, result, expectedText, "Formatted schema should contain: %s", expectedText)
			}

			// Check formatting patterns
			for _, expectedPattern := range tt.expectedFormat {
				testhelpers.AssertContains(t, result, expectedPattern, "Formatted schema should contain pattern: %s", expectedPattern)
			}
		})
	}
}

// Helper function to find string index (case-insensitive)
func findStringIndex(haystack, needle string) int {
	// Simple implementation for testing
	for i := 0; i <= len(haystack)-len(needle); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}