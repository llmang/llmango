package llmangoagents

import "fmt"

// ExampleUsage demonstrates how to use the IsToolCallingSupported function
func ExampleUsage() {
	// Example model IDs to check
	modelIDs := []string{
		"anthropic/claude-opus-4",
		"openai/gpt-4o",
		"fake/unsupported-model",
		"google/gemini-2.5-flash-preview",
		"random-model-id",
	}

	fmt.Println("Tool Calling Support Check:")
	fmt.Println("===========================")

	for _, modelID := range modelIDs {
		supported := IsToolCallingSupported(modelID)
		status := "❌ Not Supported"
		if supported {
			status = "✅ Supported"
		}
		fmt.Printf("%-40s %s\n", modelID, status)
	}
}
