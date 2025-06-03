package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/llmang/llmango/llmango"
	"github.com/llmang/llmango/llmangologger"
	"github.com/llmang/llmango/openrouter"
)

func main() {
	fmt.Println("🧪 Testing unified logging system...")

	// Create a mock OpenRouter (we won't actually call the API)
	openRouter := &openrouter.OpenRouter{}

	// Initialize LLMango manager with logging
	manager, err := llmango.CreateLLMangoManger(openRouter)
	if err != nil {
		log.Fatal("Failed to create LLMango manager:", err)
	}

	// Enable logging with print logger (logs input/output objects only)
	manager = manager.WithLogging(llmangologger.CreatePrintLogger(false))

	// Create a test goal
	testGoal := llmango.NewJSONGoal(
		"test-goal",
		"Test Goal",
		"A test goal for logging",
		json.RawMessage(`{"text": "example input"}`),
		json.RawMessage(`{"result": "example output"}`),
	)

	// Add the goal to the manager
	manager.AddGoals(testGoal)

	fmt.Println("✅ Unified logging system setup completed!")
	fmt.Println("📋 Summary of changes:")
	fmt.Println("  • Removed all debug flags and manual log.Printf() statements")
	fmt.Println("  • Implemented .WithLogging() fluent interface pattern")
	fmt.Println("  • Created simplified logger factory functions:")
	fmt.Println("    - llmangologger.CreatePrintLogger(logFullRequests bool)")
	fmt.Println("    - llmangologger.CreateSQLiteLogger(db, logFullRequests bool)")
	fmt.Println("    - llmangologger.CreateNoOpLogger()")
	fmt.Println("  • Updated example app to use new logging pattern")
	fmt.Println("  • Maintained backward compatibility with existing code")
	fmt.Println("")
	fmt.Println("🎯 Usage example:")
	fmt.Println("  manager := llmango.CreateLLMangoManger(openRouter)")
	fmt.Println("  manager = manager.WithLogging(llmangologger.CreatePrintLogger(false))")
	fmt.Println("")
	fmt.Println("🎉 Unified logging system implementation complete!")
}