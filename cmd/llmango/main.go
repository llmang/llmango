package main

import (
	"fmt"
	"os"

	"github.com/llmang/llmango/internal/cli"
	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	rootCmd := &cobra.Command{
		Use:   "llmango",
		Short: "LLMango CLI tool for generating type-safe LLM functions",
		Long: `LLMango CLI generates type-safe Go functions from LLM goals and prompts,
similar to how SQLC generates database functions from SQL queries.

The tool scans Go files and configuration files to discover goal and prompt
definitions, then generates a complete mango.go file with type-safe wrapper functions.`,
		Version: version,
	}

	// Add subcommands
	rootCmd.AddCommand(cli.NewGenerateCommand())
	rootCmd.AddCommand(cli.NewValidateCommand())
	rootCmd.AddCommand(cli.NewInitCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
