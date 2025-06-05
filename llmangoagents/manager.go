package llmangoagents

import (
	"fmt"
	"time"
)

// SYSTEM VALIDATION (replaces compilation)
func ValidateSystemWithDependencies(inputs SystemInputList) (*AgentSystemManager, error) {
	// Build dependency graph for validation only
	_, err := BuildDependencyGraph(inputs)
	if err != nil {
		return nil, fmt.Errorf("dependency validation failed: %w", err)
	}

	// Auto-generate useAgentTools for agents with subAgents
	var additionalTools []*Tool
	for _, agent := range inputs.Agents {
		if len(agent.SubAgents) > 0 {
			useAgentTool := NewUseAgentTool(&agent)
			if useAgentTool != nil {
				additionalTools = append(additionalTools, useAgentTool)
			}
		}
	}

	// Create system manager with direct references
	totalTools := len(inputs.Tools) + len(additionalTools)
	asm := &AgentSystemManager{
		CompatabillityCutoff: llmangoAgentCompatabilityTimestamp,
		Tools:                make([]*Tool, totalTools),
		Agents:               make([]*Agent, len(inputs.Agents)),
		Workflows:            make([]*Workflow, len(inputs.Workflows)),
		ActiveWorkflows:      make(map[string]*WorkflowManager),
	}

	// Copy original tools
	for i, tool := range inputs.Tools {
		asm.Tools[i] = &tool
	}
	// Add auto-generated useAgentTools
	for i, tool := range additionalTools {
		asm.Tools[len(inputs.Tools)+i] = tool
	}

	// Copy agents and add useAgentTool to their tools list if they have subAgents
	for i, agent := range inputs.Agents {
		agentCopy := agent // Create a copy
		if len(agent.SubAgents) > 0 {
			// Add useAgentTool to the agent's tools list
			agentCopy.Tools = append(agentCopy.Tools, "useAgentTool")
		}
		asm.Agents[i] = &agentCopy
	}

	// Copy workflows
	for i, workflow := range inputs.Workflows {
		asm.Workflows[i] = &workflow
	}

	return asm, nil
}

// Runtime lookup methods
func (asm *AgentSystemManager) GetWorkflow(workflowUID string) (*Workflow, error) {
	for _, wf := range asm.Workflows {
		if wf.UID == workflowUID {
			return wf, nil
		}
	}
	return nil, fmt.Errorf("workflow with UID '%s' not found", workflowUID)
}

func (asm *AgentSystemManager) GetAgent(agentUID string) (*Agent, error) {
	for _, agent := range asm.Agents {
		if agent.UID == agentUID {
			return agent, nil
		}
	}
	return nil, fmt.Errorf("agent with UID '%s' not found", agentUID)
}

func (asm *AgentSystemManager) GetTool(toolUID string) (*Tool, error) {
	for _, tool := range asm.Tools {
		if tool.Uid == toolUID || tool.Name == toolUID {
			return tool, nil
		}
	}
	return nil, fmt.Errorf("tool with UID '%s' not found", toolUID)
}

// WORKFLOW EXECUTION
func (asm *AgentSystemManager) StartNewWorkflowInstance(workflowUID string, userID int, input string) (*WorkflowInstance, error) {
	// Get workflow by UID
	workflow, err := asm.GetWorkflow(workflowUID)
	if err != nil {
		return nil, err
	}

	// Create execution context
	ctx := &WorkflowExecutionContext{
		WorkflowUUID:     workflowUID,
		UserID:           userID,
		CreatedAt:        time.Now(),
		TTL:              time.Hour * 24,
		Workflow:         workflow,
		Steps:            make(map[string]*StepExecutionContext),
		CurrentStepIndex: 0,
		GlobalKeyBank:    make(map[string]string),
		WorkflowLimits:   workflow.Options,
		SystemManager:    asm,
		ActiveRefs:       1,
	}

	// Create workflow instance that holds the context
	instance := &WorkflowInstance{
		UserId:      userID,
		WorkflowDef: workflow,
		Context:     ctx,
		Status:      "running",
	}

	// Execute workflow
	result, err := ctx.Execute(input)
	if err != nil {
		instance.Status = "failed"
		ctx.Cleanup()
		return nil, err
	}

	// Store result in context for retrieval
	ctx.GlobalKeyBank["final_result"] = result
	instance.Status = "completed"

	return instance, nil
}

//rebuild the state of the workflow based on the previous calls from database logs.
func (asm *AgentSystemManager) RebuildWorkflowManager(workflowUID string) {
	// TODO: Implement workflow state rebuilding from database logs
}