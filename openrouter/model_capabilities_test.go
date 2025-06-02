package openrouter

import (
	"testing"

	"github.com/llmang/llmango/testhelpers"
)

func TestGetModelCapabilities(t *testing.T) {
	tests := []struct {
		name                     string
		modelID                  string
		expectedStructuredOutput bool
		expectedProvider         string
		expectedSystemPrompts    bool
		expectedMinContextLength int
	}{
		{
			name:                     "OpenAI GPT-4",
			modelID:                  "openai/gpt-4",
			expectedStructuredOutput: true,
			expectedProvider:         "openai",
			expectedSystemPrompts:    true,
			expectedMinContextLength: 100000,
		},
		{
			name:                     "OpenAI GPT-3.5 Turbo",
			modelID:                  "openai/gpt-3.5-turbo",
			expectedStructuredOutput: true,
			expectedProvider:         "openai",
			expectedSystemPrompts:    true,
			expectedMinContextLength: 16000,
		},
		{
			name:                     "Anthropic Claude 3 Sonnet",
			modelID:                  "anthropic/claude-3-sonnet",
			expectedStructuredOutput: false,
			expectedProvider:         "anthropic",
			expectedSystemPrompts:    true,
			expectedMinContextLength: 100000,
		},
		{
			name:                     "Anthropic Claude 3.5 Sonnet",
			modelID:                  "anthropic/claude-3.5-sonnet",
			expectedStructuredOutput: false,
			expectedProvider:         "anthropic",
			expectedSystemPrompts:    true,
			expectedMinContextLength: 100000,
		},
		{
			name:                     "Google Gemini Pro",
			modelID:                  "google/gemini-pro",
			expectedStructuredOutput: false,
			expectedProvider:         "google",
			expectedSystemPrompts:    true,
			expectedMinContextLength: 30000,
		},
		{
			name:                     "Meta Llama 3.1",
			modelID:                  "meta-llama/llama-3.1-405b-instruct",
			expectedStructuredOutput: false,
			expectedProvider:         "meta",
			expectedSystemPrompts:    true,
			expectedMinContextLength: 100000,
		},
		{
			name:                     "Mistral 8x7B",
			modelID:                  "mistralai/mixtral-8x7b-instruct",
			expectedStructuredOutput: false,
			expectedProvider:         "mistral",
			expectedSystemPrompts:    true,
			expectedMinContextLength: 30000,
		},
		{
			name:                     "Empty model ID",
			modelID:                  "",
			expectedStructuredOutput: false,
			expectedProvider:         "unknown",
			expectedSystemPrompts:    true,
			expectedMinContextLength: 4000,
		},
		{
			name:                     "Unknown model",
			modelID:                  "unknown/model-123",
			expectedStructuredOutput: false,
			expectedProvider:         "unknown",
			expectedSystemPrompts:    true,
			expectedMinContextLength: 4000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			caps := GetModelCapabilities(tt.modelID)
			
			testhelpers.AssertEqual(t, tt.expectedStructuredOutput, caps.SupportsStructuredOutput, 
				"SupportsStructuredOutput mismatch for model %s", tt.modelID)
			testhelpers.AssertEqual(t, tt.expectedProvider, caps.Provider, 
				"Provider mismatch for model %s", tt.modelID)
			testhelpers.AssertEqual(t, tt.expectedSystemPrompts, caps.SupportsSystemPrompts, 
				"SupportsSystemPrompts mismatch for model %s", tt.modelID)
			testhelpers.AssertTrue(t, caps.MaxContextLength >= tt.expectedMinContextLength, 
				"MaxContextLength too low for model %s: got %d, expected at least %d", 
				tt.modelID, caps.MaxContextLength, tt.expectedMinContextLength)
		})
	}
}

func TestPatternMatching(t *testing.T) {
	tests := []struct {
		name                     string
		modelID                  string
		expectedStructuredOutput bool
		expectedProvider         string
		description              string
	}{
		{
			name:                     "OpenAI pattern - gpt-4 variant",
			modelID:                  "custom/gpt-4-custom",
			expectedStructuredOutput: true,
			expectedProvider:         "openai",
			description:              "Should match GPT-4 pattern",
		},
		{
			name:                     "OpenAI pattern - gpt-3.5 variant",
			modelID:                  "custom/gpt-3.5-custom",
			expectedStructuredOutput: true,
			expectedProvider:         "openai",
			description:              "Should match GPT-3.5 pattern",
		},
		{
			name:                     "Claude pattern - custom claude",
			modelID:                  "custom/claude-custom",
			expectedStructuredOutput: false,
			expectedProvider:         "anthropic",
			description:              "Should match Claude pattern",
		},
		{
			name:                     "Llama pattern - custom llama",
			modelID:                  "custom/llama-custom",
			expectedStructuredOutput: false,
			expectedProvider:         "meta",
			description:              "Should match Llama pattern",
		},
		{
			name:                     "Gemini pattern - custom gemini",
			modelID:                  "custom/gemini-custom",
			expectedStructuredOutput: false,
			expectedProvider:         "google",
			description:              "Should match Gemini pattern",
		},
		{
			name:                     "No pattern match",
			modelID:                  "custom/unknown-model",
			expectedStructuredOutput: false,
			expectedProvider:         "unknown",
			description:              "Should fall back to default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			caps := GetModelCapabilities(tt.modelID)
			
			testhelpers.AssertEqual(t, tt.expectedStructuredOutput, caps.SupportsStructuredOutput, 
				"%s: SupportsStructuredOutput mismatch", tt.description)
			testhelpers.AssertEqual(t, tt.expectedProvider, caps.Provider, 
				"%s: Provider mismatch", tt.description)
			testhelpers.AssertNotEmpty(t, caps.Notes, 
				"%s: Notes should not be empty for pattern matched models", tt.description)
		})
	}
}

func TestSupportsStructuredOutput(t *testing.T) {
	tests := []struct {
		name     string
		modelID  string
		expected bool
	}{
		{"OpenAI GPT-4", "openai/gpt-4", true},
		{"OpenAI GPT-3.5", "openai/gpt-3.5-turbo", true},
		{"Claude 3 Sonnet", "anthropic/claude-3-sonnet", false},
		{"Llama 3.1", "meta-llama/llama-3.1-405b-instruct", false},
		{"Unknown model", "unknown/model", false},
		{"Empty model", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SupportsStructuredOutput(tt.modelID)
			testhelpers.AssertEqual(t, tt.expected, result, 
				"SupportsStructuredOutput result mismatch for %s", tt.modelID)
		})
	}
}

func TestGetSupportedModels(t *testing.T) {
	supported := GetSupportedModels()
	
	testhelpers.AssertNotEmpty(t, supported, "Should have at least one supported model")
	
	// Check that all returned models actually support structured output
	for _, modelID := range supported {
		caps := GetModelCapabilities(modelID)
		testhelpers.AssertTrue(t, caps.SupportsStructuredOutput, 
			"Model %s in supported list but doesn't support structured output", modelID)
	}
	
	// Check that some known models are in the list
	expectedModels := []string{"openai/gpt-4", "openai/gpt-3.5-turbo"}
	for _, expectedModel := range expectedModels {
		found := false
		for _, supportedModel := range supported {
			if supportedModel == expectedModel {
				found = true
				break
			}
		}
		testhelpers.AssertTrue(t, found, "Expected model %s not found in supported list", expectedModel)
	}
}

func TestGetUniversalModels(t *testing.T) {
	universal := GetUniversalModels()
	
	testhelpers.AssertNotEmpty(t, universal, "Should have at least one universal model")
	
	// Check that all returned models actually require universal prompts
	for _, modelID := range universal {
		caps := GetModelCapabilities(modelID)
		testhelpers.AssertFalse(t, caps.SupportsStructuredOutput, 
			"Model %s in universal list but supports structured output", modelID)
	}
	
	// Check that some known models are in the list
	expectedModels := []string{"anthropic/claude-3-sonnet", "meta-llama/llama-3.1-405b-instruct"}
	for _, expectedModel := range expectedModels {
		found := false
		for _, universalModel := range universal {
			if universalModel == expectedModel {
				found = true
				break
			}
		}
		testhelpers.AssertTrue(t, found, "Expected model %s not found in universal list", expectedModel)
	}
}

func TestAddModelCapability(t *testing.T) {
	// Test adding a new model capability
	testModelID := "test/custom-model"
	testCaps := ModelCapabilities{
		SupportsStructuredOutput: true,
		SupportsSystemPrompts:    true,
		MaxContextLength:        50000,
		Provider:                "test",
		Notes:                   "Test model for unit testing",
	}
	
	// Add the capability
	AddModelCapability(testModelID, testCaps)
	
	// Verify it was added correctly
	retrievedCaps := GetModelCapabilities(testModelID)
	testhelpers.AssertEqual(t, testCaps.SupportsStructuredOutput, retrievedCaps.SupportsStructuredOutput, 
		"SupportsStructuredOutput mismatch for added model")
	testhelpers.AssertEqual(t, testCaps.Provider, retrievedCaps.Provider, 
		"Provider mismatch for added model")
	testhelpers.AssertEqual(t, testCaps.MaxContextLength, retrievedCaps.MaxContextLength, 
		"MaxContextLength mismatch for added model")
	testhelpers.AssertEqual(t, testCaps.Notes, retrievedCaps.Notes, 
		"Notes mismatch for added model")
	
	// Clean up - remove the test model
	delete(StructuredOutputModels, testModelID)
}

func TestGetAllModelCapabilities(t *testing.T) {
	allCaps := GetAllModelCapabilities()
	
	testhelpers.AssertTrue(t, len(allCaps) > 0, "Should return at least one model capability")
	
	// Verify it's a copy, not the original map
	originalCount := len(StructuredOutputModels)
	allCaps["test-modification"] = ModelCapabilities{}
	testhelpers.AssertEqual(t, originalCount, len(StructuredOutputModels), 
		"Original map should not be modified when returned map is modified")
	
	// Verify some known models are present
	expectedModels := []string{"openai/gpt-4", "anthropic/claude-3-sonnet", "default"}
	for _, expectedModel := range expectedModels {
		_, exists := allCaps[expectedModel]
		testhelpers.AssertTrue(t, exists, "Expected model %s not found in all capabilities", expectedModel)
	}
}

func TestModelCapabilitiesConsistency(t *testing.T) {
	// Test that the supported and universal model lists don't overlap
	supported := GetSupportedModels()
	universal := GetUniversalModels()
	
	for _, supportedModel := range supported {
		for _, universalModel := range universal {
			testhelpers.AssertNotEqual(t, supportedModel, universalModel, 
				"Model %s appears in both supported and universal lists", supportedModel)
		}
	}
	
	// Test that all models in the whitelist are either supported or universal
	allCaps := GetAllModelCapabilities()
	for modelID, caps := range allCaps {
		if modelID == "default" {
			continue // Skip default entry
		}
		
		if caps.SupportsStructuredOutput {
			found := false
			for _, supportedModel := range supported {
				if supportedModel == modelID {
					found = true
					break
				}
			}
			testhelpers.AssertTrue(t, found, 
				"Model %s supports structured output but not in supported list", modelID)
		} else {
			found := false
			for _, universalModel := range universal {
				if universalModel == modelID {
					found = true
					break
				}
			}
			testhelpers.AssertTrue(t, found, 
				"Model %s requires universal prompts but not in universal list", modelID)
		}
	}
}

func TestDefaultCapabilities(t *testing.T) {
	defaultCaps := StructuredOutputModels["default"]
	
	// Default should be conservative - no structured output support
	testhelpers.AssertFalse(t, defaultCaps.SupportsStructuredOutput, 
		"Default capabilities should not support structured output")
	testhelpers.AssertTrue(t, defaultCaps.SupportsSystemPrompts, 
		"Default capabilities should support system prompts")
	testhelpers.AssertEqual(t, "unknown", defaultCaps.Provider, 
		"Default provider should be 'unknown'")
	testhelpers.AssertTrue(t, defaultCaps.MaxContextLength > 0, 
		"Default max context length should be positive")
}