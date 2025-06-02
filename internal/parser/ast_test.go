package parser

import (
	"testing"
)

func TestParseGoFiles(t *testing.T) {
	tests := []struct {
		name          string
		dir           string
		expectedGoals int
		expectedPromp int
		expectError   bool
	}{
		{
			name:          "valid go only project",
			dir:           "../testdata/valid_projects/go_only",
			expectedGoals: 1,
			expectedPromp: 1,
			expectError:   false,
		},
		{
			name:          "config only project (no go files with goals)",
			dir:           "../testdata/valid_projects/config_only",
			expectedGoals: 0,
			expectedPromp: 0,
			expectError:   false,
		},
		{
			name:        "invalid syntax project",
			dir:         "../testdata/invalid_projects/syntax_errors",
			expectError: true,
		},
		{
			name:          "non-existent directory",
			dir:           "../testdata/non_existent",
			expectedGoals: 0,
			expectedPromp: 0,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseGoFiles(tt.dir)

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

func TestGenerateMethodName(t *testing.T) {
	tests := []struct {
		goalUID  string
		expected string
	}{
		{
			goalUID:  "test-goal",
			expected: "TestGoal",
		},
		{
			goalUID:  "user-chatbot",
			expected: "UserChatbot",
		},
		{
			goalUID:  "generate_summary",
			expected: "GenerateSummary",
		},
		{
			goalUID:  "complex-goal-name-with-many-parts",
			expected: "ComplexGoalNameWithManyParts",
		},
		{
			goalUID:  "",
			expected: "UnnamedGoal",
		},
		{
			goalUID:  "123-invalid-start",
			expected: "123InvalidStart",
		},
	}

	for _, tt := range tests {
		t.Run(tt.goalUID, func(t *testing.T) {
			result := GenerateMethodName(tt.goalUID)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestParseGoFilesValidation(t *testing.T) {
	// Test with the valid go_only project
	result, err := ParseGoFiles("../testdata/valid_projects/go_only")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Goals) != 1 {
		t.Fatalf("expected 1 goal, got %d", len(result.Goals))
	}

	goal := result.Goals[0]
	if goal.UID != "test-goal" {
		t.Errorf("expected goal UID 'test-goal', got '%s'", goal.UID)
	}

	if goal.Title != "Test Goal" {
		t.Errorf("expected goal title 'Test Goal', got '%s'", goal.Title)
	}

	if goal.InputType != "TestInput" {
		t.Errorf("expected input type 'TestInput', got '%s'", goal.InputType)
	}

	if goal.OutputType != "TestOutput" {
		t.Errorf("expected output type 'TestOutput', got '%s'", goal.OutputType)
	}

	if goal.SourceType != "go" {
		t.Errorf("expected source type 'go', got '%s'", goal.SourceType)
	}

	if len(result.Prompts) != 1 {
		t.Fatalf("expected 1 prompt, got %d", len(result.Prompts))
	}

	prompt := result.Prompts[0]
	if prompt.UID != "test-prompt" {
		t.Errorf("expected prompt UID 'test-prompt', got '%s'", prompt.UID)
	}

	if prompt.GoalUID != "test-goal" {
		t.Errorf("expected prompt goal UID 'test-goal', got '%s'", prompt.GoalUID)
	}

	if prompt.Model != "openai/gpt-4" {
		t.Errorf("expected model 'openai/gpt-4', got '%s'", prompt.Model)
	}

	if prompt.SourceType != "go" {
		t.Errorf("expected source type 'go', got '%s'", prompt.SourceType)
	}
}