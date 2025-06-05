# Agent System Integration

This document describes the integration of the llmangoagents system with the example app, creating a complete end-to-end workflow from frontend to backend to agent to LLM and back.

## Overview

The integration adds the following components to the example app:

1. **Agent Configuration** (`agents.yaml`) - Defines agents and workflows
2. **Agent System Integration** (`internal/mango/agents.go`) - Handles agent initialization and HTTP requests
3. **HTTP Endpoint** (`/agents`) - Accepts JSON requests and returns agent responses
4. **Tests** (`agents_test.go`) - Verifies the integration works correctly

## Architecture

```
Frontend/Client → HTTP Request → Backend → Agent System → LLM → Response
```

### Request Flow

1. **Client** sends POST request to `/agents` with JSON payload
2. **Backend** validates request and extracts input/workflowUID
3. **Agent System** loads workflow and executes agent
4. **Agent** processes input using configured LLM model
5. **Response** flows back through the chain to client

## Configuration

### Agent Configuration (`agents.yaml`)

```json
{
  "agents": [
    {
      "name": "example_agent",
      "systemMessage": "You are a helpful assistant...",
      "model": "anthropic/claude-3-sonnet",
      "parameters": "{\"temperature\": 0.7}",
      "tools": [],
      "preProcessors": [],
      "subAgents": [],
      "subWorkflows": []
    }
  ],
  "workflows": [
    {
      "uid": "example_workflow",
      "name": "Example Workflow",
      "description": "A simple workflow that processes user input",
      "options": {
        "maxTime": 300,
        "maxSteps": 5,
        "maxSpend": 50
      },
      "steps": [
        {
          "uid": "main_step",
          "agent": "example_agent",
          "subAgents": [],
          "allowHandoffs": false,
          "exitBehavior": "default"
        }
      ]
    }
  ]
}
```

## API Endpoint

### POST `/agents`

**Request:**
```json
{
  "input": "user message",
  "workflowUID": "example_workflow"
}
```

**Response:**
```json
{
  "output": "agent response",
  "status": "completed"
}
```

**Error Response:**
```json
{
  "output": "",
  "status": "error",
  "error": "error message"
}
```

## Integration Points

### Mango Struct Enhancement

The `Mango` struct now includes an `AgentSystem` field:

```go
type Mango struct {
    *llmango.LLMangoManager
    AgentSystem *llmangoagents.AgentSystemManager
    Debug       bool
}
```

### Initialization

Agent system is initialized during startup:

```go
// Initialize agent system
if err := mangoClient.InitializeAgents(); err != nil {
    log.Fatal("Failed to initialize agent system:", err)
}
```

### HTTP Handler

The agent HTTP handler processes requests:

```go
func (m *Mango) HandleAgentRequest(w http.ResponseWriter, r *http.Request) {
    // Parse request
    // Validate input
    // Execute workflow
    // Return response
}
```

## Testing

### Unit Tests

Run the integration tests:

```bash
cd example-app
go test -v ./... -run TestAgentSystemIntegration
```

### HTTP Endpoint Tests

Test the HTTP endpoint validation:

```bash
go test -v ./... -run TestAgentHTTPEndpointValidation
```

### End-to-End Testing

1. Start the server:
```bash
cd example-app
go run main.go
```

2. Run the test script:
```bash
./test_agent_endpoint.sh
```

3. Or test manually with curl:
```bash
curl -X POST http://localhost:8080/agents \
  -H "Content-Type: application/json" \
  -d '{"input": "Hello, can you help me?", "workflowUID": "example_workflow"}'
```

## Features

### Reuses Existing Infrastructure

- **OpenRouter Instance**: Agent system reuses the existing OpenRouter client
- **Logging**: Integrates with existing mango logging system
- **Error Handling**: Consistent error handling patterns

### Validation

- **Request Validation**: Validates required fields (input, workflowUID)
- **Agent System Validation**: Ensures agents and workflows exist
- **HTTP Method Validation**: Only accepts POST requests

### Configuration Loading

- **JSON Configuration**: Loads agent configuration from `agents.yaml`
- **Dependency Validation**: Validates agent/workflow dependencies
- **System Compilation**: Compiles configuration into runtime system

## Error Handling

The system handles various error conditions:

- Missing or invalid configuration files
- Invalid JSON requests
- Missing required fields
- Agent/workflow not found
- LLM execution errors
- Network/API errors

## Security Considerations

- **API Key Management**: Reuses existing OpenRouter API key
- **Input Validation**: Validates all user inputs
- **Error Messages**: Sanitized error messages to prevent information leakage

## Performance

- **Shared Resources**: Reuses existing OpenRouter connection
- **Efficient Parsing**: Minimal JSON parsing overhead
- **Memory Management**: Proper cleanup of workflow contexts

## Future Enhancements

Potential improvements:

1. **Authentication**: Add API key or token-based authentication
2. **Rate Limiting**: Implement request rate limiting
3. **Caching**: Cache agent responses for repeated queries
4. **Metrics**: Add detailed metrics and monitoring
5. **Streaming**: Support streaming responses for long-running workflows
6. **Multi-step Workflows**: Support more complex multi-agent workflows