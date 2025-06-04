package llmangoagents

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

//AgentSystemManager holds references for all complied tools and agents and workflows and keeps trakc of the active flows running 
//WorkflowManager - is an instance of a running workflow orchestrating the actions over time and sending out tranismission events as needed

var llmangoAgentVersion = "0.0.1"
var llmangoAgentCompatabilityTimestamp=1749060000

// DEPENDENCY GRAPH SYSTEM FOR COMPILATION
// =====================================

type DependencyNodeType string
const (
	NodeTypeAgent    DependencyNodeType = "agent"
	NodeTypeWorkflow DependencyNodeType = "workflow"
	NodeTypeTool     DependencyNodeType = "tool"
)

type DependencyNode struct {
	ID           string             // unique identifier (agent/workflow/tool UID)
	Type         DependencyNodeType
	Dependencies []string           // list of IDs this node depends on
	Dependents   []string           // list of IDs that depend on this node
	InDegree     int               // number of incoming edges (dependencies)
	Compiled     bool              // whether this node has been compiled
	CompileData  interface{}       // stores the compiled result
}

type DependencyGraph struct {
	Nodes           map[string]*DependencyNode
	CompilationOrder []string                      // result of topological sort
	CircularDeps    [][]string                    // detected circular dependencies
}

// Build dependency graph from system inputs
func BuildDependencyGraph(inputs SystemInputList) (*DependencyGraph, error) {
	graph := &DependencyGraph{
		Nodes: make(map[string]*DependencyNode),
	}
	
	// PHASE 1: Create tool registry (tools are external dependencies, not part of dependency graph)
	toolRegistry := make(map[string]bool)
	for _, tool := range inputs.Tools {
		toolRegistry[tool.UID] = true
	}
	
	// PHASE 2: Create agent and workflow nodes with their internal system dependencies
	for _, agent := range inputs.Agents {
		deps := extractAgentDependencies(agent, toolRegistry)
		graph.addNode(agent.UID, NodeTypeAgent, deps)
	}
	
	for _, workflow := range inputs.Workflows {
		deps := extractWorkflowDependencies(workflow, toolRegistry)
		graph.addNode(workflow.UID, NodeTypeWorkflow, deps)
	}
	
	// PHASE 3: Build bidirectional edges and calculate in-degrees
	if err := graph.buildEdges(); err != nil {
		return nil, err
	}
	
	// PHASE 4: Detect circular dependencies
	if cycles := graph.detectCircularDependencies(); len(cycles) > 0 {
		graph.CircularDeps = cycles
		return graph, fmt.Errorf("circular dependencies detected: %v", cycles)
	}
	
	// PHASE 5: Generate compilation order using Khan's algorithm
	order, err := graph.topologicalSort()
	if err != nil {
		return nil, err
	}
	graph.CompilationOrder = order
	
	return graph, nil
}

func (dg *DependencyGraph) addNode(id string, nodeType DependencyNodeType, deps []string) {
	dg.Nodes[id] = &DependencyNode{
		ID:           id,
		Type:         nodeType,
		Dependencies: deps,
		Dependents:   []string{},
		InDegree:     len(deps),
		Compiled:     false,
	}
}

func (dg *DependencyGraph) buildEdges() error {
	for nodeID, node := range dg.Nodes {
		for _, depID := range node.Dependencies {
			depNode, exists := dg.Nodes[depID]
			if !exists {
				return fmt.Errorf("dependency '%s' not found for node '%s'", depID, nodeID)
			}
			
			// Add bidirectional reference
			depNode.Dependents = append(depNode.Dependents, nodeID)
		}
	}
	return nil
}

// KHAN'S ALGORITHM IMPLEMENTATION
func (dg *DependencyGraph) topologicalSort() ([]string, error) {
	// Create a copy of in-degrees for manipulation
	inDegrees := make(map[string]int)
	for id, node := range dg.Nodes {
		inDegrees[id] = node.InDegree
	}
	
	// Initialize queue with nodes having no dependencies
	queue := []string{}
	for id, degree := range inDegrees {
		if degree == 0 {
			queue = append(queue, id)
		}
	}
	
	compilationOrder := []string{}
	
	// Khan's algorithm main loop
	for len(queue) > 0 {
		// Remove node from queue
		current := queue[0]
		queue = queue[1:]
		compilationOrder = append(compilationOrder, current)
		
		// Reduce in-degree of all dependent nodes
		currentNode := dg.Nodes[current]
		for _, dependentID := range currentNode.Dependents {
			inDegrees[dependentID]--
			
			// If dependent now has no incoming edges, add to queue
			if inDegrees[dependentID] == 0 {
				queue = append(queue, dependentID)
			}
		}
	}
	
	// Check if all nodes were processed (no cycles)
	if len(compilationOrder) != len(dg.Nodes) {
		return nil, fmt.Errorf("circular dependency detected - could not compile all nodes")
	}
	
	return compilationOrder, nil
}

// DEPENDENCY EXTRACTION LOGIC
func extractAgentDependencies(agent Agent, toolRegistry map[string]bool) []string {
	deps := []string{}
	
	// Agent-level internal dependencies (all string IDs before compilation)
	// SubAgents that become internal agent tools
	deps = append(deps, agent.SubAgents...)
	
	// SubWorkflows that become internal agent tools  
	deps = append(deps, agent.SubWorkflows...)
	
	// agent.Tools contains ONLY external tools (not dependencies)
	
	return deps
}

func extractWorkflowDependencies(workflow Workflow, toolRegistry map[string]bool) []string {
	deps := []string{}
	
	for _, step := range workflow.Steps {
		// Lead agent dependency (string ID)
		deps = append(deps, step.Agent)
		
		// Step-level SubAgent dependencies ([]string)
		deps = append(deps, step.SubAgents...)
	}
	
	return removeDuplicates(deps)
}

func detectCircularDependencies(dg *DependencyGraph) [][]string {
	cycles := [][]string{}
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	
	for nodeID := range dg.Nodes {
		if !visited[nodeID] {
			if cyclePath := dfsDetectCycle(dg, nodeID, visited, recStack, []string{}); len(cyclePath) > 0 {
				cycles = append(cycles, cyclePath)
			}
		}
	}
	
	return cycles
}

func dfsDetectCycle(dg *DependencyGraph, nodeID string, visited, recStack map[string]bool, path []string) []string {
	visited[nodeID] = true
	recStack[nodeID] = true
	path = append(path, nodeID)
	
	node := dg.Nodes[nodeID]
	for _, depID := range node.Dependencies {
		if !visited[depID] {
			if cycle := dfsDetectCycle(dg, depID, visited, recStack, path); len(cycle) > 0 {
				return cycle
			}
		} else if recStack[depID] {
			// Found cycle - return the cycle path
			cycleStart := -1
			for i, id := range path {
				if id == depID {
					cycleStart = i
					break
				}
			}
			if cycleStart >= 0 {
				return append(path[cycleStart:], depID)
			}
		}
	}
	
	recStack[nodeID] = false
	return []string{}
}

// COMPILATION ORCHESTRATION
func CompileSystemWithDependencies(inputs SystemInputList) (*AgentSystemManager, error) {
	// Build dependency graph
	depGraph, err := BuildDependencyGraph(inputs)
	if err != nil {
		return nil, fmt.Errorf("dependency analysis failed: %w", err)
	}
	
	asm := &AgentSystemManager{
		CompatabillityCutoff: llmangoAgentCompatabilityTimestamp,
		Tools:               []*CompiledTools{},
		Agents:              []*CompliledAgents{},
		Workflows:           []WorkflowInstantiators{},
		ActiveWorkflows:     make(map[string]*WorkflowManager),
	}
	
	// Compile in dependency order
	for _, nodeID := range depGraph.CompilationOrder {
		node := depGraph.Nodes[nodeID]
		
		switch node.Type {
		case NodeTypeTool:
			compiledTool, err := compileToolByID(nodeID, inputs.Tools, asm)
			if err != nil {
				return nil, fmt.Errorf("failed to compile tool '%s': %w", nodeID, err)
			}
			asm.Tools = append(asm.Tools, compiledTool)
			node.CompileData = compiledTool
			
		case NodeTypeAgent:
			compiledAgent, err := compileAgentByID(nodeID, inputs.Agents, asm, depGraph)
			if err != nil {
				return nil, fmt.Errorf("failed to compile agent '%s': %w", nodeID, err)
			}
			asm.Agents = append(asm.Agents, compiledAgent)
			node.CompileData = compiledAgent
			
		case NodeTypeWorkflow:
			compiledWorkflow, err := compileWorkflowByID(nodeID, inputs.Workflows, asm, depGraph)
			if err != nil {
				return nil, fmt.Errorf("failed to compile workflow '%s': %w", nodeID, err)
			}
			asm.Workflows = append(asm.Workflows, compiledWorkflow)
			node.CompileData = compiledWorkflow
		}
		
		node.Compiled = true
	}
	
	return asm, nil
}

// COMPILATION HELPERS (PSEUDOCODE)
func compileToolByID(toolID string, tools []Tool, asm *AgentSystemManager) (*CompiledTools, error) {
	// Find tool by ID and compile it
	// Tools are leaf nodes so they can always be compiled first
	return nil, fmt.Errorf("not implemented")
}

func compileAgentByID(agentID string, agents []Agent, asm *AgentSystemManager, depGraph *DependencyGraph) (*CompliledAgents, error) {
	// Find agent by ID
	// All dependencies are guaranteed to be compiled already
	// Create agent with references to compiled dependencies
	return nil, fmt.Errorf("not implemented")
}

func compileWorkflowByID(workflowID string, workflows []Workflow, asm *AgentSystemManager, depGraph *DependencyGraph) (WorkflowInstantiators, error) {
	// Find workflow by ID  
	// All agent dependencies are guaranteed to be compiled already
	// Create workflow builder with references to compiled agents
	return nil, fmt.Errorf("not implemented")
}

// UTILITY FUNCTIONS
func removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	result := []string{}
	
	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}
	
	return result
}

type WorkflowInstantiator func(*AgentSystemManager)*WorkflowInstance
type AgentSystemManager struct{
	Openrouter *openrouter.OpenRouter //allows the system to make api calls
	GlobalKeyBank map[string]string //stores global kvs for toolcalls if needed?
	CompatabillityCutoff int //unix timestamp for last point of compatability (point where users can/cannot pick back up a conversation)//for vresioning potentially?

	HTTPToolConfigs []*HTTPToolBuilderConfig `json:"customTools"` //For sending the list over the wire the rest can be "reconstructued"

	Tools           []*CompiledTools    
	Agents          []*CompliledAgents    
	Workflows       []WorkflowInstantiators 

	ActiveWorkflows map[string]*WorkflowManager
}


type SystemInputList struct{
	Tools []Tool
	CustomToolConfigs []HTTPToolBuilderConfig
	Agents []Agent
	Workflows []Workflow
}

func CreateAgentSystemManager(
	inputs SystemInputList,
)(*AgentSystemManager, error){
	rtn=&AgentSystemManager{CompatabillityCutoff: llmangoAgentCompatabilityTimestamp}

	
	return nil,nil
}



func(asm *AgentSystemManager)RebuildWorkflowManager(string)(
	//rebuild the state of the workflow based on the previous calls from database logs.



)

func(asm *AgentSystemManager)startNewWorkflowInstance(workflowUid, input){
	//get the workflow from the list of flows
	//build an instance of a workflow for a user/start the workflow

}

type WorkflowManager struct {
	SystemManager *AgentSystemManager
	UserAgentInstance *UserAgentInstance
	WorkflowUUID      string //highest level wrapper around everthing nothing higher than this.

	
	context
	StepUUID          string
	CurrentDepth      int
	Input             json.RawMessage //user data or document data etc.
}

//For extending state from previous flows we can choose, a flow, a timestamp, and the logs 

type WorkflowInstance struct{
	UserId int //user who started the workflow
	Instance *CompiledWorkflow //the running workflow
	RunningCost         int
	RunningSteps        int
	RunningTime 		int //does not take into account "wait time" when waiting for user response
	
	Status string //running, failed, stopped, user, complete

	ConnPool //Websocket stuff hwere we transmit transmission events to
}






func (ht *HTTPTool) Compile(wm *WorkflowInstance) (*Tool, error) {
	//TODO: Convert HTTPTool to runtime Tool
	//TODO: Handle different tool types (builtin, http, function)
	//TODO: Create actual function implementations
	//TODO: Validate schemas
	return nil, fmt.Errorf("not implemented")
}

func (ad *Agent) Compile(wm *WorkflowInstance, tools map[string]*Tool) (*CompiledAgent, error) {
	//TODO: Convert AgentDef to runtime Agent
	//TODO: Resolve tool references
	//TODO: Parse parameters JSON to openrouter.Parameters
	//TODO: Setup preprocessors
	return nil, fmt.Errorf("not implemented")
}

func (wd *Workflow) Compile(wm *WorkflowInstance, agents map[string]*Agent, tools map[string]*Tool) (*CompiledWorkflow, error) {
	//TODO: Convert WorkflowDef to runtime WorkflowBuilder
	//TODO: Resolve all agent and tool references
	//TODO: Build workflow steps with proper dependencies
	//TODO: Validate step configurations
	return nil, fmt.Errorf("not implemented")
}

func (as *AgentSystem) Setup(SystemConfig) (*AgentSystemManager, error) {
	//TODO: This is the main compilation entry point
	//TODO: Create AgentSystemManager
	//TODO: Compile tools first (no dependencies)
	//TODO: Compile agents (depend on tools)
	//TODO: Compile workflows (depend on agents and tools)
	//TODO: Apply global settings
	//TODO: Validate all references
	return nil, fmt.Errorf("not implemented")
}

func (asm *AgentSystemManager) AddFromWire() {
	//add a tool, agent, or workflow from "wire" over internet/from the frontend/from a service manager
	//takes a json object and creates based on that
	//TODO: Implement JSON unmarshaling and validation
	//TODO: Support incremental updates (add single tool/agent/workflow)
	//TODO: Validate against existing system state
}

func (asm *AgentSystemManager) AddFromConfig(configData []byte) error {
	//This is loading agent system from a configuration -- tools, agents, workflows
	//only http tools can be loaded from config/built tools as rest are built in
	//cannot overwrite a built in tool with a http tool name
	//process goes load tools
	//load agents
	//load workflows

	var agentSystem AgentSystem
	if err := json.Unmarshal(configData, &agentSystem); err != nil {
		return fmt.Errorf("failed to parse agent system: %w", err)
	}

	//TODO: Call agentSystem.Compile() to get runtime objects
	//TODO: Validate version compatibility
	//TODO: Process environment variable interpolation in GlobalConfig.KeyBank
	//TODO: Apply compiled system to this manager

	return nil
}



// SerializeToConfig exports current system state to JSON config
func (asm *AgentSystemManager) SerializeToConfig() ([]byte, error) {
	//TODO: Extract current tools, agents, workflows
	//TODO: Convert runtime instances back to serializable format
	//TODO: Handle sensitive data (API keys) appropriately
	//TODO: Generate metadata (timestamps, version)
	return nil, fmt.Errorf("not implemented")
}





//=========================FROM CONFIG=================================
//=====================================================================

type JSONConfig struct {
	CustomToolConfigs     []HTTPToolBuilderConfig `json:"tools"`
	Agents    []Agent                 `json:"agents"`
	Workflows []Workflow              `json:"workflows"`
}


func (asm *AgentSystemManager) LoadJSONConfig(path string) (*JSONConfig, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: %s", path)
	}

	// Open and read the file
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	// Read file contents
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	var config JSONConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse JSON config: %w", err)
	}

	// Basic validation
	if len(config.Agents) == 0 && len(config.Workflows) == 0 {
		return nil, fmt.Errorf("config must contain at least one agent or workflow")
	}

	return &config, nil
}

// Compile methods - these are the opaque compilation steps
// Developer never calls these directly, system handles it

//JSON Examples of different patterns developers would work with:
/*
// EXAMPLE 1: Simple single-agent workflow - developer creates this directly
agentSystem := &AgentSystem{
	Version: "1.0",
	Metadata: SystemMetadata{
		Name: "Simple Search Agent",
		Description: "Basic search and response agent",
	},
	GlobalConfig: GlobalSettings{
		KeyBank: map[string]string{
			"search_api_key": "${BING_API_KEY}",
		},
	},
	Tools: []*HTTPTool{
		{
			UID: "search",
			Type: "builtin",
			Name: "search",
			Config: json.RawMessage(`{"provider": "bing"}`),
		},
	},
	Agents: []*AgentDef{
		{
			UID: "search_agent",
			Name: "Search Agent",
			SystemMessage: "You are a helpful search assistant.",
			Model: "anthropic/claude-3-sonnet",
			Parameters: json.RawMessage(`{"temperature": 0.7}`),
			Tools: []string{"search"},
		},
	},
	Workflows: []*WorkflowDef{
		{
			UID: "search_workflow",
			Name: "Search Workflow",
			Steps: []*WorkflowStepDef{
				{
					UID: "search_step",
					Agent: "search_agent",
					AllowHandoffs: false,
					ExitConditions: ExitConditions{
						ToUser: true,
						MaxTurns: 1,
					},
				},
			},
		},
	},
}

// Then system calls: compiledSystem, err := agentSystem.Compile()

// EXAMPLE 2: Multi-agent customer support workflow
{
  "version": "1.0",
  "metadata": {
    "name": "Customer Support System",
    "description": "Multi-agent customer support with escalation"
  },
  "globalConfig": {
    "keyBank": {
      "customer_api": "${CUSTOMER_DB_KEY}",
      "ticket_system": "${TICKET_API_KEY}"
    },
    "defaultLimits": {
      "maxTime": 600,
      "maxSteps": 10,
      "maxSpend": 100
    }
  },
  "tools": [
    {
      "uid": "customer_lookup",
      "type": "http",
      "name": "customer lookup tool",
      "description": "Look up customer information",
      "endpoint": "https://api.company.com/customers",
      "method": "POST",
      "headers": {
        "Authorization": "Bearer ${customer_api}"
      },
      "inputSchema": {
        "type": "object",
        "properties": {
          "customerId": {"type": "string"}
        },
        "required": ["customerId"]
      }
    }
  ],
  "agents": [
    {
      "uid": "triage_agent",
      "name": "Triage Agent",
      "systemMessage": "You handle initial customer inquiries and route them appropriately.",
      "model": "anthropic/claude-3-haiku",
      "tools": ["customer_lookup"],
      "preprocessors": ["thinking"]
    }
  ],
  "workflows": [
    {
      "uid": "support_workflow",
      "name": "Customer Support Flow",
      "options": {
	  	"maxCost": .25 //25 cents
        "maxTime": 900,
        "maxSteps": 8
      },
      "steps": [
        {
          "uid": "triage_step",
          "agent": "triage_agent",
          "exitConditions": {
            "action": "default",
          }
        }
      ]
    }
  ]
}
*/
