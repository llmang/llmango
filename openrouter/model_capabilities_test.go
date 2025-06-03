package openrouter

import (
	"testing"

	"github.com/llmang/llmango/testhelpers"
)

func TestSupportsStructuredOutput(t *testing.T) {
	tests := []struct {
		name     string
		modelID  string
		expected bool
	}{
		{"OpenAI GPT-4o", "openai/gpt-4o", true},
		{"OpenAI GPT-4o Mini", "openai/gpt-4o-mini", true},
		{"OpenAI GPT-3.5 Turbo", "openai/gpt-3.5-turbo-0613", true},
		{"Qwen Model", "qwen/qwen-2.5-72b-instruct", true},
		{"Meta Llama 3.1", "meta-llama/llama-3.1-405b-instruct", true},
		{"Unknown model", "unknown/model", false},
		{"Empty model", "", false},
		{"Non-existent OpenAI model", "openai/gpt-5", false},
		{"Pattern-like but not exact", "openai/gpt-4o-custom", false},
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
		result := SupportsStructuredOutput(modelID)
		testhelpers.AssertTrue(t, result, 
			"Model %s in supported list but SupportsStructuredOutput returns false", modelID)
	}
	
	// Check that some known models are in the list
	expectedModels := []string{"openai/gpt-4o", "openai/gpt-4o-mini"}
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

func TestAddStructuredOutputModel(t *testing.T) {
	// Test adding a new model
	testModelID := "test/custom-model"
	
	// Ensure it's not supported initially
	testhelpers.AssertFalse(t, SupportsStructuredOutput(testModelID), 
		"Test model should not be supported initially")
	
	// Add the model
	AddStructuredOutputModel(testModelID)
	
	// Verify it's now supported
	testhelpers.AssertTrue(t, SupportsStructuredOutput(testModelID), 
		"Test model should be supported after adding")
	
	// Verify it appears in the supported models list
	supported := GetSupportedModels()
	found := false
	for _, modelID := range supported {
		if modelID == testModelID {
			found = true
			break
		}
	}
	testhelpers.AssertTrue(t, found, "Added model should appear in supported models list")
	
	// Clean up - remove the test model
	RemoveStructuredOutputModel(testModelID)
	
	// Verify it's no longer supported
	testhelpers.AssertFalse(t, SupportsStructuredOutput(testModelID), 
		"Test model should not be supported after removal")
}

func TestRemoveStructuredOutputModel(t *testing.T) {
	// Use a model that we know exists
	testModelID := "openai/gpt-4o"
	
	// Verify it's supported initially
	testhelpers.AssertTrue(t, SupportsStructuredOutput(testModelID), 
		"Test model should be supported initially")
	
	// Remove the model
	RemoveStructuredOutputModel(testModelID)
	
	// Verify it's no longer supported
	testhelpers.AssertFalse(t, SupportsStructuredOutput(testModelID), 
		"Test model should not be supported after removal")
	
	// Add it back for other tests
	AddStructuredOutputModel(testModelID)
	
	// Verify it's supported again
	testhelpers.AssertTrue(t, SupportsStructuredOutput(testModelID), 
		"Test model should be supported after re-adding")
}

func TestGetAllSupportedModels(t *testing.T) {
	allSupported := GetAllSupportedModels()
	regularSupported := GetSupportedModels()
	
	testhelpers.AssertEqual(t, len(regularSupported), len(allSupported), 
		"GetAllSupportedModels should return same count as GetSupportedModels")
	
	// Verify all models in both lists are the same (order might differ)
	for _, model := range allSupported {
		found := false
		for _, regModel := range regularSupported {
			if model == regModel {
				found = true
				break
			}
		}
		testhelpers.AssertTrue(t, found, "Model %s in GetAllSupportedModels but not in GetSupportedModels", model)
	}
}

func TestDirectLookupOnly(t *testing.T) {
	tests := []struct {
		name        string
		modelID     string
		expected    bool
		description string
	}{
		{
			name:        "Unknown model returns false",
			modelID:     "custom/unknown-model",
			expected:    false,
			description: "Should return false for unknown models",
		},
		{
			name:        "Pattern-like name but not in whitelist",
			modelID:     "custom/gpt-4-custom",
			expected:    false,
			description: "Should not match patterns, only exact whitelist entries",
		},
		{
			name:        "Another pattern-like name",
			modelID:     "custom/claude-custom",
			expected:    false,
			description: "Should not match patterns, only exact whitelist entries",
		},
		{
			name:        "Exact match works",
			modelID:     "openai/gpt-4o",
			expected:    true,
			description: "Should match exact entries in whitelist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SupportsStructuredOutput(tt.modelID)
			testhelpers.AssertEqual(t, tt.expected, result, 
				"%s: result mismatch", tt.description)
		})
	}
}

func TestEmptyModelID(t *testing.T) {
	result := SupportsStructuredOutput("")
	testhelpers.AssertFalse(t, result, "Empty model ID should not support structured output")
}

func TestStructuredOutputModelsConsistency(t *testing.T) {
	// Test that all models in the map can be retrieved via SupportsStructuredOutput
	for modelID := range StructuredOutputModels {
		result := SupportsStructuredOutput(modelID)
		testhelpers.AssertTrue(t, result, 
			"Model %s in StructuredOutputModels map but SupportsStructuredOutput returns false", modelID)
	}
	
	// Test that GetSupportedModels returns all models in the map
	supported := GetSupportedModels()
	testhelpers.AssertEqual(t, len(StructuredOutputModels), len(supported), 
		"GetSupportedModels should return same count as StructuredOutputModels map")
}
