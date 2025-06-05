package llmango

import (
	"testing"
)

func TestExampleUsage(t *testing.T) {
	// Test that our examples run without panicking
	t.Run("ExampleTypedGoal", func(t *testing.T) {
		goal := ExampleTypedGoal()
		if goal == nil {
			t.Error("ExampleTypedGoal returned nil")
		}
	})

	t.Run("ExampleJSONGoal", func(t *testing.T) {
		goal := ExampleJSONGoal()
		if goal == nil {
			t.Error("ExampleJSONGoal returned nil")
		}
	})

	t.Run("ExampleGoalUsage", func(t *testing.T) {
		// This should run without panicking
		ExampleGoalUsage()
	})

	t.Run("ExampleManagerIntegration", func(t *testing.T) {
		// This should run without panicking
		ExampleManagerIntegration()
	})
}
