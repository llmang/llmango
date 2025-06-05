package llmangoagents

import (
	"encoding/json"
	"fmt"
	"strings"
)

// buildContextualSystemMessage creates a context-aware system message with proper hierarchical context:
// 1. Overall workflow goal
// 2. Current step goal
// 3. Agent's specific job
// 4. Context inheritance rules based on agent level
func buildContextualSystemMessage(agent *Agent, agentCtx *AgentExecutionContext) string {
	var systemMessage strings.Builder
	
	// Determine if this is an agent subagent call (down-level) or same-level
	// Agent subagents are called via useAgentTool and have CalledByAgentUID set
	// Step subagents are same-level and don't have CalledByAgentUID set
	isAgentSubAgent := agentCtx.CalledByAgentUID != ""
	
	if isAgentSubAgent {
		// Down-level: Agent SubAgent gets fresh context - only their specific job
		systemMessage.WriteString(agent.SystemMessage)
		
		if agentCtx.AgentContext != "" {
			systemMessage.WriteString("\n\n## Agent Context:\n")
			systemMessage.WriteString(agentCtx.AgentContext)
		}
	} else {
		// Same-level: Lead agent and step subagents get full hierarchical context
		
		// 1. Overall workflow goal
		if agentCtx.ParentStepContext != nil &&
		   agentCtx.ParentStepContext.ParentWorkflowContext != nil &&
		   agentCtx.ParentStepContext.ParentWorkflowContext.Workflow != nil {
			systemMessage.WriteString("## Workflow Goal:\n")
			systemMessage.WriteString(agentCtx.ParentStepContext.ParentWorkflowContext.Workflow.Description)
			systemMessage.WriteString("\n\n")
		}
		
		// 2. Current step goal
		if agentCtx.ParentStepContext != nil && agentCtx.ParentStepContext.ParentWorkflowContext != nil {
			// Find the current step to get its description
			for _, step := range agentCtx.ParentStepContext.ParentWorkflowContext.Workflow.Steps {
				if step.UID == agentCtx.ParentStepContext.StepUID {
					if step.Description != "" {
						systemMessage.WriteString("## Step Goal:\n")
						systemMessage.WriteString(step.Description)
						systemMessage.WriteString("\n\n")
					}
					break
				}
			}
		}
		
		// 3. Agent's specific job
		systemMessage.WriteString("## Your Job:\n")
		if agent.Description != "" {
			systemMessage.WriteString(agent.Description)
			systemMessage.WriteString("\n\n")
		}
		systemMessage.WriteString(agent.SystemMessage)
		
		// Add handoff history for step agents
		if agentCtx.ParentStepContext != nil && len(agentCtx.ParentStepContext.HandoffHistory) > 0 {
			systemMessage.WriteString("\n\n## Handoff History:\n")
			for i, handoff := range agentCtx.ParentStepContext.HandoffHistory {
				systemMessage.WriteString(fmt.Sprintf("%d. %s\n", i+1, handoff))
			}
		}
		
		// Add reminder for multi-agent steps
		if agentCtx.ParentStepContext != nil {
			// Check if this step has multiple agents (lead + subAgents)
			stepAgentCount := 1 // lead agent
			if agentCtx.ParentStepContext.ParentWorkflowContext != nil {
				// Find the current step to check for subAgents
				for _, step := range agentCtx.ParentStepContext.ParentWorkflowContext.Workflow.Steps {
					if step.UID == agentCtx.ParentStepContext.StepUID {
						stepAgentCount += len(step.SubAgents)
						break
					}
				}
			}
			
			if stepAgentCount > 1 {
				systemMessage.WriteString("\n\n## Multi-Agent Step Reminder:\n")
				systemMessage.WriteString("This step has multiple agents available. If you feel the step's goal is not yet complete, make sure to call a tool to either change agents (handoffTool) or gain more information!")
			}
		}
		
		// Add agent context
		if agentCtx.AgentContext != "" {
			systemMessage.WriteString("\n\n## Agent Context:\n")
			systemMessage.WriteString(agentCtx.AgentContext)
		}
	}
	
	return systemMessage.String()
}

// setWorkflowContext sets the original user input in the workflow context
func setWorkflowContext(workflowCtx *WorkflowExecutionContext, originalUserInput string) {
	if workflowCtx != nil {
		workflowCtx.OriginalUserInput = originalUserInput
	}
}

// setStepContext sets context information for a step
func setStepContext(stepCtx *StepExecutionContext, context string) {
	if stepCtx != nil {
		stepCtx.StepContext = context
	}
}

// setAgentContext sets context information for an agent
func setAgentContext(agentCtx *AgentExecutionContext, context string) {
	if agentCtx != nil {
		agentCtx.AgentContext = context
	}
}

// buildStepContextFromInput creates step context from step input and previous results
func buildStepContextFromInput(stepInput string, previousStepResults []string) string {
	var context strings.Builder
	
	if stepInput != "" {
		context.WriteString("Step Input: ")
		context.WriteString(stepInput)
	}
	
	if len(previousStepResults) > 0 {
		if context.Len() > 0 {
			context.WriteString("\n\n")
		}
		context.WriteString("Previous Step Results:\n")
		for i, result := range previousStepResults {
			context.WriteString(fmt.Sprintf("Step %d: %s\n", i+1, result))
		}
	}
	
	return context.String()
}

// buildStepContextMessages creates mock tool call/response pairs for completed steps
// This provides clean context for subsequent steps without overwhelming conversation history
func buildStepContextMessages(workflowCtx *WorkflowExecutionContext, currentStepIndex int) []map[string]interface{} {
	var messages []map[string]interface{}
	
	if workflowCtx == nil || workflowCtx.Workflow == nil {
		return messages
	}
	
	// Iterate through completed steps (before current step)
	for i := 0; i < currentStepIndex && i < len(workflowCtx.Workflow.Steps); i++ {
		step := workflowCtx.Workflow.Steps[i]
		stepCtx, exists := workflowCtx.Steps[step.UID]
		
		if !exists || !stepCtx.IsCompleted {
			continue
		}
		
		// Create mock tool call for the step
		toolCallID := fmt.Sprintf("call_step_%s", step.UID)
		
		// Build tool call arguments
		stepInput := string(stepCtx.StepInput)
		if stepInput == "" {
			stepInput = workflowCtx.OriginalUserInput
		}
		
		arguments := map[string]interface{}{
			"step":  step.UID,
			"goal":  fmt.Sprintf("Execute %s step", step.UID),
			"input": stepInput,
		}
		
		argumentsJSON, _ := json.Marshal(arguments)
		
		// Assistant message with tool call
		assistantMsg := map[string]interface{}{
			"role": "assistant",
			"content": fmt.Sprintf("Executing step: %s", step.UID),
			"tool_calls": []map[string]interface{}{
				{
					"id":   toolCallID,
					"type": "function",
					"function": map[string]interface{}{
						"name":      "execute_workflow_step",
						"arguments": string(argumentsJSON),
					},
				},
			},
		}
		
		// Tool response message
		stepOutput := string(stepCtx.StepOutput)
		if stepOutput == "" {
			stepOutput = "Step completed successfully"
		}
		
		toolMsg := map[string]interface{}{
			"role":         "tool",
			"tool_call_id": toolCallID,
			"content":      stepOutput,
		}
		
		messages = append(messages, assistantMsg, toolMsg)
	}
	
	return messages
}

// buildAgentContextFromHandoffs creates agent context from handoff history and local state
func buildAgentContextFromHandoffs(handoffHistory []string, localKeyBank map[string]string) string {
	var context strings.Builder
	
	if len(handoffHistory) > 0 {
		context.WriteString("Agent Handoff History:\n")
		for i, handoff := range handoffHistory {
			context.WriteString(fmt.Sprintf("%d. %s\n", i+1, handoff))
		}
	}
	
	if len(localKeyBank) > 0 {
		if context.Len() > 0 {
			context.WriteString("\n\n")
		}
		context.WriteString("Available Context Keys:\n")
		for key, value := range localKeyBank {
			context.WriteString(fmt.Sprintf("- %s: %s\n", key, value))
		}
	}
	
	return context.String()
}