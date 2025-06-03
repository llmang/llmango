package openrouter

import (
	"strings"
)

// ModelCapabilities defines the capabilities of a specific LLM model
type ModelCapabilities struct {
	SupportsStructuredOutput bool   `json:"supports_structured_output"`
	SupportsSystemPrompts    bool   `json:"supports_system_prompts"`
	MaxContextLength        int    `json:"max_context_length"`
	Provider                string `json:"provider"`
	Notes                   string `json:"notes,omitempty"`
}

// StructuredOutputModels is a whitelist of models and their capabilities
// This determines which execution path (structured vs universal) to use
// Simplified list with key models for testing dual-path execution
var StructuredOutputModels = map[string]ModelCapabilities{
	// === STRUCTURED OUTPUT MODELS (🔧 Structured Path) ===
	// OpenAI Models - Support structured output via response_format
	"openai/gpt-4": {
		SupportsStructuredOutput: true,
		SupportsSystemPrompts:    true,
		MaxContextLength:        128000,
		Provider:                "openai",
		Notes:                   "🔧 Structured Path - Full JSON schema support",
	},
	"openai/gpt-4o": {
		SupportsStructuredOutput: true,
		SupportsSystemPrompts:    true,
		MaxContextLength:        128000,
		Provider:                "openai",
		Notes:                   "🔧 Structured Path - Full JSON schema support",
	},
	"openai/gpt-4o-mini": {
		SupportsStructuredOutput: true,
		SupportsSystemPrompts:    true,
		MaxContextLength:        128000,
		Provider:                "openai",
		Notes:                   "🔧 Structured Path - Full JSON schema support",
	},
	"openai/gpt-3.5-turbo": {
		SupportsStructuredOutput: true,
		SupportsSystemPrompts:    true,
		MaxContextLength:        16385,
		Provider:                "openai",
		Notes:                   "🔧 Structured Path - Full JSON schema support",
	},
	"openai/gpt-4-turbo": {
		SupportsStructuredOutput: true,
		SupportsSystemPrompts:    true,
		MaxContextLength:        128000,
		Provider:                "openai",
		Notes:                   "🔧 Structured Path - Full JSON schema support",
	},

	// === UNIVERSAL PROMPT MODELS (🌍 Universal Path) ===
	// Anthropic Models - Do not support structured output, use universal path
	"anthropic/claude-3-opus": {
		SupportsStructuredOutput: false,
		SupportsSystemPrompts:    true,
		MaxContextLength:        200000,
		Provider:                "anthropic",
		Notes:                   "🌍 Universal Path - Use universal prompts for JSON output",
	},
	"anthropic/claude-3-sonnet": {
		SupportsStructuredOutput: false,
		SupportsSystemPrompts:    true,
		MaxContextLength:        200000,
		Provider:                "anthropic",
		Notes:                   "🌍 Universal Path - Use universal prompts for JSON output",
	},
	"anthropic/claude-3.5-sonnet": {
		SupportsStructuredOutput: false,
		SupportsSystemPrompts:    true,
		MaxContextLength:        200000,
		Provider:                "anthropic",
		Notes:                   "🌍 Universal Path - Use universal prompts for JSON output",
	},

	// Meta Llama Models - Do not support structured output
	"meta-llama/llama-3.1-405b-instruct": {
		SupportsStructuredOutput: false,
		SupportsSystemPrompts:    true,
		MaxContextLength:        131072,
		Provider:                "meta",
		Notes:                   "🌍 Universal Path - Use universal prompts for JSON output",
	},
	"meta-llama/llama-3.1-70b-instruct": {
		SupportsStructuredOutput: false,
		SupportsSystemPrompts:    true,
		MaxContextLength:        131072,
		Provider:                "meta",
		Notes:                   "🌍 Universal Path - Use universal prompts for JSON output",
	},
	"meta-llama/llama-3.3-70b-instruct": {
		SupportsStructuredOutput: false,
		SupportsSystemPrompts:    true,
		MaxContextLength:        32768,
		Provider:                "meta",
		Notes:                   "🌍 Universal Path - Use universal prompts for JSON output",
	},

	// Google Models
	"google/gemini-1.5-flash": {
		SupportsStructuredOutput: false,
		SupportsSystemPrompts:    true,
		MaxContextLength:        1048576,
		Provider:                "google",
		Notes:                   "🌍 Universal Path - Use universal prompts for JSON output",
	},
	"google/gemini-1.5-pro": {
		SupportsStructuredOutput: false,
		SupportsSystemPrompts:    true,
		MaxContextLength:        2097152,
		Provider:                "google",
		Notes:                   "🌍 Universal Path - Use universal prompts for JSON output",
	},

	// Mistral Models
	"mistralai/mistral-large": {
		SupportsStructuredOutput: false,
		SupportsSystemPrompts:    true,
		MaxContextLength:        131072,
		Provider:                "mistral",
		Notes:                   "🌍 Universal Path - Use universal prompts for JSON output",
	},
	"mistralai/mistral-small": {
		SupportsStructuredOutput: false,
		SupportsSystemPrompts:    true,
		MaxContextLength:        131072,
		Provider:                "mistral",
		Notes:                   "🌍 Universal Path - Use universal prompts for JSON output",
	},

	// Cohere Models
	"cohere/command-r": {
		SupportsStructuredOutput: false,
		SupportsSystemPrompts:    true,
		MaxContextLength:        131072,
		Provider:                "cohere",
		Notes:                   "🌍 Universal Path - Use universal prompts for JSON output",
	},
	"cohere/command-r-plus": {
		SupportsStructuredOutput: false,
		SupportsSystemPrompts:    true,
		MaxContextLength:        131072,
		Provider:                "cohere",
		Notes:                   "🌍 Universal Path - Use universal prompts for JSON output",
	},

	// DeepSeek Models
	"deepseek/deepseek-chat": {
		SupportsStructuredOutput: false,
		SupportsSystemPrompts:    true,
		MaxContextLength:        65536,
		Provider:                "deepseek",
		Notes:                   "🌍 Universal Path - Use universal prompts for JSON output",
	},
	"deepseek/deepseek-r1": {
		SupportsStructuredOutput: false,
		SupportsSystemPrompts:    true,
		MaxContextLength:        131072,
		Provider:                "deepseek",
		Notes:                   "🌍 Universal Path - Use universal prompts for JSON output",
	},

	// Qwen Models
	"qwen/qwen-2.5-72b-instruct": {
		SupportsStructuredOutput: false,
		SupportsSystemPrompts:    true,
		MaxContextLength:        131072,
		Provider:                "alibaba",
		Notes:                   "🌍 Universal Path - Use universal prompts for JSON output",
	},

	// Default fallback for unknown models
	"default": {
		SupportsStructuredOutput: false,
		SupportsSystemPrompts:    true,
		MaxContextLength:        4096,
		Provider:                "unknown",
		Notes:                   "🌍 Universal Path - Default fallback",
	},
}

// GetModelCapabilities returns the capabilities for a given model ID
// If the model is not in the whitelist, returns default capabilities (universal path)
func GetModelCapabilities(modelID string) ModelCapabilities {
	if modelID == "" {
		return StructuredOutputModels["default"]
	}

	// Direct lookup first
	if caps, exists := StructuredOutputModels[modelID]; exists {
		return caps
	}

	// Try pattern matching for model families
	modelLower := strings.ToLower(modelID)
	
	// OpenAI pattern matching
	if strings.Contains(modelLower, "gpt-4") || strings.Contains(modelLower, "gpt-3.5") {
		if strings.Contains(modelLower, "openai/") || strings.Contains(modelLower, "gpt") {
			return ModelCapabilities{
				SupportsStructuredOutput: true,
				SupportsSystemPrompts:    true,
				MaxContextLength:        128000,
				Provider:                "openai",
				Notes:                   "🔧 Structured Path - Pattern matched OpenAI model",
			}
		}
	}

	// Anthropic pattern matching
	if strings.Contains(modelLower, "claude") {
		return ModelCapabilities{
			SupportsStructuredOutput: false,
			SupportsSystemPrompts:    true,
			MaxContextLength:        200000,
			Provider:                "anthropic",
			Notes:                   "🌍 Universal Path - Pattern matched Anthropic model",
		}
	}

	// Llama pattern matching
	if strings.Contains(modelLower, "llama") {
		return ModelCapabilities{
			SupportsStructuredOutput: false,
			SupportsSystemPrompts:    true,
			MaxContextLength:        8192,
			Provider:                "meta",
			Notes:                   "🌍 Universal Path - Pattern matched Llama model",
		}
	}

	// Gemini pattern matching
	if strings.Contains(modelLower, "gemini") {
		return ModelCapabilities{
			SupportsStructuredOutput: false,
			SupportsSystemPrompts:    true,
			MaxContextLength:        32768,
			Provider:                "google",
			Notes:                   "🌍 Universal Path - Pattern matched Google model",
		}
	}

	// Default fallback for unknown models
	return StructuredOutputModels["default"]
}

// SupportsStructuredOutput is a convenience function to check if a model supports structured output
func SupportsStructuredOutput(modelID string) bool {
	return GetModelCapabilities(modelID).SupportsStructuredOutput
}

// GetSupportedModels returns a list of all models that support structured output
func GetSupportedModels() []string {
	var supported []string
	for modelID, caps := range StructuredOutputModels {
		if caps.SupportsStructuredOutput && modelID != "default" {
			supported = append(supported, modelID)
		}
	}
	return supported
}

// GetUniversalModels returns a list of all models that require universal prompts
func GetUniversalModels() []string {
	var universal []string
	for modelID, caps := range StructuredOutputModels {
		if !caps.SupportsStructuredOutput && modelID != "default" {
			universal = append(universal, modelID)
		}
	}
	return universal
}

// AddModelCapability allows adding new model capabilities at runtime
// This is useful for testing or adding support for new models
func AddModelCapability(modelID string, capabilities ModelCapabilities) {
	StructuredOutputModels[modelID] = capabilities
}

// GetAllModelCapabilities returns a copy of all model capabilities
// This is useful for debugging or API endpoints
func GetAllModelCapabilities() map[string]ModelCapabilities {
	result := make(map[string]ModelCapabilities)
	for k, v := range StructuredOutputModels {
		result[k] = v
	}
	return result
}