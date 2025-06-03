# LLMango Agents

Multi-agent system for complex LLM workflows with tool support and agent orchestration.

## Features (🚧 In Development)

### Tool System 🚧
Framework for LLM tool calling and function execution:

```go
// Tool interface for extensible functionality
type Tool interface {
    Name() string
    Description() string
    Execute(input json.RawMessage) (json.RawMessage, error)
}
```

### Built-in Tools 🚧
- [`bingsearchtool.go`](bingsearchtool.go) - Web search capabilities
- [`greptool.go`](greptool.go) - Text search and pattern matching
- [`sqltool.go`](sqltool.go) - Database query execution
- [`vectorembedtool.go`](vectorembedtool.go) - Vector embedding and similarity search

### Agent Orchestration 🚧
Multi-agent workflows with coordination and communication:

- Agent-to-agent communication
- Workflow orchestration
- State management across agents
- Tool sharing and coordination

### Schema Generation 🚧
Automatic tool schema generation for LLM integration:

- [`generateschemafromstruct.go`](generateschemafromstruct.go) - Go struct to JSON schema conversion
- [`toolsupport.go`](toolsupport.go) - Tool registration and management

## Key Components

- [`agents.go`](agents.go) - Core agent framework and orchestration
- [`toolsupport.go`](toolsupport.go) - Tool registration and schema generation
- Various tool implementations for common LLM tasks

## Status: 🚧 In Development

Agent system architecture planned but implementation incomplete. Tool framework partially implemented.

**Next Steps:**
- Complete agent orchestration system
- Implement agent-to-agent communication
- Add workflow management
- Comprehensive testing and examples