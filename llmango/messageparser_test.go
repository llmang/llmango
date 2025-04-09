package llmango

import (
	"reflect"
	"testing"

	"github.com/llmang/llmango/openrouter"
)

type TestInput struct {
	Name           string `json:"name"`
	Age            int    `json:"age"`
	PreviousItems  string `json:"previousItems"`
	EmptyString    string `json:"emptyString"`
	ZeroValue      int    `json:"zeroValue"`
	NonEmptyString string `json:"nonEmptyString"`
}

func TestParseMessageIfStatements(t *testing.T) {
	// Test input with various fields
	input := TestInput{
		Name:           "John",
		Age:            30,
		PreviousItems:  "item1, item2, item3",
		EmptyString:    "",
		ZeroValue:      0,
		NonEmptyString: "value",
	}

	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name: "Basic if statement with existing variable",
			message: `Test message with {{#if previousItems}}
The following items have already been used: {{previousItems}}
{{/if}}`,
			expected: `Test message with 
The following items have already been used: {{previousItems}}
`,
		},
		{
			name:     "If statement with empty variable",
			message:  `Test message with {{#if emptyString}}This should be removed{{/if}}`,
			expected: `Test message with `,
		},
		{
			name:     "If statement with non-existent variable",
			message:  `Test message with {{#if nonExistentVar}}This should be removed{{/if}}`,
			expected: `Test message with `,
		},
		{
			name:     "If statement with zero value",
			message:  `Test message with {{#if zeroValue}}Zero is considered non-empty{{/if}}`,
			expected: `Test message with Zero is considered non-empty`,
		},
		{
			name:     "Multiple if statements",
			message:  `{{#if name}}Name: {{name}}{{/if}} {{#if emptyString}}Empty{{/if}} {{#if nonEmptyString}}NonEmpty{{/if}}`,
			expected: `Name: {{name}}  NonEmpty`,
		},
		{
			name:     "Nested content without actual nesting",
			message:  `{{#if name}}Name: {{#if nonEmptyString}}Value exists{{/if}}{{/if}}`,
			expected: `Name: {{#if nonEmptyString}}Value exists{{/if}}`,
		},
		{
			name:     "Malformed if statement",
			message:  `Test {{#if }}Malformed{{/if}}`,
			expected: `Test {{#if }}Malformed{{/if}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Save original for later comparison
			originalMessage := openrouter.Message{
				Role:    "user",
				Content: test.message,
			}
			originalMessages := []openrouter.Message{originalMessage}

			// Process the message
			result, err := ParseMessageIfStatements(input, originalMessages)
			if err != nil {
				t.Fatalf("ParseMessageIfStatements returned error: %v", err)
			}

			// Check the result
			if len(result) != 1 {
				t.Fatalf("Expected 1 message, got %d", len(result))
			}
			if result[0].Content != test.expected {
				t.Errorf("Expected content:\n%s\n\nGot:\n%s", test.expected, result[0].Content)
			}

			// Check that the original wasn't modified
			if originalMessages[0].Content != test.message {
				t.Errorf("Original message was modified")
			}
		})
	}
}

func TestInsertVariableValuesIntoPromptMessagesCopy(t *testing.T) {
	// Test input
	input := TestInput{
		Name:          "John",
		Age:           30,
		PreviousItems: "item1, item2, item3",
	}

	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name:     "Basic variable replacement",
			message:  "Hello {{name}}, you are {{age}} years old.",
			expected: "Hello John, you are 30 years old.",
		},
		{
			name:     "Missing variable",
			message:  "Hello {{nonExistent}}",
			expected: "Hello {{nonExistent}}",
		},
		{
			name:     "Multiple variables",
			message:  "Name: {{name}}, Age: {{age}}, Items: {{previousItems}}",
			expected: "Name: John, Age: 30, Items: item1, item2, item3",
		},
		{
			name:     "No variables",
			message:  "Plain text with no variables",
			expected: "Plain text with no variables",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Save original for later comparison
			originalMessage := openrouter.Message{
				Role:    "user",
				Content: test.message,
			}
			originalMessages := []openrouter.Message{originalMessage}

			// Process the message
			result, err := InsertVariableValuesIntoPromptMessagesCopy(input, originalMessages)
			if err != nil {
				t.Fatalf("InsertVariableValuesIntoPromptMessagesCopy returned error: %v", err)
			}

			// Check the result
			if len(result) != 1 {
				t.Fatalf("Expected 1 message, got %d", len(result))
			}
			if result[0].Content != test.expected {
				t.Errorf("Expected content:\n%s\n\nGot:\n%s", test.expected, result[0].Content)
			}

			// Check that the original wasn't modified
			if originalMessages[0].Content != test.message {
				t.Errorf("Original message was modified")
			}
		})
	}
}

func TestParseMessages(t *testing.T) {
	// Test input
	input := TestInput{
		Name:           "John",
		Age:            30,
		PreviousItems:  "item1, item2, item3",
		EmptyString:    "",
		NonEmptyString: "value",
	}

	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name: "Conditional block followed by variable replacement",
			message: `{{#if previousItems}}
Previous items: {{previousItems}}
{{/if}}
Hello {{name}}, you are {{age}} years old.`,
			expected: `
Previous items: item1, item2, item3

Hello John, you are 30 years old.`,
		},
		{
			name: "Empty conditional block",
			message: `{{#if emptyString}}
This should be removed
{{/if}}
Hello {{name}}!`,
			expected: `
Hello John!`,
		},
		{
			name: "Multiple conditional blocks with variables",
			message: `{{#if name}}Name: {{name}}{{/if}}
{{#if emptyString}}Empty{{/if}}
{{#if nonEmptyString}}NonEmpty: {{nonEmptyString}}{{/if}}`,
			expected: `Name: John

NonEmpty: value`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Save original for later comparison
			originalMessage := openrouter.Message{
				Role:    "user",
				Content: test.message,
			}
			originalMessages := []openrouter.Message{originalMessage}

			// Create a copy for deep equality check later
			originalMessagesCopy := make([]openrouter.Message, len(originalMessages))
			for i, msg := range originalMessages {
				originalMessagesCopy[i] = openrouter.Message{
					Role:    msg.Role,
					Content: msg.Content,
				}
			}

			// Process the message
			result, err := ParseMessages(input, originalMessages)
			if err != nil {
				t.Fatalf("ParseMessages returned error: %v", err)
			}

			// Check the result
			if len(result) != 1 {
				t.Fatalf("Expected 1 message, got %d", len(result))
			}
			if result[0].Content != test.expected {
				t.Errorf("Expected content:\n%s\n\nGot:\n%s", test.expected, result[0].Content)
			}

			// Check that the original wasn't modified
			if !reflect.DeepEqual(originalMessages, originalMessagesCopy) {
				t.Errorf("Original messages were modified")
			}
		})
	}
}

func TestComplexConditionalExample(t *testing.T) {
	// Test data matching the example in the requirement
	input := struct {
		PreviousItems   string `json:"previousItems"`
		Count           int    `json:"count"`
		GeneralCategory string `json:"generalCategory"`
		CourseLanguage  string `json:"courseLanguage"`
		EmptyPrevItems  string `json:"emptyPrevItems"`
	}{
		PreviousItems:   "1. Going to a cafe\n2. Meeting a friend",
		Count:           5,
		GeneralCategory: "daily activities",
		CourseLanguage:  "Spanish",
		EmptyPrevItems:  "",
	}

	// The example prompt from the requirement
	examplePrompt := `Our goal is to create a list of conversation topics based on the current category we are in, these topics should be related to the category, they should not be overly specific as we will make the more specific later but they should define a location and event. Doing x at y. Going to the x. Meeting an x at Y. Having a x with y. 

			These should be distinct topics from one another.
{{#if previousItems}}
The following items have already been used so do not repeat or reuse them, You must come up with new unique new topics so expand upon this list with new ideas. 
For each item you will give a short description title of the general scenario and the scenario itself which is a sentence or two that describes the scenario. 
ONLY return your NEW ideas.
{{previousItems}}
{{/if}}

Create a list of {{count}} conversation topics for {{generalCategory}}, The scenarios will be for users set in the country where most speak {{courseLanguage}} BUT DO NOT GIVE SPECIFICS ABOUT THAT COUNTRY. give broad scenarios`

	// Test with previousItems
	t.Run("With previousItems", func(t *testing.T) {
		messages := []openrouter.Message{
			{
				Role:    "user",
				Content: examplePrompt,
			},
		}

		// Process the message
		result, err := ParseMessages(input, messages)
		if err != nil {
			t.Fatalf("ParseMessages returned error: %v", err)
		}

		// The expected result should have the previousItems block included and variables replaced
		expected := `Our goal is to create a list of conversation topics based on the current category we are in, these topics should be related to the category, they should not be overly specific as we will make the more specific later but they should define a location and event. Doing x at y. Going to the x. Meeting an x at Y. Having a x with y. 

			These should be distinct topics from one another.

The following items have already been used so do not repeat or reuse them, You must come up with new unique new topics so expand upon this list with new ideas. 
For each item you will give a short description title of the general scenario and the scenario itself which is a sentence or two that describes the scenario. 
ONLY return your NEW ideas.
1. Going to a cafe
2. Meeting a friend


Create a list of 5 conversation topics for daily activities, The scenarios will be for users set in the country where most speak Spanish BUT DO NOT GIVE SPECIFICS ABOUT THAT COUNTRY. give broad scenarios`

		if result[0].Content != expected {
			t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result[0].Content)
		}
	})

	// Test with empty previousItems to ensure conditional block is removed
	t.Run("With empty previousItems", func(t *testing.T) {
		// Modify input to have empty previousItems
		inputWithEmptyPrevItems := struct {
			PreviousItems   string `json:"previousItems"`
			Count           int    `json:"count"`
			GeneralCategory string `json:"generalCategory"`
			CourseLanguage  string `json:"courseLanguage"`
		}{
			PreviousItems:   "",
			Count:           5,
			GeneralCategory: "daily activities",
			CourseLanguage:  "Spanish",
		}

		messages := []openrouter.Message{
			{
				Role:    "user",
				Content: examplePrompt,
			},
		}

		// Process the message
		result, err := ParseMessages(inputWithEmptyPrevItems, messages)
		if err != nil {
			t.Fatalf("ParseMessages returned error: %v", err)
		}

		// The expected result should have the previousItems block removed and variables replaced
		expected := `Our goal is to create a list of conversation topics based on the current category we are in, these topics should be related to the category, they should not be overly specific as we will make the more specific later but they should define a location and event. Doing x at y. Going to the x. Meeting an x at Y. Having a x with y. 

			These should be distinct topics from one another.


Create a list of 5 conversation topics for daily activities, The scenarios will be for users set in the country where most speak Spanish BUT DO NOT GIVE SPECIFICS ABOUT THAT COUNTRY. give broad scenarios`

		if result[0].Content != expected {
			t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result[0].Content)
		}
	})
}
