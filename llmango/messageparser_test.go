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
	// Add new field for testing insertMessages
	InsertMessages string `json:"insertMessages"`
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
		{
			name: "If true with else",
			message: `Test {{#if nonEmptyString}}
If Block
{{:else}}
Else Block
{{/if}} End`,
			expected: `Test 
If Block
 End`,
		},
		{
			name: "If false (empty string) with else",
			message: `Test {{#if emptyString}}
If Block
{{:else}}
Else Block
{{/if}} End`,
			expected: `Test 
Else Block
 End`,
		},
		{
			name: "If false (non-existent var) with else",
			message: `Test {{#if nonExistentVar}}
If Block
{{:else}}
Else Block
{{/if}} End`,
			expected: `Test 
Else Block
 End`,
		},
		{
			name:     "If true without else",
			message:  `Test {{#if nonEmptyString}}If Block{{/if}} End`,
			expected: `Test If Block End`,
		},
		{
			name:     "If false (empty string) without else",
			message:  `Test {{#if emptyString}}If Block{{/if}} End`,
			expected: `Test  End`, // Content removed, space remains
		},
		{
			name:     "If false (non-existent var) without else",
			message:  `Test {{#if nonExistentVar}}If Block{{/if}} End`,
			expected: `Test  End`, // Content removed, space remains
		},
		{
			name:     "Multiple if/else blocks",
			message:  `{{#if name}}Name: {{name}}{{:else}}No Name{{/if}} {{#if emptyString}}Empty{{:else}}Not Empty{{/if}}`,
			expected: `Name: {{name}} Not Empty`,
		},
		{
			name:     "Malformed else (ignored)", // Current regex won't match this structure correctly, treated as no-else
			message:  `Test {{#if emptyString}}If Block {{:else malformed}} Else Block{{/if}} End`,
			expected: `Test  End`, // behaves as if no else block was present
		},
		{
			name:     "Else outside if (ignored)",
			message:  `Test {{:else}} Orphaned Else {{/if}} End`,
			expected: `Test {{:else}} Orphaned Else {{/if}} End`,
		},
		{
			name:     "Zero value treated as truthy",
			message:  `Test {{#if zeroValue}}Zero If{{:else}}Zero Else{{/if}} End`,
			expected: `Test Zero If End`,
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
		{
			name:     "If true with else and variables",
			message:  `{{#if name}}Hello {{name}}{{:else}}No name provided{{/if}}. You are {{age}}.`,
			expected: `Hello John. You are 30.`,
		},
		{
			name:     "If false with else and variables",
			message:  `{{#if emptyString}}Got empty: {{emptyString}}{{:else}}Not empty: {{nonEmptyString}}{{/if}}`,
			expected: `Not empty: value`,
		},
		{
			name:     "If false (non-existent) with else and variables",
			message:  `{{#if nonExistent}}If: {{nonExistent}}{{:else}}Else: {{name}}{{/if}}`,
			expected: `Else: John`,
		},
		{
			name:     "Variable substitution only in kept block (if)",
			message:  `{{#if name}}Name: {{name}}, Age: {{age}}{{:else}}Else with {{age}}{{/if}}`,
			expected: `Name: John, Age: 30`,
		},
		{
			name:     "Variable substitution only in kept block (else)",
			message:  `{{#if emptyString}}If with {{name}}{{:else}}Else: {{nonEmptyString}}, Age: {{age}}{{/if}}`,
			expected: `Else: value, Age: 30`,
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

func TestComplexConditionalWithElseExample(t *testing.T) {
	// Test data
	inputWithItems := struct {
		PreviousItems   string `json:"previousItems"`
		Count           int    `json:"count"`
		GeneralCategory string `json:"generalCategory"`
		CourseLanguage  string `json:"courseLanguage"`
	}{
		PreviousItems:   "1. Item A\n2. Item B",
		Count:           3,
		GeneralCategory: "Test Category",
		CourseLanguage:  "Klingon",
	}
	inputWithoutItems := struct {
		PreviousItems   string `json:"previousItems"`
		Count           int    `json:"count"`
		GeneralCategory string `json:"generalCategory"`
		CourseLanguage  string `json:"courseLanguage"`
	}{
		PreviousItems:   "", // Empty
		Count:           3,
		GeneralCategory: "Test Category",
		CourseLanguage:  "Klingon",
	}

	// Modified prompt using if/else
	promptWithElse := `Generate {{count}} topics for {{generalCategory}} (language: {{courseLanguage}}).
{{#if previousItems}}
Avoid these used topics:
{{previousItems}}
{{:else}}
This is the first set of topics for this category.
{{/if}}
Ensure topics are unique.`

	messages := []openrouter.Message{{Role: "user", Content: promptWithElse}}

	// Test with previousItems (if block should be kept)
	t.Run("With previousItems and else", func(t *testing.T) {
		result, err := ParseMessages(inputWithItems, messages)
		if err != nil {
			t.Fatalf("ParseMessages returned error: %v", err)
		}
		expected := `Generate 3 topics for Test Category (language: Klingon).

Avoid these used topics:
1. Item A
2. Item B

Ensure topics are unique.`
		if result[0].Content != expected {
			t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result[0].Content)
		}
	})

	// Test without previousItems (else block should be kept)
	t.Run("Without previousItems and else", func(t *testing.T) {
		result, err := ParseMessages(inputWithoutItems, messages)
		if err != nil {
			t.Fatalf("ParseMessages returned error: %v", err)
		}
		expected := `Generate 3 topics for Test Category (language: Klingon).

This is the first set of topics for this category.

Ensure topics are unique.`
		if result[0].Content != expected {
			t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result[0].Content)
		}
	})
}

func TestInsertMessagesFeature(t *testing.T) {
	// Test input with a field containing valid JSON messages array
	validMessagesJSON := `[{"role":"user","content":"Hello there"},{"role":"assistant","content":"Hi, how can I help?"}]`

	// Test input with invalid JSON
	invalidMessagesJSON := `not valid json`

	// Test input with valid JSON but invalid message format
	invalidFormatJSON := `[{"role":"invalid_role","content":"Test"},{"role":"user","content":""}]`

	// Test inputs
	tests := []struct {
		name           string
		input          interface{}
		messages       []openrouter.Message
		expectedResult []openrouter.Message
	}{
		{
			name: "Basic insertion - exact match",
			input: struct {
				InsertMessages string `json:"insertMessages"`
			}{
				InsertMessages: validMessagesJSON,
			},
			messages: []openrouter.Message{
				{Role: "user", Content: "{{insertMessages}}"},
			},
			expectedResult: []openrouter.Message{
				{Role: "user", Content: "Hello there"},
				{Role: "assistant", Content: "Hi, how can I help?"},
			},
		},
		{
			name: "Mixed scenario - messages before and after",
			input: struct {
				InsertMessages string `json:"insertMessages"`
			}{
				InsertMessages: validMessagesJSON,
			},
			messages: []openrouter.Message{
				{Role: "user", Content: "First message"},
				{Role: "assistant", Content: "{{insertMessages}}"},
				{Role: "user", Content: "Last message"},
			},
			expectedResult: []openrouter.Message{
				{Role: "user", Content: "First message"},
				{Role: "user", Content: "Hello there"},
				{Role: "assistant", Content: "Hi, how can I help?"},
				{Role: "user", Content: "Last message"},
			},
		},
		{
			name: "Not exact match - process as variable",
			input: struct {
				InsertMessages string `json:"insertMessages"`
			}{
				InsertMessages: validMessagesJSON,
			},
			messages: []openrouter.Message{
				{Role: "user", Content: "This is {{insertMessages}} as part of text"},
			},
			expectedResult: []openrouter.Message{
				{Role: "user", Content: "This is " + validMessagesJSON + " as part of text"},
			},
		},
		{
			name: "Invalid JSON - fallback to string",
			input: struct {
				InsertMessages string `json:"insertMessages"`
			}{
				InsertMessages: invalidMessagesJSON,
			},
			messages: []openrouter.Message{
				{Role: "user", Content: "{{insertMessages}}"},
			},
			expectedResult: []openrouter.Message{
				{Role: "user", Content: invalidMessagesJSON},
			},
		},
		{
			name: "Invalid message format - fallback to string",
			input: struct {
				InsertMessages string `json:"insertMessages"`
			}{
				InsertMessages: invalidFormatJSON,
			},
			messages: []openrouter.Message{
				{Role: "user", Content: "{{insertMessages}}"},
			},
			expectedResult: []openrouter.Message{
				{Role: "user", Content: invalidFormatJSON},
			},
		},
		{
			name: "Empty messages array - preserve original message",
			input: struct {
				InsertMessages string `json:"insertMessages"`
			}{
				InsertMessages: "[]",
			},
			messages: []openrouter.Message{
				{Role: "user", Content: "{{insertMessages}}"},
			},
			expectedResult: []openrouter.Message{
				{Role: "user", Content: "[]"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create a copy of original messages for comparison
			originalMessages := make([]openrouter.Message, len(test.messages))
			for i, msg := range test.messages {
				originalMessages[i] = openrouter.Message{
					Role:    msg.Role,
					Content: msg.Content,
				}
			}

			// Process the messages
			result, err := ParseMessages(test.input, test.messages)
			if err != nil {
				t.Fatalf("ParseMessages returned error: %v", err)
			}

			// Check the result length
			if len(result) != len(test.expectedResult) {
				t.Fatalf("Expected %d messages, got %d", len(test.expectedResult), len(result))
			}

			// Check each message in the result
			for i, msg := range result {
				expected := test.expectedResult[i]
				if msg.Role != expected.Role || msg.Content != expected.Content {
					t.Errorf("Message %d: expected {%s, %s}, got {%s, %s}",
						i, expected.Role, expected.Content, msg.Role, msg.Content)
				}
			}

			// Check that original messages weren't modified
			for i, msg := range test.messages {
				if msg.Role != originalMessages[i].Role || msg.Content != originalMessages[i].Content {
					t.Errorf("Original message %d was modified", i)
				}
			}
		})
	}
}
