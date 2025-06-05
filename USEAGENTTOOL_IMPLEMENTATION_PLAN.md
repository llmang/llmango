# useAgentTool Implementation Plan

## Current Status
❌ **useAgentTool is NOT implemented** in the llmango codebase.

## What Exists
- Agent struct has `SubAgents []string` field
- Basic tool infrastructure exists but incomplete
- Agent execution has TODOs for tool handling
- Example config shows agents with subAgents but no way to use them

## Simple Implementation Plan

### Step 1: Auto-generate useAgentTool
**File**: `llmangoagents/tools_builtin.go`
- Add function to create useAgentTool for agents with subAgents
- Tool schema allows selecting which subAgent to call and what input to send

### Step 2: Add useAgentTool during agent system creation
**File**: `llmangoagents/builder.go` or `llmangoagents/manager.go`
- When creating agent system, check each agent for subAgents
- If subAgents exist, auto-add useAgentTool to that agent's tools

### Step 3: Complete tool call processing
**File**: `llmangoagents/agent_execution.go`
- Complete the TODO on line 84 to handle tool calls from LLM responses
- Add special handling for useAgentTool calls

### Step 4: Add subAgent execution
**File**: `llmangoagents/agent_execution.go`
- Add function to execute subAgents
- Handle the "key" emission and parsing internally

## Key Design Points
- useAgentTool is automatically added (not user-configurable)
- Tool emits internal key that gets parsed by system
- SubAgent execution reuses existing agent execution logic
- No external configuration needed - purely internal mechanism

## Files to Modify
1. `llmangoagents/tools_builtin.go` - Add useAgentTool generator
2. `llmangoagents/builder.go` - Auto-add useAgentTool during system creation
3. `llmangoagents/agent_execution.go` - Complete tool handling + subAgent execution

## Test with Existing Config
The `example-app/agents.json` already has:
- `organizer_agent` with `subAgents: ["sentiment_agent", "correction_agent"]`
- This should automatically get useAgentTool added