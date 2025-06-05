package llmangoagents

import (
	"encoding/json"
	"fmt"
	"strings"
)

//Agent specific tools
//Thinking - ability to pause and think
//
//When subAgents
//    -- Transmit Back partial data
//    -- Optional Handoff to other agent

// Cloud function tool
//allow users to have arbitrary endpoints for tools
//input json output json.
//basically just ability to use lambda or cloudflare workers to host funcs if they are heavy/to allow for 3rd party functions easier. create func from url instead of something else pre provides self contained workspace.

//searchtool
//bing or google

type searchTool struct{}

func (s *searchTool) Run(input json.RawMessage) (json.RawMessage, error) {
	// TODO: Implement search functionality
	return json.RawMessage(`{"result": "search not implemented"}`), nil
}

func (s *searchTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"query": {
				"type": "string",
				"description": "The search query to execute"
			}
		},
		"required": ["query"]
	}`)
}

func (s *searchTool) Description() string {
	return "Search for information using a query string"
}

func (s *searchTool) Name() string {
	return "search"
}

func NewBingTool(manager *WorkflowManager) *Tool {
	return &Tool{
		Uid:  "bing_search",
		Name: "bing_search",
		Description: "Search using Bing search engine",
		Function: func(secrets map[string]string, input json.RawMessage) (json.RawMessage, error) {
			//add the user id
			//filter the message
			//log the usage in the manager
			//do the action and return
			// TODO: Implement actual Bing search functionality
			return json.RawMessage(`{"result": "bing search not implemented"}`), nil
		},
		InputSchema:  `{"type": "object", "properties": {"query": {"type": "string"}}, "required": ["query"]}`,
		OutputSchema: `{"type": "object", "properties": {"result": {"type": "string"}}}`,
	}
}

// NewUseAgentTool creates a useAgentTool for agents with subAgents
// This tool allows an agent to delegate work to its subAgents
func NewUseAgentTool(agent *Agent) *Tool {
	if len(agent.SubAgents) == 0 {
		return nil // No subAgents, no tool needed
	}

	// Create enum of available subAgents for the schema
	subAgentEnums := make([]string, len(agent.SubAgents))
	for i, subAgent := range agent.SubAgents {
		subAgentEnums[i] = fmt.Sprintf(`"%s"`, subAgent)
	}
	enumStr := strings.Join(subAgentEnums, ", ")

	inputSchema := fmt.Sprintf(`{
		"type": "object",
		"properties": {
			"agent": {
				"type": "string",
				"description": "The UID of the subAgent to invoke",
				"enum": [%s]
			},
			"input": {
				"type": "string",
				"description": "The input to pass to the subAgent"
			}
		},
		"required": ["agent", "input"]
	}`, enumStr)

	return &Tool{
		Uid:         fmt.Sprintf("useAgentTool_%s", agent.UID),
		Name:        "useAgentTool",
		Description: fmt.Sprintf("Delegate work to one of the available subAgents: %s", strings.Join(agent.SubAgents, ", ")),
		Function: func(secrets map[string]string, input json.RawMessage) (json.RawMessage, error) {
			// Parse the input to get agent and input
			var toolInput struct {
				Agent string `json:"agent"`
				Input string `json:"input"`
			}
			if err := json.Unmarshal(input, &toolInput); err != nil {
				return nil, fmt.Errorf("invalid useAgentTool input: %v", err)
			}

			// Emit the special key that will be parsed by the system
			key := fmt.Sprintf("@@AGENT_CALL:%s:%s@@", toolInput.Agent, toolInput.Input)
			
			// Create the response struct and marshal it properly to handle escaping
			response := struct {
				AgentCallKey string `json:"agentCallKey"`
			}{
				AgentCallKey: key,
			}
			
			result, err := json.Marshal(response)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal useAgentTool response: %v", err)
			}
			
			return json.RawMessage(result), nil
		},
		InputSchema:  inputSchema,
		OutputSchema: `{"type": "object", "properties": {"agentCallKey": {"type": "string"}}}`,
	}
}

// NewHandoffTool creates a handoff tool for step agents to transfer work to other step agents
// This tool allows agents within the same step to handoff work to each other
func NewHandoffTool(stepAgents []string, currentAgentUID string, systemManager *AgentSystemManager) *Tool {
	if len(stepAgents) <= 1 {
		return nil // No other agents to handoff to
	}

	// Filter out the current agent from the available handoff targets
	var availableAgents []string
	var agentDescriptions []string
	
	for _, agentUID := range stepAgents {
		if agentUID != currentAgentUID {
			availableAgents = append(availableAgents, agentUID)
			
			// Get agent description for better tool description
			if agent, err := systemManager.GetAgent(agentUID); err == nil {
				description := agent.Description
				if description == "" {
					description = agent.Name // Fallback to name if no description
				}
				agentDescriptions = append(agentDescriptions, fmt.Sprintf("%s: %s", agentUID, description))
			} else {
				agentDescriptions = append(agentDescriptions, agentUID)
			}
		}
	}

	if len(availableAgents) == 0 {
		return nil // No other agents available
	}

	// Create enum of available agents for the schema
	agentEnums := make([]string, len(availableAgents))
	for i, agentUID := range availableAgents {
		agentEnums[i] = fmt.Sprintf(`"%s"`, agentUID)
	}
	enumStr := strings.Join(agentEnums, ", ")

	inputSchema := fmt.Sprintf(`{
		"type": "object",
		"properties": {
			"agent": {
				"type": "string",
				"description": "The UID of the step agent to handoff work to. Available: %s",
				"enum": [%s]
			},
			"input": {
				"type": "string",
				"description": "The input/context to pass to the target agent"
			},
			"reason": {
				"type": "string",
				"description": "Brief explanation of why this handoff is needed"
			}
		},
		"required": ["agent", "input", "reason"]
	}`, strings.Join(agentDescriptions, "; "), enumStr)

	return &Tool{
		Uid:         fmt.Sprintf("handoffTool_%s", currentAgentUID),
		Name:        "handoffTool",
		Description: fmt.Sprintf("Handoff work to another agent in this step. Available agents: %s", strings.Join(agentDescriptions, "; ")),
		Function: func(secrets map[string]string, input json.RawMessage) (json.RawMessage, error) {
			// Parse the input to get agent, input, and reason
			var toolInput struct {
				Agent  string `json:"agent"`
				Input  string `json:"input"`
				Reason string `json:"reason"`
			}
			if err := json.Unmarshal(input, &toolInput); err != nil {
				return nil, fmt.Errorf("invalid handoffTool input: %v", err)
			}

			// Emit the special key that will be parsed by the system
			key := fmt.Sprintf("@@HANDOFF:%s:%s:%s@@", toolInput.Agent, toolInput.Input, toolInput.Reason)
			
			// Create the response struct and marshal it properly to handle escaping
			response := struct {
				HandoffKey string `json:"handoffKey"`
			}{
				HandoffKey: key,
			}
			
			result, err := json.Marshal(response)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal handoffTool response: %v", err)
			}
			
			return json.RawMessage(result), nil
		},
		InputSchema:  inputSchema,
		OutputSchema: `{"type": "object", "properties": {"handoffKey": {"type": "string"}}}`,
	}
}

// TODO: Implement additional built-in tools:
// - scrape tool (crawl4ai scrape of pages)
// - perplexity search type research tools
// - grok research agent tool
// - realtime data pulling (twitter, weather, etc.)
// - language focused tools (get words, get sentences for word, search grammar books)