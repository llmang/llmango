# Step 1: Minimal Agent System Implementation

## Immediate Next Actions

### 1. Fix Current Compilation Issues

Before we can proceed, we need to ensure the refactored package compiles correctly. Based on the current state, we need to:

#### A. Remove unused sync/atomic import from types.go
```go
// Remove this line from types.go imports:
"sync/atomic"
```

#### B. Verify all type references are correct
- Ensure all `UID` vs `Uid` field references are consistent
- Check that all function signatures match between declarations and implementations

### 2. Create Minimal Test Configuration

#### A. Create test configuration file
**File: `llmangoagents/test_config.go`**
```go
package llmangoagents

func GetMinimalTestConfig() SystemInputList {
    return SystemInputList{
        Tools: []Tool{}, // No tools for minimal test
        CustomToolConfigs: []HTTPToolBuilderConfig{},
        Agents: []Agent{
            {
                Name:          "test_agent",
                SystemMessage: "You are a helpful assistant. Respond concisely to user questions.",
                Model:         "anthropic/claude-3-haiku",
                Parameters:    `{"temperature": 0.7}`,
                Tools:         []string{},
                PreProcessors: []string{},
                SubAgents:     []string{},
                SubWorkflows:  []string{},
            },
        },
        Workflows: []Workflow{
            {
                UID:         "test_workflow",
                Name:        "Test Workflow",
                Description: "Minimal single-agent workflow for testing",
                Options: WorkflowLimits{
                    MaxTime:  60,
                    MaxSteps: 1,
                    MaxSpend: 10,
                },
                Steps: []*WorkflowStep{
                    {
                        UID:           "step1",
                        Agent:         "test_agent",
                        SubAgents:     []string{},
                        AllowHandoffs: false,
                        ExitBehavior:  "default",
                    },
                },
            },
        },
    }
}
```

### 3. Create Basic Unit Tests

#### A. Test agent system creation
**File: `llmangoagents/basic_test.go`**
```go
package llmangoagents

import (
    "testing"
)

func TestMinimalSystemCreation(t *testing.T) {
    config := GetMinimalTestConfig()
    
    // Test system creation
    system, err := CreateAgentSystemManager(config)
    if err != nil {
        t.Fatalf("Failed to create agent system: %v", err)
    }
    
    if system == nil {
        t.Fatal("System is nil")
    }
    
    // Verify agents were loaded
    if len(system.Agents) != 1 {
        t.Fatalf("Expected 1 agent, got %d", len(system.Agents))
    }
    
    // Verify workflows were loaded
    if len(system.Workflows) != 1 {
        t.Fatalf("Expected 1 workflow, got %d", len(system.Workflows))
    }
}

func TestAgentLookup(t *testing.T) {
    config := GetMinimalTestConfig()
    system, err := CreateAgentSystemManager(config)
    if err != nil {
        t.Fatalf("Failed to create system: %v", err)
    }
    
    // Test agent lookup
    agent, err := system.GetAgent("test_agent")
    if err != nil {
        t.Fatalf("Failed to get agent: %v", err)
    }
    
    if agent.Name != "test_agent" {
        t.Fatalf("Expected agent name 'test_agent', got '%s'", agent.Name)
    }
}

func TestWorkflowLookup(t *testing.T) {
    config := GetMinimalTestConfig()
    system, err := CreateAgentSystemManager(config)
    if err != nil {
        t.Fatalf("Failed to create system: %v", err)
    }
    
    // Test workflow lookup
    workflow, err := system.GetWorkflow("test_workflow")
    if err != nil {
        t.Fatalf("Failed to get workflow: %v", err)
    }
    
    if workflow.UID != "test_workflow" {
        t.Fatalf("Expected workflow UID 'test_workflow', got '%s'", workflow.UID)
    }
}
```

### 4. Create Mock LLM for Testing

#### A. Mock OpenRouter for testing
**File: `llmangoagents/mock_llm.go`**
```go
package llmangoagents

import (
    "github.com/llmang/llmango/openrouter"
)

type MockOpenRouter struct {
    Response string
    Error    error
}

func (m *MockOpenRouter) GenerateNonStreamingChatResponse(req *openrouter.OpenRouterRequest) (*openrouter.OpenRouterResponse, error) {
    if m.Error != nil {
        return nil, m.Error
    }
    
    content := m.Response
    if content == "" {
        content = "Mock response from test agent"
    }
    
    return &openrouter.OpenRouterResponse{
        Choices: []openrouter.Choice{
            {
                Message: openrouter.Message{
                    Role:    "assistant",
                    Content: &content,
                },
            },
        },
    }, nil
}

func CreateTestSystemWithMockLLM(response string) (*AgentSystemManager, error) {
    config := GetMinimalTestConfig()
    system, err := CreateAgentSystemManager(config)
    if err != nil {
        return nil, err
    }
    
    // Replace with mock
    system.Openrouter = &MockOpenRouter{Response: response}
    
    return system, nil
}
```

### 5. Test Workflow Execution

#### A. Test complete workflow execution
**File: `llmangoagents/execution_test.go`**
```go
package llmangoagents

import (
    "testing"
)

func TestMinimalWorkflowExecution(t *testing.T) {
    // Create system with mock LLM
    system, err := CreateTestSystemWithMockLLM("Hello! I'm a test response.")
    if err != nil {
        t.Fatalf("Failed to create test system: %v", err)
    }
    
    // Test workflow execution
    instance, err := system.StartNewWorkflowInstance("test_workflow", 1, "Hello, how are you?")
    if err != nil {
        t.Fatalf("Failed to start workflow: %v", err)
    }
    
    if instance == nil {
        t.Fatal("Workflow instance is nil")
    }
    
    if instance.Status != "completed" {
        t.Fatalf("Expected status 'completed', got '%s'", instance.Status)
    }
    
    // Check that we got a response
    result := instance.Context.GlobalKeyBank["final_result"]
    if result == "" {
        t.Fatal("No final result in workflow context")
    }
    
    t.Logf("Workflow result: %s", result)
}
```

### 6. Integration with Example App

#### A. Update example-app go.mod
```go
module github.com/llmang/llmango/example-app

go 1.24.2

require (
    github.com/llmang/llmango v0.1.0
    // Add local reference to llmangoagents
)

replace github.com/llmang/llmango => ../
```

#### B. Create agent configuration
**File: `example-app/agents.yaml`**
```yaml
agents:
  - name: "example_agent"
    systemMessage: "You are a helpful assistant for the example application. Provide clear, concise responses."
    model: "anthropic/claude-3-haiku"
    parameters: '{"temperature": 0.7}'
    tools: []
    preprocessors: []
    subAgents: []
    subWorkflows: []

workflows:
  - uid: "example_workflow"
    name: "Example Workflow"
    description: "Simple example workflow for testing"
    options:
      maxTime: 120
      maxSteps: 1
      maxSpend: 25
    steps:
      - uid: "main_step"
        agent: "example_agent"
        subAgents: []
        allowHandoffs: false
        exitBehavior: "default"
```

#### C. Initialize agents in main.go
**File: `example-app/internal/mango/agents.go`**
```go
package mango

import (
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    
    "github.com/llmang/llmango/llmangoagents"
    "github.com/llmang/llmango/openrouter"
    "gopkg.in/yaml.v3"
)

type AgentConfig struct {
    Agents    []llmangoagents.Agent    `yaml:"agents"`
    Workflows []llmangoagents.Workflow `yaml:"workflows"`
}

func (m *Mango) InitializeAgents() error {
    // Load agent configuration
    data, err := os.ReadFile("agents.yaml")
    if err != nil {
        return fmt.Errorf("failed to read agents.yaml: %w", err)
    }
    
    var config AgentConfig
    if err := yaml.Unmarshal(data, &config); err != nil {
        return fmt.Errorf("failed to parse agents.yaml: %w", err)
    }
    
    // Create system input
    systemInput := llmangoagents.SystemInputList{
        Tools:             []llmangoagents.Tool{},
        CustomToolConfigs: []llmangoagents.HTTPToolBuilderConfig{},
        Agents:            config.Agents,
        Workflows:         config.Workflows,
    }
    
    // Create agent system
    agentSystem, err := llmangoagents.CreateAgentSystemManager(systemInput)
    if err != nil {
        return fmt.Errorf("failed to create agent system: %w", err)
    }
    
    // Set up OpenRouter (reuse from existing mango setup)
    agentSystem.Openrouter = m.OpenRouter
    
    m.AgentSystem = agentSystem
    return nil
}

type AgentRequest struct {
    Input       string `json:"input"`
    WorkflowUID string `json:"workflowUID"`
}

type AgentResponse struct {
    Output string `json:"output"`
    Status string `json:"status"`
    Error  string `json:"error,omitempty"`
}

func (m *Mango) HandleAgentRequest(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    var req AgentRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    // Execute workflow
    instance, err := m.AgentSystem.StartNewWorkflowInstance(req.WorkflowUID, 1, req.Input)
    if err != nil {
        resp := AgentResponse{
            Status: "error",
            Error:  err.Error(),
        }
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(resp)
        return
    }
    
    // Get result
    result := instance.Context.GlobalKeyBank["final_result"]
    
    resp := AgentResponse{
        Output: result,
        Status: instance.Status,
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}
```

### 7. Testing Checklist

#### Before proceeding to frontend:
- [ ] `go build` succeeds in llmangoagents package
- [ ] All unit tests pass: `go test -v ./llmangoagents`
- [ ] Mock LLM tests work correctly
- [ ] Example app compiles with agent integration
- [ ] Agent configuration loads successfully
- [ ] HTTP endpoint responds to test requests

#### Test Commands:
```bash
# Test package compilation
cd llmangoagents && go build .

# Run unit tests
cd llmangoagents && go test -v .

# Test example app
cd example-app && go build .

# Test agent endpoint (after starting server)
curl -X POST http://localhost:8080/agents \
  -H "Content-Type: application/json" \
  -d '{"input": "Hello, how are you?", "workflowUID": "example_workflow"}'
```

This step-by-step approach ensures we build incrementally and test at each level before moving to the next phase.