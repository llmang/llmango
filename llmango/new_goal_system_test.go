package llmango

import (
	"encoding/json"
	"errors"
	"testing"
)

// Test data structures
type UserInput struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type UserOutput struct {
	Message string  `json:"message"`
	Valid   bool    `json:"valid"`
	Score   float64 `json:"score"`
}

// Test validators
func validateUserInput(input UserInput) error {
	if input.Name == "" {
		return errors.New("name is required")
	}
	if input.Age < 0 {
		return errors.New("age must be non-negative")
	}
	return nil
}

func validateUserOutput(output UserOutput) error {
	if output.Message == "" {
		return errors.New("message is required")
	}
	if output.Score < 0 || output.Score > 1 {
		return errors.New("score must be between 0 and 1")
	}
	return nil
}

func TestNewGoal(t *testing.T) {
	inputExample := UserInput{Name: "John", Age: 30}
	outputExample := UserOutput{Message: "Hello John", Valid: true, Score: 0.95}

	validator := TypedValidator[UserInput, UserOutput]{
		ValidateInput:  validateUserInput,
		ValidateOutput: validateUserOutput,
	}

	goal := NewGoal("test-typed", "Test Typed Goal", "A test goal with validators",
		inputExample, outputExample, validator)

	// Test basic properties
	if goal.UID != "test-typed" {
		t.Errorf("Expected UID 'test-typed', got %s", goal.UID)
	}

	if goal.IsSchemaValidated {
		t.Error("Typed goal should not be schema validated")
	}

	if goal.InputValidator == nil {
		t.Error("Input validator should be set")
	}

	if goal.OutputValidator == nil {
		t.Error("Output validator should be set")
	}

	// Test that examples were marshaled correctly
	var unmarshaledInput UserInput
	if err := json.Unmarshal(goal.InputExample, &unmarshaledInput); err != nil {
		t.Fatalf("Failed to unmarshal input example: %v", err)
	}

	if unmarshaledInput.Name != inputExample.Name || unmarshaledInput.Age != inputExample.Age {
		t.Errorf("Input example mismatch. Expected: %+v, Got: %+v", inputExample, unmarshaledInput)
	}

	var unmarshaledOutput UserOutput
	if err := json.Unmarshal(goal.OutputExample, &unmarshaledOutput); err != nil {
		t.Fatalf("Failed to unmarshal output example: %v", err)
	}

	if unmarshaledOutput.Message != outputExample.Message ||
		unmarshaledOutput.Valid != outputExample.Valid ||
		unmarshaledOutput.Score != outputExample.Score {
		t.Errorf("Output example mismatch. Expected: %+v, Got: %+v", outputExample, unmarshaledOutput)
	}
}

func TestNewGoalWithoutValidators(t *testing.T) {
	inputExample := UserInput{Name: "Jane", Age: 25}
	outputExample := UserOutput{Message: "Hello Jane", Valid: true, Score: 0.88}

	goal := NewGoal("test-no-validators", "Test Goal No Validators", "A test goal without validators",
		inputExample, outputExample)

	if goal.IsSchemaValidated {
		t.Error("Typed goal should not be schema validated")
	}

	if goal.InputValidator != nil {
		t.Error("Input validator should be nil when not provided")
	}

	if goal.OutputValidator != nil {
		t.Error("Output validator should be nil when not provided")
	}
}

func TestNewGoalValidatorWrappers(t *testing.T) {
	inputExample := UserInput{Name: "Bob", Age: 35}
	outputExample := UserOutput{Message: "Hello Bob", Valid: true, Score: 0.92}

	validator := TypedValidator[UserInput, UserOutput]{
		ValidateInput:  validateUserInput,
		ValidateOutput: validateUserOutput,
	}

	goal := NewGoal("test-validators", "Test Validators", "Test validator wrappers",
		inputExample, outputExample, validator)

	// Test valid input
	validInputJSON, _ := json.Marshal(UserInput{Name: "Alice", Age: 28})
	if err := goal.InputValidator(validInputJSON); err != nil {
		t.Errorf("Valid input should pass validation: %v", err)
	}

	// Test invalid input (empty name)
	invalidInputJSON, _ := json.Marshal(UserInput{Name: "", Age: 28})
	if err := goal.InputValidator(invalidInputJSON); err == nil {
		t.Error("Invalid input should fail validation")
	}

	// Test invalid input (negative age)
	invalidAgeJSON, _ := json.Marshal(UserInput{Name: "Charlie", Age: -5})
	if err := goal.InputValidator(invalidAgeJSON); err == nil {
		t.Error("Invalid age should fail validation")
	}

	// Test valid output
	validOutputJSON, _ := json.Marshal(UserOutput{Message: "Valid", Valid: true, Score: 0.75})
	if err := goal.OutputValidator(validOutputJSON); err != nil {
		t.Errorf("Valid output should pass validation: %v", err)
	}

	// Test invalid output (empty message)
	invalidOutputJSON, _ := json.Marshal(UserOutput{Message: "", Valid: true, Score: 0.75})
	if err := goal.OutputValidator(invalidOutputJSON); err == nil {
		t.Error("Invalid output should fail validation")
	}

	// Test invalid output (score out of range)
	invalidScoreJSON, _ := json.Marshal(UserOutput{Message: "Test", Valid: true, Score: 1.5})
	if err := goal.OutputValidator(invalidScoreJSON); err == nil {
		t.Error("Invalid score should fail validation")
	}
}

func TestNewJSONGoal(t *testing.T) {
	inputJSON := json.RawMessage(`{"name": "John", "age": 30}`)
	outputJSON := json.RawMessage(`{"message": "Hello John", "valid": true, "score": 0.95}`)

	goal := NewJSONGoal("test-json", "Test JSON Goal", "A test JSON goal", inputJSON, outputJSON)

	// Test basic properties
	if goal.UID != "test-json" {
		t.Errorf("Expected UID 'test-json', got %s", goal.UID)
	}

	if !goal.IsSchemaValidated {
		t.Error("JSON goal should be schema validated")
	}

	if goal.InputValidator == nil {
		t.Error("Input validator should be set for JSON goal")
	}

	if goal.OutputValidator == nil {
		t.Error("Output validator should be set for JSON goal")
	}

	// Test that examples are stored correctly
	if string(goal.InputExample) != string(inputJSON) {
		t.Errorf("Input example mismatch. Expected: %s, Got: %s", string(inputJSON), string(goal.InputExample))
	}

	if string(goal.OutputExample) != string(outputJSON) {
		t.Errorf("Output example mismatch. Expected: %s, Got: %s", string(outputJSON), string(goal.OutputExample))
	}
}

func TestJSONGoalSchemaValidation(t *testing.T) {
	inputJSON := json.RawMessage(`{"name": "John", "age": 30}`)
	outputJSON := json.RawMessage(`{"message": "Hello", "valid": true, "score": 0.95}`)

	goal := NewJSONGoal("test-schema", "Test Schema", "Test schema validation", inputJSON, outputJSON)

	// Test valid input (matches schema)
	validInput := json.RawMessage(`{"name": "Alice", "age": 25}`)
	if err := goal.InputValidator(validInput); err != nil {
		t.Errorf("Valid input should pass schema validation: %v", err)
	}

	// Test invalid input (missing required field)
	invalidInput := json.RawMessage(`{"name": "Bob"}`)
	if err := goal.InputValidator(invalidInput); err == nil {
		t.Error("Invalid input should fail schema validation")
	}

	// Test invalid input (wrong type)
	wrongTypeInput := json.RawMessage(`{"name": "Charlie", "age": "thirty"}`)
	if err := goal.InputValidator(wrongTypeInput); err == nil {
		t.Error("Wrong type input should fail schema validation")
	}

	// Test valid output
	validOutput := json.RawMessage(`{"message": "Test", "valid": false, "score": 0.5}`)
	if err := goal.OutputValidator(validOutput); err != nil {
		t.Errorf("Valid output should pass schema validation: %v", err)
	}

	// Test invalid output (missing required field)
	invalidOutput := json.RawMessage(`{"message": "Test", "valid": true}`)
	if err := goal.OutputValidator(invalidOutput); err == nil {
		t.Error("Invalid output should fail schema validation")
	}
}

func TestGoalValidation(t *testing.T) {
	// Test typed goal with validators (should have no warnings)
	typedGoal := NewGoal("typed", "Typed", "Typed goal",
		UserInput{Name: "test", Age: 25},
		UserOutput{Message: "test", Valid: true, Score: 0.5},
		TypedValidator[UserInput, UserOutput]{
			ValidateInput:  validateUserInput,
			ValidateOutput: validateUserOutput,
		})

	warnings := typedGoal.Validate()
	if len(warnings) != 0 {
		t.Errorf("Typed goal with validators should have no warnings, got: %v", warnings)
	}

	// Test typed goal without validators (should warn)
	typedGoalNoValidators := NewGoal("typed-no-val", "Typed No Val", "Typed goal no validators",
		UserInput{Name: "test", Age: 25},
		UserOutput{Message: "test", Valid: true, Score: 0.5})

	warnings = typedGoalNoValidators.Validate()
	if len(warnings) == 0 {
		t.Error("Typed goal without validators should have warnings")
	}

	// Test JSON goal (should have no warnings)
	jsonGoal := NewJSONGoal("json", "JSON", "JSON goal",
		json.RawMessage(`{"name": "test", "age": 25}`),
		json.RawMessage(`{"message": "test", "valid": true, "score": 0.5}`))

	warnings = jsonGoal.Validate()
	if len(warnings) != 0 {
		t.Errorf("JSON goal should have no warnings, got: %v", warnings)
	}
}
