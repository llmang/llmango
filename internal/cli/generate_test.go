package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/llmang/llmango/internal/parser"
)

func TestRunGenerate(t *testing.T) {
	tests := []struct {
		name        string
		setupDir    func(string) error
		opts        *parser.GenerateOptions
		expectError bool
		validate    func(string, *testing.T)
	}{
		{
			name: "generate from go files",
			setupDir: func(dir string) error {
				// Copy test files
				return copyTestProject("../testdata/valid_projects/go_only", dir)
			},
			opts: &parser.GenerateOptions{
				InputDir:    "",
				OutputFile:  "mango.go",
				PackageName: "mango",
			},
			expectError: false,
			validate: func(dir string, t *testing.T) {
				content, err := os.ReadFile(filepath.Join(dir, "mango.go"))
				if err != nil {
					t.Fatalf("Failed to read generated file: %v", err)
				}
				contentStr := string(content)
				if !strings.Contains(contentStr, "func (m *Mango) TestGoal") {
					t.Error("Generated file should contain TestGoal method")
				}
			},
		},
		{
			name: "generate from config files",
			setupDir: func(dir string) error {
				return copyTestProject("../testdata/valid_projects/config_only", dir)
			},
			opts: &parser.GenerateOptions{
				InputDir:    "",
				OutputFile:  "mango.go",
				PackageName: "mango",
			},
			expectError: false,
			validate: func(dir string, t *testing.T) {
				content, err := os.ReadFile(filepath.Join(dir, "mango.go"))
				if err != nil {
					t.Fatalf("Failed to read generated file: %v", err)
				}
				contentStr := string(content)
				if !strings.Contains(contentStr, "func (m *Mango) ConfigGoal") {
					t.Error("Generated file should contain ConfigGoal method")
				}
			},
		},
		{
			name: "validate only mode",
			setupDir: func(dir string) error {
				return copyTestProject("../testdata/valid_projects/go_only", dir)
			},
			opts: &parser.GenerateOptions{
				InputDir:    "",
				OutputFile:  "mango.go",
				PackageName: "mango",
				Validate:    true,
			},
			expectError: false,
			validate: func(dir string, t *testing.T) {
				// Should not create mango.go in validate mode
				if _, err := os.Stat(filepath.Join(dir, "mango.go")); err == nil {
					t.Error("mango.go should not be created in validate mode")
				}
			},
		},
		{
			name: "invalid syntax should error",
			setupDir: func(dir string) error {
				return copyTestProject("../testdata/invalid_projects/syntax_errors", dir)
			},
			opts: &parser.GenerateOptions{
				InputDir:    "",
				OutputFile:  "mango.go",
				PackageName: "mango",
			},
			expectError: true,
			validate:    func(dir string, t *testing.T) {}, // No validation needed for error case
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

			// Setup test directory
			if err := tt.setupDir(tmpDir); err != nil {
				t.Fatalf("Failed to setup test directory: %v", err)
			}

			// Update options with correct paths
			opts := *tt.opts
			opts.InputDir = tmpDir
			opts.OutputFile = filepath.Join(tmpDir, filepath.Base(opts.OutputFile))

			// Run generate
			err = runGenerate(&opts)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Run validation
			tt.validate(tmpDir, t)
		})
	}
}

func TestRunGenerateNonExistentDirectory(t *testing.T) {
	opts := &parser.GenerateOptions{
		InputDir:    "/non/existent/directory",
		OutputFile:  "mango.go",
		PackageName: "mango",
	}

	err := runGenerate(opts)
	if err == nil {
		t.Error("expected error for non-existent directory")
	}

	if !strings.Contains(err.Error(), "input directory does not exist") {
		t.Errorf("expected specific error message, got: %v", err)
	}
}

func TestRunGenerateCustomPackageName(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "llmango_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Setup test directory
	if err := copyTestProject("../testdata/valid_projects/go_only", tmpDir); err != nil {
		t.Fatalf("Failed to setup test directory: %v", err)
	}

	opts := &parser.GenerateOptions{
		InputDir:    tmpDir,
		OutputFile:  filepath.Join(tmpDir, "custom.go"),
		PackageName: "custompackage",
	}

	err = runGenerate(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify custom package name
	content, err := os.ReadFile(opts.OutputFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "package custompackage") {
		t.Error("Generated file should contain custom package name")
	}
}

// Helper function to copy test projects
func copyTestProject(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		// Copy file
		srcFile, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(dstPath, srcFile, info.Mode())
	})
}

func TestRunGenerateOutputDirectoryCreation(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "llmango_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Setup test directory
	if err := copyTestProject("../testdata/valid_projects/go_only", tmpDir); err != nil {
		t.Fatalf("Failed to setup test directory: %v", err)
	}

	// Use nested output directory that doesn't exist
	outputDir := filepath.Join(tmpDir, "nested", "output", "dir")
	outputFile := filepath.Join(outputDir, "mango.go")

	opts := &parser.GenerateOptions{
		InputDir:    tmpDir,
		OutputFile:  outputFile,
		PackageName: "mango",
	}

	err = runGenerate(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify output directory was created
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		t.Error("output directory should have been created")
	}

	// Verify file was created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("output file should have been created")
	}
}