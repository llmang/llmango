package openrouter

import (
	"encoding/json"
	"errors"
	"fmt"
)

// ErrorResponse matches the OpenRouter API error structure
type ErrorResponse struct {
	Details struct {
		Code     int                    `json:"code"`
		Message  string                 `json:"message"`
		Metadata map[string]interface{} `json:"metadata,omitempty"`
	} `json:"error"`
}

// Error implements the error interface for ErrorResponse
func (e *ErrorResponse) Error() string {
	msg := fmt.Sprintf("OpenRouter error (code %d): %s", e.Details.Code, e.Details.Message)
	if len(e.Details.Metadata) > 0 {
		meta, _ := json.Marshal(e.Details.Metadata)
		msg += fmt.Sprintf(" | Metadata: %s", string(meta))
	}
	return msg
}

// ModerationErrorMetadata defines the structure for moderation-related errors
type ModerationErrorMetadata struct {
	Reasons      []string `json:"reasons"`
	FlaggedInput string   `json:"flagged_input"`
	ProviderName string   `json:"provider_name"`
	ModelSlug    string   `json:"model_slug"`
}

// Error implements the error interface for ModerationErrorMetadata
func (m *ModerationErrorMetadata) Error() string {
	return fmt.Sprintf("OpenRouter moderation error (provider: %s, model: %s): input '%s' was flagged for: %v",
		m.ProviderName, m.ModelSlug, m.FlaggedInput, m.Reasons)
}

// ProviderErrorMetadata defines the structure for provider-related errors
type ProviderErrorMetadata struct {
	ProviderName string      `json:"provider_name"`
	Raw          interface{} `json:"raw"`
}

// Error implements the error interface for ProviderErrorMetadata
func (p *ProviderErrorMetadata) Error() string {
	rawStr := ""
	if p.Raw != nil {
		rawBytes, _ := json.Marshal(p.Raw)
		rawStr = string(rawBytes)
	}
	return fmt.Sprintf("OpenRouter provider error (provider: %s): %s",
		p.ProviderName, rawStr)
}

// Standard error codes as defined in OpenRouter documentation
var (
	ErrBadRequest          = errors.New("bad request (invalid or missing params, CORS)")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInsufficientCredits = errors.New("insufficient credits")
	ErrModerationFlag      = errors.New("input flagged by moderation")
	ErrTimeout             = errors.New("request timed out")
	ErrRateLimited         = errors.New("rate limited")
	ErrModelDown           = errors.New("model unavailable")
	ErrNoProviders         = errors.New("no available providers meet requirements")
	ErrNoResponse          = errors.New("no message response received from API")
)

// ErrorCodeToError maps OpenRouter error codes to standardized errors
var ErrorCodeToError = map[int]error{
	400: ErrBadRequest,
	401: ErrInvalidCredentials,
	402: ErrInsufficientCredits,
	403: ErrModerationFlag,
	408: ErrTimeout,
	429: ErrRateLimited,
	502: ErrModelDown,
	503: ErrNoProviders,
}

// IsModerationError checks if an error is a moderation error and returns parsed metadata
func IsModerationError(err error) (*ModerationErrorMetadata, bool) {
	var orErr *ErrorResponse
	if errors.As(err, &orErr) && orErr.Details.Code == 403 {
		meta := &ModerationErrorMetadata{}
		if b, err := json.Marshal(orErr.Details.Metadata); err == nil {
			json.Unmarshal(b, meta)
			return meta, true
		}
	}
	return nil, false
}

// IsProviderError checks if an error is a provider error and returns parsed metadata
func IsProviderError(err error) (*ProviderErrorMetadata, bool) {
	var orErr *ErrorResponse
	if errors.As(err, &orErr) && orErr.Details.Code == 502 {
		meta := &ProviderErrorMetadata{}
		if b, err := json.Marshal(orErr.Details.Metadata); err == nil {
			json.Unmarshal(b, meta)
			return meta, true
		}
	}
	return nil, false
}

// ExtractChoiceError extracts error information from a choice if present
func ExtractChoiceError(choice *BaseChoice) *ErrorResponse {
	if choice == nil || choice.Error == nil {
		return nil
	}

	return &ErrorResponse{
		Details: struct {
			Code     int                    `json:"code"`
			Message  string                 `json:"message"`
			Metadata map[string]interface{} `json:"metadata,omitempty"`
		}{
			Code:     choice.Error.Code,
			Message:  choice.Error.Message,
			Metadata: choice.Error.Metadata,
		},
	}
}

// HasChoiceErrors checks if any choices in a response contain errors
func HasChoiceErrors(choices []*BaseChoice) bool {
	for _, choice := range choices {
		if choice != nil && choice.Error != nil {
			return true
		}
	}
	return false
}

// GetChoiceErrors extracts all errors from choices
func GetChoiceErrors(choices []*BaseChoice) []*ErrorResponse {
	var errors []*ErrorResponse
	for _, choice := range choices {
		if err := ExtractChoiceError(choice); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

// IsNoChoicesError checks if the response has no choices (which indicates an error)
func IsNoChoicesError(choicesCount int) bool {
	return choicesCount == 0
}

// ValidateNonStreamingResponse checks a non-streaming response for various error conditions
// and returns standardized errors when problems are found
func ValidateNonStreamingResponse(respBody []byte, statusCode int) (*NonStreamingChatResponse, error) {
	// First, check for HTTP error status
	if statusCode != 200 {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Details.Message != "" {
			stdErr, exists := ErrorCodeToError[errResp.Details.Code]
			if !exists {
				stdErr = fmt.Errorf("unhandled error code %d: %s",
					errResp.Details.Code, errResp.Details.Message)
			}
			return nil, fmt.Errorf("%w: %s", stdErr, errResp.Details.Message)
		}
		return nil, fmt.Errorf("API error (status %d): %s", statusCode, string(respBody))
	}

	// Parse the response body
	var response NonStreamingChatResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("error parsing non-streaming chat response: %w\nBody: %s", err, string(respBody))
	}

	// Check for error response instead of accessing nonexistent Error field
	var errResp ErrorResponse
	if json.Unmarshal(respBody, &errResp) == nil && errResp.Details.Message != "" {
		stdErr, exists := ErrorCodeToError[errResp.Details.Code]
		if !exists {
			stdErr = fmt.Errorf("unhandled error code %d: %s",
				errResp.Details.Code, errResp.Details.Message)
		}
		return nil, fmt.Errorf("%w: %s", stdErr, errResp.Details.Message)
	}

	// Check for no choices
	if len(response.Choices) == 0 {
		return nil, ErrNoResponse
	}

	// Check for errors in choices
	for _, choice := range response.Choices {
		if choice.Error != nil {
			// Use standard error codes if available
			if stdErr, exists := ErrorCodeToError[choice.Error.Code]; exists {
				return &response, fmt.Errorf("%w: %s", stdErr, choice.Error.Message)
			}
			// Fall back to ErrorResponse if no standard error exists
			return &response, &ErrorResponse{
				Details: struct {
					Code     int                    `json:"code"`
					Message  string                 `json:"message"`
					Metadata map[string]interface{} `json:"metadata,omitempty"`
				}{
					Code:     choice.Error.Code,
					Message:  choice.Error.Message,
					Metadata: choice.Error.Metadata,
				},
			}
		}
	}

	// Check for error finish reasons in first choice
	if response.Choices[0].FinishReason != nil {
		reason := *response.Choices[0].FinishReason
		if reason == "content_filter" {
			return &response, ErrModerationFlag
		} else if reason == "length" {
			// Not necessarily an error, but we'll include it in the response
			return &response, nil
		}
	}

	return &response, nil
}
