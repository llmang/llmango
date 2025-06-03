package openrouter

// NOTE on API Specification Implementation:
// This file implements Go structs based on the OpenRouter API specification
// provided (https://openrouter.ai/docs/api-reference).

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type OpenRouter struct {
	ApiKey string
}

func CreateOpenRouter(apiKey string) (*OpenRouter, error) {
	if apiKey != "" {
		return &OpenRouter{ApiKey: apiKey}, nil
	}

	return nil, errors.New("failed to create openrouter as the api key was empty")
}

type OpenRouterParameters struct {
	Temperature       float64 `json:"temperature,omitempty"`
	TopP              float64 `json:"top_p,omitempty"`
	FrequencyPenalty  float64 `json:"frequency_penalty,omitempty"`
	PresencePenalty   float64 `json:"presence_penalty,omitempty"`
	RepetitionPenalty float64 `json:"repetition_penalty,omitempty"`
	TopK              int     `json:"top_k,omitempty"`
}

// Message represents a single message in the chat conversation.
type Message struct {
	Role       string  `json:"role"`    // "user", "assistant", "system", or "tool"
	Content    string  `json:"content"` // Simple string content
	Name       *string `json:"name,omitempty"`
	ToolCallID *string `json:"tool_call_id,omitempty"` // Required if role is "tool"
}

// Tool defines a tool (currently only "function" type is supported).
type Tool struct {
}

// ToolChoiceFunction specifies a function to be called.
type ToolChoiceFunction struct {
	Name string `json:"name"`
}

// OpenRouterRequest represents the request body sent to the OpenRouter API.
type OpenRouterRequest struct {
	// Either Messages or Prompt is required.
	Messages []Message `json:"messages,omitempty"`
	Prompt   *string   `json:"prompt,omitempty"`

	Model *string `json:"model,omitempty"` // Uses user's default if unspecified
	Parameters
}

var JSONObjectResponseFormat = "json_object"
var JSONSchemaStringResponseFormat = "json_schema"

type Parameters struct {
	// Allows to force the model to produce specific output format
	// See models page and note on this docs page for which models support it
	ResponseFormat json.RawMessage `json:"response_format,omitempty"`

	Stop []string `json:"stop,omitempty"` // String(s) to stop generation at
	// Note: Spec says string | string[], using []string for simplicity. API likely handles it.

	Stream    *bool `json:"stream,omitempty"`     // Enable streaming
	MaxTokens *int  `json:"max_tokens,omitempty"` // Range: [1, context_length)

	// Tool calling
	Tools []struct {
		Type     string `json:"type"` // Should be "function"
		Function struct {
			Description *string        `json:"description,omitempty"`
			Name        string         `json:"name"`
			Parameters  map[string]any `json:"parameters"` // JSON Schema object
		} `json:"function"`
	} `json:"tools,omitempty"`

	ToolChoice any `json:"tool_choice,omitempty"` // "none", "auto", or {"type": "function", "function": {"name": "..."}}

	// LLM Parameters (Optional)
	Temperature       *float64        `json:"temperature,omitempty"`        // Range: [0, 2]
	TopP              *float64        `json:"top_p,omitempty"`              // Range: (0, 1]
	TopK              *int            `json:"top_k,omitempty"`              // Range: [1, Infinity) Not available for OpenAI models
	FrequencyPenalty  *float64        `json:"frequency_penalty,omitempty"`  // Range: [-2, 2]
	PresencePenalty   *float64        `json:"presence_penalty,omitempty"`   // Range: [-2, 2]
	RepetitionPenalty *float64        `json:"repetition_penalty,omitempty"` // Range: (0, 2]
	Seed              *int            `json:"seed,omitempty"`               // Integer only
	LogitBias         map[int]float64 `json:"logit_bias,omitempty"`         // { token_id: bias }
	TopLogprobs       *int            `json:"top_logprobs,omitempty"`       // Integer only
	MinP              *float64        `json:"min_p,omitempty"`              // Range: [0, 1]
	TopA              *float64        `json:"top_a,omitempty"`              // Range: [0, 1]

	// OpenRouter-only parameters (Optional)
	Transforms []string `json:"transforms,omitempty"` // Prompt transforms
	Models     []string `json:"models,omitempty"`     // Model routing list
	Route      *string  `json:"route,omitempty"`      // Model routing strategy ("fallback")
	// Provider   *ProviderPreferences `json:"provider,omitempty"` // Inlined below
	ProviderOrder             []string `json:"provider_order,omitempty"`              // List of provider names to try in order
	ProviderAllowFallbacks    *bool    `json:"provider_allow_fallbacks,omitempty"`    // Default: true. Allow backup providers
	ProviderRequireParameters *bool    `json:"provider_require_parameters,omitempty"` // Default: false. Only use providers supporting all request parameters
	ProviderDataCollection    *string  `json:"provider_data_collection,omitempty"`    // Default: "allow". Control data storage ("allow" | "deny")
	ProviderIgnore            []string `json:"provider_ignore,omitempty"`             // List of provider names to skip
	ProviderQuantizations     []string `json:"provider_quantizations,omitempty"`      // List of quantization levels to filter by (e.g., ["int4", "int8"])
	ProviderSort              *string  `json:"provider_sort,omitempty"`               // Sort providers by "price" or "throughput"
}

// --- OpenRouter Response Structs ---

// ResponseUsage contains token usage information.
type ResponseUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// OpenRouterBaseResponse holds fields common to all top-level response objects.
type OpenRouterBaseResponse struct {
	ID                string         `json:"id"`
	Created           int64          `json:"created"`
	Model             string         `json:"model"`
	Object            string         `json:"object"` // "chat.completion" or "chat.completion.chunk"
	SystemFingerprint *string        `json:"system_fingerprint,omitempty"`
	Usage             *ResponseUsage `json:"usage,omitempty"` // Present non-streaming or in final stream chunk
}

// GenerationStatsResponse represents the response from the /api/v1/generation endpoint
type GenerationStatsResponse struct {
	Data GenerationStats `json:"data"`
}

// GenerationStats contains detailed information about a generation
type GenerationStats struct {
	ID                     string  `json:"id"`
	TotalCost              float64 `json:"total_cost"`
	CreatedAt              string  `json:"created_at"`
	Model                  string  `json:"model"`
	Origin                 string  `json:"origin"`
	Usage                  float64 `json:"usage"`
	IsByok                 bool    `json:"is_byok"`
	UpstreamID             string  `json:"upstream_id"`
	CacheDiscount          float64 `json:"cache_discount"`
	AppID                  int     `json:"app_id"`
	Streamed               bool    `json:"streamed"`
	Cancelled              bool    `json:"cancelled"`
	ProviderName           string  `json:"provider_name"`
	Latency                int     `json:"latency"`
	ModerationLatency      int     `json:"moderation_latency"`
	GenerationTime         int     `json:"generation_time"`
	FinishReason           string  `json:"finish_reason"`
	NativeFinishReason     string  `json:"native_finish_reason"`
	TokensPrompt           int     `json:"tokens_prompt"`
	TokensCompletion       int     `json:"tokens_completion"`
	NativeTokensPrompt     int     `json:"native_tokens_prompt"`
	NativeTokensCompletion int     `json:"native_tokens_completion"`
	NativeTokensReasoning  int     `json:"native_tokens_reasoning"`
	NumMediaPrompt         int     `json:"num_media_prompt"`
	NumMediaCompletion     int     `json:"num_media_completion"`
	NumSearchResults       int     `json:"num_search_results"`
}

// BaseChoice holds fields common to all types of choices within a response.
type BaseChoice struct {
	FinishReason       *string   `json:"finish_reason"`        // "stop", "length", "tool_calls", etc. or null
	NativeFinishReason *string   `json:"native_finish_reason"` // Provider-specific reason or null
	Error              *struct { // Optional error details for this specific choice
		Code     int            `json:"code"`
		Message  string         `json:"message"`
		Metadata map[string]any `json:"metadata,omitempty"`
	} `json:"error,omitempty"`
}

// ToolCallFunction represents the function details in a tool call
type ToolCallFunction struct {
	Name      string `json:"name"`      // Function name
	Arguments string `json:"arguments"` // JSON string arguments
}

// ToolCall represents a single tool call in a message
type ToolCall struct {
	ID       string           `json:"id"`   // ID of the tool call
	Type     string           `json:"type"` // Should be "function"
	Function ToolCallFunction `json:"function"`
}

// ResponseMessage represents the message content in a chat response
type ResponseMessage struct {
	Content   *string    `json:"content"` // Message content or null
	Role      string     `json:"role"`    // Usually "assistant"
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// NonStreamingChatChoice represents a choice in a standard chat completion response
type NonStreamingChatChoice struct {
	BaseChoice
	Message ResponseMessage `json:"message"`
}

// StreamingChatChoice represents a choice chunk in a streaming chat response.
type StreamingChatChoice struct {
	BaseChoice                    // Finish reasons usually null until the final chunk
	Delta      StreamingChatDelta `json:"delta"`
}

// StreamingChatDelta represents the delta changes in a streaming chat response.
type StreamingChatDelta struct {
	Content   *string         `json:"content"`              // Content delta (token chunk) or null
	Role      *string         `json:"role,omitempty"`       // Usually present only in the first chunk
	ToolCalls []ToolCallDelta `json:"tool_calls,omitempty"` // Tool calls delta
}

// ToolCallDelta represents a single tool call delta in a streaming response.
type ToolCallDelta struct {
	Index    *int          `json:"index,omitempty"` // Index for incremental tool call updates
	ID       string        `json:"id"`              // ID of the tool call
	Type     string        `json:"type"`            // Should be "function"
	Function FunctionDelta `json:"function"`
}

// FunctionDelta represents the function details delta in a tool call.
type FunctionDelta struct {
	Name      *string `json:"name,omitempty"` // Function name (often only in first chunk)
	Arguments string  `json:"arguments"`      // Argument chunk (JSON string delta)
}

// PromptCompletionChoice represents a choice when the input was a simple prompt string.
type PromptCompletionChoice struct {
	BaseChoice
	Text string `json:"text"` // The generated text
}

// --- Main Response Struct Definitions ---

// NonStreamingChatResponse represents a standard (non-streamed) chat completion response.
type NonStreamingChatResponse struct {
	OpenRouterBaseResponse
	Choices []NonStreamingChatChoice `json:"choices"`
}

// StreamingChatResponse represents a streaming chat completion response chunk.
type StreamingChatResponse struct {
	OpenRouterBaseResponse
	Choices []StreamingChatChoice `json:"choices"` // Note: Final chunk might have empty Choices and only Usage in Base
}

// PromptCompletionResponse represents a response when the input was a simple prompt string.
type PromptCompletionResponse struct {
	OpenRouterBaseResponse
	Choices []PromptCompletionChoice `json:"choices"`
}

// --- Helper Function for NON-Streaming HTTP Request ---

// autoConfigureProviderRequirements automatically sets ProviderRequireParameters
// to true when ResponseFormat is detected (structured output)
func (r *OpenRouterRequest) autoConfigureProviderRequirements() {
	if r.Parameters.ResponseFormat != nil && len(r.Parameters.ResponseFormat) > 2 {
		// Check if it's more than just "{}" - meaningful structured output
		responseStr := string(r.Parameters.ResponseFormat)
		if responseStr != "{}" && responseStr != "" {
			if r.Parameters.ProviderRequireParameters == nil {
				requireParams := true
				r.Parameters.ProviderRequireParameters = &requireParams
				log.Printf("🔧 Auto-detected structured output: setting require_parameters=true")
			}
		}
	}
}

// executeOpenRouterRequest handles sending the request and basic response/error handling
// for non-streaming requests. It returns the response body bytes on success.
func (o *OpenRouter) executeOpenRouterRequest(request *OpenRouterRequest) ([]byte, error) {
	if o.ApiKey == "" {
		return nil, errors.New("API KEY is empty in openrouter instance")
	}
	
	// Auto-configure provider requirements based on request content
	request.autoConfigureProviderRequirements()
	
	// Ensure stream is not accidentally set for this helper
	if request.Stream != nil && *request.Stream {
		return nil, errors.New("executeOpenRouterRequest is for non-streaming requests; use GenerateStreamingChatResponse for streaming")
	}

	// Marshal the request body to JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	// Create the new HTTP request
	req, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.ApiKey)
	// TODO: Add other optional headers like "HTTP-Referer", "X-Title" if needed

	// Send the request with a timeout
	client := &http.Client{Timeout: 5 * time.Minute} // Consider making timeout configurable
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	return body, nil
}

// --- Specific Response Generation Functions ---
// GenerateNonStreamingChatResponse sends a request expected to yield a standard chat response.
// Assumes request.Stream is false or nil, and request.Messages is used.
func (o *OpenRouter) GenerateNonStreamingChatResponse(request *OpenRouterRequest) (*NonStreamingChatResponse, error) {
	// Explicitly set stream to false if nil
	if request.Stream == nil {
		stream := false
		request.Stream = &stream
	} else if *request.Stream {
		return nil, errors.New("GenerateNonStreamingChatResponse called with Stream=true; use GenerateStreamingChatResponse instead")
	}

	resp, err := o.executeOpenRouterRequest(request)
	if err != nil {
		return nil, err // Error already formatted by executeOpenRouterRequest
	}

	// Use the validator function to check for all error conditions
	// and parse the response in one step
	response, err := ValidateNonStreamingResponse(resp, http.StatusOK)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// GeneratePromptCompletionResponse sends a request expected to yield a simple prompt completion.
// Assumes request.Stream is false or nil, and request.Prompt is used.
func (o *OpenRouter) GeneratePromptCompletionResponse(request *OpenRouterRequest) (*PromptCompletionResponse, error) {
	// Explicitly set stream to false if nil
	if request.Stream == nil {
		stream := false
		request.Stream = &stream
	} else if *request.Stream {
		return nil, errors.New("GeneratePromptCompletionResponse called with Stream=true; stream is not supported for prompt completions")
	}
	// Ensure prompt is provided
	if request.Prompt == nil || *request.Prompt == "" {
		return nil, errors.New("GeneratePromptCompletionResponse requires the Prompt field to be set")
	}

	body, err := o.executeOpenRouterRequest(request)
	if err != nil {
		return nil, err
	}

	var response PromptCompletionResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("error parsing prompt completion response: %w\nBody: %s", err, string(body))
	}

	return &response, nil
}

// GenerateStreamingChatResponse sends a request and returns a channel to receive streaming chat chunks.
// Assumes request.Stream is explicitly set to true.
// The caller MUST read from the channel until it is closed.
// Errors encountered during streaming will cause the channel to be closed prematurely.
// Context can be used to cancel the request and clean up resources.
func (o *OpenRouter) GenerateStreamingChatResponse(ctx context.Context, request *OpenRouterRequest) (<-chan *StreamingChatResponse, error) {
	// Ensure stream is explicitly set to true
	if request.Stream == nil || !*request.Stream {
		stream := true
		request.Stream = &stream
		// Alternatively, return an error:
		// return nil, errors.New("GenerateStreamingChatResponse requires the Stream field to be explicitly set to true")
	}

	// Auto-configure provider requirements based on request content
	request.autoConfigureProviderRequirements()

	// Marshal the request body to JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("error marshaling streaming request: %w", err)
	}

	// Create the new HTTP request with context
	req, err := http.NewRequestWithContext(ctx, "POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating streaming request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.ApiKey)
	req.Header.Set("Accept", "text/event-stream") // Important for SSE
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	// Send the request
	// Use a client with a timeout for idle connections
	client := &http.Client{
		Timeout: 0, // Overall timeout is handled by context
		Transport: &http.Transport{
			IdleConnTimeout: 90 * time.Second,
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending streaming request: %w", err)
	}

	// Check for non-200 status codes *before* starting the goroutine
	if resp.StatusCode != http.StatusOK {
		// Read the body to attempt to get an error message
		bodyBytes, readErr := io.ReadAll(resp.Body)
		resp.Body.Close() // Ensure body is closed even on error

		if readErr == nil {
			var errResp ErrorResponse
			if err := json.Unmarshal(bodyBytes, &errResp); err == nil && errResp.Details.Message != "" {
				// Check if this is a known error code
				if stdErr, exists := ErrorCodeToError[errResp.Details.Code]; exists {
					return nil, fmt.Errorf("%w: %s", stdErr, errResp.Details.Message)
				}
				// If code isn't recognized, return the full ErrorResponse
				return nil, &errResp
			}
		}
		// Fallback error if we can't parse the response
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	// Create a buffered channel to send results back (prevents blocking for short periods)
	resultsChan := make(chan *StreamingChatResponse, 10)

	// Start a goroutine to process the streaming response body
	go func() {
		// Ensure resources are cleaned up when the goroutine exits
		defer resp.Body.Close()
		defer close(resultsChan)

		// Setup a timeout for inactivity
		const maxIdleTime = 2 * time.Minute
		idleTimer := time.NewTimer(maxIdleTime)
		defer idleTimer.Stop()

		// Create a context that gets canceled if the parent context is canceled
		// or if we've been idle too long
		streamCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		// Monitor for context cancellation
		go func() {
			<-streamCtx.Done()
			// Force the response body to close, which will cause scanner.Scan() to return false
			resp.Body.Close()
		}()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			// Reset idle timer on each message
			if !idleTimer.Stop() {
				select {
				case <-idleTimer.C:
				default:
				}
			}
			idleTimer.Reset(maxIdleTime)

			line := scanner.Text()
			if line == "" { // Skip empty lines typical in SSE
				continue
			}

			// Check for the SSE "data: " prefix
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			// Trim the prefix
			jsonData := strings.TrimPrefix(line, "data: ")

			// Check for the special [DONE] message
			if jsonData == "[DONE]" {
				break // End of stream signal
			}

			// Unmarshal the JSON data for this chunk
			var chunk StreamingChatResponse
			if err := json.Unmarshal([]byte(jsonData), &chunk); err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing stream chunk JSON: %v\nJSON Data: %s\n", err, jsonData)
				break // Stop processing on error
			}

			// Send the chunk to the channel with context awareness to prevent blocking forever
			select {
			case resultsChan <- &chunk:
				// Successfully sent
			case <-streamCtx.Done():
				return // Context was canceled, exit goroutine
			case <-time.After(30 * time.Second):
				// If we can't send for 30 seconds, assume receiver is not processing and bail
				fmt.Fprintf(os.Stderr, "Warning: Stream consumer not reading responses for 30 seconds, closing connection\n")
				return
			}
		}

		// Check for errors during scanning (e.g., connection closed abruptly)
		if err := scanner.Err(); err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, io.ErrClosedPipe) {
			// Log the scanner error, but only if it's not due to our context cancellation
			fmt.Fprintf(os.Stderr, "Error reading stream: %v\n", err)
		}
	}()

	// Return the channel immediately
	return resultsChan, nil
}

var ErrGenerationIDNotFound = errors.New("generation information for this generation ID was not found, either this id is invalid or you requested the endpoint too soon")

// GetGenerationStats retrieves detailed information about a generation by its ID
// WARNING: You must wait around 400 ms before calling the generation stats endpoint else you will get a 404 error
func (o *OpenRouter) GetGenerationStats(generationID string) (*GenerationStats, error) {
	// time.Sleep(800 * time.Millisecond)
	if o.ApiKey == "" {
		return nil, errors.New("API KEY is empty in openrouter instance")
	}

	if generationID == "" {
		return nil, errors.New("generation ID cannot be empty")
	}

	// Create the HTTP request with the generation ID as a query parameter
	req, err := http.NewRequest("GET", "https://openrouter.ai/api/v1/generation", nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Add the query parameter
	q := req.URL.Query()
	q.Add("id", generationID)
	req.URL.RawQuery = q.Encode()

	// Set headers
	req.Header.Set("Authorization", "Bearer "+o.ApiKey)
	// Send the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Read and parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// Check for non-success status code
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, ErrGenerationIDNotFound
		}
		return nil, fmt.Errorf("API returned non-success status: %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse the response JSON
	var statsResponse GenerationStatsResponse
	if err := json.Unmarshal(body, &statsResponse); err != nil {
		return nil, fmt.Errorf("error parsing generation stats response: %w\nBody: %s", err, string(body))
	}
	return &statsResponse.Data, nil
}
