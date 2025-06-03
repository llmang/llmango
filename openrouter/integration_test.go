package openrouter

import (
	"encoding/json"
	"testing"
)

// TestJSONResponseFormatIntegration tests the complete integration of our new JSON-based
// response format generation with real-world examples from the dual-path execution system
func TestJSONResponseFormatIntegration(t *testing.T) {
	// Test case 1: Sentiment analysis example (from our example app)
	sentimentExample := json.RawMessage(`{"sentiment": "positive", "confidence": 0.95, "reasoning": "Contains positive language"}`)
	
	responseFormat, err := UseOpenRouterJsonFormatFromJSON(sentimentExample, "sentiment_analysis")
	if err != nil {
		t.Fatalf("Failed to generate response format for sentiment analysis: %v", err)
	}
	
	// Validate the response format
	if err := ValidateJSONResponseFormat(responseFormat); err != nil {
		t.Fatalf("Invalid response format generated: %v", err)
	}
	
	// Parse and verify structure
	var format map[string]interface{}
	if err := json.Unmarshal(responseFormat, &format); err != nil {
		t.Fatalf("Failed to parse generated response format: %v", err)
	}
	
	jsonSchema := format["json_schema"].(map[string]interface{})
	schemaName := jsonSchema["name"].(string)
	
	if schemaName != "sentiment_analysis" {
		t.Errorf("Expected schema name 'sentiment_analysis', got '%s'", schemaName)
	}
	
	// Test case 2: Text summary example (from our example app)
	summaryExample := json.RawMessage(`{"summary": "Brief summary", "key_points": ["Point 1", "Point 2"], "word_count": 150}`)
	
	responseFormat2, err := UseOpenRouterJsonFormatFromJSON(summaryExample, "text_summary")
	if err != nil {
		t.Fatalf("Failed to generate response format for text summary: %v", err)
	}
	
	if err := ValidateJSONResponseFormat(responseFormat2); err != nil {
		t.Fatalf("Invalid response format generated for summary: %v", err)
	}
	
	// Test case 3: Complex nested example
	complexExample := json.RawMessage(`{
		"result": {
			"status": "success",
			"data": {
				"items": [
					{"id": 1, "name": "item1", "active": true},
					{"id": 2, "name": "item2", "active": false}
				],
				"metadata": {
					"total": 2,
					"timestamp": "2023-01-01T00:00:00Z"
				}
			}
		},
		"performance": {
			"duration_ms": 150.5,
			"memory_usage": 1024
		}
	}`)
	
	responseFormat3, err := UseOpenRouterJsonFormatFromJSON(complexExample, "complex_response")
	if err != nil {
		t.Fatalf("Failed to generate response format for complex example: %v", err)
	}
	
	if err := ValidateJSONResponseFormat(responseFormat3); err != nil {
		t.Fatalf("Invalid response format generated for complex example: %v", err)
	}
	
	t.Logf("✅ All integration tests passed!")
	t.Logf("📊 Generated %d different response formats successfully", 3)
}

// TestComparisonWithOriginalFunction compares our new JSON-based function
// with the original struct-based function to ensure compatibility
func TestComparisonWithOriginalFunction(t *testing.T) {
	// Define a struct that matches our JSON example
	type SentimentResult struct {
		Sentiment  string  `json:"sentiment"`
		Confidence float64 `json:"confidence"`
		Reasoning  string  `json:"reasoning"`
	}
	
	// Create an instance
	structExample := SentimentResult{
		Sentiment:  "positive",
		Confidence: 0.95,
		Reasoning:  "Contains positive language",
	}
	
	// Generate response format using original function
	originalFormat, err := UseOpenRouterJsonFormat(structExample, "sentiment_analysis")
	if err != nil {
		t.Fatalf("Original function failed: %v", err)
	}
	
	// Generate response format using our new function
	jsonExample := json.RawMessage(`{"sentiment": "positive", "confidence": 0.95, "reasoning": "Contains positive language"}`)
	newFormat, err := UseOpenRouterJsonFormatFromJSON(jsonExample, "sentiment_analysis")
	if err != nil {
		t.Fatalf("New function failed: %v", err)
	}
	
	// Both should be valid
	if err := ValidateJSONResponseFormat(originalFormat); err != nil {
		t.Fatalf("Original format invalid: %v", err)
	}
	
	if err := ValidateJSONResponseFormat(newFormat); err != nil {
		t.Fatalf("New format invalid: %v", err)
	}
	
	// Parse both formats
	var originalParsed, newParsed map[string]interface{}
	json.Unmarshal(originalFormat, &originalParsed)
	json.Unmarshal(newFormat, &newParsed)
	
	// Compare structure (both should have same basic structure)
	if originalParsed["type"] != newParsed["type"] {
		t.Errorf("Type mismatch: original=%v, new=%v", originalParsed["type"], newParsed["type"])
	}
	
	originalSchema := originalParsed["json_schema"].(map[string]interface{})
	newSchema := newParsed["json_schema"].(map[string]interface{})
	
	if originalSchema["name"] != newSchema["name"] {
		t.Errorf("Schema name mismatch: original=%v, new=%v", originalSchema["name"], newSchema["name"])
	}
	
	if originalSchema["strict"] != newSchema["strict"] {
		t.Errorf("Strict mode mismatch: original=%v, new=%v", originalSchema["strict"], newSchema["strict"])
	}
	
	t.Logf("✅ Both functions generate compatible OpenRouter response formats!")
}