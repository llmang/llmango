package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/llmang/llmango/llmango"
)

// Demo of the completed dual-mode goal system
func main() {
	fmt.Println("=== LLMango Dual-Mode Goal System Demo ===")

	// 1. Create a typed goal (for developers)
	fmt.Println("1. Creating Typed Goal (Sentiment Analysis)")
	typedGoal := createSentimentGoal()
	fmt.Printf("   UID: %s\n", typedGoal.UID)
	fmt.Printf("   IsSchemaValidated: %v\n", typedGoal.IsSchemaValidated)
	fmt.Printf("   Has Validators: %v\n\n", typedGoal.InputValidator != nil)

	// 2. Create a JSON goal (for frontend/dynamic use)
	fmt.Println("2. Creating JSON Goal (Entity Extraction)")
	jsonGoal := createEntityExtractionGoal()
	fmt.Printf("   UID: %s\n", jsonGoal.UID)
	fmt.Printf("   IsSchemaValidated: %v\n", jsonGoal.IsSchemaValidated)
	fmt.Printf("   Has Validators: %v\n\n", jsonGoal.InputValidator != nil)

	// 3. Test validation on both goal types
	fmt.Println("3. Testing Validation")
	testValidation(typedGoal, jsonGoal)

	// 4. Show goal conversion
	fmt.Println("4. Converting Typed Goal to JSON Goal")
	convertedGoal, err := llmango.ConvertTypedGoalToJSON(typedGoal)
	if err != nil {
		log.Printf("Conversion failed: %v", err)
	} else {
		fmt.Printf("   Converted successfully! IsSchemaValidated: %v\n\n", convertedGoal.IsSchemaValidated)
	}

	// 5. Manager integration
	fmt.Println("5. Manager Integration")
	testManagerIntegration(typedGoal, jsonGoal)

	fmt.Println("=== Demo Complete ===")
}

func createSentimentGoal() *llmango.Goal {
	type SentimentInput struct {
		Text string `json:"text"`
	}

	type SentimentOutput struct {
		Sentiment  string  `json:"sentiment"`
		Confidence float64 `json:"confidence"`
		Reasoning  string  `json:"reasoning"`
	}

	inputExample := SentimentInput{
		Text: "I love this new product! It works perfectly.",
	}

	outputExample := SentimentOutput{
		Sentiment:  "positive",
		Confidence: 0.92,
		Reasoning:  "Contains positive language like 'love' and 'perfectly'",
	}

	validator := llmango.TypedValidator[SentimentInput, SentimentOutput]{
		ValidateInput: func(input SentimentInput) error {
			if input.Text == "" {
				return fmt.Errorf("text is required")
			}
			return nil
		},
		ValidateOutput: func(output SentimentOutput) error {
			if output.Sentiment == "" {
				return fmt.Errorf("sentiment is required")
			}
			return nil
		},
	}

	return llmango.NewGoal(
		"sentiment-classifier",
		"Sentiment Classification",
		"Analyzes text sentiment with confidence scoring",
		inputExample,
		outputExample,
		validator,
	)
}

func createEntityExtractionGoal() *llmango.Goal {
	inputJSON := json.RawMessage(`{
		"text": "John Smith works at Apple Inc. in Cupertino.",
		"entity_types": ["person", "organization", "location"]
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
			}
		],
		"total_entities": 2
	}`)

	return llmango.NewJSONGoal(
		"entity-extractor",
		"Named Entity Extraction",
		"Extracts and classifies named entities from text",
		inputJSON,
		outputJSON,
	)
}

func testValidation(typedGoal, jsonGoal *llmango.Goal) {
	// Test typed goal
	validInput := json.RawMessage(`{"text": "This is a great product!"}`)
	if err := typedGoal.InputValidator(validInput); err != nil {
		fmt.Printf("   Typed goal validation failed: %v\n", err)
	} else {
		fmt.Printf("   Typed goal validation: ✓ PASSED\n")
	}

	invalidInput := json.RawMessage(`{"text": ""}`)
	if err := typedGoal.InputValidator(invalidInput); err != nil {
		fmt.Printf("   Typed goal invalid input: ✓ CORRECTLY REJECTED (%v)\n", err)
	}

	// Test JSON goal
	validEntityInput := json.RawMessage(`{"text": "Alice works at Microsoft.", "entity_types": ["person", "organization"]}`)
	if err := jsonGoal.InputValidator(validEntityInput); err != nil {
		fmt.Printf("   JSON goal validation failed: %v\n", err)
	} else {
		fmt.Printf("   JSON goal validation: ✓ PASSED\n")
	}

	invalidEntityInput := json.RawMessage(`{"text": "Alice works at Microsoft."}`)
	if err := jsonGoal.InputValidator(invalidEntityInput); err != nil {
		fmt.Printf("   JSON goal invalid input: ✓ CORRECTLY REJECTED (%v)\n\n", err)
	}
}

func testManagerIntegration(typedGoal, jsonGoal *llmango.Goal) {
	manager, err := llmango.CreateLLMangoManger(nil)
	if err != nil {
		log.Printf("Failed to create manager: %v", err)
		return
	}

	manager.AddGoals(typedGoal, jsonGoal)
	fmt.Printf("   Added %d goals to manager\n", len(manager.Goals.Snapshot()))

	// Retrieve goals
	if retrieved, exists := manager.Goals.Get("sentiment-classifier"); exists {
		fmt.Printf("   Retrieved typed goal: %s\n", retrieved.UID)
	}

	if retrieved, exists := manager.Goals.Get("entity-extractor"); exists {
		fmt.Printf("   Retrieved JSON goal: %s\n", retrieved.UID)
	}

	// Test reconstruction
	for _, goal := range manager.Goals.Snapshot() {
		if err := goal.ReconstructValidators(); err != nil {
			fmt.Printf("   Failed to reconstruct %s: %v\n", goal.UID, err)
		} else {
			fmt.Printf("   Successfully reconstructed validators for %s\n", goal.UID)
		}
	}
	fmt.Println()
}