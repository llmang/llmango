package llmango

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/llmang/llmango/openrouter"
)

func Run[I, R any](l *LLMangoManager, g *Goal, input *I) (*R, error) {
	inputoutput, ok := g.InputOutput.(*InputOutput[I, R])
	if !ok || inputoutput == nil {
		return nil, fmt.Errorf("goal '%s' has invalid or missing InputOutput configuration for types %T -> %T", g.UID, *new(I), *new(R))
	}

	requestStartTime := float64(time.Now().UnixNano()) / 1e9
	var res R
	validPrompts := make(map[string]*Prompt)
	totalWeight := 0

	for _, promptUID := range g.PromptUIDs {
		if !l.Prompts.Exists(promptUID) {
			log.Printf("WARN: prompt %s not found in manager, skipping.", promptUID)
			continue
		}
		prompt := l.Prompts.Get(promptUID)
		if prompt == nil {
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
		hasBasePrompt := false
		for _, pUID := range g.PromptUIDs {
			if l.Prompts.Exists(pUID) {
				p := l.Prompts.Get(pUID)
				if p != nil && !p.IsCanary {
					hasBasePrompt = true
					break
				}
			}
		}
		if hasBasePrompt {
			return nil, fmt.Errorf("no valid prompts available for goal %s", g.UID)
		} else {
			return nil, fmt.Errorf("no valid prompts available for goal %s and no base prompt exists or is loaded", g.UID)
		}
	}

	randWeight := rand.Intn(totalWeight)
	currentWeight := 0

	promptUIDs := make([]string, 0, len(validPrompts))
	for uid := range validPrompts {
		promptUIDs = append(promptUIDs, uid)
	}

	for _, uid := range promptUIDs {
		prompt := validPrompts[uid]
		currentWeight += prompt.Weight
		if randWeight < currentWeight {
			selectedPrompt = prompt
			if selectedPrompt.IsCanary {
				selectedPrompt.TotalRuns++
			}
			break
		}
	}

	if selectedPrompt == nil {
		return nil, errors.New("failed to select prompt after weighted random selection")
	}

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

	responseFormat, err := openrouter.UseOpenRouterJsonFormat(inputoutput.OutputExample, g.Title)
	if err != nil {
		return nil, fmt.Errorf("failed to create JSON schema format: %w", err)
	}

	routerRequest.Parameters.ResponseFormat = responseFormat

	openrouterResponse, err := l.OpenRouter.GenerateNonStreamingChatResponse(routerRequest)

	requestTimeElapsed := float64(time.Now().UnixNano())/1e9 - requestStartTime

	if l.RetryRateLimit && errors.Is(err, openrouter.ErrRateLimited) {
		curDelay := BASE_BACKOFF_DELAY
		for i := range MAX_BACKOFF_ATTEMPTS {
			log.Printf("Rate limited. Retrying in %v (Attempt %d/%d)", curDelay, i+1, MAX_BACKOFF_ATTEMPTS)
			time.Sleep(curDelay)
			requestTimeStartRetry := float64(time.Now().UnixNano()) / 1e9
			openrouterResponse, err = l.OpenRouter.GenerateNonStreamingChatResponse(routerRequest)
			requestTimeElapsed += float64(time.Now().UnixNano())/1e9 - requestTimeStartRetry

			if err == nil || !errors.Is(err, openrouter.ErrRateLimited) {
				break
			}
			curDelay *= 2
		}

		if errors.Is(err, openrouter.ErrRateLimited) {
			log.Printf("Max rate limit retries reached for goal %s.", g.UID)
			err = fmt.Errorf("%w: for goal %s", ErrMaxRateLimitRetries, g.UID)
		}
	}

	logErr := err

	if logErr != nil {
		if l.Logging != nil && l.Logging.LogResponse != nil {
			logEntry, createLogErr := l.createLogObject(g.UID, selectedPrompt.UID, input, routerRequest, openrouterResponse, nil, requestTimeElapsed, selectedPrompt.IsCanary, logErr)
			if createLogErr != nil {
				log.Printf("Failed to create log object after API error: %v (Original Error: %v)", createLogErr, logErr)
			} else {
				go func(mangoLog *LLMangoLog) {
					if logErr := l.Logging.LogResponse(mangoLog); logErr != nil {
						log.Printf("Failed to log API error response: %v", logErr)
					}
				}(logEntry)
			}
		}
		return nil, fmt.Errorf("error generating response from OpenRouter for goal %s: %w", g.UID, logErr)
	}

	if openrouterResponse == nil {
		logErr = errors.New("received nil response from OpenRouter without error")
		if l.Logging != nil && l.Logging.LogResponse != nil {
			logEntry, createLogErr := l.createLogObject(g.UID, selectedPrompt.UID, input, routerRequest, nil, nil, requestTimeElapsed, selectedPrompt.IsCanary, logErr)
			if createLogErr == nil {
				go func(mangoLog *LLMangoLog) {
					if logErr := l.Logging.LogResponse(mangoLog); logErr != nil {
						log.Printf("Failed to log nil response: %v", logErr)
					}
				}(logEntry)
			} else {
				log.Printf("Failed to create log object for nil response: %v", createLogErr)
			}
		}
		return nil, logErr
	}

	if len(openrouterResponse.Choices) == 0 || openrouterResponse.Choices[0].Message.Content == nil {
		logErr = errors.New("llm response had 0 choices or nil content")
		if l.Logging != nil && l.Logging.LogResponse != nil {
			logEntry, createLogErr := l.createLogObject(g.UID, selectedPrompt.UID, input, routerRequest, openrouterResponse, nil, requestTimeElapsed, selectedPrompt.IsCanary, logErr)
			if createLogErr == nil {
				go func(mangoLog *LLMangoLog) {
					if logErr := l.Logging.LogResponse(mangoLog); logErr != nil {
						log.Printf("Failed to log empty choices/content response: %v", logErr)
					}
				}(logEntry)
			} else {
				log.Printf("Failed to create log object for empty choices/content: %v", createLogErr)
			}
		}
		return nil, logErr
	}

	content := *openrouterResponse.Choices[0].Message.Content
	if errUnmarshal := json.Unmarshal([]byte(content), &res); errUnmarshal != nil {
		logErr = fmt.Errorf("failed to decode response content into target struct: %w, content: %s", errUnmarshal, content)
		if l.Logging != nil && l.Logging.LogResponse != nil {
			logEntry, createLogErr := l.createLogObject(g.UID, selectedPrompt.UID, input, routerRequest, openrouterResponse, nil, requestTimeElapsed, selectedPrompt.IsCanary, logErr)
			if createLogErr == nil {
				go func(mangoLog *LLMangoLog) {
					if logErr := l.Logging.LogResponse(mangoLog); logErr != nil {
						log.Printf("Failed to log decoding error response: %v", logErr)
					}
				}(logEntry)
			} else {
				log.Printf("Failed to create log object for decoding error: %v", createLogErr)
			}
		}
		return nil, logErr
	}

	if l.Logging != nil && l.Logging.LogResponse != nil {
		logEntry, createLogErr := l.createLogObject(g.UID, selectedPrompt.UID, input, routerRequest, openrouterResponse, &res, requestTimeElapsed, selectedPrompt.IsCanary, nil) // Pass nil for error
		if createLogErr != nil {
			log.Printf("Failed to create log object for successful response: %v", createLogErr)
		} else {
			go func(mangoLog *LLMangoLog) {
				if logErr := l.Logging.LogResponse(mangoLog); logErr != nil {
					log.Printf("Failed to log successful response: %v", logErr)
				}
			}(logEntry)
		}
	}

	return &res, nil
}
