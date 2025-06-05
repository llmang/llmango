package llmangoagents

import (
	"sync/atomic"
)

// Context cleanup methods
func (ctx *WorkflowExecutionContext) Cleanup() {
	atomic.StoreInt32(&ctx.ActiveRefs, 0)

	// Clean up all step contexts
	for _, stepCtx := range ctx.Steps {
		stepCtx.Cleanup()
	}

	// Clear maps
	ctx.Steps = nil
	ctx.GlobalKeyBank = nil
	ctx.SystemManager = nil
}

func (ctx *StepExecutionContext) Cleanup() {
	// Clean up all agent contexts
	for _, agentCtx := range ctx.Agents {
		agentCtx.Cleanup()
	}

	// Clear references
	ctx.Agents = nil
	ctx.HandoffHistory = nil
	ctx.ParentWorkflowContext = nil
}

func (ctx *AgentExecutionContext) Cleanup() {
	// Clear state
	ctx.ConversationHistory = nil
	ctx.ToolCallHistory = nil
	ctx.PreprocessorResults = nil
	ctx.LocalKeyBank = nil
	ctx.ParentStepContext = nil
}