package openrouter

import (
	"encoding/json"
	"testing"
)

func TestAutoConfigureProviderRequirements(t *testing.T) {
	tests := []struct {
		name           string
		responseFormat json.RawMessage
		existingValue  *bool
		expectedValue  *bool
		description    string
	}{
		{
			name:           "Auto-set when ResponseFormat present and ProviderRequireParameters nil",
			responseFormat: json.RawMessage(`{"type": "json_schema"}`),
			existingValue:  nil,
			expectedValue:  boolPtr(true),
			description:    "Should automatically set to true when structured output detected",
		},
		{
			name:           "No change when ResponseFormat present but ProviderRequireParameters already set",
			responseFormat: json.RawMessage(`{"type": "json_schema"}`),
			existingValue:  boolPtr(false),
			expectedValue:  boolPtr(false),
			description:    "Should preserve existing value when already set",
		},
		{
			name:           "No change when ResponseFormat empty",
			responseFormat: nil,
			existingValue:  nil,
			expectedValue:  nil,
			description:    "Should not set anything when no structured output",
		},
		{
			name:           "No change when ResponseFormat empty JSON",
			responseFormat: json.RawMessage(`{}`),
			existingValue:  nil,
			expectedValue:  nil,
			description:    "Should not set anything for empty JSON response format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &OpenRouterRequest{
				Parameters: Parameters{
					ResponseFormat:            tt.responseFormat,
					ProviderRequireParameters: tt.existingValue,
				},
			}

			// Call the auto-detection function
			request.autoConfigureProviderRequirements()

			// Check the result
			if tt.expectedValue == nil {
				if request.Parameters.ProviderRequireParameters != nil {
					t.Errorf("Expected ProviderRequireParameters to be nil, got %v", 
						*request.Parameters.ProviderRequireParameters)
				}
			} else {
				if request.Parameters.ProviderRequireParameters == nil {
					t.Errorf("Expected ProviderRequireParameters to be %v, got nil", *tt.expectedValue)
				} else if *request.Parameters.ProviderRequireParameters != *tt.expectedValue {
					t.Errorf("Expected ProviderRequireParameters to be %v, got %v", 
						*tt.expectedValue, *request.Parameters.ProviderRequireParameters)
				}
			}
		})
	}
}

func TestStructuredOutputAutoRequiresParameters(t *testing.T) {
	// Test that when we use structured output, require_parameters is automatically set
	exampleJSON := json.RawMessage(`{"name": "John", "age": 30}`)
	
	responseFormat, err := UseOpenRouterJsonFormatFromJSON(exampleJSON, "TestSchema")
	if err != nil {
		t.Fatalf("Failed to create response format: %v", err)
	}

	request := &OpenRouterRequest{
		Messages: []Message{{Role: "user", Content: "Test"}},
		Parameters: Parameters{
			ResponseFormat: responseFormat,
		},
	}

	// Simulate what happens in executeOpenRouterRequest
	request.autoConfigureProviderRequirements()

	// Verify that ProviderRequireParameters was automatically set to true
	if request.Parameters.ProviderRequireParameters == nil {
		t.Error("Expected ProviderRequireParameters to be automatically set, but it was nil")
	} else if !*request.Parameters.ProviderRequireParameters {
		t.Error("Expected ProviderRequireParameters to be true, but it was false")
	}
}

// Helper function to create bool pointers
func boolPtr(b bool) *bool {
	return &b
}