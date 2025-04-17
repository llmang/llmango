package llmango

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/llmang/llmango/openrouter"
)

var ErrLoggerNotInitialized = errors.New("there is no error logger initialized for llmango so no logs can be queried, setup an error logger with log getter to fix this")

// LogFilter represents the filtering options for retrieving logs
type LLmangoLogFilter struct {
	MinTimestamp *int    `json:"minTimestamp,omitempty"`
	MaxTimestamp *int    `json:"maxTimestamp,omitempty"`
	GoalUID      *string `json:"goalUID,omitempty"`
	PromptUID    *string `json:"promptUID,omitempty"`
	Limit        *int    `json:"limit"`
	Offset       *int    `json:"offset"`
	IncludeRaw   bool    `json:"includeRaw"`
}

// LLMangoLog represents a single log entry.
// UserID and Metadata must be set by custom loggers, they are not part of the default log message passed into the logger.
type LLMangoLog struct {
	Timestamp      int     `json:"timestamp"`
	GoalUID        string  `json:"goalUID"`
	PromptUID      string  `json:"promptUID"`
	RawRequest     string  `json:"rawInput,omitempty"`
	InputObject    string  `json:"inputObject"`
	RawResponse    string  `json:"rawOutput,omitempty"`
	OutputObject   string  `json:"outputObject"`
	InputTokens    int     `json:"inputTokens"`
	OutputTokens   int     `json:"outputTokens"`
	Cost           float64 `json:"cost"`
	RequestTime    float64 `json:"requestTime"`
	GenerationTime float64 `json:"generationTime"`
	Error          string  `json:"error"`

	UserID   string `json:"userID"`
	Metadata any    `json:"metadata,omitempty"`
}

type Logging struct {
	LogResponse func(*LLMangoLog) error                            //logger
	GetLogs     func(*LLmangoLogFilter) ([]LLMangoLog, int, error) //log reteriver
}

// WARNING: IF YOU DO NOT SLEEP YOU MAY HIT THE GENERATION ENDPOINT TOO FAST RESULTING IN A 404 ERROR.
// This gathers cost and usage from openrouter getGeneration endpoint.
// createLogObject builds a log entry from request/response data
func (mang *LLMangoManager) createLogObject(
	goalUID string,
	promptUID string,
	input any,
	request *openrouter.OpenRouterRequest,
	response *openrouter.NonStreamingChatResponse,
	output any,
	requestTime float64,
	includeRawData bool,
	err error,
) (*LLMangoLog, error) {
	// Convert input and output to JSON
	requestJSONString, _ := json.Marshal(request)
	responseJSONString, _ := json.Marshal(response)
	inputJSONString, _ := json.Marshal(input)
	outputJSONString, _ := json.Marshal(output)

	// Create base log object
	logObject := &LLMangoLog{
		GoalUID:      goalUID,
		PromptUID:    promptUID,
		InputObject:  string(inputJSONString),
		OutputObject: string(outputJSONString),
		RequestTime:  requestTime,
	}

	// Conditionally include raw request/response strings
	if includeRawData {
		logObject.RawRequest = string(requestJSONString)
		logObject.RawResponse = string(responseJSONString)
	}

	// If there's a response, populate fields from it
	if response != nil {
		if mang == nil || mang.OpenRouter == nil {
			return nil, errors.New("OpenRouter client is not initialized")
		}

		logObject.Timestamp = int(response.Created)

		if response.Usage != nil {
			logObject.InputTokens = response.Usage.PromptTokens
			logObject.OutputTokens = response.Usage.CompletionTokens
		}

		// Get actual cost and generation time from OpenRouter API
		if response.ID == "" {
			return nil, errors.New("response ID is empty, cannot retrieve generation stats")
		}

		//WARNING: IF YOU DO NOT SLEEP YOU MAY HIT THE GENERATION ENDPOINT TOO FAST RESULTING IN A 404 ERROR
		time.Sleep(800 * time.Millisecond)
		stats, apiErr := mang.OpenRouter.GetGenerationStats(response.ID)
		if apiErr != nil {
			return nil, fmt.Errorf("failed to get OpenRouter generation stats: %w", apiErr)
		}

		logObject.Cost = stats.TotalCost
		logObject.GenerationTime = float64(stats.GenerationTime)
	}

	// Add error if there is one
	if err != nil {
		logObject.Error = err.Error()
	}

	return logObject, nil
}
