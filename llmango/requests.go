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
	res, _, err := RunRaw[I, R](l, g, input)
	return res, err
}
func RunRaw[I, R any](l *LLMangoManager, g *Goal, input *I) (*R, *openrouter.NonStreamingChatResponse, error) {
	// Validate input using the goal's validator
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal input for goal '%s': %w", g.UID, err)
	}

	if g.InputValidator != nil {
		if err := g.InputValidator(inputJSON); err != nil {
			return nil, nil, fmt.Errorf("input validation failed for goal '%s': %w", g.UID, err)
		}
	}

	requestStartTime := float64(time.Now().UnixNano()) / 1e9
	var res R
	validPrompts := make(map[string]*Prompt)
	totalWeight := 0

	for _, promptUID := range g.PromptUIDs {
		if !l.Prompts.Exists(promptUID) {
			continue
		}
		prompt, ok := l.Prompts.Get(promptUID)
		if !ok || prompt == nil {
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
				p, ok := l.Prompts.Get(pUID)
				if ok && p != nil && !p.IsCanary {
					hasBasePrompt = true
					break
				}
			}
		}
		if hasBasePrompt {
			return nil, nil, fmt.Errorf("no valid prompts available for goal %s", g.UID)
		} else {
			return nil, nil, fmt.Errorf("no valid prompts available for goal %s and no base prompt exists or is loaded", g.UID)
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
		return nil, nil, errors.New("failed to select prompt after weighted random selection")
	}

	updatedMessages, err := ParseMessages(input, selectedPrompt.Messages)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update prompt messages with err: %w", err)
	}

	routerRequest := &openrouter.OpenRouterRequest{
		Messages:   updatedMessages,
		Prompt:     nil,
		Model:      &selectedPrompt.Model,
		Parameters: selectedPrompt.Parameters,
	}

	// Check if model supports structured output to determine execution path
	supportsStructuredOutput := openrouter.SupportsStructuredOutput(selectedPrompt.Model)

	if supportsStructuredOutput {
		// Generate response format from output example for structured output
		var outputExample R
		if err := json.Unmarshal(g.OutputExample, &outputExample); err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal output example for goal '%s': %w", g.UID, err)
		}

		responseFormat, err := openrouter.UseOpenRouterJsonFormat(outputExample, g.Title)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create JSON schema format: %w", err)
		}

		routerRequest.Parameters.ResponseFormat = responseFormat
	} else {
		// For models that don't support structured output, use universal prompts
		// Generate schema for validation from output example
		schema, err := openrouter.GenerateSchemaFromJSONExample(g.OutputExample)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate schema for universal path: %w", err)
		}

		// Convert schema to map for universal prompt generation
		schemaMap := make(map[string]interface{})
		schemaBytes, err := json.Marshal(schema)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal schema: %w", err)
		}
		if err := json.Unmarshal(schemaBytes, &schemaMap); err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal schema to map: %w", err)
		}

		// Extract existing system prompt from messages
		existingSystemPrompt := ""
		for _, msg := range updatedMessages {
			if msg.Role == "system" && msg.Content != "" {
				existingSystemPrompt = msg.Content
				break
			}
		}

		// Create universal system prompt
		universalPrompt := openrouter.CreateUniversalCompatibilityPrompt(
			existingSystemPrompt,
			schemaMap,
			g.InputExample,
			g.OutputExample,
		)

		// Update messages with universal system prompt
		updatedMessages = injectUniversalPromptIntoMessages(updatedMessages, universalPrompt)
		routerRequest.Messages = updatedMessages
	}

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
			logEntry, createLogErr := l.createLogObject(g.UID, selectedPrompt.UID, input, routerRequest, openrouterResponse, nil, requestTimeElapsed, true, logErr)
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
		return nil, nil, fmt.Errorf("error generating response from OpenRouter for goal %s: %w", g.UID, logErr)
	}

	if openrouterResponse == nil {
		logErr = errors.New("received nil response from OpenRouter without error")
		if l.Logging != nil && l.Logging.LogResponse != nil {
			logEntry, createLogErr := l.createLogObject(g.UID, selectedPrompt.UID, input, routerRequest, nil, nil, requestTimeElapsed, true, logErr)
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
		return nil, nil, logErr
	}

	if len(openrouterResponse.Choices) == 0 || openrouterResponse.Choices[0].Message.Content == nil {
		logErr = errors.New("llm response had 0 choices or nil content")
		if l.Logging != nil && l.Logging.LogResponse != nil {
			logEntry, createLogErr := l.createLogObject(g.UID, selectedPrompt.UID, input, routerRequest, openrouterResponse, nil, requestTimeElapsed, true, logErr)
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
		return nil, nil, logErr
	}

	content := *openrouterResponse.Choices[0].Message.Content

	// Handle response differently based on whether structured output was used
	var finalContent string
	if !supportsStructuredOutput {
		// For universal compatibility path, clean the JSON response
		cleanedJSON := openrouter.PseudoStructuredResponseCleaner(content)
		if cleanedJSON == "" {
			logErr = fmt.Errorf("failed to extract valid JSON from universal compatibility response: %s", content)
			if l.Logging != nil && l.Logging.LogResponse != nil {
				logEntry, createLogErr := l.createLogObject(g.UID, selectedPrompt.UID, input, routerRequest, openrouterResponse, nil, requestTimeElapsed, true, logErr)
				if createLogErr == nil {
					go func(mangoLog *LLMangoLog) {
						if logErr := l.Logging.LogResponse(mangoLog); logErr != nil {
							log.Printf("Failed to log JSON extraction error response: %v", logErr)
						}
					}(logEntry)
				} else {
					log.Printf("Failed to create log object for JSON extraction error: %v", createLogErr)
				}
			}
			return nil, nil, logErr
		}
		finalContent = cleanedJSON
	} else {
		// For structured output path, use content directly
		finalContent = content
	}

	// Validate output using the goal's validator
	outputJSON := json.RawMessage(finalContent)
	if g.OutputValidator != nil {
		if err := g.OutputValidator(outputJSON); err != nil {
			logErr = fmt.Errorf("output validation failed for goal '%s': %w", g.UID, err)
			if l.Logging != nil && l.Logging.LogResponse != nil {
				logEntry, createLogErr := l.createLogObject(g.UID, selectedPrompt.UID, input, routerRequest, openrouterResponse, nil, requestTimeElapsed, true, logErr)
				if createLogErr == nil {
					go func(mangoLog *LLMangoLog) {
						if logErr := l.Logging.LogResponse(mangoLog); logErr != nil {
							log.Printf("Failed to log validation error response: %v", logErr)
						}
					}(logEntry)
				} else {
					log.Printf("Failed to create log object for validation error: %v", createLogErr)
				}
			}
			return nil, nil, logErr
		}
	}

	if errUnmarshal := json.Unmarshal([]byte(finalContent), &res); errUnmarshal != nil {
		logErr = fmt.Errorf("failed to decode response content into target struct: %w, content: %s", errUnmarshal, finalContent)
		if l.Logging != nil && l.Logging.LogResponse != nil {
			logEntry, createLogErr := l.createLogObject(g.UID, selectedPrompt.UID, input, routerRequest, openrouterResponse, nil, requestTimeElapsed, true, logErr)
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
		return nil, nil, logErr
	}

	if l.Logging != nil && l.Logging.LogResponse != nil {
		logEntry, createLogErr := l.createLogObject(g.UID, selectedPrompt.UID, input, routerRequest, openrouterResponse, &res, requestTimeElapsed, true, nil) // Pass nil for error
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

	return &res, openrouterResponse, nil
}

// injectUniversalPromptIntoMessages merges the universal system prompt with existing messages
// Uses the collision strategy from universal_prompts.go
func injectUniversalPromptIntoMessages(messages []openrouter.Message, universalPrompt string) []openrouter.Message {
	var result []openrouter.Message
	systemPromptInjected := false

	for _, msg := range messages {
		if msg.Role == "system" && !systemPromptInjected {
			// Merge with existing system prompt using collision strategy
			existingContent := ""
			if msg.Content != "" {
				existingContent = msg.Content
			}

			mergedContent := openrouter.MergeSystemPrompts(existingContent, universalPrompt)
			result = append(result, openrouter.Message{
				Role:    "system",
				Content: mergedContent,
			})
			systemPromptInjected = true
		} else {
			result = append(result, msg)
		}
	}

	// If no system message was found, add the universal prompt as the first message
	if !systemPromptInjected {
		systemMsg := openrouter.Message{
			Role:    "system",
			Content: universalPrompt,
		}
		result = append([]openrouter.Message{systemMsg}, result...)
	}

	return result
}
