package cli

import (
	"github.com/llmang/llmango/internal/parser"
	"github.com/spf13/cobra"
)

// NewValidateCommand creates the validate command
func NewValidateCommand() *cobra.Command {
	var opts parser.GenerateOptions

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate goal and prompt definitions without generating code",
		Long: `Validate scans the current directory for LLM goal and prompt definitions
and validates them for correctness without generating any code.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Validate = true
			return runGenerate(&opts)
		},
	}

	cmd.Flags().StringVarP(&opts.InputDir, "input", "i", ".", "Input directory to scan for goals and prompts")
	cmd.Flags().StringVarP(&opts.ConfigFile, "config", "c", "", "Specific config file to use (optional)")

	return cmd
}