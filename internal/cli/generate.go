package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/llmang/llmango/internal/generator"
	"github.com/llmang/llmango/internal/parser"
	"github.com/spf13/cobra"
)

// NewGenerateCommand creates the generate command
func NewGenerateCommand() *cobra.Command {
	var opts parser.GenerateOptions

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate type-safe LLM functions from goals and prompts",
		Long: `Generate scans the current directory for LLM goal and prompt definitions
in Go files and configuration files, then generates a mango.go file with
type-safe wrapper functions.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenerate(&opts)
		},
	}

	cmd.Flags().StringVarP(&opts.InputDir, "input", "i", "", "Input directory to scan for goals and prompts (default: mango if exists, otherwise current directory)")
	cmd.Flags().StringVarP(&opts.OutputFile, "output", "o", "", "Output file path (default: mango/mango.go if mango dir exists, otherwise mango.go)")
	cmd.Flags().StringVarP(&opts.ConfigFile, "config", "c", "", "Specific config file to use (optional)")
	cmd.Flags().StringVarP(&opts.PackageName, "package", "p", "mango", "Package name for generated code")
	cmd.Flags().StringVar(&opts.GoSourceDir, "go-source", "", "Directory to scan for Go goal definitions (default: same as input dir)")
	cmd.Flags().BoolVar(&opts.Validate, "validate", false, "Validate only, don't generate code")

	return cmd
}

// runGenerate executes the generate command
func runGenerate(opts *parser.GenerateOptions) error {
	// Set smart defaults based on project structure
	if err := setSmartDefaults(opts); err != nil {
		return err
	}

	// Ensure input directory exists
	if _, err := os.Stat(opts.InputDir); os.IsNotExist(err) {
		return fmt.Errorf("input directory does not exist: %s", opts.InputDir)
	}

	// Determine Go source directory
	goSourceDir := opts.GoSourceDir
	if goSourceDir == "" {
		goSourceDir = opts.InputDir
	}

	// Parse Go files (exclude the output file to avoid conflicts)
	fmt.Printf("Scanning Go files in %s...\n", goSourceDir)
	
	// Always exclude the output file basename to prevent scanning generated code
	excludeFiles := []string{filepath.Base(opts.OutputFile)}
	fmt.Printf("Excluding files: %v\n", excludeFiles)
	
	goResult, err := parser.ParseGoFilesWithExclusions(goSourceDir, excludeFiles)
	if err != nil {
		return fmt.Errorf("failed to parse Go files: %w", err)
	}

	// Parse config files
	fmt.Printf("Scanning config files in %s...\n", opts.InputDir)
	configResult, err := parser.ParseConfigFiles(opts.InputDir)
	if err != nil {
		return fmt.Errorf("failed to parse config files: %w", err)
	}

	// Merge results
	result := parser.MergeResults(goResult, configResult)

	// Print summary
	fmt.Printf("Found %d goals and %d prompts\n", len(result.Goals), len(result.Prompts))

	// Print errors and warnings
	errorCount := 0
	warningCount := 0
	for _, parseErr := range result.Errors {
		if parseErr.Type == "error" {
			fmt.Printf("ERROR: %s: %s\n", parseErr.File, parseErr.Message)
			errorCount++
		} else {
			fmt.Printf("WARNING: %s: %s\n", parseErr.File, parseErr.Message)
			warningCount++
		}
	}

	if errorCount > 0 {
		return fmt.Errorf("found %d errors, cannot generate code", errorCount)
	}

	if warningCount > 0 {
		fmt.Printf("Found %d warnings\n", warningCount)
	}

	// Validate mode - don't generate code
	if opts.Validate {
		fmt.Println("Validation completed successfully")
		return nil
	}

	// Generate code
	fmt.Printf("Generating code to %s...\n", opts.OutputFile)

	// Ensure output directory exists
	outputDir := filepath.Dir(opts.OutputFile)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate the mango.go file
	if err := generator.GenerateMangoFile(result, opts); err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	fmt.Printf("Successfully generated %s\n", opts.OutputFile)
	return nil
}

// setSmartDefaults sets intelligent defaults based on project structure
func setSmartDefaults(opts *parser.GenerateOptions) error {
	// Set input directory default
	if opts.InputDir == "" {
		opts.InputDir = "." // Always scan current directory for config files
	}

	// Set Go source directory default - check mango directories first
	if opts.GoSourceDir == "" {
		mangoDirs := []string{"mango", "internal/mango", "./mango", "./internal/mango"}
		found := false
		for _, dir := range mangoDirs {
			if _, err := os.Stat(dir); err == nil {
				opts.GoSourceDir = dir
				found = true
				break
			}
		}
		if !found {
			opts.GoSourceDir = "." // Fall back to current directory
		}
	}

	// Set output file default
	if opts.OutputFile == "" {
		// Check for mango directories for output location
		if _, err := os.Stat("mango"); err == nil {
			opts.OutputFile = "mango/mango.go"
		} else if _, err := os.Stat("internal/mango"); err == nil {
			opts.OutputFile = "internal/mango/mango.go"
		} else {
			opts.OutputFile = "mango.go"
		}
	}

	return nil
}
