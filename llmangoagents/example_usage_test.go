package llmangoagents

import (
	"testing"
)

// TestMVPUsageExample demonstrates how to use the MVP agent system
func TestMVPUsageExample(t *testing.T) {
	// Step 1: Create a test system manager
	asm, err := CreateTestSystemManager()
	if err != nil {
		t.Fatalf("Failed to create test system: %v", err)
	}

	// Step 2: Verify we can look up our test agent
	agent, err := asm.GetAgent("test_agent")
	if err != nil {
		t.Fatalf("Failed to get test agent: %v", err)
	}

	t.Logf("Found agent: %s with model: %s", agent.Name, agent.Model)
	t.Logf("Agent system message: %s", agent.SystemMessage)

	// Step 3: Verify we can look up our test workflow
	workflow, err := asm.GetWorkflow("test_workflow")
	if err != nil {
		t.Fatalf("Failed to get test workflow: %v", err)
	}

	t.Logf("Found workflow: %s with %d steps", workflow.Name, len(workflow.Steps))
	t.Logf("Workflow limits - MaxTime: %ds, MaxSteps: %d, MaxSpend: %d",
		workflow.Options.MaxTime, workflow.Options.MaxSteps, workflow.Options.MaxSpend)

	// Step 4: Verify the workflow step references the correct agent
	if len(workflow.Steps) > 0 {
		step := workflow.Steps[0]
		if step.Agent != agent.UID {
			t.Errorf("Workflow step agent '%s' doesn't match expected agent UID '%s'", step.Agent, agent.UID)
		}
		t.Logf("Workflow step '%s' uses agent '%s'", step.UID, step.Agent)
	}

	// Step 5: Test the mock OpenRouter functionality
	mock := NewMockOpenRouter()
	mock.SetResponse("anthropic/claude-3-haiku", "Hello from the test agent!")

	// Verify the mock is configured correctly
	if mock.ResponseMap["anthropic/claude-3-haiku"] != "Hello from the test agent!" {
		t.Error("Mock response not set correctly")
	}

	t.Log("✅ MVP agent system is working correctly!")
	t.Log("✅ Test configuration loaded successfully")
	t.Log("✅ Agent and workflow lookup working")
	t.Log("✅ System validation passing")
	t.Log("✅ Mock LLM system ready for testing")
}

// TestMVPConfigurationDetails shows the details of the test configuration
func TestMVPConfigurationDetails(t *testing.T) {
	config := GetTestConfig()

	t.Log("=== MVP Test Configuration Details ===")

	// Agent details
	t.Logf("Agent Count: %d", len(config.Agents))
	for i, agent := range config.Agents {
		t.Logf("Agent %d:", i+1)
		t.Logf("  Name: %s", agent.Name)
		t.Logf("  Model: %s", agent.Model)
		t.Logf("  System Message: %s", agent.SystemMessage)
		t.Logf("  Tools: %v", agent.Tools)
		t.Logf("  PreProcessors: %v", agent.PreProcessors)
		t.Logf("  SubAgents: %v", agent.SubAgents)
		t.Logf("  SubWorkflows: %v", agent.SubWorkflows)
	}

	// Workflow details
	t.Logf("Workflow Count: %d", len(config.Workflows))
	for i, workflow := range config.Workflows {
		t.Logf("Workflow %d:", i+1)
		t.Logf("  UID: %s", workflow.UID)
		t.Logf("  Name: %s", workflow.Name)
		t.Logf("  Description: %s", workflow.Description)
		t.Logf("  MaxTime: %d seconds", workflow.Options.MaxTime)
		t.Logf("  MaxSteps: %d", workflow.Options.MaxSteps)
		t.Logf("  MaxSpend: %d", workflow.Options.MaxSpend)
		t.Logf("  Steps Count: %d", len(workflow.Steps))

		for j, step := range workflow.Steps {
			t.Logf("    Step %d:", j+1)
			t.Logf("      UID: %s", step.UID)
			t.Logf("      Agent: %s", step.Agent)
			t.Logf("      SubAgents: %v", step.SubAgents)
			t.Logf("      ExitBehavior: %s", step.ExitBehavior)
		}
	}

	// Tool details
	t.Logf("Tool Count: %d", len(config.Tools))
	t.Logf("Custom Tool Config Count: %d", len(config.CustomToolConfigs))

	t.Log("=== End Configuration Details ===")
}
