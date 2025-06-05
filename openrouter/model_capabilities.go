package openrouter

// StructuredOutputModels is a simple set of models that support structured output
// If a model is in this set, it supports structured output. If not, it doesn't.
// This determines which execution path (structured vs universal) to use
var StructuredOutputModels = map[string]struct{}{
	"qwen/qwen3-30b-a3b":                        {},
	"qwen/qwen3-32b":                            {},
	"qwen/qwen3-235b-a22b":                      {},
	"openai/o4-mini-high":                       {},
	"openai/o3":                                 {},
	"openai/o4-mini":                            {},
	"openai/gpt-4.1":                            {},
	"openai/gpt-4.1-mini":                       {},
	"openai/gpt-4.1-nano":                       {},
	"meta-llama/llama-4-maverick:free":          {},
	"meta-llama/llama-4-maverick":               {},
	"meta-llama/llama-4-scout:free":             {},
	"meta-llama/llama-4-scout":                  {},
	"google/gemini-2.5-pro-exp-03-25":           {},
	"mistralai/mistral-small-3.1-24b-instruct":  {},
	"google/gemma-3-4b-it:free":                 {},
	"google/gemma-3-12b-it:free":                {},
	"cohere/command-a":                          {},
	"openai/gpt-4o-mini-search-preview":         {},
	"openai/gpt-4o-search-preview":              {},
	"google/gemma-3-27b-it:free":                {},
	"qwen/qwq-32b":                              {},
	"openai/gpt-4.5-preview":                    {},
	"google/gemini-2.0-flash-lite-001":          {},
	"mistralai/mistral-saba":                    {},
	"openai/o3-mini-high":                       {},
	"google/gemini-2.0-flash-001":               {},
	"openai/o3-mini":                            {},
	"mistralai/mistral-small-24b-instruct-2501": {},
	"deepseek/deepseek-r1-distill-llama-70b":    {},
	"deepseek/deepseek-r1":                      {},
	"mistralai/codestral-2501":                  {},
	"deepseek/deepseek-chat":                    {},
	"openai/o1":                                 {},
	"cohere/command-r7b-12-2024":                {},
	"meta-llama/llama-3.3-70b-instruct":         {},
	"openai/gpt-4o-2024-11-20":                  {},
	"mistralai/mistral-large-2411":              {},
	"mistralai/mistral-large-2407":              {},
	"mistralai/pixtral-large-2411":              {},
	"qwen/qwen-2.5-7b-instruct":                 {},
	"google/gemini-flash-1.5-8b":                {},
	"qwen/qwen-2.5-72b-instruct":                {},
	"mistralai/pixtral-12b":                     {},
	"cohere/command-r-plus-08-2024":             {},
	"cohere/command-r-08-2024":                  {},
	"openai/chatgpt-4o-latest":                  {},
	"openai/gpt-4o-2024-08-06":                  {},
	"meta-llama/llama-3.1-405b-instruct":        {},
	"meta-llama/llama-3.1-70b-instruct":         {},
	"mistralai/mistral-nemo":                    {},
	"openai/gpt-4o-mini":                        {},
	"openai/gpt-4o-mini-2024-07-18":             {},
	"01-ai/yi-large":                            {},
	"openai/gpt-4o":                             {},
	"openai/gpt-4o:extended":                    {},
	"openai/gpt-4o-2024-05-13":                  {},
	"mistralai/mixtral-8x22b-instruct":          {},
	"google/gemini-pro-1.5":                     {},
	"cohere/command-r-plus":                     {},
	"cohere/command-r-plus-04-2024":             {},
	"cohere/command":                            {},
	"cohere/command-r":                          {},
	"cohere/command-r-03-2024":                  {},
	"mistralai/mistral-large":                   {},
	"openai/gpt-3.5-turbo-0613":                 {},
	"openai/gpt-4-turbo-preview":                {},
	"mistralai/mistral-medium":                  {},
	"mistralai/mistral-small":                   {},
	"mistralai/mistral-tiny":                    {},
	"openai/gpt-3.5-turbo-1106":                 {},
	"openai/gpt-4-1106-preview":                 {},
	"openai/gpt-4-32k-0314":                     {},
	"openai/gpt-4-0314":                         {},
}

// SupportsStructuredOutput checks if a model supports structured output
// Simply checks if the model is in the StructuredOutputModels set
func SupportsStructuredOutput(modelID string) bool {
	if modelID == "" {
		return false
	}
	_, exists := StructuredOutputModels[modelID]
	return exists
}

// GetSupportedModels returns a list of all models that support structured output
func GetSupportedModels() []string {
	var supported []string
	for modelID := range StructuredOutputModels {
		supported = append(supported, modelID)
	}
	return supported
}

// AddStructuredOutputModel allows adding new models that support structured output at runtime
// This is useful for testing or adding support for new models
func AddStructuredOutputModel(modelID string) {
	StructuredOutputModels[modelID] = struct{}{}
}

// RemoveStructuredOutputModel removes a model from the structured output support list
// This is useful for testing
func RemoveStructuredOutputModel(modelID string) {
	delete(StructuredOutputModels, modelID)
}

// GetAllSupportedModels returns a copy of all models that support structured output
// This is useful for debugging or API endpoints
func GetAllSupportedModels() []string {
	return GetSupportedModels()
}
