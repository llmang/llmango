package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/llmang/llmango/internal/parser"
)

func TestGenerateMangoFile(t *testing.T) {
	// Create test data
	result := &parser.ParseResult{
		Goals: []parser.DiscoveredGoal{
			{
				UID:        "test-goal",
				Title:      "Test Goal",
				InputType:  "TestInput",
				OutputType: "TestOutput",
				VarName:    "testGoal",
				SourceType: "go",
			},
		},
		Prompts: []parser.DiscoveredPrompt{
			{
				UID:        "test-prompt",
				GoalUID:    "test-goal",
				VarName:    "testPrompt",
				SourceType: "go",
			},
		},
	}

	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "llmango_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	outputFile := filepath.Join(tmpDir, "mango.go")
	opts := &parser.GenerateOptions{
		OutputFile:  outputFile,
		PackageName: "testmango",
	}

	// Generate the file
	err = GenerateMangoFile(result, opts)
	if err != nil {
		t.Fatalf("GenerateMangoFile failed: %v", err)
	}

	// Read the generated file
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	contentStr := string(content)

	// Verify package declaration
	if !strings.Contains(contentStr, "package testmango") {
		t.Error("Generated file should contain correct package declaration")
	}

	// Verify imports
	if !strings.Contains(contentStr, `"github.com/llmang/llmango/llmango"`) {
		t.Error("Generated file should import llmango package")
	}

	if !strings.Contains(contentStr, `"github.com/llmang/llmango/openrouter"`) {
		t.Error("Generated file should import openrouter package")
	}

	// Verify Mango struct
	if !strings.Contains(contentStr, "type Mango struct") {
		t.Error("Generated file should contain Mango struct")
	}

	// Verify CreateMango function
	if !strings.Contains(contentStr, "func CreateMango(or *openrouter.OpenRouter) (*Mango, error)") {
		t.Error("Generated file should contain CreateMango function")
	}

	// Verify goal registration
	if !strings.Contains(contentStr, "&testGoal,") {
		t.Error("Generated file should register testGoal")
	}

	// Verify prompt registration
	if !strings.Contains(contentStr, "&testPrompt,") {
		t.Error("Generated file should register testPrompt")
	}

	// Verify generated method
	if !strings.Contains(contentStr, "func (m *Mango) TestGoal(input *TestInput) (*TestOutput, error)") {
		t.Error("Generated file should contain TestGoal method")
	}

	// Verify generated raw method
	if !strings.Contains(contentStr, "func (m *Mango) TestGoalRaw(input *TestInput) (*TestOutput, *openrouter.NonStreamingChatResponse, error)") {
		t.Error("Generated file should contain TestGoalRaw method")
	}
}

func TestGenerateUniqueMethodName(t *testing.T) {
	tests := []struct {
		name     string
		goalUID  string
		existing map[string]bool
		expected string
	}{
		{
			name:     "unique name",
			goalUID:  "test-goal",
			existing: map[string]bool{},
			expected: "TestGoal",
		},
		{
			name:     "conflicting name",
			goalUID:  "test-goal",
			existing: map[string]bool{"TestGoal": true},
			expected: "TestGoal1",
		},
		{
			name:     "multiple conflicts",
			goalUID:  "test-goal",
			existing: map[string]bool{"TestGoal": true, "TestGoal1": true, "TestGoal2": true},
			expected: "TestGoal3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateUniqueMethodName(tt.goalUID, tt.existing)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestValidateGoalPromptRelationships(t *testing.T) {
	tests := []struct {
		name          string
		result        *parser.ParseResult
		expectedErrors int
	}{
		{
			name: "valid relationships",
			result: &parser.ParseResult{
				Goals: []parser.DiscoveredGoal{
					{UID: "goal1"},
					{UID: "goal2"},
				},
				Prompts: []parser.DiscoveredPrompt{
					{UID: "prompt1", GoalUID: "goal1"},
					{UID: "prompt2", GoalUID: "goal2"},
				},
			},
			expectedErrors: 0,
		},
		{
			name: "invalid relationships",
			result: &parser.ParseResult{
				Goals: []parser.DiscoveredGoal{
					{UID: "goal1"},
				},
				Prompts: []parser.DiscoveredPrompt{
					{UID: "prompt1", GoalUID: "goal1"},
					{UID: "prompt2", GoalUID: "nonexistent"},
				},
			},
			expectedErrors: 1,
		},
		{
			name: "multiple invalid relationships",
			result: &parser.ParseResult{
				Goals: []parser.DiscoveredGoal{
					{UID: "goal1"},
				},
				Prompts: []parser.DiscoveredPrompt{
					{UID: "prompt1", GoalUID: "nonexistent1"},
					{UID: "prompt2", GoalUID: "nonexistent2"},
				},
			},
			expectedErrors: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateGoalPromptRelationships(tt.result)
			if len(errors) != tt.expectedErrors {
				t.Errorf("expected %d errors, got %d", tt.expectedErrors, len(errors))
			}

			// Verify all errors are of type "error"
			for _, err := range errors {
				if err.Type != "error" {
					t.Errorf("expected error type 'error', got '%s'", err.Type)
				}
			}
		})
	}
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    `simple string`,
			expected: `simple string`,
		},
		{
			input:    `string with "quotes"`,
			expected: `string with \"quotes\"`,
		},
		{
			input:    `string with \backslash`,
			expected: `string with \\backslash`,
		},
		{
			input:    "string with\nnewline",
			expected: `string with\nnewline`,
		},
		{
			input:    "string with\ttab",
			expected: `string with\ttab`,
		},
		{
			input:    "string with\rcarriage return",
			expected: `string with\rcarriage return`,
		},
		{
			input:    `complex "string" with` + "\n" + `multiple` + "\t" + `special chars`,
			expected: `complex \"string\" with\nmultiple\tspecial chars`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := SanitizeString(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGenerateMangoFileWithConfigGoals(t *testing.T) {
	// Test with config-generated goals and prompts
	result := &parser.ParseResult{
		Goals: []parser.DiscoveredGoal{
			{
				UID:        "config-goal",
				Title:      "Config Goal",
				InputType:  "ConfigInput",
				OutputType: "ConfigOutput",
				VarName:    "configGoal",
				SourceType: "config",
			},
		},
		Prompts: []parser.DiscoveredPrompt{
			{
				UID:        "config-prompt",
				GoalUID:    "config-goal",
				VarName:    "configPrompt",
				SourceType: "config",
				Model:      "openai/gpt-4",
				Weight:     100,
			},
		},
	}

	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "llmango_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	outputFile := filepath.Join(tmpDir, "mango.go")
	opts := &parser.GenerateOptions{
		OutputFile:  outputFile,
		PackageName: "testmango",
	}

	// Generate the file
	err = GenerateMangoFile(result, opts)
	if err != nil {
		t.Fatalf("GenerateMangoFile failed: %v", err)
	}

	// Read the generated file
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	contentStr := string(content)

	// Verify config-generated variables are included
	if !strings.Contains(contentStr, "// Config-generated goals and prompts") {
		t.Error("Generated file should contain config-generated section")
	}

	if !strings.Contains(contentStr, "var configGoal = llmango.Goal{") {
		t.Error("Generated file should contain config goal variable")
	}

	if !strings.Contains(contentStr, "var configPrompt = llmango.Prompt{") {
		t.Error("Generated file should contain config prompt variable")
	}
}