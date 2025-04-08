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
	IncludeRaw   bool    `json:"includeRaw"`
	Limit        int     `json:"limit"`
	Offset       int     `json:"offset"`
}

// LogObject represents a single log entry
type LLMangoLog struct {
	Timestamp      int     `json:"timestamp"`
	GoalUID        string  `json:"goalUID"`
	PromptUID      string  `json:"promptUID"`
	RawInput       string  `json:"rawInput"`
	InputObject    string  `json:"inputObject"`
	RawOutput      string  `json:"rawOutput"`
	OutputObject   string  `json:"outputObject"`
	InputTokens    int     `json:"inputTokens"`
	OutputTokens   int     `json:"outputTokens"`
	Cost           float64 `json:"cost"`
	RequestTime    float64 `json:"requestTime"`
	GenerationTime float64 `json:"generationTime"`
	Error          string  `json:"error"`
}

type Logging struct {
	LogPercentage              int                                                //0-100, we always log canaries though
	LogFullInputOutputMessages bool                                               //this logs the entire input and output nt just the specific vals we got and sent
	LogResponse                func(*LLMangoLog) error                            //logger
	GetLogs                    func(*LLmangoLogFilter) ([]LLMangoLog, int, error) //log reteriver
}

//What we need to handle

//we need to safely handle rate limits I thikn we want to just use exponential backoff as this is now easier given I honeslty don't know how rate limits work in openrouter? Using the single semaphore and pushing delay to it when we hit rate limit could be a good strategy though? When we do a runall we ill still run into issues if we don't use a semaphore? potentially we need "minDelay for the semaphore that gets updated by any user, the minDelay for the runall controls the amoutn of delay and gets updated as we go thorugh"

// createLogObject builds a log entry from request/response data
func createLogObject(
	manager *LLMangoManager,
	goalInfo *GoalInfo,
	promptUID string,
	input interface{},
	output interface{},
	response *openrouter.NonStreamingChatResponse,
	requestTime float64,
	err error,
) (*LLMangoLog, error) {
	// Convert input and output to JSON
	inputJSON, _ := json.Marshal(input)
	outputJSON, _ := json.Marshal(output)

	// Create base log object
	logObject := &LLMangoLog{
		GoalUID:     goalInfo.UID,
		PromptUID:   promptUID,
		RawInput:    string(inputJSON),
		InputObject: string(inputJSON),
		RequestTime: requestTime,
	}

	// If there's a response, populate fields from it
	if response != nil {
		logObject.Timestamp = int(response.Created)
		logObject.RawOutput = string(outputJSON)
		logObject.OutputObject = string(outputJSON)

		if response.Usage != nil {
			logObject.InputTokens = response.Usage.PromptTokens
			logObject.OutputTokens = response.Usage.CompletionTokens
		}

		// Get actual cost and generation time from OpenRouter API
		if response.ID == "" {
			return nil, errors.New("response ID is empty, cannot retrieve generation stats")
		}

		if manager == nil || manager.OpenRouter == nil {
			return nil, errors.New("OpenRouter client is not initialized")
		}
		//WARNING: IF YOU DO NOT SLEEP YOU MAY HIT THE GENERATION ENDPOINT TOO FAST RESULTING IN A 404 ERROR
		time.Sleep(800 * time.Millisecond)
		stats, apiErr := manager.OpenRouter.GetGenerationStats(response.ID)
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
