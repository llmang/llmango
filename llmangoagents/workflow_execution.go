package llmangoagents

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/llmang/llmango/openrouter"
)

// Execute runs the workflow with the given input
func (ctx *WorkflowExecutionContext) Execute(input string) (string, error) {
	currentInput := input

	// Set the original user input for context inheritance
	setWorkflowContext(ctx, input)

	// Run through each step in sequence
	for i, step := range ctx.Workflow.Steps {
		ctx.CurrentStepIndex = i
		ctx.CurrentStepUID = step.UID

		// Create step context
		stepCtx := &StepExecutionContext{
			StepUID:               step.UID,
			ParentWorkflowContext: ctx,
			Agents:                make(map[string]*AgentExecutionContext),
			LeadAgentUID:          step.Agent,
			StepInput:             json.RawMessage(currentInput),
			ConversationHistory:   []openrouter.Message{}, // Shared conversation context
			ToolCallHistory:       []ToolCallRecord{},     // Shared tool call history
		}

		// Build step context from previous results
		var previousResults []string
		for j := 0; j < i; j++ {
			if prevStep, exists := ctx.Steps[ctx.Workflow.Steps[j].UID]; exists && prevStep.IsCompleted {
				previousResults = append(previousResults, string(prevStep.StepOutput))
			}
		}
		stepContext := buildStepContextFromInput(currentInput, previousResults)
		setStepContext(stepCtx, stepContext)

		// Generate handoff tools for step agents if subAgents exist
		if len(step.SubAgents) > 0 {
			// Create list of all agents in this step (lead agent + subAgents)
			allStepAgents := []string{step.Agent}
			allStepAgents = append(allStepAgents, step.SubAgents...)

			// Generate handoff tools for each agent in the step
			for _, agentUID := range allStepAgents {
				handoffTool := NewHandoffTool(allStepAgents, agentUID, ctx.SystemManager)
				if handoffTool != nil {
					// Add the handoff tool to the system manager temporarily for this step
					ctx.SystemManager.Tools = append(ctx.SystemManager.Tools, handoffTool)

					// Add handoffTool to the agent's tools list
					agent, err := ctx.SystemManager.GetAgent(agentUID)
					if err == nil {
						// Create a copy of the agent with the handoff tool added
						agentCopy := *agent
						agentCopy.Tools = append(agentCopy.Tools, "handoffTool")

						// Update the agent in the system manager
						for j, existingAgent := range ctx.SystemManager.Agents {
							if existingAgent.UID == agentUID {
								ctx.SystemManager.Agents[j] = &agentCopy
								break
							}
						}
					}
				}
			}
		}

		// Register step context
		ctx.Steps[step.UID] = stepCtx

		// Execute step
		result, err := stepCtx.Execute(currentInput)
		if err != nil {
			return "", fmt.Errorf("step '%s' failed: %w", step.UID, err)
		}

		stepCtx.StepOutput = json.RawMessage(result)
		stepCtx.IsCompleted = true
		currentInput = result
	}

	return currentInput, nil
}

// Execute runs a single step with the given input
func (stepCtx *StepExecutionContext) Execute(input string) (string, error) {
	// Start with the lead agent
	currentAgentUID := stepCtx.LeadAgentUID
	currentInput := input

	// Execute agents in the step, handling handoffs
	for {
		// Get current agent from system manager
		currentAgent, err := stepCtx.ParentWorkflowContext.SystemManager.GetAgent(currentAgentUID)
		if err != nil {
			return "", err
		}

		// Create or get agent execution context
		agentCtx, exists := stepCtx.Agents[currentAgentUID]
		if !exists {
			agentCtx = &AgentExecutionContext{
				AgentUID:            currentAgent.UID,
				ParentStepContext:   stepCtx,
				ConversationHistory: []openrouter.Message{}, // Individual agent history (not used for step agents)
				ToolCallHistory:     []ToolCallRecord{},     // Individual agent history (not used for step agents)
				PreprocessorResults: make(map[string]interface{}),
				LocalKeyBank:        make(map[string]string),
				MaxCalls:            10, // Default max calls
			}

			// Build agent context from handoffs and local state
			agentContext := buildAgentContextFromHandoffs(stepCtx.HandoffHistory, agentCtx.LocalKeyBank)
			setAgentContext(agentCtx, agentContext)

			// Ensure this agent has handoff tools if this step has subAgents
			// Find the current step to check for subAgents
			for _, step := range stepCtx.ParentWorkflowContext.Workflow.Steps {
				if step.UID == stepCtx.StepUID && len(step.SubAgents) > 0 {
					// Create list of all agents in this step (lead agent + subAgents)
					allStepAgents := []string{step.Agent}
					allStepAgents = append(allStepAgents, step.SubAgents...)

					// Check if this agent needs a handoff tool and doesn't have one
					needsHandoffTool := false
					for _, agentUID := range allStepAgents {
						if agentUID == currentAgentUID {
							needsHandoffTool = true
							break
						}
					}

					if needsHandoffTool {
						// Check if handoff tool already exists for this agent
						handoffToolExists := false
						for _, tool := range stepCtx.ParentWorkflowContext.SystemManager.Tools {
							if tool.Uid == fmt.Sprintf("handoffTool_%s", currentAgentUID) {
								handoffToolExists = true
								break
							}
						}

						if !handoffToolExists {
							// Generate handoff tool for this agent
							handoffTool := NewHandoffTool(allStepAgents, currentAgentUID, stepCtx.ParentWorkflowContext.SystemManager)
							if handoffTool != nil {
								// Add the handoff tool to the system manager
								stepCtx.ParentWorkflowContext.SystemManager.Tools = append(stepCtx.ParentWorkflowContext.SystemManager.Tools, handoffTool)

								// Add handoffTool to the agent's tools list
								agentCopy := *currentAgent
								agentCopy.Tools = append(agentCopy.Tools, "handoffTool")

								// Update the agent in the system manager
								for j, existingAgent := range stepCtx.ParentWorkflowContext.SystemManager.Agents {
									if existingAgent.UID == currentAgentUID {
										stepCtx.ParentWorkflowContext.SystemManager.Agents[j] = &agentCopy
										break
									}
								}
							}
						}
					}
					break
				}
			}

			// Register agent context
			stepCtx.Agents[currentAgent.UID] = agentCtx
		}

		// Execute current agent
		result, err := agentCtx.Execute(currentInput, currentAgent)
		if err != nil {
			return "", err
		}

		// Check if the result contains a handoff key
		if handoffTarget := parseHandoffKey(result); handoffTarget != "" {
			// Record the handoff in history
			handoffRecord := fmt.Sprintf("%s -> %s", currentAgentUID, handoffTarget)
			stepCtx.HandoffHistory = append(stepCtx.HandoffHistory, handoffRecord)

			// Continue with the handoff target agent
			currentAgentUID = handoffTarget
			currentInput = result // Pass the result as input to the next agent
			continue
		}

		// No handoff, return the result
		return result, nil
	}
}

// parseHandoffKey extracts the target agent UID from a handoff key
// The result is a JSON object: {"handoffKey":"@@HANDOFF:target_agent:input:reason@@"}
func parseHandoffKey(result string) string {
	// First try to parse as JSON to extract the handoffKey
	var handoffResult struct {
		HandoffKey string `json:"handoffKey"`
	}

	if err := json.Unmarshal([]byte(result), &handoffResult); err == nil && handoffResult.HandoffKey != "" {
		// Parse the handoff key: @@HANDOFF:target_agent:input:reason@@
		handoffKey := handoffResult.HandoffKey
		if strings.HasPrefix(handoffKey, "@@HANDOFF:") && strings.HasSuffix(handoffKey, "@@") {
			content := strings.TrimPrefix(handoffKey, "@@HANDOFF:")
			content = strings.TrimSuffix(content, "@@")
			parts := strings.SplitN(content, ":", 3)
			if len(parts) >= 1 {
				return parts[0] // Return target agent UID
			}
		}
	}

	// Fallback: look for handoff key pattern directly in the result
	if strings.Contains(result, "@@HANDOFF:") {
		start := strings.Index(result, "@@HANDOFF:")
		if start != -1 {
			end := strings.Index(result[start:], "@@")
			if end != -1 {
				handoffKey := result[start : start+end+2]
				content := strings.TrimPrefix(handoffKey, "@@HANDOFF:")
				content = strings.TrimSuffix(content, "@@")
				parts := strings.SplitN(content, ":", 3)
				if len(parts) >= 1 {
					return parts[0] // Return target agent UID
				}
			}
		}
	}

	return ""
}
