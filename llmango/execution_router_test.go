package llmango

import (
	"encoding/json"
	"testing"

	"github.com/llmang/llmango/openrouter"
	"github.com/llmang/llmango/testhelpers"
)

func TestExecuteGoalWithDualPath(t *testing.T) {
	// Create a test manager
	manager, err := CreateLLMangoManger(nil)
	testhelpers.RequireNoError(t, err, "Failed to create manager")

	// Create test goals
	typedGoal := createTestTypedGoal()
	jsonGoal := createTestJSONGoal()

	// Add goals to manager
	manager.AddGoals(typedGoal, jsonGoal)

	// Create test prompts
	structuredPrompt := createTestPrompt("openai/gpt-4o", "structured-prompt")
	universalPrompt := createTestPrompt("anthropic/claude-3-sonnet", "universal-prompt")

	// Add prompts to manager
	manager.AddPrompts(structuredPrompt, universalPrompt)

	// Link prompts to goals
	typedGoal.PromptUIDs = []string{structuredPrompt.UID}
	jsonGoal.PromptUIDs = []string{universalPrompt.UID}

	tests := []struct {
		name                string
		goalUID             string
		expectedPathType    string
		expectedModelSupport bool
	}{
		{
			name:                "Typed goal with structured output model",
			goalUID:             typedGoal.UID,
			expectedPathType:    "structured",
			expectedModelSupport: true,
		},
		{
			name:                "JSON goal with universal compatibility model",
			goalUID:             jsonGoal.UID,
			expectedPathType:    "universal",
			expectedModelSupport: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get the goal and its prompt
			goal, exists := manager.Goals.Get(tt.goalUID)
			testhelpers.AssertTrue(t, exists, "Goal should exist")
			testhelpers.AssertNotEmpty(t, goal.PromptUIDs, "Goal should have prompts")

			prompt, exists := manager.Prompts.Get(goal.PromptUIDs[0])
			testhelpers.AssertTrue(t, exists, "Prompt should exist")

			// Check model capabilities
			supportsStructuredOutput := openrouter.SupportsStructuredOutput(prompt.Model)
			testhelpers.AssertEqual(t, tt.expectedModelSupport, supportsStructuredOutput,
				"Model support should match expected for %s", prompt.Model)

			// Note: We can't actually execute the goal without a real API key and network call
			// But we can test the path selection logic
			if supportsStructuredOutput {
				testhelpers.AssertEqual(t, "structured", tt.expectedPathType, "Should use structured path")
			} else {
				testhelpers.AssertEqual(t, "universal", tt.expectedPathType, "Should use universal path")
			}
		})
	}
}

func TestSelectPromptForGoal(t *testing.T) {
	manager, err := CreateLLMangoManger(nil)
	testhelpers.RequireNoError(t, err, "Failed to create manager")

	// Create test goal and prompts
	goal := createTestTypedGoal()
	prompt1 := createTestPrompt("openai/gpt-4o", "prompt-1")
	prompt2 := createTestPrompt("anthropic/claude-3-sonnet", "prompt-2")

	// Set up prompt weights
	prompt1.Weight = 50
	prompt2.Weight = 30

	// Add to manager
	manager.AddGoals(goal)
	manager.AddPrompts(prompt1, prompt2)

	// Link prompts to goal
	goal.PromptUIDs = []string{prompt1.UID, prompt2.UID}

	// Test prompt selection
	selectedPrompt, err := manager.selectPromptForGoal(goal)
	testhelpers.RequireNoError(t, err, "Should select a prompt successfully")
	testhelpers.AssertNotNil(t, selectedPrompt, "Selected prompt should not be nil")

	// Should select one of the valid prompts
	validUIDs := []string{prompt1.UID, prompt2.UID}
	found := false
	for _, uid := range validUIDs {
		if selectedPrompt.UID == uid {
			found = true
			break
		}
	}
	testhelpers.AssertTrue(t, found, "Selected prompt should be one of the valid prompts")
}

func TestSelectPromptForGoalNoValidPrompts(t *testing.T) {
	manager, err := CreateLLMangoManger(nil)
	testhelpers.RequireNoError(t, err, "Failed to create manager")

	// Create test goal with no prompts
	goal := createTestTypedGoal()
	goal.PromptUIDs = []string{"non-existent-prompt"}

	manager.AddGoals(goal)

	// Test prompt selection should fail
	_, err = manager.selectPromptForGoal(goal)
	testhelpers.AssertError(t, err, "Should fail when no valid prompts exist")
	testhelpers.AssertContains(t, err.Error(), "no valid prompts available", "Error should mention no valid prompts")
}

func TestInjectUniversalPrompt(t *testing.T) {
	manager, err := CreateLLMangoManger(nil)
	testhelpers.RequireNoError(t, err, "Failed to create manager")

	universalPrompt := "You must respond with valid JSON matching the schema."

	tests := []struct {
		name            string
		inputMessages   []openrouter.Message
		expectedCount   int
		expectedSystem  bool
		description     string
	}{
		{
			name: "Messages with existing system prompt",
			inputMessages: []openrouter.Message{
				{Role: "system", Content: "You are a helpful assistant."},
				{Role: "user", Content: "Hello"},
			},
			expectedCount:  2,
			expectedSystem: true,
			description:    "Should merge with existing system prompt",
		},
		{
			name: "Messages without system prompt",
			inputMessages: []openrouter.Message{
				{Role: "user", Content: "Hello"},
				{Role: "assistant", Content: "Hi there!"},
			},
			expectedCount:  3,
			expectedSystem: true,
			description:    "Should add system prompt as first message",
		},
		{
			name: "Empty messages",
			inputMessages: []openrouter.Message{},
			expectedCount:  1,
			expectedSystem: true,
			description:    "Should add system prompt to empty messages",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.injectUniversalPrompt(tt.inputMessages, universalPrompt)
			
			testhelpers.AssertEqual(t, tt.expectedCount, len(result), 
				"%s: Message count mismatch", tt.description)

			if tt.expectedSystem {
				testhelpers.AssertEqual(t, "system", result[0].Role, 
					"%s: First message should be system", tt.description)
				testhelpers.AssertContains(t, result[0].Content, "JSON", 
					"%s: System message should contain universal prompt", tt.description)
			}
		})
	}
}

func TestInjectUniversalPromptMerging(t *testing.T) {
	manager, err := CreateLLMangoManger(nil)
	testhelpers.RequireNoError(t, err, "Failed to create manager")

	existingSystemPrompt := "You are a helpful assistant."
	universalPrompt := "You must respond with valid JSON."

	messages := []openrouter.Message{
		{Role: "system", Content: existingSystemPrompt},
		{Role: "user", Content: "Hello"},
	}

	result := manager.injectUniversalPrompt(messages, universalPrompt)

	testhelpers.AssertEqual(t, 2, len(result), "Should have same number of messages")
	testhelpers.AssertEqual(t, "system", result[0].Role, "First message should be system")
	
	// Check that both prompts are present in the merged content
	mergedContent := result[0].Content
	testhelpers.AssertContains(t, mergedContent, "helpful assistant", 
		"Should contain original system prompt")
	testhelpers.AssertContains(t, mergedContent, "valid JSON", 
		"Should contain universal prompt")
}

// Helper functions for creating test data

func createTestTypedGoal() *Goal {
	type TestInput struct {
		Text string `json:"text"`
	}
	type TestOutput struct {
		Result string `json:"result"`
	}

	inputExample := TestInput{Text: "test input"}
	outputExample := TestOutput{Result: "test output"}

	return NewGoal(
		"test-typed-goal",
		"Test Typed Goal",
		"A test goal using typed structs",
		inputExample,
		outputExample,
	)
}

func createTestJSONGoal() *Goal {
	inputJSON := json.RawMessage(`{"text": "test input"}`)
	outputJSON := json.RawMessage(`{"result": "test output"}`)

	return NewJSONGoal(
		"test-json-goal",
		"Test JSON Goal",
		"A test goal using JSON",
		inputJSON,
		outputJSON,
	)
}

func createTestPrompt(model, uid string) *Prompt {
	return &Prompt{
		UID:    uid,
		Model:  model,
		Weight: 100,
		Messages: []openrouter.Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: "Process this: {{.text}}"},
		},
		Parameters: openrouter.Parameters{
			Temperature: &[]float64{0.7}[0],
		},
	}
}