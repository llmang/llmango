package llmangoagents

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/llmang/llmango/openrouter"
)

//AgentSystemManager holds references for all complied tools and agents and workflows and keeps trakc of the active flows running
//WorkflowManager - is an instance of a running workflow orchestrating the actions over time and sending out tranismission events as needed

var llmangoAgentVersion = "0.0.1"
var llmangoAgentCompatabilityTimestamp = 1749060000

// EXECUTION CONTEXT HIERARCHY FOR ISOLATION
// ==========================================

type WorkflowExecutionContext struct {
	WorkflowUUID string
	UserID       int
	CreatedAt    time.Time
	TTL          time.Duration

	// Workflow reference (lookup-based)
	Workflow *Workflow

	// Step management
	Steps            map[string]*StepExecutionContext
	CurrentStepUID   string
	CurrentStepIndex int
	MaxSteps         int
	CurrentStepCount int

	// Global workflow state
	GlobalKeyBank  map[string]string
	WorkflowLimits WorkflowLimits

	// System access for runtime lookup
	SystemManager *AgentSystemManager

	// Reference counting for cleanup
	ActiveRefs int32
}

type StepExecutionContext struct {
	StepUID               string
	ParentWorkflowContext *WorkflowExecutionContext

	// Agent management within step
	Agents            map[string]*AgentExecutionContext
	LeadAgentUID      string
	MaxAgentCalls     int
	CurrentAgentCalls int

	// Step-specific state
	AllowHandoffs  bool
	HandoffHistory []string
	StepInput      json.RawMessage
	StepOutput     json.RawMessage

	// Cleanup
	IsCompleted bool
}

type AgentExecutionContext struct {
	AgentUID          string
	ParentStepContext *StepExecutionContext

	// Agent state
	ConversationHistory []openrouter.Message
	ToolCallHistory     []ToolCallRecord
	PreprocessorResults map[string]interface{}

	// Execution tracking
	CallCount int
	MaxCalls  int

	// Isolation
	LocalKeyBank map[string]string

	// Agent relationship (for sub-agents, this points to the calling agent)
	CalledByAgentUID string // empty for lead agents, set for sub-agents
}

type ToolCallRecord struct {
	ToolUID   string
	Input     json.RawMessage
	Output    json.RawMessage
	Error     error
	Timestamp time.Time
}

type ToolExecutionContext struct {
	ToolUID        string
	ParentAgentCtx *AgentExecutionContext
	CallIndex      int

	// Tool-specific isolation
	Secrets    map[string]string
	LocalState map[string]interface{}
}

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

// Preprocessor function type with context
type PreprocessorFunc func(*Agent, string, *AgentExecutionContext) string

// SYSTEM VALIDATION (replaces compilation)
func ValidateSystemWithDependencies(inputs SystemInputList) (*AgentSystemManager, error) {
	// Build dependency graph for validation only
	_, err := BuildDependencyGraph(inputs)
	if err != nil {
		return nil, fmt.Errorf("dependency validation failed: %w", err)
	}

	// Create system manager with direct references
	asm := &AgentSystemManager{
		CompatabillityCutoff: llmangoAgentCompatabilityTimestamp,
		Tools:                make([]*Tool, len(inputs.Tools)),
		Agents:               make([]*Agent, len(inputs.Agents)),
		Workflows:            make([]*Workflow, len(inputs.Workflows)),
		ActiveWorkflows:      make(map[string]*WorkflowManager),
	}

	// Copy tools
	for i, tool := range inputs.Tools {
		asm.Tools[i] = &tool
	}
	// Copy agents
	for i, agent := range inputs.Agents {
		asm.Agents[i] = &agent
	}
	// Copy workflows
	for i, workflow := range inputs.Workflows {
		asm.Workflows[i] = &workflow
	}

	return asm, nil
}

type WorkflowInstantiator func(*AgentSystemManager) *WorkflowInstance
type AgentSystemManager struct {
	Openrouter           *openrouter.OpenRouter //allows the system to make api calls
	GlobalKeyBank        map[string]string      //stores global kvs for toolcalls if needed?
	CompatabillityCutoff int                    //unix timestamp for last point of compatability (point where users can/cannot pick back up a conversation)//for vresioning potentially?

	HTTPToolConfigs []*HTTPToolBuilderConfig `json:"customTools"` //For sending the list over the wire the rest can be "reconstructued"

	Tools     []*Tool
	Agents    []*Agent
	Workflows []*Workflow

	ActiveWorkflows map[string]*WorkflowManager
}

type SystemInputList struct {
	Tools             []Tool
	CustomToolConfigs []HTTPToolBuilderConfig
	Agents            []Agent
	Workflows         []Workflow
}

func CreateAgentSystemManager(
	inputs SystemInputList,
) (*AgentSystemManager, error) {
	return ValidateSystemWithDependencies(inputs)
}

func (asm *AgentSystemManager) RebuildWorkflowManager(string)

//rebuild the state of the workflow based on the previous calls from database logs.

// Runtime lookup methods
func (asm *AgentSystemManager) GetWorkflow(workflowUID string) (*Workflow, error) {
	for _, wf := range asm.Workflows {
		if wf.UID == workflowUID {
			return wf, nil
		}
	}
	return nil, fmt.Errorf("workflow with UID '%s' not found", workflowUID)
}

func (asm *AgentSystemManager) GetAgent(agentName string) (*Agent, error) {
	for _, agent := range asm.Agents {
		if agent.Name == agentName {
			return agent, nil
		}
	}
	return nil, fmt.Errorf("agent with name '%s' not found", agentName)
}

func (asm *AgentSystemManager) GetTool(toolUID string) (*Tool, error) {
	for _, tool := range asm.Tools {
		if tool.Uid == toolUID {
			return tool, nil
		}
	}
	return nil, fmt.Errorf("tool with UID '%s' not found", toolUID)
}

// LEGACY: Keep for backward compatibility
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

// Execute runs the workflow with the given input
func (ctx *WorkflowExecutionContext) Execute(input string) (string, error) {
	currentInput := input

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
			AllowHandoffs:         step.AllowHandoffs,
			StepInput:             json.RawMessage(currentInput),
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
	// Get lead agent from system manager
	leadAgent, err := stepCtx.ParentWorkflowContext.SystemManager.GetAgent(stepCtx.LeadAgentUID)
	if err != nil {
		return "", err
	}

	// Create agent execution context
	agentCtx := &AgentExecutionContext{
		AgentUID:            leadAgent.Name,
		ParentStepContext:   stepCtx,
		ConversationHistory: []openrouter.Message{},
		ToolCallHistory:     []ToolCallRecord{},
		PreprocessorResults: make(map[string]interface{}),
		LocalKeyBank:        make(map[string]string),
		MaxCalls:            10, // Default max calls
	}

	// Register agent context
	stepCtx.Agents[leadAgent.Name] = agentCtx

	// Execute agent
	result, err := agentCtx.Execute(input, leadAgent)
	if err != nil {
		return "", err
	}

	return result, nil
}

// Execute runs the agent with the given input
func (agentCtx *AgentExecutionContext) Execute(input string, agent *Agent) (string, error) {
	// Track call count
	agentCtx.CallCount++
	if agentCtx.CallCount > agentCtx.MaxCalls {
		return "", fmt.Errorf("agent '%s' exceeded max calls (%d)", agent.Name, agentCtx.MaxCalls)
	}

	currentInput := input

	// Run preprocessors first
	for _, preprocessorName := range agent.PreProcessors {
		// TODO: Look up and run actual preprocessor function
		// For now, just pass through
		processedInput := runPreprocessor(preprocessorName, agent, currentInput, agentCtx)
		if processedInput != "" {
			currentInput = processedInput
		}
	}

	// Check for action phrases in preprocessor results
	if action := parseActionPhrases(currentInput); action != "" {
		switch action {
		case "@@ABORT":
			return "", fmt.Errorf("workflow aborted by agent")
		case "@@RETURN":
			return currentInput, nil
		}
	}

	// Build system message
	systemMessage := agent.SystemMessage

	// Create conversation messages
	messages := []openrouter.Message{
		{Role: "system", Content: systemMessage},
	}

	// Add conversation history
	messages = append(messages, agentCtx.ConversationHistory...)

	// Add current input
	messages = append(messages, openrouter.Message{
		Role:    "user",
		Content: currentInput,
	})

	// Make LLM request
	req := &openrouter.OpenRouterRequest{
		Model:    &agent.Model,
		Messages: messages,
		// TODO: Add tools if agent has them
	}

	response, err := agentCtx.ParentStepContext.ParentWorkflowContext.SystemManager.Openrouter.GenerateNonStreamingChatResponse(req)
	if err != nil {
		return "", err
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from LLM")
	}

	// Update conversation history
	responseMsg := openrouter.Message{
		Role:    response.Choices[0].Message.Role,
		Content: *response.Choices[0].Message.Content,
	}
	agentCtx.ConversationHistory = append(agentCtx.ConversationHistory,
		openrouter.Message{Role: "user", Content: currentInput},
		responseMsg,
	)

	// Handle tool calls if present
	// TODO: Check if response has tool calls and process them

	return *response.Choices[0].Message.Content, nil
}

// Helper functions
func runPreprocessor(preprocessorName string, agent *Agent, input string, agentCtx *AgentExecutionContext) string {
	// TODO: Implement actual preprocessor lookup and execution
	return ""
}

func parseActionPhrases(input string) string {
	// Simple action phrase detection
	if strings.Contains(input, "@@ABORT") {
		return "@@ABORT"
	}
	if strings.Contains(input, "@@RETURN") {
		return "@@RETURN"
	}
	return ""
}

type WorkflowManager struct {
	SystemManager *AgentSystemManager
	WorkflowUUID  string
	StepUUID      string
	CurrentDepth  int
	Input         json.RawMessage
}

type WorkflowInstance struct {
	UserId       int
	WorkflowDef  *Workflow                 // Reference to the workflow definition
	Context      *WorkflowExecutionContext // The context tree/graph
	RunningCost  int
	RunningSteps int
	RunningTime  int
	Status       string
}
