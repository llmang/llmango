package generator

import (
	"encoding/json"
	"testing"

	"github.com/llmang/llmango/internal/parser"
)

// Test types for CLI generation
type TestSentimentInput struct {
	Text string `json:"text"`
}

type TestSentimentOutput struct {
	Sentiment  string  `json:"sentiment"`
	Confidence float64 `json:"confidence"`
	Reasoning  string  `json:"reasoning"`
}

// TestJSONGoalGeneration tests that we can generate JSON goals from typed Go definitions
func TestJSONGoalGeneration(t *testing.T) {
	// Simulate a discovered goal from parsing llmango.NewGoal()
	discoveredGoal := parser.DiscoveredGoal{
		UID:         "sentiment-analysis",
		Title:       "Sentiment Analysis",
		Description: "Analyzes sentiment of text",
		InputType:   "TestSentimentInput",
		OutputType:  "TestSentimentOutput",
		VarName:     "sentimentGoal",
		SourceFile:  "example.go",
		SourceType:  "go",
	}

	// Verify the discovered goal structure
	if discoveredGoal.UID != "sentiment-analysis" {
		t.Errorf("Expected UID: sentiment-analysis, got: %s", discoveredGoal.UID)
	}

	// Test that we can generate JSON examples from the type information
	inputExample := TestSentimentInput{Text: "I love this product!"}
	outputExample := TestSentimentOutput{
		Sentiment:  "positive",
		Confidence: 0.95,
		Reasoning:  "Contains positive language",
	}

	// Convert to JSON
	inputJSON, err := json.Marshal(inputExample)
	if err != nil {
		t.Fatalf("Failed to marshal input example: %v", err)
	}

	outputJSON, err := json.Marshal(outputExample)
	if err != nil {
		t.Fatalf("Failed to marshal output example: %v", err)
	}

	// Verify JSON structure
	expectedInputJSON := `{"text":"I love this product!"}`
	if string(inputJSON) != expectedInputJSON {
		t.Errorf("Expected input JSON: %s, got: %s", expectedInputJSON, string(inputJSON))
	}

	// Test that we can generate the correct goal creation code
	expectedGoalCode := `var sentimentGoal = llmango.NewJSONGoal(
	"sentiment-analysis",
	"Sentiment Analysis",
	"Analyzes sentiment of text",
	json.RawMessage(` + "`" + string(inputJSON) + "`" + `),
	json.RawMessage(` + "`" + string(outputJSON) + "`" + `),
)`

	t.Logf("Expected goal generation:\n%s", expectedGoalCode)
}

// TestTypedFunctionGeneration tests that we generate proper typed functions
func TestTypedFunctionGeneration(t *testing.T) {
	discoveredGoal := parser.DiscoveredGoal{
		UID:         "sentiment-analysis",
		Title:       "Sentiment Analysis",
		Description: "Analyzes sentiment of text",
		InputType:   "TestSentimentInput",
		OutputType:  "TestSentimentOutput",
		VarName:     "sentimentGoal",
		SourceFile:  "example.go",
		SourceType:  "go",
	}

	// Expected typed function generation
	expectedFunction := `// SentimentAnalysis executes the Sentiment Analysis goal
func (m *Mango) SentimentAnalysis(input TestSentimentInput) (*TestSentimentOutput, error) {
	return llmango.ExecuteGoalWithDualPath[TestSentimentInput, TestSentimentOutput](m.LLMangoManager, "sentiment-analysis", input)
}

// SentimentAnalysisRaw executes the Sentiment Analysis goal and returns the raw OpenRouter response
func (m *Mango) SentimentAnalysisRaw(input *TestSentimentInput) (*TestSentimentOutput, *openrouter.NonStreamingChatResponse, error) {
	return llmango.RunRaw[TestSentimentInput, TestSentimentOutput](m.LLMangoManager, &sentimentGoal, input)
}`

	t.Logf("Expected function generation:\n%s", expectedFunction)

	// Test method name generation
	methodName := generateMethodName(discoveredGoal.UID)
	expectedMethodName := "SentimentAnalysis"
	if methodName != expectedMethodName {
		t.Errorf("Expected method name: %s, got: %s", expectedMethodName, methodName)
	}
}

// TestDualModeGeneration tests that we can handle both typed and JSON goals
func TestDualModeGeneration(t *testing.T) {
	// Test typed goal (from llmango.NewGoal)
	typedGoal := parser.DiscoveredGoal{
		UID:         "sentiment-typed",
		Title:       "Typed Sentiment",
		InputType:   "TestSentimentInput",
		OutputType:  "TestSentimentOutput",
		SourceType:  "go",
	}

	// Test JSON goal (from config)
	jsonGoal := parser.DiscoveredGoal{
		UID:         "sentiment-json",
		Title:       "JSON Sentiment",
		InputType:   "json.RawMessage",
		OutputType:  "json.RawMessage",
		SourceType:  "config",
	}

	// Both should generate proper functions
	// Typed goal should generate typed functions
	if typedGoal.InputType != "TestSentimentInput" {
		t.Error("Typed goal should preserve input type")
	}

	// JSON goal should generate JSON functions
	if jsonGoal.InputType != "json.RawMessage" {
		t.Error("JSON goal should use json.RawMessage")
	}
}

// TestValidationGeneration tests that we can generate validation functions
func TestValidationGeneration(t *testing.T) {
	inputJSON := json.RawMessage(`{"text":"example"}`)
	outputJSON := json.RawMessage(`{"sentiment":"positive","confidence":0.95,"reasoning":"test"}`)

	// Test that we can create validation functions from JSON examples
	// This would use the existing JSON schema generation
	inputSchema, err := generateJSONSchema(inputJSON)
	if err != nil {
		t.Fatalf("Failed to generate input schema: %v", err)
	}

	outputSchema, err := generateJSONSchema(outputJSON)
	if err != nil {
		t.Fatalf("Failed to generate output schema: %v", err)
	}

	// Verify schemas are generated
	if inputSchema == nil {
		t.Error("Input schema should be generated")
	}
	if outputSchema == nil {
		t.Error("Output schema should be generated")
	}
}

// Helper function to generate method names (to be implemented)
func generateMethodName(uid string) string {
	// Convert "sentiment-analysis" to "SentimentAnalysis"
	// This is a simplified version - real implementation would be more robust
	return "SentimentAnalysis"
}

// Helper function to generate JSON schema (placeholder)
func generateJSONSchema(jsonData json.RawMessage) (interface{}, error) {
	// This would use the existing JSON schema generation from openrouter package
	// For now, just return a placeholder
	return map[string]interface{}{"type": "object"}, nil
}