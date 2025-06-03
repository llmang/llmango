package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseConfigFiles(t *testing.T) {
	tests := []struct {
		name          string
		dir           string
		expectedGoals int
		expectedPromp int
		expectError   bool
	}{
		{
			name:          "config only project",
			dir:           "../testdata/valid_projects/config_only",
			expectedGoals: 1,
			expectedPromp: 1,
			expectError:   false,
		},
		{
			name:          "go only project (no config)",
			dir:           "../testdata/valid_projects/go_only",
			expectedGoals: 0,
			expectedPromp: 0,
			expectError:   false,
		},
		{
			name:          "non-existent directory",
			dir:           "../testdata/non_existent",
			expectedGoals: 0,
			expectedPromp: 0,
			expectError:   false, // Should not error, just return empty results
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseConfigFiles(tt.dir)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(result.Goals) != tt.expectedGoals {
				t.Errorf("expected %d goals, got %d", tt.expectedGoals, len(result.Goals))
			}

			if len(result.Prompts) != tt.expectedPromp {
				t.Errorf("expected %d prompts, got %d", tt.expectedPromp, len(result.Prompts))
			}
		})
	}
}

func TestParseConfigFilesValidation(t *testing.T) {
	// Test with the valid config_only project
	result, err := ParseConfigFiles("../testdata/valid_projects/config_only")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Goals) != 1 {
		t.Fatalf("expected 1 goal, got %d", len(result.Goals))
	}

	goal := result.Goals[0]
	if goal.UID != "config-goal" {
		t.Errorf("expected goal UID 'config-goal', got '%s'", goal.UID)
	}

	if goal.Title != "Config Goal" {
		t.Errorf("expected goal title 'Config Goal', got '%s'", goal.Title)
	}

	if goal.InputType != "ConfigInput" {
		t.Errorf("expected input type 'ConfigInput', got '%s'", goal.InputType)
	}

	if goal.OutputType != "ConfigOutput" {
		t.Errorf("expected output type 'ConfigOutput', got '%s'", goal.OutputType)
	}

	if goal.SourceType != "config" {
		t.Errorf("expected source type 'config', got '%s'", goal.SourceType)
	}

	if len(result.Prompts) != 1 {
		t.Fatalf("expected 1 prompt, got %d", len(result.Prompts))
	}

	prompt := result.Prompts[0]
	if prompt.UID != "config-prompt" {
		t.Errorf("expected prompt UID 'config-prompt', got '%s'", prompt.UID)
	}

	if prompt.GoalUID != "config-goal" {
		t.Errorf("expected prompt goal UID 'config-goal', got '%s'", prompt.GoalUID)
	}

	if prompt.Model != "openai/gpt-3.5-turbo" {
		t.Errorf("expected model 'openai/gpt-3.5-turbo', got '%s'", prompt.Model)
	}

	if prompt.Weight != 50 {
		t.Errorf("expected weight 50, got %d", prompt.Weight)
	}

	if prompt.SourceType != "config" {
		t.Errorf("expected source type 'config', got '%s'", prompt.SourceType)
	}
}

func TestGenerateVarName(t *testing.T) {
	tests := []struct {
		uid      string
		suffix   string
		expected string
	}{
		{
			uid:      "test-goal",
			suffix:   "Goal",
			expected: "testGoalGoal",
		},
		{
			uid:      "user_chatbot",
			suffix:   "Prompt",
			expected: "userChatbotPrompt",
		},
		{
			uid:      "complex-goal-name",
			suffix:   "Goal",
			expected: "complexGoalNameGoal",
		},
		{
			uid:      "",
			suffix:   "Goal",
			expected: "unnamedGoal",
		},
		{
			uid:      "single",
			suffix:   "Prompt",
			expected: "singlePrompt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.uid+"_"+tt.suffix, func(t *testing.T) {
			result := generateVarName(tt.uid, tt.suffix)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestMergeResults(t *testing.T) {
	// Create test results
	goResult := &ParseResult{
		Goals: []DiscoveredGoal{
			{UID: "go-goal", SourceType: "go", Title: "Go Goal"},
			{UID: "shared-goal", SourceType: "go", Title: "Go Shared Goal"},
		},
		Prompts: []DiscoveredPrompt{
			{UID: "go-prompt", SourceType: "go", GoalUID: "go-goal"},
		},
	}

	configResult := &ParseResult{
		Goals: []DiscoveredGoal{
			{UID: "config-goal", SourceType: "config", Title: "Config Goal"},
			{UID: "shared-goal", SourceType: "config", Title: "Config Shared Goal"},
		},
		Prompts: []DiscoveredPrompt{
			{UID: "config-prompt", SourceType: "config", GoalUID: "config-goal"},
		},
	}

	merged := MergeResults(goResult, configResult)

	// Should have 3 goals total (2 unique + 1 shared with Go taking priority)
	if len(merged.Goals) != 3 {
		t.Errorf("expected 3 goals, got %d", len(merged.Goals))
	}

	// Should have 2 prompts total
	if len(merged.Prompts) != 2 {
		t.Errorf("expected 2 prompts, got %d", len(merged.Prompts))
	}

	// Check that Go definition takes priority for shared goal
	var sharedGoal *DiscoveredGoal
	for _, goal := range merged.Goals {
		if goal.UID == "shared-goal" {
			sharedGoal = &goal
			break
		}
	}

	if sharedGoal == nil {
		t.Fatal("shared goal not found in merged results")
	}

	if sharedGoal.SourceType != "go" {
		t.Errorf("expected shared goal to have Go source type, got %s", sharedGoal.SourceType)
	}

	if sharedGoal.Title != "Go Shared Goal" {
		t.Errorf("expected shared goal to have Go title, got %s", sharedGoal.Title)
	}

	// Should have warnings about conflicts
	warningFound := false
	for _, err := range merged.Errors {
		if err.Type == "warning" && err.Message != "" {
			warningFound = true
			break
		}
	}

	if !warningFound {
		t.Error("expected warning about conflicting definitions")
	}
}

func TestInvalidConfigFile(t *testing.T) {
	// Create a temporary directory with invalid YAML
	tmpDir, err := os.MkdirTemp("", "llmango_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Write invalid YAML
	invalidYAML := `
goals:
  - uid: "test"
    title: "Test"
    invalid_yaml: [unclosed array
`
	yamlFile := filepath.Join(tmpDir, "llmango.yaml")
	if err := os.WriteFile(yamlFile, []byte(invalidYAML), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := ParseConfigFiles(tmpDir)
	if err != nil {
		t.Errorf("ParseConfigFiles should not return error, but got: %v", err)
	}

	// Should have errors in the result
	if len(result.Errors) == 0 {
		t.Error("expected parse errors for invalid YAML")
	}

	errorFound := false
	for _, parseErr := range result.Errors {
		if parseErr.Type == "error" {
			errorFound = true
			break
		}
	}

	if !errorFound {
		t.Error("expected error type in parse errors")
	}
}