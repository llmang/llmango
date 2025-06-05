package llmango

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Example usage of the new goal system - focused on LLM tasks: extraction, classification, generation

// Example 1: Text Classification (Typed Goal)
type SentimentInput struct {
	Text string `json:"text"`
}

type SentimentOutput struct {
	Sentiment  string  `json:"sentiment"`
	Confidence float64 `json:"confidence"`
	Reasoning  string  `json:"reasoning"`
}

// Custom validators for typed goals
func validateSentimentInput(input SentimentInput) error {
	if input.Text == "" {
		return errors.New("text is required for sentiment analysis")
	}
	if len(input.Text) > 5000 {
		return errors.New("text too long, maximum 5000 characters")
	}
	return nil
}

func validateSentimentOutput(output SentimentOutput) error {
	validSentiments := []string{"positive", "negative", "neutral"}
	isValid := false
	for _, valid := range validSentiments {
		if output.Sentiment == valid {
			isValid = true
			break
		}
	}
	if !isValid {
		return errors.New("sentiment must be positive, negative, or neutral")
	}
	if output.Confidence < 0 || output.Confidence > 1 {
		return errors.New("confidence must be between 0 and 1")
	}
	if output.Reasoning == "" {
		return errors.New("reasoning is required")
	}
	return nil
}

// ExampleTypedGoal demonstrates creating a typed goal for sentiment classification
func ExampleTypedGoal() *Goal {
	// Create input/output examples
	inputExample := SentimentInput{
		Text: "I love this new product! It works perfectly and exceeded my expectations.",
	}

	outputExample := SentimentOutput{
		Sentiment:  "positive",
		Confidence: 0.92,
		Reasoning:  "The text contains positive language like 'love', 'perfectly', and 'exceeded expectations'",
	}

	// Create validator
	validator := TypedValidator[SentimentInput, SentimentOutput]{
		ValidateInput:  validateSentimentInput,
		ValidateOutput: validateSentimentOutput,
	}

	// Create goal using NewGoal (standard way for developers)
	goal := NewGoal(
		"sentiment-classifier",
		"Sentiment Classification",
		"Analyzes text sentiment with confidence scoring and reasoning",
		inputExample,
		outputExample,
		validator,
	)

	fmt.Printf("Created typed goal: %s (IsSchemaValidated: %v)\n", goal.UID, goal.IsSchemaValidated)
	return goal
}

// ExampleJSONGoal demonstrates creating a JSON goal for entity extraction
func ExampleJSONGoal() *Goal {
	// JSON examples for entity extraction task (could come from frontend)
	inputJSON := json.RawMessage(`{
		"text": "John Smith works at Apple Inc. in Cupertino, California. He can be reached at john.smith@apple.com or (555) 123-4567.",
		"entity_types": ["person", "organization", "location", "email", "phone"]
	}`)

	outputJSON := json.RawMessage(`{
		"entities": [
			{
				"text": "John Smith",
				"type": "person",
				"start": 0,
				"end": 10,
				"confidence": 0.98
			},
			{
				"text": "Apple Inc.",
				"type": "organization",
				"start": 20,
				"end": 30,
				"confidence": 0.95
			},
			{
				"text": "Cupertino, California",
				"type": "location",
				"start": 34,
				"end": 55,
				"confidence": 0.89
			},
			{
				"text": "john.smith@apple.com",
				"type": "email",
				"start": 78,
				"end": 98,
				"confidence": 0.99
			}
		],
		"total_entities": 4
	}`)

	// Create goal using NewJSONGoal (standard way for frontend/dynamic use)
	goal := NewJSONGoal(
		"entity-extractor",
		"Named Entity Extraction",
		"Extracts and classifies named entities from text with position and confidence",
		inputJSON,
		outputJSON,
	)

	fmt.Printf("Created JSON goal: %s (IsSchemaValidated: %v)\n", goal.UID, goal.IsSchemaValidated)
	return goal
}

// ExampleGoalUsage demonstrates how both goal types work the same way at runtime
func ExampleGoalUsage() {
	// Create both types of goals
	typedGoal := ExampleTypedGoal()
	jsonGoal := ExampleJSONGoal()

	// Both goals have the same runtime interface
	fmt.Println("\n=== Typed Goal Validation ===")

	// Test typed goal validation (sentiment analysis)
	validInput := json.RawMessage(`{"text": "This movie was absolutely fantastic!"}`)
	if err := typedGoal.InputValidator(validInput); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
	} else {
		fmt.Println("Input validation passed!")
	}

	invalidInput := json.RawMessage(`{"text": ""}`)
	if err := typedGoal.InputValidator(invalidInput); err != nil {
		fmt.Printf("Expected validation failure: %v\n", err)
	}

	fmt.Println("\n=== JSON Goal Validation ===")

	// Test JSON goal validation (entity extraction)
	validEntityInput := json.RawMessage(`{"text": "Alice Johnson works at Microsoft.", "entity_types": ["person", "organization"]}`)
	if err := jsonGoal.InputValidator(validEntityInput); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
	} else {
		fmt.Println("Input validation passed!")
	}

	invalidEntityInput := json.RawMessage(`{"text": "Alice Johnson works at Microsoft."}`)
	if err := jsonGoal.InputValidator(invalidEntityInput); err != nil {
		fmt.Printf("Expected validation failure: %v\n", err)
	}

	// Show goal validation warnings
	fmt.Println("\n=== Goal Validation Warnings ===")

	// Goal without validators
	goalNoValidators := NewGoal("no-validators", "No Validators", "Goal without validators",
		SentimentInput{Text: "This is a test"},
		SentimentOutput{Sentiment: "neutral", Confidence: 0.5, Reasoning: "Test reasoning"})

	warnings := goalNoValidators.Validate()
	if len(warnings) > 0 {
		fmt.Printf("Warnings for goal without validators: %v\n", warnings)
	}

	// Goal with validators
	warnings = typedGoal.Validate()
	if len(warnings) == 0 {
		fmt.Println("No warnings for properly configured typed goal")
	}

	warnings = jsonGoal.Validate()
	if len(warnings) == 0 {
		fmt.Println("No warnings for JSON goal")
	}
}

// ExampleManagerIntegration shows how goals work with LLMangoManager
func ExampleManagerIntegration() {
	// Create manager
	manager, err := CreateLLMangoManger(nil)
	if err != nil {
		fmt.Printf("Failed to create manager: %v\n", err)
		return
	}

	// Create and add goals
	typedGoal := ExampleTypedGoal()
	jsonGoal := ExampleJSONGoal()

	manager.AddGoals(typedGoal, jsonGoal)

	fmt.Printf("\n=== Manager Integration ===\n")
	fmt.Printf("Added %d goals to manager\n", len(manager.Goals.Snapshot()))

	// Retrieve goals
	retrievedTyped, exists := manager.Goals.Get("sentiment-classifier")
	if exists {
		fmt.Printf("Retrieved typed goal: %s (IsSchemaValidated: %v)\n",
			retrievedTyped.UID, retrievedTyped.IsSchemaValidated)
	}

	retrievedJSON, exists := manager.Goals.Get("entity-extractor")
	if exists {
		fmt.Printf("Retrieved JSON goal: %s (IsSchemaValidated: %v)\n",
			retrievedJSON.UID, retrievedJSON.IsSchemaValidated)
	}
}
