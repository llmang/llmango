# LLMango Agents

Multi-agent system for complex LLM workflows with step-based orchestration and agent collaboration.

## Architecture Overview

### Workflows → Steps → Agents
The system is built on a hierarchical structure where:

**Workflows** are composed of **Steps**, which each contain:
- A **Lead Agent** (planner and orchestrator for the step)
- **AssistantAgents** array (helpers that can be used as tools or for handoffs)
- Tools and capabilities specific to that step

### Common Workflow Pattern
```
Step 1: Generator Agent → Step 2: Validator Agent → Step 3: Formatter Agent → 
Step 4: DB Accessor Agent → Step 5: Final Validator → Return to User
```

## Core Components

### Workflow System 🚧
Multi-step orchestration with agent collaboration:

```go
type Workflow struct {
    Steps []WorkflowStep
    Name  string
    ID    string
}

type WorkflowStep struct {
    LeadAgent      Agent              // Primary orchestrator for this step
    AssistantAgents []Agent           // Helper agents for tools or handoffs
    AllowHandoffs  bool              // Enable agent-to-agent communication
    Tools          []Tool            // Step-specific tools
    SubWorkflows   []Workflow        // Nested workflow capabilities
}
```

### Agent Roles & Collaboration
- **Lead Agent**: Planner and orchestrator for each step
- **Assistant Agents**: Can be used as:
  - Single-use tools
  - Full agents allowing handoffs before returning to lead agent
- **Handoff Management**: Agents can pass control between each other within a step

### Preprocessor Modifiers 🚧
Enhance agent capabilities with modular preprocessors:

#### Thinking Preprocessor
```go
type ThinkingPreprocessor struct{}
// Makes agents think/reason before acting
```

#### Data Retrieval/Augmenter
```go
type DataAugmenterPreprocessor struct{}
// Based on user/action query:
// 1. Generates 1+ subqueries  
// 2. Searches vector database
// 3. Augments context with relevant data
```

### Tool System 🚧
Framework for LLM tool calling and function execution:

```go
type Tool interface {
    Name() string
    Description() string
    Execute(input json.RawMessage) (json.RawMessage, error)
}
```

### Built-in Tools 🚧
- Web search capabilities
- Text search and pattern matching  
- Database query execution
- Vector embedding and similarity search

### Schema Generation 🚧
Automatic tool schema generation for LLM integration:
- Go struct to JSON schema conversion
- Tool registration and management

## Key Files

- [`agents.go`](agents.go) - Core agent framework and workflow orchestration
- [`toolsupport.go`](toolsupport.go) - Tool registration and schema generation
- [`builtintools.go`](builtintools.go) - Standard tool implementations
- [`patterns.md`](patterns.md) - Agent interaction patterns and best practices

## Workflow Execution Flow

1. **Step Initialization**: Lead agent receives task and context
2. **Planning Phase**: Lead agent decides on approach using available tools/agents
3. **Execution Phase**: 
   - Use tools directly, OR
   - Delegate to assistant agents with potential handoffs
   - Assistant agents can collaborate before returning to lead agent
4. **Step Completion**: Lead agent consolidates results and passes to next step
5. **Workflow Completion**: Final step returns results to user

## Status: 🚧 In Development

New workflow-step architecture designed and being implemented.

**Current Focus:**
- Implement step-based workflow execution
- Build preprocessor modifier system
- Create handoff management between agents
- Add comprehensive agent collaboration patterns

**Next Steps:**
- Complete workflow orchestration engine
- Implement thinking and data augmentation preprocessors
- Add step validation and error handling
- Build workflow composition tools
- Comprehensive testing and examples

### Design Constraints
1. Agents must not be used in child toolcalls of themselves (build/lint/CLI/frontend validation)
2. Each step must have exactly one lead agent
3. Assistant agents can only hand off within their step scope
4. Preprocessors are applied in order and can modify agent context

