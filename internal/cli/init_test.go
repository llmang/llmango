package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunInit(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		validate    func(string, *testing.T)
	}{
		{
			name:        "default package name",
			packageName: "mango",
			validate: func(dir string, t *testing.T) {
				validateInitFiles(dir, "mango", t)
			},
		},
		{
			name:        "custom package name",
			packageName: "custompackage",
			validate: func(dir string, t *testing.T) {
				validateInitFiles(dir, "custompackage", t)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tmpDir, err := os.MkdirTemp("", "llmango_test")
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

			// Run init
			err = runInit(tt.packageName)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Run validation
			tt.validate(tmpDir, t)
		})
	}
}

func validateInitFiles(dir, packageName string, t *testing.T) {
	// Check that llmango.yaml was created
	yamlFile := filepath.Join(dir, "llmango.yaml")
	if _, err := os.Stat(yamlFile); os.IsNotExist(err) {
		t.Error("llmango.yaml should have been created")
	}

	// Check that example.go was created
	goFile := filepath.Join(dir, "example.go")
	if _, err := os.Stat(goFile); os.IsNotExist(err) {
		t.Error("example.go should have been created")
	}

	// Validate YAML content
	yamlContent, err := os.ReadFile(yamlFile)
	if err != nil {
		t.Fatalf("Failed to read llmango.yaml: %v", err)
	}

	yamlStr := string(yamlContent)
	if !strings.Contains(yamlStr, "goals:") {
		t.Error("llmango.yaml should contain goals section")
	}

	if !strings.Contains(yamlStr, "prompts:") {
		t.Error("llmango.yaml should contain prompts section")
	}

	if !strings.Contains(yamlStr, "example-goal") {
		t.Error("llmango.yaml should contain example goal")
	}

	if !strings.Contains(yamlStr, "example-prompt") {
		t.Error("llmango.yaml should contain example prompt")
	}

	// Validate Go file content
	goContent, err := os.ReadFile(goFile)
	if err != nil {
		t.Fatalf("Failed to read example.go: %v", err)
	}

	goStr := string(goContent)
	expectedPackage := "package " + packageName
	if !strings.Contains(goStr, expectedPackage) {
		t.Errorf("example.go should contain '%s'", expectedPackage)
	}

	if !strings.Contains(goStr, "type ExampleInput struct") {
		t.Error("example.go should contain ExampleInput type")
	}

	if !strings.Contains(goStr, "type ExampleOutput struct") {
		t.Error("example.go should contain ExampleOutput type")
	}

	if !strings.Contains(goStr, "var exampleGoal = llmango.Goal{") {
		t.Error("example.go should contain example goal definition")
	}

	if !strings.Contains(goStr, "var examplePrompt = llmango.Prompt{") {
		t.Error("example.go should contain example prompt definition")
	}

	if !strings.Contains(goStr, `UID:         "example-goal"`) {
		t.Error("example.go should contain correct goal UID")
	}

	if !strings.Contains(goStr, `GoalUID: exampleGoal.UID`) {
		t.Error("example.go should reference goal UID in prompt")
	}
}

func TestRunInitInExistingProject(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "llmango_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create existing files
	existingYAML := filepath.Join(tmpDir, "llmango.yaml")
	existingGo := filepath.Join(tmpDir, "example.go")

	if err := os.WriteFile(existingYAML, []byte("existing content"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(existingGo, []byte("existing content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Change to temp directory
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Run init (should overwrite existing files)
	err = runInit("mango")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify files were overwritten
	yamlContent, err := os.ReadFile(existingYAML)
	if err != nil {
		t.Fatal(err)
	}

	if string(yamlContent) == "existing content" {
		t.Error("llmango.yaml should have been overwritten")
	}

	goContent, err := os.ReadFile(existingGo)
	if err != nil {
		t.Fatal(err)
	}

	if string(goContent) == "existing content" {
		t.Error("example.go should have been overwritten")
	}
}

func TestRunInitValidatesGeneratedFiles(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "llmango_test")
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

	// Run init
	err = runInit("testpackage")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Try to parse the generated files to ensure they're valid
	// This is a basic smoke test to ensure the generated content is syntactically correct

	// Test YAML parsing
	yamlContent, err := os.ReadFile("llmango.yaml")
	if err != nil {
		t.Fatal(err)
	}

	// Basic YAML validation - should contain required fields
	yamlStr := string(yamlContent)
	requiredYAMLFields := []string{"uid:", "title:", "description:", "input_type:", "output_type:", "goal_uid:", "model:", "messages:"}
	for _, field := range requiredYAMLFields {
		if !strings.Contains(yamlStr, field) {
			t.Errorf("Generated YAML should contain field: %s", field)
		}
	}

	// Test Go file structure
	goContent, err := os.ReadFile("example.go")
	if err != nil {
		t.Fatal(err)
	}

	goStr := string(goContent)
	requiredGoElements := []string{
		"package testpackage",
		"import (",
		"github.com/llmang/llmango/llmango",
		"github.com/llmang/llmango/openrouter",
		"type ExampleInput struct",
		"type ExampleOutput struct",
		"var exampleGoal = llmango.Goal{",
		"var examplePrompt = llmango.Prompt{",
		"InputOutput: llmango.InputOutput[ExampleInput, ExampleOutput]{",
		"Messages: []openrouter.Message{",
	}

	for _, element := range requiredGoElements {
		if !strings.Contains(goStr, element) {
			t.Errorf("Generated Go file should contain: %s", element)
		}
	}
}