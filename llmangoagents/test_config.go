package llmangoagents

import (
	"encoding/json"
)

// GetTestConfig returns a minimal test configuration for testing the agent system
func GetTestConfig() SystemInputList {
	// Create a simple test agent
	testAgent := Agent{
		UID:           "test_agent",
		Name:          "Test Agent",
		SystemMessage: "You are a helpful assistant. Respond concisely to user questions.",
		Model:         "anthropic/claude-3-haiku",
		Parameters:    `{"temperature": 0.7}`,
		Tools:         []string{}, // No tools for simplicity
		PreProcessors: []string{}, // No preprocessors for simplicity
		SubAgents:     []string{}, // No sub-agents for simplicity
		SubWorkflows:  []string{}, // No sub-workflows for simplicity
	}

	// Create a simple workflow step
	testStep := &WorkflowStep{
		UID:          "test_step",
		Agent:        "test_agent", // This should match the agent UID
		SubAgents:    []string{},
		ExitBehavior: "default",
	}

	// Create a simple workflow with limits
	testWorkflow := Workflow{
		UID:         "test_workflow",
		Name:        "test_workflow",
		Description: "A simple test workflow with one step",
		Options: WorkflowLimits{
			MaxTime:  60, // 60 seconds max time
			MaxSteps: 1,  // 1 max step
			MaxSpend: 10, // 10 max spend
		},
		Steps: []*WorkflowStep{testStep},
	}

	// Return the complete system input list
	return SystemInputList{
		Tools:             []Tool{}, // No tools for simplicity
		CustomToolConfigs: []HTTPToolBuilderConfig{}, // No custom tools
		Agents:            []Agent{testAgent},
		Workflows:         []Workflow{testWorkflow},
	}
}

// GetTestConfigJSON returns the test configuration as JSON bytes
func GetTestConfigJSON() ([]byte, error) {
	config := GetTestConfig()
	
	jsonConfig := JSONConfig{
		CustomToolConfigs: config.CustomToolConfigs,
		Agents:            config.Agents,
		Workflows:         config.Workflows,
	}
	
	return json.MarshalIndent(jsonConfig, "", "  ")
}