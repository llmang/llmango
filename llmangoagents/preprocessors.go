package llmangoagents

import (
	"strings"
)

//agent preprocessors run before the agent invokes its first action

func preThink(ag *Agent, input string) string {
	// TODO: This function needs to be updated to work with the new execution context system
	// The original implementation referenced ag.Manager which doesn't exist in the new structure
	// For now, return empty string to avoid compilation errors
	return ""
}

func preRag(ag *Agent, input string) string {
	// TODO: This function needs to be updated to work with the new execution context system
	// The original implementation had syntax errors and referenced non-existent fields
	// For now, return empty string to avoid compilation errors
	return ""
}

//========================================================
//================Prevalidation tool======================

// PreValidateTool definitions
const (
	PreValidateToolRedirect   = "prevalidate_redirect"
	PreValidateToolError      = "prevalidate_error"
	PreValidateToolReturnBack = "prevalidate_returnback"
)

// PreValidateToolList is the list of special toolcalls for prevalidation
var PreValidateToolList = []string{
	PreValidateToolRedirect,
	PreValidateToolError,
	PreValidateToolReturnBack,
}

func preValidate(ag *Agent, input string) string {
	// TODO: This function needs to be updated to work with the new execution context system
	// The original implementation referenced ag.BuildSystemMessage() and ag.Manager which don't exist
	// For now, return empty string to avoid compilation errors
	return ""
}

//========================================================
//================Post-processors========================

//these are ran after the step completes but before the next step.
//they can inject extra context messages along with the message.

func SummarizationPostProcesssor() string {
	return "@@CONTEXTMESSAGE: _________"
}

// Helper function to check if a string contains any of the action phrases
func containsActionPhrase(input string) bool {
	actionPhrases := []string{"@@REDIRECT:", "@@ERROR:", "@@RETURNBACK", "@@ABORT", "@@RETURN"}
	for _, phrase := range actionPhrases {
		if strings.Contains(input, phrase) {
			return true
		}
	}
	return false
}

// Extract action phrase from input
func extractActionPhrase(input string) string {
	if strings.Contains(input, "@@REDIRECT:") {
		return "@@REDIRECT"
	}
	if strings.Contains(input, "@@ERROR:") {
		return "@@ERROR"
	}
	if strings.Contains(input, "@@RETURNBACK") {
		return "@@RETURNBACK"
	}
	if strings.Contains(input, "@@ABORT") {
		return "@@ABORT"
	}
	if strings.Contains(input, "@@RETURN") {
		return "@@RETURN"
	}
	return ""
}
