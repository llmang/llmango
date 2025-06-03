package llmangoagents

import "testing"

func TestIsToolCallingSupported(t *testing.T) {
	// Test cases for models that should support tool calling
	supportedModels := []string{
		"anthropic/claude-opus-4",
		"anthropic/claude-sonnet-4",
		"openai/gpt-4o",
		"openai/gpt-4-turbo",
		"google/gemini-2.5-flash-preview",
		"mistralai/mistral-large-2411",
	}

	for _, model := range supportedModels {
		if !IsToolCallingSupported(model) {
			t.Errorf("Expected model %s to support tool calling, but it doesn't", model)
		}
	}

	// Test cases for models that should NOT support tool calling (non-existent models)
	unsupportedModels := []string{
		"fake/model-id",
		"nonexistent/model",
		"",
		"random-string",
	}

	for _, model := range unsupportedModels {
		if IsToolCallingSupported(model) {
			t.Errorf("Expected model %s to NOT support tool calling, but it does", model)
		}
	}
}

func TestToolCallingSupportedModelsMapStructure(t *testing.T) {
	// Verify the map is not nil
	if toolCallingSupportedModels == nil {
		t.Error("toolCallingSupportedModels map should not be nil")
	}

	// Verify the map has entries
	if len(toolCallingSupportedModels) == 0 {
		t.Error("toolCallingSupportedModels map should have entries")
	}

	// Verify a few known models exist in the map
	knownModels := []string{
		"anthropic/claude-opus-4",
		"openai/gpt-4o",
		"google/gemini-2.5-flash-preview",
	}

	for _, model := range knownModels {
		if _, exists := toolCallingSupportedModels[model]; !exists {
			t.Errorf("Expected model %s to exist in toolCallingSupportedModels map", model)
		}
	}
}

// Benchmark the lookup performance
func BenchmarkIsToolCallingSupported(b *testing.B) {
	modelID := "anthropic/claude-opus-4"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsToolCallingSupported(modelID)
	}
}