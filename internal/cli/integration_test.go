package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/llmang/llmango/internal/parser"
)

// TestEndToEndWorkflow tests the complete init -> generate -> validate workflow
func TestEndToEndWorkflow(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "llmango_e2e_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Step 1: Initialize project
	t.Log("Step 1: Initializing project")
	err = runInit("e2etest")
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Verify init created files
	if _, err := os.Stat("llmango.yaml"); os.IsNotExist(err) {
		t.Fatal("Init should have created llmango.yaml")
	}
	if _, err := os.Stat("example.go"); os.IsNotExist(err) {
		t.Fatal("Init should have created example.go")
	}

	// Step 2: Generate mango.go
	t.Log("Step 2: Generating mango.go")
	generateOpts := &parser.GenerateOptions{
		InputDir:    tmpDir,
		OutputFile:  filepath.Join(tmpDir, "mango.go"),
		PackageName: "e2etest",
	}

	err = runGenerate(generateOpts)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify mango.go was created
	if _, err := os.Stat("mango.go"); os.IsNotExist(err) {
		t.Fatal("Generate should have created mango.go")
	}

	// Step 3: Validate the generated project
	t.Log("Step 3: Validating project")
	validateOpts := &parser.GenerateOptions{
		InputDir: tmpDir,
		Validate: true,
	}

	err = runGenerate(validateOpts)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	// Step 4: Verify the generated mango.go content
	t.Log("Step 4: Verifying generated content")
	content, err := os.ReadFile("mango.go")
	if err != nil {
		t.Fatalf("Failed to read mango.go: %v", err)
	}

	contentStr := string(content)

	// Verify package declaration
	if !strings.Contains(contentStr, "package e2etest") {
		t.Error("Generated file should have correct package name")
	}

	// Verify imports
	expectedImports := []string{
		`"github.com/llmang/llmango/llmango"`,
		`"github.com/llmang/llmango/openrouter"`,
	}
	for _, imp := range expectedImports {
		if !strings.Contains(contentStr, imp) {
			t.Errorf("Generated file should contain import: %s", imp)
		}
	}

	// Verify struct and constructor
	if !strings.Contains(contentStr, "type Mango struct") {
		t.Error("Generated file should contain Mango struct")
	}
	if !strings.Contains(contentStr, "func CreateMango(or *openrouter.OpenRouter) (*Mango, error)") {
		t.Error("Generated file should contain CreateMango function")
	}

	// Verify generated methods (should have both config and Go-defined goals)
	if !strings.Contains(contentStr, "func (m *Mango) ExampleGoal(") {
		t.Error("Generated file should contain ExampleGoal method")
	}
	if !strings.Contains(contentStr, "func (m *Mango) ExampleGoalRaw(") {
		t.Error("Generated file should contain ExampleGoalRaw method")
	}

	t.Log("End-to-end workflow completed successfully")
}

// TestMixedSourcesWorkflow tests a project with both Go and config definitions
func TestMixedSourcesWorkflow(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "llmango_mixed_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a project with both Go and config definitions
	// First, create Go definitions
	goContent := `package mixed

import (
	"github.com/llmang/llmango/llmango"
	"github.com/llmang/llmango/openrouter"
)

type GoInput struct {
	Message string ` + "`json:\"message\"`" + `
}

type GoOutput struct {
	Response string ` + "`json:\"response\"`" + `
}

type ConfigInput struct {
	Text string ` + "`json:\"text\"`" + `
}

type ConfigOutput struct {
	Result string ` + "`json:\"result\"`" + `
}

var goGoal = llmango.Goal{
	UID:         "go-goal",
	Title:       "Go Goal",
	Description: "A goal defined in Go",
	InputOutput: llmango.InputOutput[GoInput, GoOutput]{
		InputExample: GoInput{Message: "Hello"},
		OutputExample: GoOutput{Response: "Hi"},
	},
}

var goPrompt = llmango.Prompt{
	UID:     "go-prompt",
	GoalUID: goGoal.UID,
	Model:   "openai/gpt-4",
	Messages: []openrouter.Message{
		{Role: "user", Content: "{{message}}"},
	},
}
`

	// Create YAML config
	yamlContent := `goals:
  - uid: "config-goal"
    title: "Config Goal"
    description: "A goal defined in config"
    input_type: "ConfigInput"
    output_type: "ConfigOutput"

prompts:
  - uid: "config-prompt"
    goal_uid: "config-goal"
    model: "openai/gpt-3.5-turbo"
    messages:
      - role: "user"
        content: "Process: {{text}}"
`

	// Write files
	if err := os.WriteFile(filepath.Join(tmpDir, "definitions.go"), []byte(goContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "llmango.yaml"), []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Generate mango.go
	generateOpts := &parser.GenerateOptions{
		InputDir:    tmpDir,
		OutputFile:  filepath.Join(tmpDir, "mango.go"),
		PackageName: "mixed",
	}

	err = runGenerate(generateOpts)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify generated content includes both sources
	content, err := os.ReadFile(filepath.Join(tmpDir, "mango.go"))
	if err != nil {
		t.Fatalf("Failed to read mango.go: %v", err)
	}

	contentStr := string(content)

	// Should contain methods for both Go and config goals
	if !strings.Contains(contentStr, "func (m *Mango) GoGoal(") {
		t.Error("Generated file should contain GoGoal method")
	}
	if !strings.Contains(contentStr, "func (m *Mango) ConfigGoal(") {
		t.Error("Generated file should contain ConfigGoal method")
	}

	// Should register both goals and prompts
	if !strings.Contains(contentStr, "&goGoal,") {
		t.Error("Generated file should register goGoal")
	}
	if !strings.Contains(contentStr, "&configGoalGoal,") {
		t.Error("Generated file should register config goal")
	}
}

// TestConflictResolution tests that Go definitions take priority over config
func TestConflictResolution(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "llmango_conflict_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create conflicting definitions
	goContent := `package conflict

import (
	"github.com/llmang/llmango/llmango"
	"github.com/llmang/llmango/openrouter"
)

type TestInput struct {
	Message string ` + "`json:\"message\"`" + `
}

type TestOutput struct {
	Response string ` + "`json:\"response\"`" + `
}

var conflictGoal = llmango.Goal{
	UID:         "conflict-goal",
	Title:       "Go Version",
	Description: "Go version of the goal",
	InputOutput: llmango.InputOutput[TestInput, TestOutput]{
		InputExample: TestInput{Message: "Go"},
		OutputExample: TestOutput{Response: "Go Response"},
	},
}

var conflictPrompt = llmango.Prompt{
	UID:     "conflict-prompt",
	GoalUID: conflictGoal.UID,
	Model:   "openai/gpt-4",
	Messages: []openrouter.Message{
		{Role: "user", Content: "Go: {{message}}"},
	},
}
`

	yamlContent := `goals:
  - uid: "conflict-goal"
    title: "Config Version"
    description: "Config version of the goal"
    input_type: "TestInput"
    output_type: "TestOutput"

prompts:
  - uid: "conflict-prompt"
    goal_uid: "conflict-goal"
    model: "openai/gpt-3.5-turbo"
    messages:
      - role: "user"
        content: "Config: {{message}}"
`

	// Write files
	if err := os.WriteFile(filepath.Join(tmpDir, "conflict.go"), []byte(goContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "llmango.yaml"), []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Validate (should show warnings)
	validateOpts := &parser.GenerateOptions{
		InputDir: tmpDir,
		Validate: true,
	}

	// Capture the validation - it should succeed but with warnings
	err = runGenerate(validateOpts)
	if err != nil {
		t.Fatalf("Validate should succeed with warnings, but got error: %v", err)
	}

	// Generate and verify Go definitions take priority
	generateOpts := &parser.GenerateOptions{
		InputDir:    tmpDir,
		OutputFile:  filepath.Join(tmpDir, "mango.go"),
		PackageName: "conflict",
	}

	err = runGenerate(generateOpts)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, "mango.go"))
	if err != nil {
		t.Fatalf("Failed to read mango.go: %v", err)
	}

	contentStr := string(content)

	// Should use Go definitions (register &conflictGoal, not generated config goal)
	if !strings.Contains(contentStr, "&conflictGoal,") {
		t.Error("Should register Go-defined goal")
	}
	if !strings.Contains(contentStr, "&conflictPrompt,") {
		t.Error("Should register Go-defined prompt")
	}

	// Should NOT contain config-generated variables for conflicting items
	if strings.Contains(contentStr, "var conflictGoalGoal = llmango.Goal{") {
		t.Error("Should not generate config goal variable when Go version exists")
	}
}