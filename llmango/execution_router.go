package llmango

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/llmang/llmango/openrouter"
)

// ExecuteGoalWithDualPath executes a goal using the appropriate execution path
// based on the model's capabilities (structured output vs universal compatibility)
func (m *LLMangoManager) ExecuteGoalWithDualPath(goalUID string, input json.RawMessage) (json.RawMessage, error) {
	log.Printf("🚀 ExecuteGoalWithDualPath: Starting execution for goal '%s'", goalUID)
	log.Printf("📥 Input type: %T, length: %d bytes", input, len(input))
	log.Printf("📥 Input content: %s", string(input))
	
	goal, exists := m.Goals.Get(goalUID)
	if !exists {
		log.Printf("❌ Goal '%s' not found in manager", goalUID)
		return nil, fmt.Errorf("goal with UID '%s' not found", goalUID)
	}
	
	log.Printf("✅ Found goal '%s': %s", goal.UID, goal.Title)
	log.Printf("🎯 Goal has %d prompt UIDs: %v", len(goal.PromptUIDs), goal.PromptUIDs)

	// Select prompt using existing logic
	selectedPrompt, err := m.selectPromptForGoal(goal)
	if err != nil {
		log.Printf("❌ Failed to select prompt for goal '%s': %v", goalUID, err)
		return nil, fmt.Errorf("failed to select prompt for goal '%s': %w", goalUID, err)
	}
	
	log.Printf("✅ Selected prompt '%s' for model '%s'", selectedPrompt.UID, selectedPrompt.Model)
	log.Printf("📝 Prompt has %d messages", len(selectedPrompt.Messages))

	// Get model capabilities to determine execution path
	capabilities := openrouter.GetModelCapabilities(selectedPrompt.Model)
	
	log.Printf("🔍 Model capabilities for '%s': structured_output=%v",
		selectedPrompt.Model, capabilities.SupportsStructuredOutput)

	// Choose execution path based on model capabilities
	if capabilities.SupportsStructuredOutput {
		log.Printf("🎯 Using STRUCTURED OUTPUT path for goal '%s'", goalUID)
		return m.executeWithStructuredOutput(goal, selectedPrompt, input)
	} else {
		log.Printf("🌐 Using UNIVERSAL COMPATIBILITY path for goal '%s'", goalUID)
		return m.executeWithUniversalCompatibility(goal, selectedPrompt, input)
	}
}

// executeWithStructuredOutput uses the existing structured output path
// Enhanced with better error handling and fallback to universal path
func (m *LLMangoManager) executeWithStructuredOutput(goal *Goal, prompt *Prompt, input json.RawMessage) (json.RawMessage, error) {
	log.Printf("🎯 executeWithStructuredOutput: Starting for goal '%s'", goal.UID)
	
	// Validate input using the goal's validator
	if goal.InputValidator != nil {
		log.Printf("🔍 Validating input using goal validator...")
		if err := goal.InputValidator(input); err != nil {
			log.Printf("❌ Input validation failed: %v", err)
			return nil, fmt.Errorf("input validation failed for goal '%s': %w", goal.UID, err)
		}
		log.Printf("✅ Input validation passed")
	} else {
		log.Printf("⚠️ No input validator found for goal '%s'", goal.UID)
	}

	// Parse messages with input variables
	log.Printf("📝 Parsing messages with input variables...")
	log.Printf("📝 Original messages count: %d", len(prompt.Messages))
	for i, msg := range prompt.Messages {
		log.Printf("📝 Message %d [%s]: %s", i, msg.Role, msg.Content)
	}
	
	updatedMessages, err := ParseMessages(input, prompt.Messages)
	if err != nil {
		log.Printf("❌ Failed to parse messages: %v", err)
		return nil, fmt.Errorf("failed to update prompt messages: %w", err)
	}
	
	log.Printf("✅ Successfully parsed messages, count: %d", len(updatedMessages))
	for i, msg := range updatedMessages {
		log.Printf("✅ Updated message %d [%s]: %s", i, msg.Role, msg.Content)
	}

	// Create router request
	routerRequest := &openrouter.OpenRouterRequest{
		Messages:   updatedMessages,
		Model:      &prompt.Model,
		Parameters: prompt.Parameters,
	}
	
	log.Printf("🚀 STRUCTURED PATH - Sending request to OpenRouter:")
	log.Printf("📤 Model: %s", prompt.Model)
	log.Printf("📤 Messages count: %d", len(updatedMessages))
	for i, msg := range updatedMessages {
		log.Printf("📤 Message %d [%s]: %s", i, msg.Role, msg.Content)
	}

	// Generate JSON schema for structured output
	var outputExample interface{}
	if err := json.Unmarshal(goal.OutputExample, &outputExample); err != nil {
		// If we can't unmarshal the output example, fall back to universal path
		log.Printf("❌ Failed to unmarshal output example for structured path, falling back to universal: %v", err)
		return m.executeWithUniversalCompatibility(goal, prompt, input)
	}
	
	responseFormat, err := openrouter.UseOpenRouterJsonFormat(outputExample, goal.Title)
	if err != nil {
		// If schema generation fails, fall back to universal path
		log.Printf("❌ Failed to generate JSON schema for structured path, falling back to universal: %v", err)
		return m.executeWithUniversalCompatibility(goal, prompt, input)
	}

	routerRequest.Parameters.ResponseFormat = responseFormat
	log.Printf("📤 Response format schema: %+v", responseFormat)

	// Execute request
	log.Printf("🌐 Sending request to OpenRouter...")
	response, err := m.OpenRouter.GenerateNonStreamingChatResponse(routerRequest)
	if err != nil {
		log.Printf("❌ OpenRouter request failed: %v", err)
		return nil, fmt.Errorf("structured output execution failed: %w", err)
	}

	log.Printf("📥 Received response from OpenRouter")
	log.Printf("📥 Response choices count: %d", len(response.Choices))
	if len(response.Choices) == 0 || response.Choices[0].Message.Content == nil {
		log.Printf("❌ Empty response from OpenRouter")
		return nil, fmt.Errorf("received empty response from structured output execution")
	}

	content := *response.Choices[0].Message.Content
	log.Printf("📥 Raw response content: %s", content)
	outputJSON := json.RawMessage(content)

	// Validate output using the goal's validator
	if goal.OutputValidator != nil {
		if err := goal.OutputValidator(outputJSON); err != nil {
			return nil, fmt.Errorf("output validation failed for goal '%s': %w", goal.UID, err)
		}
	}

	return outputJSON, nil
}

// executeWithUniversalCompatibility uses universal prompts for models that don't support structured output
func (m *LLMangoManager) executeWithUniversalCompatibility(goal *Goal, prompt *Prompt, input json.RawMessage) (json.RawMessage, error) {
	log.Printf("Using universal compatibility path for goal '%s'", goal.UID)
	
	// Validate input using the goal's validator
	if goal.InputValidator != nil {
		if err := goal.InputValidator(input); err != nil {
			return nil, fmt.Errorf("input validation failed for goal '%s': %w", goal.UID, err)
		}
	}

	// Generate schema for validation from output example
	schema, err := openrouter.GenerateSchemaFromJSONExample(goal.OutputExample)
	if err != nil {
		return nil, fmt.Errorf("failed to generate schema for universal path: %w", err)
	}

	// Convert schema to map for universal prompt generation
	schemaMap := make(map[string]interface{})
	schemaBytes, err := json.Marshal(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema: %w", err)
	}
	if err := json.Unmarshal(schemaBytes, &schemaMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema to map: %w", err)
	}

	// Extract existing system prompt from messages
	existingSystemPrompt := ""
	for _, msg := range prompt.Messages {
		if msg.Role == "system" && msg.Content != "" {
			existingSystemPrompt = msg.Content
			break
		}
	}

	// Create universal system prompt using existing universal_prompts.go
	universalPrompt := openrouter.CreateUniversalCompatibilityPrompt(
		existingSystemPrompt,
		schemaMap,
		goal.InputExample,
		goal.OutputExample,
	)

	// Create updated messages with universal system prompt
	updatedMessages := m.injectUniversalPrompt(prompt.Messages, universalPrompt)

	// Parse messages with input variables
	finalMessages, err := ParseMessages(input, updatedMessages)
	if err != nil {
		return nil, fmt.Errorf("failed to update prompt messages: %w", err)
	}

	// Create router request (no ResponseFormat for universal path)
	routerRequest := &openrouter.OpenRouterRequest{
		Messages:   finalMessages,
		Model:      &prompt.Model,
		Parameters: prompt.Parameters,
	}
	
	log.Printf("🌐 UNIVERSAL PATH - Sending request to OpenRouter:")
	log.Printf("📤 Model: %s", prompt.Model)
	log.Printf("📤 Final messages count: %d", len(finalMessages))
	for i, msg := range finalMessages {
		log.Printf("📤 Message %d [%s]: %s", i, msg.Role, msg.Content)
	}

	// Execute request
	log.Printf("🌐 Sending request to OpenRouter...")
	response, err := m.OpenRouter.GenerateNonStreamingChatResponse(routerRequest)
	if err != nil {
		log.Printf("❌ OpenRouter request failed: %v", err)
		return nil, fmt.Errorf("universal compatibility execution failed: %w", err)
	}

	log.Printf("📥 Received response from OpenRouter")
	log.Printf("📥 Response choices count: %d", len(response.Choices))
	if len(response.Choices) == 0 || response.Choices[0].Message.Content == nil {
		log.Printf("❌ Empty response from OpenRouter")
		return nil, fmt.Errorf("received empty response from universal compatibility execution")
	}

	content := *response.Choices[0].Message.Content
	log.Printf("📥 Raw response content: %s", content)

	// Extract and clean JSON from response using existing cleaner
	log.Printf("🧹 Cleaning JSON from response...")
	cleanedJSON := openrouter.PseudoStructuredResponseCleaner(content)
	log.Printf("🧹 Cleaned JSON: %s", cleanedJSON)
	
	if cleanedJSON == "" {
		log.Printf("❌ Failed to extract JSON from response")
		return nil, fmt.Errorf("failed to extract valid JSON from response: %s", content)
	}

	// Validate JSON against schema
	outputJSON := json.RawMessage(cleanedJSON)
	log.Printf("🔍 Validating JSON against schema...")
	if err := openrouter.ValidateJSONAgainstSchema(outputJSON, schema); err != nil {
		log.Printf("❌ JSON validation failed: %v", err)
		log.Printf("❌ Invalid JSON: %s", cleanedJSON)
		return nil, fmt.Errorf("response validation failed for universal path: %w", err)
	}
	log.Printf("✅ JSON validation passed")

	// Validate output using the goal's validator
	if goal.OutputValidator != nil {
		if err := goal.OutputValidator(outputJSON); err != nil {
			return nil, fmt.Errorf("output validation failed for goal '%s': %w", goal.UID, err)
		}
	}

	return outputJSON, nil
}

// injectUniversalPrompt merges the universal system prompt with existing messages
// Uses the collision strategy from universal_prompts.go
func (m *LLMangoManager) injectUniversalPrompt(messages []openrouter.Message, universalPrompt string) []openrouter.Message {
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

// selectPromptForGoal selects a prompt for the given goal using existing logic
// This is extracted from the existing Run function in requests.go
func (m *LLMangoManager) selectPromptForGoal(goal *Goal) (*Prompt, error) {
	validPrompts := make(map[string]*Prompt)
	totalWeight := 0

	for _, promptUID := range goal.PromptUIDs {
		if !m.Prompts.Exists(promptUID) {
			log.Printf("WARN: prompt %s not found in manager, skipping.", promptUID)
			continue
		}
		prompt, ok := m.Prompts.Get(promptUID)
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

	if len(validPrompts) == 0 {
		hasBasePrompt := false
		for _, pUID := range goal.PromptUIDs {
			if m.Prompts.Exists(pUID) {
				p, ok := m.Prompts.Get(pUID)
				if ok && p != nil && !p.IsCanary {
					hasBasePrompt = true
					break
				}
			}
		}
		if hasBasePrompt {
			return nil, fmt.Errorf("no valid prompts available for goal %s", goal.UID)
		} else {
			return nil, fmt.Errorf("no valid prompts available for goal %s and no base prompt exists or is loaded", goal.UID)
		}
	}

	// Weighted random selection (simplified for now)
	// In a full implementation, this would use the same logic as requests.go
	for _, prompt := range validPrompts {
		if prompt.IsCanary {
			prompt.TotalRuns++
		}
		return prompt, nil
	}

	return nil, fmt.Errorf("failed to select prompt after weighted random selection")
}