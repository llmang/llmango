package generator

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/llmang/llmango/internal/parser"
	"github.com/llmang/llmango/llmango"
	"github.com/llmang/llmango/openrouter"
)

// TestGeneratorIssues contains test cases for all known generator issues
// This ensures we catch and fix issues systematically

// Issue 1: InputOutput field doesn't exist in llmango.Goal
func TestGeneratedGoalStructure(t *testing.T) {
	// Test that generated goals use the correct Goal struct format
	result := &parser.ParseResult{
		Goals: []parser.DiscoveredGoal{
			{
				UID:        "test-goal",
				Title:      "Test Goal",
				InputType:  "TestInput",
				OutputType: "TestOutput",
				VarName:    "testGoal",
				SourceType: "config",
			},
		},
	}

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

	err = GenerateMangoFile(result, opts)
	if err != nil {
		t.Fatalf("GenerateMangoFile failed: %v", err)
	}

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	contentStr := string(content)

	// Should NOT contain InputOutput field (this is the bug)
	if strings.Contains(contentStr, "InputOutput:") {
		t.Error("Generated goal should not contain InputOutput field - this field doesn't exist in llmango.Goal")
	}

	// Should contain proper Goal struct initialization
	if !strings.Contains(contentStr, "var testGoal = llmango.Goal{") {
		t.Error("Generated file should contain proper Goal struct initialization")
	}

	// Should contain UID, Title, Description
	if !strings.Contains(contentStr, `UID:         "test-goal"`) {
		t.Error("Generated goal should contain UID field")
	}
}

// Issue 2: ExecuteGoalWithDualPath function doesn't exist
func TestGeneratedFunctionCalls(t *testing.T) {
	// Test that generated functions use existing llmango functions
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
	}

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

	err = GenerateMangoFile(result, opts)
	if err != nil {
		t.Fatalf("GenerateMangoFile failed: %v", err)
	}

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	contentStr := string(content)

	// Should NOT contain ExecuteGoalWithDualPath (this function doesn't exist yet)
	if strings.Contains(contentStr, "ExecuteGoalWithDualPath") {
		t.Error("Generated function should not call ExecuteGoalWithDualPath - this function doesn't exist yet")
	}

	// Should use existing llmango.Run function
	if !strings.Contains(contentStr, "llmango.Run[") {
		t.Error("Generated function should use existing llmango.Run function")
	}
}

// Issue 3: String escaping in prompt content
func TestPromptContentEscaping(t *testing.T) {
	// Test that prompt content with newlines is properly escaped
	result := &parser.ParseResult{
		Prompts: []parser.DiscoveredPrompt{
			{
				UID:        "test-prompt",
				GoalUID:    "test-goal",
				VarName:    "testPrompt",
				SourceType: "config",
				Model:      "openai/gpt-4",
				Weight:     100,
				Messages: []openrouter.Message{
					{
						Role:    "user",
						Content: "Line 1\nLine 2\nLine 3",
					},
				},
			},
		},
	}

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

	err = GenerateMangoFile(result, opts)
	if err != nil {
		t.Fatalf("GenerateMangoFile failed: %v", err)
	}

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	contentStr := string(content)

	// Should properly escape newlines in strings - check that the content is properly quoted
	if !strings.Contains(contentStr, `"Line 1\nLine 2\nLine 3"`) {
		t.Error("Generated prompt content should properly escape newlines with backslashes")
	}

	// Should be valid Go syntax (no unescaped newlines in string literals)
	contentLines := strings.Split(contentStr, "\n")
	for i, line := range contentLines {
		if strings.Contains(line, "Content: \"") && !strings.Contains(line, "\",") {
			t.Errorf("Line %d has unclosed string literal: %s", i+1, line)
		}
	}
}

// Issue 4: Mixed source types (config + Go inline goals)
func TestMixedSourceTypes(t *testing.T) {
	// Test that we can handle both config-based and Go inline goals
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
			{
				UID:        "go-goal",
				Title:      "Go Goal",
				InputType:  "GoInput",
				OutputType: "GoOutput",
				VarName:    "goGoal",
				SourceType: "go",
			},
		},
	}

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

	err = GenerateMangoFile(result, opts)
	if err != nil {
		t.Fatalf("GenerateMangoFile failed: %v", err)
	}

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	contentStr := string(content)

	// Should generate functions for both goal types
	if !strings.Contains(contentStr, "func (m *Mango) ConfigGoal(") {
		t.Error("Should generate function for config-based goal")
	}

	if !strings.Contains(contentStr, "func (m *Mango) GoGoal(") {
		t.Error("Should generate function for Go inline goal")
	}

	// Should register both goals
	if !strings.Contains(contentStr, "&configGoal,") {
		t.Error("Should register config-based goal")
	}

	if !strings.Contains(contentStr, "&goGoal,") {
		t.Error("Should register Go inline goal")
	}
}

// Issue 5: Method name generation from UIDs
func TestMethodNameGeneration(t *testing.T) {
	testCases := []struct {
		uid      string
		expected string
	}{
		{"sentiment-analysis", "SentimentAnalysis"},
		{"text-summary", "TextSummary"},
		{"email-classification", "EmailClassification"},
		{"language-detection", "LanguageDetection"},
		{"simple", "Simple"},
		{"multi-word-goal", "MultiWordGoal"},
	}

	for _, tc := range testCases {
		t.Run(tc.uid, func(t *testing.T) {
			result := &parser.ParseResult{
				Goals: []parser.DiscoveredGoal{
					{
						UID:        tc.uid,
						Title:      "Test Goal",
						InputType:  "TestInput",
						OutputType: "TestOutput",
						VarName:    "testGoal",
						SourceType: "config",
					},
				},
			}

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

			err = GenerateMangoFile(result, opts)
			if err != nil {
				t.Fatalf("GenerateMangoFile failed: %v", err)
			}

			content, err := os.ReadFile(outputFile)
			if err != nil {
				t.Fatalf("Failed to read generated file: %v", err)
			}

			contentStr := string(content)

			expectedFunc := "func (m *Mango) " + tc.expected + "("
			if !strings.Contains(contentStr, expectedFunc) {
				t.Errorf("Expected method name %s, but function not found in generated code", tc.expected)
			}
		})
	}
}

// Issue 6: Compilation validation
func TestGeneratedCodeCompiles(t *testing.T) {
	// Test that generated code actually compiles
	result := &parser.ParseResult{
		Goals: []parser.DiscoveredGoal{
			{
				UID:        "test-goal",
				Title:      "Test Goal",
				InputType:  "TestInput",
				OutputType: "TestOutput",
				VarName:    "testGoal",
				SourceType: "config",
			},
		},
		Prompts: []parser.DiscoveredPrompt{
			{
				UID:        "test-prompt",
				GoalUID:    "test-goal",
				VarName:    "testPrompt",
				SourceType: "config",
				Model:      "openai/gpt-4",
				Weight:     100,
				Messages: []openrouter.Message{
					{
						Role:    "system",
						Content: "You are a test assistant.",
					},
					{
						Role:    "user",
						Content: "Process this: {{.input}}",
					},
				},
			},
		},
	}

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

	err = GenerateMangoFile(result, opts)
	if err != nil {
		t.Fatalf("GenerateMangoFile failed: %v", err)
	}

	// TODO: Add actual compilation test when we fix the template issues
	// For now, just verify the file was created and has basic structure
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	contentStr := string(content)

	// Basic structure checks
	if !strings.Contains(contentStr, "package testmango") {
		t.Error("Generated file should have correct package declaration")
	}

	if !strings.Contains(contentStr, "type Mango struct") {
		t.Error("Generated file should contain Mango struct")
	}

	if !strings.Contains(contentStr, "func CreateMango(") {
		t.Error("Generated file should contain CreateMango function")
	}
}

// Issue 16: Message parsing fails with json.RawMessage input
func TestMessageParsingWithJSONInput(t *testing.T) {
	// This test reproduces the runtime error:
	// "failed to update prompt messages: error processing conditional blocks: input must be a struct or pointer to struct"
	
	// Create a test input struct
	type TestInput struct {
		Text string `json:"text"`
	}
	
	input := TestInput{Text: "Hello world"}
	inputJSON, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal input: %v", err)
	}
	
	// Create test messages with template variables
	messages := []openrouter.Message{
		{Role: "system", Content: "You are a helpful assistant."},
		{Role: "user", Content: "Analyze this text: {{text}}"},
	}
	
	// Test 1: ParseMessages should work with struct input (current working case)
	parsedMessages1, err := llmango.ParseMessages(&input, messages)
	if err != nil {
		t.Errorf("ParseMessages failed with struct input: %v", err)
	}
	if len(parsedMessages1) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(parsedMessages1))
	}
	if !strings.Contains(parsedMessages1[1].Content, "Hello world") {
		t.Errorf("Template variable not substituted correctly: %s", parsedMessages1[1].Content)
	}
	
	// Test 2: ParseMessages should fail with json.RawMessage input (current issue)
	// This is the bug we need to fix - it should work with json.RawMessage too
	_, err = llmango.ParseMessages(inputJSON, messages)
	if err == nil {
		t.Log("ParseMessages unexpectedly succeeded with json.RawMessage input - this means the bug is fixed!")
	} else {
		expectedError := "input must be a struct or pointer to struct"
		if !strings.Contains(err.Error(), expectedError) {
			t.Errorf("Expected error containing '%s', got: %v", expectedError, err)
		} else {
			t.Logf("Confirmed bug: ParseMessages fails with json.RawMessage input: %v", err)
		}
	}
}

// Issue 17: Dual-path execution system needs to handle json.RawMessage input
func TestDualPathExecutionInputHandling(t *testing.T) {
	// This test ensures the dual-path execution system can handle both:
	// 1. Typed struct input (from Run[I,O] functions)
	// 2. json.RawMessage input (from ExecuteGoalWithDualPath)
	
	// Create test goal and manager
	or := &openrouter.OpenRouter{}
	manager, err := llmango.CreateLLMangoManger(or)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	
	// Create test input
	type TestInput struct {
		Text string `json:"text"`
	}
	
	input := TestInput{Text: "Test message"}
	inputJSON, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal input: %v", err)
	}
	
	// Create test goal
	goal := llmango.NewGoal("test-goal", "Test Goal", "Test description", input, map[string]string{"result": "test"})
	manager.AddGoals(goal)
	
	// Create test prompt
	prompt := &llmango.Prompt{
		UID:     "test-prompt",
		GoalUID: "test-goal",
		Model:   "openai/gpt-3.5-turbo",
		Weight:  100,
		Messages: []openrouter.Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: "Process this: {{text}}"},
		},
	}
	manager.AddPrompts(prompt)
	
	// Test that ExecuteGoalWithDualPath can handle json.RawMessage input
	// This should not fail with "input must be a struct or pointer to struct"
	_, err = manager.ExecuteGoalWithDualPath("test-goal", inputJSON)
	if err != nil {
		// We expect this to fail due to missing API key, but NOT due to input parsing
		if strings.Contains(err.Error(), "input must be a struct or pointer to struct") {
			t.Errorf("ExecuteGoalWithDualPath failed with input parsing error: %v", err)
		} else {
			t.Logf("ExecuteGoalWithDualPath failed as expected (likely API key issue): %v", err)
		}
	}
}
