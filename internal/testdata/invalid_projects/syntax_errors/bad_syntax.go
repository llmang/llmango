package mango

import (
	"github.com/llmang/llmango/llmango"
	"github.com/llmang/llmango/openrouter"
)

type BadInput struct {
	Message string `json:"message"`
}

type BadOutput struct {
	Response string `json:"response"`
}

// This has invalid syntax - missing closing brace
var badGoal = llmango.Goal{
	UID:         "bad-goal",
	Title:       "Bad Goal",
	Description: "A goal with syntax errors",
	InputOutput: llmango.InputOutput[BadInput, BadOutput]{
		InputExample: BadInput{
			Message: "Hello",
		},
		OutputExample: BadOutput{
			Response: "Hi there!",
		},
	},
// Missing closing brace here