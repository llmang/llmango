package mango

import (
	"github.com/llmang/llmango/llmango"
	"github.com/llmang/llmango/openrouter"
)

type TestInput struct {
	Message string `json:"message"`
}

type TestOutput struct {
	Response string `json:"response"`
}

var testGoal = llmango.Goal{
	UID:         "test-goal",
	Title:       "Test Goal",
	Description: "A test goal for unit testing",
	InputOutput: llmango.InputOutput[TestInput, TestOutput]{
		InputExample: TestInput{
			Message: "Hello",
		},
		OutputExample: TestOutput{
			Response: "Hi there!",
		},
	},
}

var testPrompt = llmango.Prompt{
	UID:     "test-prompt",
	GoalUID: "test-goal",
	Model:   "openai/gpt-4",
	Weight:  100,
	Messages: []openrouter.Message{
		{Role: "system", Content: "You are a test assistant."},
		{Role: "user", Content: "{{message}}"},
	},
}