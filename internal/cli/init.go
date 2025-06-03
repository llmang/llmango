package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// NewInitCommand creates the init command
func NewInitCommand() *cobra.Command {
	var packageName string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new LLMango project",
		Long: `Init creates a new LLMango project with example configuration files
and directory structure.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(packageName)
		},
	}

	cmd.Flags().StringVarP(&packageName, "package", "p", "mango", "Package name for the generated code")

	return cmd
}

// runInit executes the init command
func runInit(packageName string) error {
	// Create example config file
	configContent := `# LLMango Configuration
goals:
  - uid: "example-goal"
    title: "Example Goal"
    description: "An example goal for demonstration"
    input_type: "ExampleInput"
    output_type: "ExampleOutput"

prompts:
  - uid: "example-prompt"
    goal_uid: "example-goal"
    model: "openai/gpt-4"
    weight: 100
    messages:
      - role: "system"
        content: "You are a helpful assistant."
      - role: "user"
        content: "{{userMessage}}"
`

	// Create example Go file
	goContent := fmt.Sprintf(`package %s

import (
	"github.com/llmang/llmango/llmango"
	"github.com/llmang/llmango/openrouter"
)

// Example input and output types
type ExampleInput struct {
	UserMessage string ` + "`json:\"userMessage\"`" + `
}

type ExampleOutput struct {
	Response string ` + "`json:\"response\"`" + `
}

// Example goal definition
var exampleGoal = llmango.Goal{
	UID:         "example-goal",
	Title:       "Example Goal",
	Description: "An example goal for demonstration",
	InputOutput: llmango.InputOutput[ExampleInput, ExampleOutput]{
		InputExample: ExampleInput{
			UserMessage: "Hello, how are you?",
		},
		OutputExample: ExampleOutput{
			Response: "I'm doing well, thank you!",
		},
	},
}

// Example prompt definition
var examplePrompt = llmango.Prompt{
	UID:     "example-prompt",
	GoalUID: exampleGoal.UID,
	Model:   "openai/gpt-4",
	Weight:  100,
	Messages: []openrouter.Message{
		{Role: "system", Content: "You are a helpful assistant."},
		{Role: "user", Content: "{{userMessage}}"},
	},
}
`, packageName)

	// Write config file
	if err := os.WriteFile("llmango.yaml", []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to create llmango.yaml: %w", err)
	}

	// Write example Go file
	exampleFile := filepath.Join("example.go")
	if err := os.WriteFile(exampleFile, []byte(goContent), 0644); err != nil {
		return fmt.Errorf("failed to create example.go: %w", err)
	}

	fmt.Println("LLMango project initialized successfully!")
	fmt.Println("Created files:")
	fmt.Println("  - llmango.yaml (configuration file)")
	fmt.Println("  - example.go (example goal and prompt definitions)")
	fmt.Println("")
	fmt.Println("Next steps:")
	fmt.Printf("  1. Run 'llmango generate' to generate %s.go\n", packageName)
	fmt.Println("  2. Customize your goals and prompts")
	fmt.Println("  3. Integrate the generated code into your application")

	return nil
}