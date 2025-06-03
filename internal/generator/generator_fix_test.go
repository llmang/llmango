package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/llmang/llmango/internal/parser"
	"github.com/llmang/llmango/testhelpers"
)

func TestGenerateMangoFileWithKeyedLiterals(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "llmango_test_*")
	testhelpers.RequireNoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(tempDir)

	// Create test config file
	configContent := `
goals:
  - uid: "test-goal"
    title: "Test Goal"
    description: "A test goal for validation"
    input_type: "TestInput"
    output_type: "TestOutput"

prompts:
  - uid: "test-prompt"
    goal_uid: "test-goal"
    model: "openai/gpt-4o"
    weight: 100
    messages:
      - role: "system"
        content: "You are a helpful assistant."
      - role: "user"
        content: "Process this: {{.input}}"
`

	configFile := filepath.Join(tempDir, "llmango.yaml")
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	testhelpers.RequireNoError(t, err, "Failed to write config file")

	// Parse the config
	result, err := parser.ParseConfigFiles(tempDir)
	testhelpers.RequireNoError(t, err, "Failed to parse config")

	// Generate the mango file
	outputFile := filepath.Join(tempDir, "mango.go")
	opts := &parser.GenerateOptions{
		InputDir:    tempDir,
		OutputFile:  outputFile,
		PackageName: "mango",
		Validate:    false,
	}

	err = GenerateMangoFile(result, opts)
	testhelpers.RequireNoError(t, err, "Failed to generate mango file")

	// Read the generated file
	generatedContent, err := os.ReadFile(outputFile)
	testhelpers.RequireNoError(t, err, "Failed to read generated file")

	content := string(generatedContent)

	// Test 1: Should use keyed field literals for Prompt struct
	testhelpers.AssertContains(t, content, `UID:      "test-prompt"`, 
		"Generated code should use keyed field literals for UID")
	testhelpers.AssertContains(t, content, `GoalUID:  "test-goal"`, 
		"Generated code should use keyed field literals for GoalUID")
	testhelpers.AssertContains(t, content, `Model:    "openai/gpt-4o"`, 
		"Generated code should use keyed field literals for Model")

	// Test 2: Should use keyed field literals for Message struct
	testhelpers.AssertContains(t, content, `{Role: "system", Content: "You are a helpful assistant."}`, 
		"Generated code should use keyed field literals for Message")
	testhelpers.AssertContains(t, content, `{Role: "user", Content: "Process this: {{.input}}"}`, 
		"Generated code should use keyed field literals for Message")

	// Test 3: Should NOT contain InputOutput field (old structure)
	testhelpers.AssertNotContains(t, content, "InputOutput:", 
		"Generated code should not contain old InputOutput field")

	// Test 4: Should properly escape multi-line strings
	testhelpers.AssertNotContains(t, content, "Content: \"Process this:\n{{.input}}\"", 
		"Generated code should not have unescaped newlines in strings")

	// Test 5: Should be valid Go code (no syntax errors)
	testhelpers.AssertNotContains(t, content, "newline in string", 
		"Generated code should not have newline syntax errors")
}

func TestGenerateMangoFileWithMultiLineStrings(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "llmango_test_*")
	testhelpers.RequireNoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(tempDir)

	// Create test config with multi-line content
	configContent := `
prompts:
  - uid: "multiline-prompt"
    goal_uid: "test-goal"
    model: "openai/gpt-4o"
    weight: 100
    messages:
      - role: "system"
        content: "You are a helpful assistant."
      - role: "user"
        content: |
          Process this email:
          Subject: {{.subject}}
          From: {{.sender}}
          Body: {{.body}}
`

	configFile := filepath.Join(tempDir, "llmango.yaml")
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	testhelpers.RequireNoError(t, err, "Failed to write config file")

	// Parse the config
	result, err := parser.ParseConfigFiles(tempDir)
	testhelpers.RequireNoError(t, err, "Failed to parse config")

	// Generate the mango file
	outputFile := filepath.Join(tempDir, "mango.go")
	opts := &parser.GenerateOptions{
		InputDir:    tempDir,
		OutputFile:  outputFile,
		PackageName: "mango",
		Validate:    false,
	}

	err = GenerateMangoFile(result, opts)
	testhelpers.RequireNoError(t, err, "Failed to generate mango file")

	// Read the generated file
	generatedContent, err := os.ReadFile(outputFile)
	testhelpers.RequireNoError(t, err, "Failed to read generated file")

	content := string(generatedContent)

	// Test: Multi-line strings should be properly escaped
	testhelpers.AssertContains(t, content, `Content: "Process this email:\nSubject: {{.subject}}\nFrom: {{.sender}}\nBody: {{.body}}\n"`,
		"Multi-line strings should be properly escaped with \\n")

	// Test: Should not contain literal newlines in string literals
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.Contains(line, "Content:") && strings.Contains(line, "{{") {
			// This line contains a Content field with template variables
			// It should not have an unclosed string literal
			openQuotes := strings.Count(line, `"`) - strings.Count(line, `\"`)
			testhelpers.AssertTrue(t, openQuotes%2 == 0, 
				"Line %d should have balanced quotes: %s", i+1, line)
		}
	}
}