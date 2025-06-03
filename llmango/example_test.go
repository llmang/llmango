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
		if goal.IsSchemaValidated {
			t.Error("Typed goal should not be schema validated")
		}
	})

	t.Run("ExampleJSONGoal", func(t *testing.T) {
		goal := ExampleJSONGoal()
		if goal == nil {
			t.Error("ExampleJSONGoal returned nil")
		}
		if !goal.IsSchemaValidated {
			t.Error("JSON goal should be schema validated")
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
