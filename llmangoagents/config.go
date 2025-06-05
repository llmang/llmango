package llmangoagents

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Configuration types
type JSONConfig struct {
	CustomToolConfigs []HTTPToolBuilderConfig `json:"tools"`
	Agents            []Agent                 `json:"agents"`
	Workflows         []Workflow              `json:"workflows"`
}

// Configuration methods
func (asm *AgentSystemManager) AddFromWire() {
	//add a tool, agent, or workflow from "wire" over internet/from the frontend/from a service manager
	//takes a json object and creates based on that
	//TODO: Implement JSON unmarshaling and validation
	//TODO: Support incremental updates (add single tool/agent/workflow)
	//TODO: Validate against existing system state
}

func (asm *AgentSystemManager) AddFromConfig(configData []byte) error {
	//TODO: This is loading agent system from a configuration -- tools, agents, workflows
	//TODO: Implement JSON unmarshaling and validation
	//TODO: Process environment variable interpolation in GlobalConfig.KeyBank
	//TODO: Apply compiled system to this manager
	return fmt.Errorf("not implemented")
}

// SerializeToConfig exports current system state to JSON config
func (asm *AgentSystemManager) SerializeToConfig() ([]byte, error) {
	//TODO: Extract current tools, agents, workflows
	//TODO: Convert runtime instances back to serializable format
	//TODO: Handle sensitive data (API keys) appropriately
	//TODO: Generate metadata (timestamps, version)
	return nil, fmt.Errorf("not implemented")
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