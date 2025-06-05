# LLMango Agents Development & Testing Plan

## Overview
Create a systematic, incremental development process to build and test the agent framework, starting with the simplest possible implementation and expanding from there.

## Phase 1: Minimal Viable Agent System

### Goal
Create a single agent workflow with:
- ✅ Single agent (no tools)
- ✅ Single workflow step
- ✅ Input from frontend → Agent processing → Response back to frontend
- ✅ Full framework instantiation and validation
- ✅ Complete request/response cycle

### Development Process

#### Step 1: Framework Integration Setup
1. **Add llmangoagents to example app** (similar to llmango integration)
   - Update `example-app/go.mod` to include llmangoagents
   - Create agent configuration in `example-app/llmango.yaml`
   - Initialize agent system in `example-app/main.go`

#### Step 2: Minimal Agent Configuration
Create the simplest possible agent system:
```yaml
# example-app/agents.yaml
agents:
  - name: "simple_agent"
    systemMessage: "You are a helpful assistant. Respond concisely to user questions."
    model: "anthropic/claude-3-haiku"
    parameters: "{\"temperature\": 0.7}"
    tools: []
    preprocessors: []
    subAgents: []
    subWorkflows: []

workflows:
  - uid: "simple_workflow"
    name: "Simple Workflow"
    description: "Single agent, single step workflow"
    options:
      maxTime: 60
      maxSteps: 1
      maxSpend: 10
    steps:
      - uid: "step1"
        agent: "simple_agent"
        subAgents: []
        allowHandoffs: false
        exitBehavior: "default"
```

#### Step 3: Implementation Layers

##### Layer 1: Core Framework Functions
**File: `llmangoagents/minimal_test.go`**
```go
func TestMinimalAgentCreation(t *testing.T) {
    // Test creating a single agent
}

func TestMinimalWorkflowCreation(t *testing.T) {
    // Test creating a single workflow
}

func TestSystemValidation(t *testing.T) {
    // Test system validation with minimal config
}
```

##### Layer 2: System Integration
**File: `llmangoagents/integration_test.go`**
```go
func TestMinimalWorkflowExecution(t *testing.T) {
    // Test complete workflow execution
}

func TestAgentSystemManager(t *testing.T) {
    // Test system manager with minimal setup
}
```

##### Layer 3: Example App Integration
**File: `example-app/agents_test.go`**
```go
func TestExampleAppAgentIntegration(t *testing.T) {
    // Test loading agents in example app
}

func TestEndToEndWorkflow(t *testing.T) {
    // Test complete request/response cycle
}
```

#### Step 4: Frontend Integration

##### Backend API Endpoint
**File: `example-app/internal/mango/agents.go`**
```go
type AgentRequest struct {
    Input string `json:"input"`
    WorkflowUID string `json:"workflowUID"`
}

type AgentResponse struct {
    Output string `json:"output"`
    Status string `json:"status"`
    Error string `json:"error,omitempty"`
}

func (m *Mango) HandleAgentRequest(w http.ResponseWriter, r *http.Request) {
    // 1. Parse request
    // 2. Get workflow from agent system
    // 3. Create workflow instance
    // 4. Execute workflow
    // 5. Return response
}
```

##### Frontend Interface
**File: `llmangofrontend/svelte/src/routes/agents/+page.svelte`**
- Simple form with text input
- Submit button
- Response display area
- Status indicators

### Testing Strategy

#### Level 1: Unit Tests
```bash
cd llmangoagents
go test -v ./...
```
**Tests:**
- Agent creation and validation
- Workflow creation and validation
- System manager initialization
- Basic execution context creation

#### Level 2: Integration Tests
```bash
cd llmangoagents
go test -v -tags=integration ./...
```
**Tests:**
- Complete workflow execution
- Agent-to-LLM communication
- Context management and cleanup
- Error handling and recovery

#### Level 3: Example App Tests
```bash
cd example-app
go test -v ./...
```
**Tests:**
- Configuration loading
- Agent system initialization
- HTTP endpoint functionality
- End-to-end request/response

#### Level 4: Frontend Tests
```bash
cd llmangofrontend/svelte
npm test
```
**Tests:**
- UI component functionality
- API communication
- Error handling
- User experience flow

### Implementation Order

#### Week 1: Core Framework
1. **Day 1-2**: Fix any remaining compilation issues
2. **Day 3-4**: Implement minimal agent/workflow creation
3. **Day 5**: Write and pass unit tests

#### Week 2: System Integration
1. **Day 1-2**: Implement system validation and manager
2. **Day 3-4**: Implement basic workflow execution
3. **Day 5**: Write and pass integration tests

#### Week 3: Example App Integration
1. **Day 1-2**: Add agents to example app configuration
2. **Day 3-4**: Implement backend API endpoint
3. **Day 5**: Write and pass example app tests

#### Week 4: Frontend & E2E
1. **Day 1-2**: Create frontend agent interface
2. **Day 3-4**: Implement complete request/response cycle
3. **Day 5**: End-to-end testing and refinement

### Success Criteria

#### Minimal Viable Product (MVP) Checklist:
- [ ] Agent system compiles without errors
- [ ] Unit tests pass for core components
- [ ] Integration tests pass for workflow execution
- [ ] Example app loads agent configuration
- [ ] Backend API endpoint accepts requests
- [ ] Frontend can send input and display response
- [ ] Complete request cycle: Frontend → Backend → Agent → LLM → Response → Frontend

#### Quality Gates:
1. **Code Quality**: All tests pass, no compilation errors
2. **Functionality**: Complete request/response cycle works
3. **Performance**: Response time < 10 seconds for simple queries
4. **Reliability**: System handles errors gracefully
5. **Usability**: Frontend provides clear feedback to users

### Expansion Path

Once MVP is working:
1. **Add Tools**: Implement tool calling system
2. **Add Preprocessors**: Implement thinking and validation
3. **Add Multi-Step Workflows**: Chain multiple agents
4. **Add Handoffs**: Agent-to-agent communication
5. **Add Advanced Features**: Sub-agents, complex workflows

### Risk Mitigation

#### Potential Issues & Solutions:
1. **Compilation Errors**: Incremental testing at each step
2. **LLM Integration**: Mock LLM responses for testing
3. **Configuration Complexity**: Start with hardcoded values
4. **Frontend Complexity**: Use existing UI patterns
5. **Performance Issues**: Profile and optimize incrementally

This plan ensures we build the agent system incrementally, with testing at every level, and clear success criteria for each phase.