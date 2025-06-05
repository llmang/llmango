package llmangoagents

import (
	"encoding/json"
	"time"

	"github.com/llmang/llmango/openrouter"
)

// Version and compatibility
var llmangoAgentVersion = "0.0.1"
var llmangoAgentCompatabilityTimestamp = 1749060000

// Function type definitions
type PreprocessorFunc func(*Agent, string, *AgentExecutionContext) string
type ToolConstructor func(*WorkflowManager) *Tool
type WorkflowInstantiator func(*AgentSystemManager) *WorkflowInstance

// Core system configuration types
type GlobalSettings struct {
	KeyBank             map[string]string `json:"keyBank"`             // Environment variable mapping
	CompatibilityCutoff int               `json:"compatibilityCutoff"` // Unix timestamp
	DefaultLimits       WorkflowLimits    `json:"defaultLimits"`
}

type WorkflowLimits struct {
	MaxTime  int `json:"maxTime"`  // seconds
	MaxSteps int `json:"maxSteps"` // number of steps
	MaxSpend int `json:"maxSpend"` // cost units
}

// Tool system types
type HTTPToolBuilderConfig struct {
	UID         string `json:"uid"`
	Type        string `json:"type"` // "builtin" | "http" | "function"
	Name        string `json:"name"`
	Description string `json:"description"`

	Endpoint        string `json:"endpoint"`        // POST endpoint for the tool
	ExtraHeaders    string `json:"extraHeaders"`    // Headers as JSON or key:value\nkey:value
	RequiredSecrets string `json:"requiredSecrets"` // comma-separated list of required secrets

	InputSchema  json.RawMessage `json:"inputSchema"`  // JSON schema for validation
	OutputSchema json.RawMessage `json:"outputSchema"` // JSON schema for validation
}

// you will only get the secrets from required secrets this is to provide transparency and saftey regarding secret sharing
type Tool struct {
	Uid             string
	Name            string
	Description     string
	Function        func(map[string]string, json.RawMessage) (json.RawMessage, error)
	RequiredSecrets string
	InputSchema     string
	OutputSchema    string
}

// Agent system types
type Agent struct {
	UID           string   `json:"uid"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`   // Brief description for tool handoffs
	SystemMessage string   `json:"systemMessage"`
	Model         string   `json:"model"`
	Parameters    string   `json:"parameters"`
	Tools         []string `json:"tools"`         //abilities for the agent.
	PreProcessors []string `json:"preprocessors"` //in order of their usage
	SubAgents     []string `json:"subAgents"`
	SubWorkflows  []string `json:"subWorkflows"`
}

// Workflow system types
type Workflow struct {
	UID         string          `json:"uid"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Options     WorkflowLimits  `json:"options"`
	Steps       []*WorkflowStep `json:"steps"`
}

type WorkflowStep struct {
	UID          string   `json:"uid"`
	Description  string   `json:"description"`  // Description of what this step accomplishes
	Agent        string   `json:"agent"`        // Reference to agent UID
	SubAgents    []string `json:"subAgents"`    // References to agent UIDs
	ExitBehavior string   `json:"exitBehavior"` // "default", "return|s4", "user"
}

// System management types
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

// Execution tracking types
type ToolCallRecord struct {
	ToolUID   string
	Input     json.RawMessage
	Output    json.RawMessage
	Error     error
	Timestamp time.Time
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

	// Context for tiered system messages
	OriginalUserInput string

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
	HandoffHistory []string
	StepInput      json.RawMessage
	StepOutput     json.RawMessage

	// Context for tiered system messages
	StepContext string

	// Shared conversation context for all step agents
	ConversationHistory []openrouter.Message
	ToolCallHistory     []ToolCallRecord

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

	// Tool call tracking for proper conversation flow
	LastToolCalls []openrouter.ToolCall

	// Context for tiered system messages
	AgentContext string
}

type ToolExecutionContext struct {
	ToolUID        string
	ParentAgentCtx *AgentExecutionContext
	CallIndex      int

	// Tool-specific isolation
	Secrets    map[string]string
	LocalState map[string]interface{}
}
