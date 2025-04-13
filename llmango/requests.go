package llmango

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"time"

	"github.com/llmang/llmango/openrouter"
)

func Run[I, R any](l *LLMangoManager, g *Goal[I, R], input *I) (*R, error) {
	// Record start time for request timing
	requestStartTime := float64(time.Now().UnixNano()) / 1e9
	var res R
	validPrompts := make(map[string]*Prompt)
	totalWeight := 0

	for _, promptUID := range g.PromptUIDs {
		prompt, exists := l.Prompts[promptUID]
		if !exists {
			continue
		}

		if prompt.Weight > 0 {
			if prompt.IsCanary {
				if prompt.TotalRuns < prompt.MaxRuns {
					validPrompts[promptUID] = prompt
					totalWeight += prompt.Weight
				}
			} else {
				validPrompts[promptUID] = prompt
				totalWeight += prompt.Weight
			}
		}
	}

	var selectedPrompt *Prompt
	if len(validPrompts) == 0 {
		return nil, errors.New("there are no valid prompts for this goal, canaries may all have ran out and no base prompt is available")
	}

	randWeight := rand.Intn(totalWeight)
	currentWeight := 0

	for _, prompt := range validPrompts {
		currentWeight += prompt.Weight
		if randWeight < currentWeight {
			selectedPrompt = prompt
			if selectedPrompt.IsCanary {
				selectedPrompt.TotalRuns++ // Directly increment the reference
			}
			break
		}
	}

	if selectedPrompt == nil {
		return nil, errors.New("failed to select prompt after looping over prompts")
	}

	//here we have to use a helper func to replace {{val}} with struct field vals
	//we want to reflect the input vals into a map of string to val then loop over them and regex the prompt mesages for {{string}} where not /{{}} valid everything ofcourse
	updatedMessages, err := ParseMessages(input, selectedPrompt.Messages)
	if err != nil {
		return nil, fmt.Errorf("failed to update prompt messages with err: %w", err)
	}

	routerRequest := &openrouter.OpenRouterRequest{
		Messages:   updatedMessages,
		Prompt:     nil,
		Model:      &selectedPrompt.Model,
		Parameters: selectedPrompt.Parameters,
	}

	//we need to add the json struct schme of the output and turn on the paramters for strucuttred json output
	// ResponseFormatType *string `json:"response_format_type,omitempty

	// For structured schema generation using the existing function from structured_responses.go
	schemaDef, err := GenerateSchemaForType(g.ExampleOutput)
	if err != nil {
		return nil, err
	}

	// Convert the Definition to JSON for the schema
	schemaBytes, err := json.Marshal(schemaDef)
	if err != nil {
		return nil, err
	}

	// Use the schema in the response format
	schemaObject := struct {
		Name   string          `json:"name"`
		Schema json.RawMessage `json:"schema"`
		Strict bool            `json:"strict"`
	}{
		Name:   regexp.MustCompile(`[^a-zA-Z0-9]+`).ReplaceAllString(g.Title, "_"),
		Strict: true,
		Schema: schemaBytes,
	}

	responseFormat := struct {
		Type   string      `json:"type"`
		Schema interface{} `json:"json_schema"`
	}{
		Type:   "json_schema",
		Schema: schemaObject,
	}

	// Add it to the request object
	bytes, err := json.Marshal(responseFormat)
	if err != nil {
		return nil, err
	}

	routerRequest.Parameters.ResponseFormat = bytes

	openrouterResponse, err := l.OpenRouter.GenerateNonStreamingChatResponse(routerRequest)

	// Calculate request time elapsed so far
	requestTimeElapsed := float64(time.Now().UnixNano())/1e9 - requestStartTime

	//EXponEnTiALY BACKOFF RETRIES
	if errors.Is(err, openrouter.ErrRateLimited) {
		curDelay := BASE_BACKOFF_DELAY
		for range MAX_BACKOFF_ATTEMPTS {
			time.Sleep(curDelay)
			openrouterResponse, err = l.OpenRouter.GenerateNonStreamingChatResponse(routerRequest)
			if err == nil || !errors.Is(err, openrouter.ErrRateLimited) {
				break
			}
			curDelay = curDelay * 2
		}
		if err != nil && errors.Is(err, openrouter.ErrRateLimited) {
			err = ErrMaxRateLimitRetries
		}
	}

	if openrouterResponse.Choices != nil && openrouterResponse.Choices[0].Message.Content != nil {
		if err := json.Unmarshal([]byte(*openrouterResponse.Choices[0].Message.Content), &res); err != nil {
			return nil, fmt.Errorf("failed to decode response content: %w", err)
		}
	}

	if openrouterResponse.Choices == nil {
		err = errors.New("llm response had 0 choices in object, error occured")
	}

	// Log in a separate goroutine if logging is enabled
	if l.Logging != nil && l.Logging.LogResponse != nil {
		// Create log object
		logEntry, logErr := createLogObject(l, &g.GoalInfo, selectedPrompt.UID, input, &res, openrouterResponse, requestTimeElapsed, err)
		if logErr != nil {
			fmt.Printf("Failed to create log object: %v", logErr)
		} else {
			// Log asynchronously
			go func(log *LLMangoLog) {
				if logErr := l.Logging.LogResponse(log); logErr != nil {
					fmt.Printf("Failed to log response: %v", logErr)
				}
			}(logEntry)
		}
	}

	// Return early if there was an error generating the response
	if err != nil {
		return nil, err
	}

	return &res, nil
}
