package llmangoagents

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/llmang/llmango/openrouter"
)

// Execute runs the agent with the given input
func (agentCtx *AgentExecutionContext) Execute(input string, agent *Agent) (string, error) {
	fmt.Printf("🤖 Agent '%s' starting execution with input: %s\n", agent.UID, input)

	// Track call count
	agentCtx.CallCount++
	if agentCtx.CallCount > agentCtx.MaxCalls {
		return "", fmt.Errorf("agent '%s' exceeded max calls (%d)", agent.UID, agentCtx.MaxCalls)
	}

	currentInput := input

	// Run preprocessors first
	for _, preprocessorName := range agent.PreProcessors {
		// TODO: Look up and run actual preprocessor function
		// For now, just pass through
		processedInput := runPreprocessor(preprocessorName, agent, currentInput, agentCtx)
		if processedInput != "" {
			currentInput = processedInput
		}
	}

	// Check for action phrases in preprocessor results
	if action := parseActionPhrases(currentInput); action != "" {
		switch action {
		case "@@ABORT":
			return "", fmt.Errorf("workflow aborted by agent")
		case "@@RETURN":
			return currentInput, nil
		}
	}

	// Build contextual system message
	systemMessage := buildContextualSystemMessage(agent, agentCtx)

	// Create conversation messages starting with system message
	messages := []openrouter.Message{
		{Role: "system", Content: systemMessage},
	}

	// Add original user input
	messages = append(messages, openrouter.Message{
		Role:    "user",
		Content: agentCtx.ParentStepContext.ParentWorkflowContext.OriginalUserInput,
	})

	// Add step context messages (completed steps as tool calls) - TODO: implement later
	// For now, we'll focus on fixing the step-level conversation context

	// Add step-level conversation history (shared among all step agents)
	if agentCtx.ParentStepContext != nil {
		messages = append(messages, agentCtx.ParentStepContext.ConversationHistory...)
	}

	// Add current input if it's different from original user input
	if currentInput != agentCtx.ParentStepContext.ParentWorkflowContext.OriginalUserInput {
		messages = append(messages, openrouter.Message{
			Role:    "user",
			Content: currentInput,
		})
	}

	// Prepare tools for the request using OpenAI format
	var tools []struct {
		Type        string         `json:"type"`
		Name        string         `json:"name"`
		Description string         `json:"description"`
		Parameters  map[string]any `json:"parameters"`
	}

	for _, toolName := range agent.Tools {
		tool, err := agentCtx.ParentStepContext.ParentWorkflowContext.SystemManager.GetTool(toolName)
		if err != nil {
			// Skip tools that can't be found, but log it
			continue
		}

		// Parse the input schema JSON into map[string]any
		var parameters map[string]any
		if err := json.Unmarshal([]byte(tool.InputSchema), &parameters); err != nil {
			// Skip tools with invalid schema
			continue
		}

		// Convert our tool to OpenAI format
		openAITool := struct {
			Type        string         `json:"type"`
			Name        string         `json:"name"`
			Description string         `json:"description"`
			Parameters  map[string]any `json:"parameters"`
		}{
			Type:        "function",
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  parameters,
		}
		tools = append(tools, openAITool)
	}

	// Make LLM request
	req := &openrouter.OpenRouterRequest{
		Model:    &agent.Model,
		Messages: messages,
	}

	fmt.Printf("🔧 Agent '%s' has %d tools available\n", agent.UID, len(tools))

	// Add tools if any are available (convert to OpenRouter format)
	if len(tools) > 0 {
		var openRouterTools []struct {
			Type     string `json:"type"`
			Function struct {
				Description *string        `json:"description,omitempty"`
				Name        string         `json:"name"`
				Parameters  map[string]any `json:"parameters"`
			} `json:"function"`
		}

		for _, tool := range tools {
			fmt.Printf("🛠️  Adding tool: %s\n", tool.Name)
			openRouterTool := struct {
				Type     string `json:"type"`
				Function struct {
					Description *string        `json:"description,omitempty"`
					Name        string         `json:"name"`
					Parameters  map[string]any `json:"parameters"`
				} `json:"function"`
			}{
				Type: "function",
				Function: struct {
					Description *string        `json:"description,omitempty"`
					Name        string         `json:"name"`
					Parameters  map[string]any `json:"parameters"`
				}{
					Description: &tool.Description,
					Name:        tool.Name,
					Parameters:  tool.Parameters,
				},
			}
			openRouterTools = append(openRouterTools, openRouterTool)
		}
		req.Tools = openRouterTools
	}

	fmt.Printf("🌐 Making LLM request for agent '%s'...\n", agent.UID)
	response, err := agentCtx.ParentStepContext.ParentWorkflowContext.SystemManager.Openrouter.GenerateNonStreamingChatResponse(req)
	if err != nil {
		fmt.Printf("❌ LLM request failed for agent '%s': %v\n", agent.UID, err)
		return "", err
	}

	fmt.Printf("✅ LLM response received for agent '%s'\n", agent.UID)

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from LLM")
	}

	// Update step-level conversation history with agent name for clarity
	responseMsg := openrouter.Message{
		Role:    response.Choices[0].Message.Role,
		Content: fmt.Sprintf("[%s]: %s", agent.Name, *response.Choices[0].Message.Content),
	}
	newMessages := []openrouter.Message{
		{Role: "user", Content: currentInput},
		responseMsg,
	}
	
	// Save to step-level conversation history (shared among all step agents)
	if agentCtx.ParentStepContext != nil {
		agentCtx.ParentStepContext.ConversationHistory = append(agentCtx.ParentStepContext.ConversationHistory, newMessages...)
	}

	// Handle tool calls if present
	if len(response.Choices[0].Message.ToolCalls) > 0 {
		fmt.Printf("🔧 Agent '%s' received %d tool calls\n", agent.UID, len(response.Choices[0].Message.ToolCalls))

		// Debug: Log the full tool call structure
		for i, toolCall := range response.Choices[0].Message.ToolCalls {
			fmt.Printf("📋 Tool Call %d:\n", i)
			fmt.Printf("  ID: %s\n", toolCall.ID)
			fmt.Printf("  Type: %s\n", toolCall.Type)
			fmt.Printf("  Function Name: %s\n", toolCall.Function.Name)
			fmt.Printf("  Function Arguments (raw): %q\n", toolCall.Function.Arguments)
			fmt.Printf("  Function Arguments (pretty):\n%s\n", toolCall.Function.Arguments)
		}

		// Store the tool calls for later use in conversation
		agentCtx.LastToolCalls = response.Choices[0].Message.ToolCalls

		return agentCtx.processToolCalls(response.Choices[0].Message.ToolCalls, agent)
	}

	finalResponse := *response.Choices[0].Message.Content
	fmt.Printf("💬 Agent '%s' final response: %s\n", agent.UID, finalResponse)
	return finalResponse, nil
}

// Helper functions
func runPreprocessor(preprocessorName string, agent *Agent, input string, agentCtx *AgentExecutionContext) string {
	// TODO: Implement actual preprocessor lookup and execution
	return ""
}

func parseActionPhrases(input string) string {
	// Simple action phrase detection
	if strings.Contains(input, "@@ABORT") {
		return "@@ABORT"
	}
	if strings.Contains(input, "@@RETURN") {
		return "@@RETURN"
	}
	return ""
}

// processToolCalls handles tool calls from LLM responses
func (agentCtx *AgentExecutionContext) processToolCalls(toolCalls []openrouter.ToolCall, agent *Agent) (string, error) {
	// Process each tool call and add tool responses to conversation
	// Note: The assistant message with tool calls was already added in Execute()
	for _, toolCall := range toolCalls {
		fmt.Printf("🔧 Processing tool call: %s\n", toolCall.Function.Name)
		fmt.Printf("🔧 Tool arguments: %q\n", toolCall.Function.Arguments)

		// Look up the tool
		tool, err := agentCtx.ParentStepContext.ParentWorkflowContext.SystemManager.GetTool(toolCall.Function.Name)
		if err != nil {
			fmt.Printf("❌ Tool lookup failed: %v\n", err)
			return "", fmt.Errorf("tool '%s' not found: %v", toolCall.Function.Name, err)
		}

		fmt.Printf("✅ Found tool: %s (UID: %s)\n", tool.Name, tool.Uid)

		var toolResult string

		// Check if this is a useAgentTool call (special handling)
		if toolCall.Function.Name == "useAgentTool" {
			// Execute the tool first
			result, err := tool.Function(agentCtx.ParentStepContext.ParentWorkflowContext.GlobalKeyBank, json.RawMessage(toolCall.Function.Arguments))
			if err != nil {
				fmt.Printf("❌ Tool execution failed: %v\n", err)
				return "", fmt.Errorf("tool '%s' execution failed: %v", toolCall.Function.Name, err)
			}

			// Handle the useAgentTool call and get the subAgent response
			subAgentResult, err := agentCtx.handleUseAgentToolCall(result, agent)
			if err != nil {
				return "", err
			}
			toolResult = subAgentResult
		} else if toolCall.Function.Name == "handoffTool" {
			// Execute the handoff tool
			result, err := tool.Function(agentCtx.ParentStepContext.ParentWorkflowContext.GlobalKeyBank, json.RawMessage(toolCall.Function.Arguments))
			if err != nil {
				fmt.Printf("❌ Tool execution failed: %v\n", err)
				return "", fmt.Errorf("tool '%s' execution failed: %v", toolCall.Function.Name, err)
			}
			
			// For handoff tools, we return the handoff key directly without continuing conversation
			// The step execution will handle the actual handoff
			fmt.Printf("🔄 Handoff tool executed, returning result: %s\n", string(result))
			return string(result), nil
		} else {
			// Execute regular tool
			result, err := tool.Function(agentCtx.ParentStepContext.ParentWorkflowContext.GlobalKeyBank, json.RawMessage(toolCall.Function.Arguments))
			if err != nil {
				fmt.Printf("❌ Tool execution failed: %v\n", err)
				return "", fmt.Errorf("tool '%s' execution failed: %v", toolCall.Function.Name, err)
			}
			toolResult = string(result)
		}

		fmt.Printf("✅ Tool execution result: %q\n", toolResult)

		// Record the tool call
		record := ToolCallRecord{
			ToolUID: tool.Uid,
			Input:   json.RawMessage(toolCall.Function.Arguments),
			Output:  json.RawMessage(toolResult),
			Error:   err,
		}
		agentCtx.ToolCallHistory = append(agentCtx.ToolCallHistory, record)

		// Add tool response message to step-level conversation history
		toolMsg := openrouter.Message{
			Role:       "tool",
			Content:    toolResult,
			ToolCallID: &toolCall.ID,
		}
		if agentCtx.ParentStepContext != nil {
			agentCtx.ParentStepContext.ConversationHistory = append(agentCtx.ParentStepContext.ConversationHistory, toolMsg)
		}
	}

	// Now continue the conversation - make another LLM request with the updated context
	fmt.Printf("🔄 Continuing conversation after tool calls for agent '%s'\n", agent.UID)
	return agentCtx.continueAfterToolCalls(agent)
}

// continueAfterToolCalls makes another LLM request with the updated conversation context
func (agentCtx *AgentExecutionContext) continueAfterToolCalls(agent *Agent) (string, error) {
	// Build contextual system message
	systemMessage := buildContextualSystemMessage(agent, agentCtx)

	// We need to construct the messages manually to include tool_calls
	// Start with system message
	messages := []map[string]interface{}{
		{
			"role":    "system",
			"content": systemMessage,
		},
	}

	// Add original user input
	messages = append(messages, map[string]interface{}{
		"role":    "user",
		"content": agentCtx.ParentStepContext.ParentWorkflowContext.OriginalUserInput,
	})

	// Add step context messages (completed steps as tool calls)
	if agentCtx.ParentStepContext != nil && agentCtx.ParentStepContext.ParentWorkflowContext != nil {
		currentStepIndex := agentCtx.ParentStepContext.ParentWorkflowContext.CurrentStepIndex
		stepContextMsgs := buildStepContextMessages(agentCtx.ParentStepContext.ParentWorkflowContext, currentStepIndex)
		messages = append(messages, stepContextMsgs...)
	}

	// Use step-level conversation history instead of agent-level
	stepConversationHistory := agentCtx.ParentStepContext.ConversationHistory
	
	// Add conversation history, but we need to find where to inject tool_calls
	// The tool_calls should be in the assistant message that preceded the tool responses
	var assistantMsgWithToolCallsIndex = -1

	// Find the assistant message that should have tool_calls (the one before tool messages)
	for i := len(stepConversationHistory) - 1; i >= 0; i-- {
		msg := stepConversationHistory[i]
		if msg.Role == "tool" {
			// Found a tool message, the assistant message before it should have tool_calls
			for j := i - 1; j >= 0; j-- {
				if stepConversationHistory[j].Role == "assistant" {
					assistantMsgWithToolCallsIndex = j
					break
				}
			}
			break
		}
	}

	// Add conversation history with proper tool_calls injection
	for i, msg := range stepConversationHistory {
		if i == assistantMsgWithToolCallsIndex && len(agentCtx.LastToolCalls) > 0 {
			// This assistant message should have tool_calls
			var toolCalls []map[string]interface{}
			for _, tc := range agentCtx.LastToolCalls {
				toolCall := map[string]interface{}{
					"id":   tc.ID,
					"type": tc.Type,
					"function": map[string]interface{}{
						"name":      tc.Function.Name,
						"arguments": tc.Function.Arguments,
					},
				}
				toolCalls = append(toolCalls, toolCall)
			}

			messages = append(messages, map[string]interface{}{
				"role":       "assistant",
				"content":    msg.Content,
				"tool_calls": toolCalls,
			})
		} else {
			// Regular message
			msgMap := map[string]interface{}{
				"role":    msg.Role,
				"content": msg.Content,
			}
			if msg.ToolCallID != nil {
				msgMap["tool_call_id"] = *msg.ToolCallID
			}
			messages = append(messages, msgMap)
		}
	}

	// Prepare tools for the request (same as before)
	var tools []struct {
		Type        string         `json:"type"`
		Name        string         `json:"name"`
		Description string         `json:"description"`
		Parameters  map[string]any `json:"parameters"`
	}

	for _, toolName := range agent.Tools {
		tool, err := agentCtx.ParentStepContext.ParentWorkflowContext.SystemManager.GetTool(toolName)
		if err != nil {
			continue
		}

		var parameters map[string]any
		if err := json.Unmarshal([]byte(tool.InputSchema), &parameters); err != nil {
			continue
		}

		openAITool := struct {
			Type        string         `json:"type"`
			Name        string         `json:"name"`
			Description string         `json:"description"`
			Parameters  map[string]any `json:"parameters"`
		}{
			Type:        "function",
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  parameters,
		}
		tools = append(tools, openAITool)
	}

	// Create the request structure manually to include tool_calls
	requestBody := map[string]interface{}{
		"model":    agent.Model,
		"messages": messages,
	}

	// Add tools if any are available
	if len(tools) > 0 {
		var openRouterTools []map[string]interface{}
		for _, tool := range tools {
			openRouterTool := map[string]interface{}{
				"type": "function",
				"function": map[string]interface{}{
					"description": tool.Description,
					"name":        tool.Name,
					"parameters":  tool.Parameters,
				},
			}
			openRouterTools = append(openRouterTools, openRouterTool)
		}
		requestBody["tools"] = openRouterTools
	}

	// Marshal to JSON and create a raw request
	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	fmt.Printf("🌐 Making follow-up LLM request for agent '%s'...\n", agent.UID)
	fmt.Printf("🔍 Request JSON: %s\n", string(requestJSON))

	// Log the messages being sent for debugging
	fmt.Printf("📝 Follow-up conversation messages for agent '%s':\n", agent.UID)
	for i, msg := range messages {
		msgJSON, _ := json.MarshalIndent(msg, "  ", "  ")
		fmt.Printf("  [%d] %s\n", i, string(msgJSON))
	}

	// Make raw HTTP request to OpenRouter
	response, err := agentCtx.makeRawOpenRouterRequest(requestJSON)
	if err != nil {
		fmt.Printf("❌ Follow-up LLM request failed for agent '%s': %v\n", agent.UID, err)
		return "", err
	}

	fmt.Printf("✅ Follow-up LLM response received for agent '%s'\n", agent.UID)

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from LLM")
	}

	// Update step-level conversation history with the final response (with agent name)
	responseMsg := openrouter.Message{
		Role:    response.Choices[0].Message.Role,
		Content: fmt.Sprintf("[%s]: %s", agent.Name, *response.Choices[0].Message.Content),
	}
	if agentCtx.ParentStepContext != nil {
		agentCtx.ParentStepContext.ConversationHistory = append(agentCtx.ParentStepContext.ConversationHistory, responseMsg)
	}

	// Check if there are more tool calls (recursive)
	if len(response.Choices[0].Message.ToolCalls) > 0 {
		fmt.Printf("🔧 Agent '%s' made additional tool calls, processing recursively\n", agent.UID)
		return agentCtx.processToolCalls(response.Choices[0].Message.ToolCalls, agent)
	}

	finalResponse := *response.Choices[0].Message.Content
	fmt.Printf("💬 Agent '%s' final response: %s\n", agent.UID, finalResponse)
	return finalResponse, nil
}

// handleUseAgentToolCall processes the special useAgentTool response
func (agentCtx *AgentExecutionContext) handleUseAgentToolCall(toolResult json.RawMessage, parentAgent *Agent) (string, error) {
	fmt.Printf("🔍 useAgentTool result (raw): %q\n", string(toolResult))
	fmt.Printf("🔍 useAgentTool result (pretty):\n%s\n", string(toolResult))

	// Parse the tool result to get the agent call key
	var result struct {
		AgentCallKey string `json:"agentCallKey"`
	}
	if err := json.Unmarshal(toolResult, &result); err != nil {
		fmt.Printf("❌ JSON parsing error: %v\n", err)
		fmt.Printf("❌ Failed to parse this JSON: %q\n", string(toolResult))
		return "", fmt.Errorf("failed to parse useAgentTool result: %v", err)
	}

	fmt.Printf("🔑 Extracted agent call key: %q\n", result.AgentCallKey)

	// Parse the agent call key: @@AGENT_CALL:agent_uid:input_data@@
	if !strings.HasPrefix(result.AgentCallKey, "@@AGENT_CALL:") || !strings.HasSuffix(result.AgentCallKey, "@@") {
		return "", fmt.Errorf("invalid agent call key format: %s", result.AgentCallKey)
	}

	// Extract agent UID and input
	keyContent := strings.TrimPrefix(result.AgentCallKey, "@@AGENT_CALL:")
	keyContent = strings.TrimSuffix(keyContent, "@@")
	parts := strings.SplitN(keyContent, ":", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid agent call key content: %s", keyContent)
	}

	agentUID := parts[0]
	input := parts[1]

	// Validate that the agent is in the parent's subAgents list
	found := false
	for _, subAgentUID := range parentAgent.SubAgents {
		if subAgentUID == agentUID {
			found = true
			break
		}
	}
	if !found {
		return "", fmt.Errorf("agent '%s' is not a subAgent of '%s'", agentUID, parentAgent.UID)
	}

	// Execute the subAgent
	return agentCtx.executeSubAgent(agentUID, input)
}

// executeSubAgent executes a subAgent and returns its response
func (agentCtx *AgentExecutionContext) executeSubAgent(agentUID, input string) (string, error) {
	// Get the subAgent
	subAgent, err := agentCtx.ParentStepContext.ParentWorkflowContext.SystemManager.GetAgent(agentUID)
	if err != nil {
		return "", fmt.Errorf("subAgent '%s' not found: %v", agentUID, err)
	}

	// Create a new agent execution context for the subAgent
	subAgentCtx := &AgentExecutionContext{
		AgentUID:            agentUID,
		ParentStepContext:   agentCtx.ParentStepContext,
		ConversationHistory: []openrouter.Message{}, // Fresh conversation for subAgent
		ToolCallHistory:     []ToolCallRecord{},
		PreprocessorResults: make(map[string]interface{}),
		CallCount:           0,
		MaxCalls:            10, // Reasonable limit for subAgent calls
		LocalKeyBank:        make(map[string]string),
		CalledByAgentUID:    agentCtx.AgentUID, // Track who called this subAgent
	}

	// Build agent context for subAgent (fresh context since it's down-level)
	subAgentContext := buildAgentContextFromHandoffs([]string{}, subAgentCtx.LocalKeyBank)
	setAgentContext(subAgentCtx, subAgentContext)

	// Execute the subAgent
	result, err := subAgentCtx.Execute(input, subAgent)
	if err != nil {
		return "", fmt.Errorf("subAgent '%s' execution failed: %v", agentUID, err)
	}

	return result, nil
}

// makeRawOpenRouterRequest makes a raw HTTP request to OpenRouter with custom JSON
func (agentCtx *AgentExecutionContext) makeRawOpenRouterRequest(requestJSON []byte) (*openrouter.NonStreamingChatResponse, error) {
	openRouter := agentCtx.ParentStepContext.ParentWorkflowContext.SystemManager.Openrouter
	if openRouter.ApiKey == "" {
		return nil, fmt.Errorf("API KEY is empty in openrouter instance")
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(requestJSON))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+openRouter.ApiKey)

	// Send the request
	client := &http.Client{Timeout: 5 * time.Minute}
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

	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse the response
	var response openrouter.NonStreamingChatResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("error parsing response: %w\nBody: %s", err, string(body))
	}

	return &response, nil
}
