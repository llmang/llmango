package llmangoagents

import (
	"testing"
)

// TestGetTestConfig verifies that the test configuration is valid
func TestGetTestConfig(t *testing.T) {
	config := GetTestConfig()

	// Verify we have the expected agent
	if len(config.Agents) != 1 {
		t.Fatalf("Expected 1 agent, got %d", len(config.Agents))
	}

	agent := config.Agents[0]
	if agent.Name != "Test Agent" {
		t.Errorf("Expected agent name 'Test Agent', got '%s'", agent.Name)
	}

	if agent.Model != "anthropic/claude-3-haiku" {
		t.Errorf("Expected model 'anthropic/claude-3-haiku', got '%s'", agent.Model)
	}

	if agent.SystemMessage == "" {
		t.Error("Expected non-empty system message")
	}

	// Verify we have the expected workflow
	if len(config.Workflows) != 1 {
		t.Fatalf("Expected 1 workflow, got %d", len(config.Workflows))
	}

	workflow := config.Workflows[0]
	if workflow.UID != "test_workflow" {
		t.Errorf("Expected workflow UID 'test_workflow', got '%s'", workflow.UID)
	}

	if workflow.Name != "test_workflow" {
		t.Errorf("Expected workflow name 'test_workflow', got '%s'", workflow.Name)
	}

	// Verify workflow limits
	if workflow.Options.MaxTime != 60 {
		t.Errorf("Expected MaxTime 60, got %d", workflow.Options.MaxTime)
	}

	if workflow.Options.MaxSteps != 1 {
		t.Errorf("Expected MaxSteps 1, got %d", workflow.Options.MaxSteps)
	}

	if workflow.Options.MaxSpend != 10 {
		t.Errorf("Expected MaxSpend 10, got %d", workflow.Options.MaxSpend)
	}

	// Verify workflow has one step
	if len(workflow.Steps) != 1 {
		t.Fatalf("Expected 1 workflow step, got %d", len(workflow.Steps))
	}

	step := workflow.Steps[0]
	if step.UID != "test_step" {
		t.Errorf("Expected step UID 'test_step', got '%s'", step.UID)
	}

	if step.Agent != "test_agent" {
		t.Errorf("Expected step agent 'test_agent', got '%s'", step.Agent)
	}

	if step.AllowHandoffs != false {
		t.Errorf("Expected AllowHandoffs false, got %t", step.AllowHandoffs)
	}
}

// TestGetTestConfigJSON verifies that the test configuration can be serialized to JSON
func TestGetTestConfigJSON(t *testing.T) {
	jsonData, err := GetTestConfigJSON()
	if err != nil {
		t.Fatalf("Failed to get test config JSON: %v", err)
	}

	if len(jsonData) == 0 {
		t.Error("Expected non-empty JSON data")
	}

	// Verify it's valid JSON by checking it contains expected fields
	jsonStr := string(jsonData)
	expectedFields := []string{
		"\"agents\"",
		"\"workflows\"",
		"\"test_agent\"",
		"\"test_workflow\"",
		"\"anthropic/claude-3-haiku\"",
	}

	for _, field := range expectedFields {
		if !contains(jsonStr, field) {
			t.Errorf("Expected JSON to contain %s", field)
		}
	}
}

// TestCreateAgentSystemManager verifies that we can create an agent system manager from test config
func TestCreateAgentSystemManager(t *testing.T) {
	config := GetTestConfig()

	asm, err := CreateAgentSystemManager(config)
	if err != nil {
		t.Fatalf("Failed to create agent system manager: %v", err)
	}

	if asm == nil {
		t.Fatal("Expected non-nil agent system manager")
	}

	// Verify the system has our test components
	if len(asm.Agents) != 1 {
		t.Errorf("Expected 1 agent in system, got %d", len(asm.Agents))
	}

	if len(asm.Workflows) != 1 {
		t.Errorf("Expected 1 workflow in system, got %d", len(asm.Workflows))
	}

	if len(asm.Tools) != 0 {
		t.Errorf("Expected 0 tools in system, got %d", len(asm.Tools))
	}
}

// TestAgentLookup verifies that we can look up agents by name
func TestAgentLookup(t *testing.T) {
	asm, err := CreateTestSystemManager()
	if err != nil {
		t.Fatalf("Failed to create test system manager: %v", err)
	}

	// Test successful lookup
	agent, err := asm.GetAgent("test_agent")
	if err != nil {
		t.Fatalf("Failed to get test agent: %v", err)
	}

	if agent == nil {
		t.Fatal("Expected non-nil agent")
	}

	if agent.Name != "Test Agent" {
		t.Errorf("Expected agent name 'Test Agent', got '%s'", agent.Name)
	}

	// Test failed lookup
	_, err = asm.GetAgent("nonexistent_agent")
	if err == nil {
		t.Error("Expected error when looking up nonexistent agent")
	}
}

// TestWorkflowLookup verifies that we can look up workflows by UID
func TestWorkflowLookup(t *testing.T) {
	asm, err := CreateTestSystemManager()
	if err != nil {
		t.Fatalf("Failed to create test system manager: %v", err)
	}

	// Test successful lookup
	workflow, err := asm.GetWorkflow("test_workflow")
	if err != nil {
		t.Fatalf("Failed to get test workflow: %v", err)
	}

	if workflow == nil {
		t.Fatal("Expected non-nil workflow")
	}

	if workflow.UID != "test_workflow" {
		t.Errorf("Expected workflow UID 'test_workflow', got '%s'", workflow.UID)
	}

	// Test failed lookup
	_, err = asm.GetWorkflow("nonexistent_workflow")
	if err == nil {
		t.Error("Expected error when looking up nonexistent workflow")
	}
}

// TestSystemValidation verifies that the system validation works correctly
func TestSystemValidation(t *testing.T) {
	// Test valid configuration
	validConfig := GetTestConfig()
	_, err := CreateAgentSystemManager(validConfig)
	if err != nil {
		t.Errorf("Expected valid config to pass validation, got error: %v", err)
	}

	// Test invalid configuration - workflow referencing nonexistent agent
	invalidConfig := GetTestConfig()
	invalidConfig.Workflows[0].Steps[0].Agent = "nonexistent_agent"
	
	_, err = CreateAgentSystemManager(invalidConfig)
	if err == nil {
		t.Error("Expected invalid config to fail validation")
	}
}

// TestMockOpenRouter verifies that the mock OpenRouter works correctly
func TestMockOpenRouter(t *testing.T) {
	mock := NewMockOpenRouter()

	// Test default response
	if mock.DefaultResponse == "" {
		t.Error("Expected non-empty default response")
	}

	// Test setting custom response
	customResponse := "Custom test response"
	mock.SetDefaultResponse(customResponse)
	if mock.DefaultResponse != customResponse {
		t.Errorf("Expected default response '%s', got '%s'", customResponse, mock.DefaultResponse)
	}

	// Test setting model-specific response
	modelResponse := "Model-specific response"
	mock.SetResponse("test-model", modelResponse)
	if mock.ResponseMap["test-model"] != modelResponse {
		t.Errorf("Expected model response '%s', got '%s'", modelResponse, mock.ResponseMap["test-model"])
	}

	// Test error configuration
	mock.SetError(true, "Test error")
	if !mock.ShouldError {
		t.Error("Expected ShouldError to be true")
	}
	if mock.ErrorMessage != "Test error" {
		t.Errorf("Expected error message 'Test error', got '%s'", mock.ErrorMessage)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		 containsAt(s, substr, 1)))
}

func containsAt(s, substr string, start int) bool {
	if start >= len(s) {
		return false
	}
	if start+len(substr) <= len(s) && s[start:start+len(substr)] == substr {
		return true
	}
	return containsAt(s, substr, start+1)
}