# LLMango CLI Tool

The LLMango CLI tool generates type-safe Go functions from LLM goals and prompts, similar to how SQLC generates database functions from SQL queries.

## Installation

Build the CLI tool from source:

```bash
go build -o llmango-cli ./cmd/llmango
```

## Quick Start

1. **Initialize a new project:**
   ```bash
   llmango-cli init
   ```
   This creates:
   - `llmango.yaml` - Configuration file with example goal and prompt
   - `example.go` - Go file with example goal and prompt definitions

2. **Generate type-safe functions:**
   ```bash
   llmango-cli generate
   ```
   This creates `mango.go` with type-safe wrapper functions.

3. **Validate your definitions:**
   ```bash
   llmango-cli validate
   ```

## Commands

### `llmango-cli init`

Initialize a new LLMango project with example files.

**Flags:**
- `--package, -p`: Package name for generated code (default: "mango")

**Example:**
```bash
llmango-cli init --package myapp
```

### `llmango-cli generate`

Generate type-safe LLM functions from goals and prompts.

**Flags:**
- `--input, -i`: Input directory to scan (default: ".")
- `--output, -o`: Output file path (default: "mango.go")
- `--config, -c`: Specific config file to use (optional)
- `--package, -p`: Package name for generated code (default: "mango")
- `--validate`: Validate only, don't generate code

**Example:**
```bash
llmango-cli generate --output internal/mango/mango.go --package mango
```

### `llmango-cli validate`

Validate goal and prompt definitions without generating code.

**Flags:**
- `--input, -i`: Input directory to scan (default: ".")
- `--config, -c`: Specific config file to use (optional)

## Configuration

### YAML Configuration (`llmango.yaml`)

```yaml
goals:
  - uid: "summarize-text"
    title: "Summarize Text"
    description: "Summarizes long text into key points"
    input_type: "SummarizeInput"
    output_type: "SummarizeOutput"

prompts:
  - uid: "summarize-prompt-v1"
    goal_uid: "summarize-text"
    model: "openai/gpt-4"
    weight: 100
    messages:
      - role: "system"
        content: "You are a helpful assistant that summarizes text."
      - role: "user"
        content: "Summarize this text: {{text}}"
```

### Go Definitions

```go
package mango

import (
    "github.com/llmang/llmango/llmango"
    "github.com/llmang/llmango/openrouter"
)

type SummarizeInput struct {
    Text string `json:"text"`
}

type SummarizeOutput struct {
    Summary string `json:"summary"`
}

var summarizeGoal = llmango.Goal{
    UID:         "summarize-text",
    Title:       "Summarize Text",
    Description: "Summarizes long text into key points",
    InputOutput: llmango.InputOutput[SummarizeInput, SummarizeOutput]{
        InputExample: SummarizeInput{
            Text: "Long text to summarize...",
        },
        OutputExample: SummarizeOutput{
            Summary: "Key points from the text",
        },
    },
}

var summarizePrompt = llmango.Prompt{
    UID:     "summarize-prompt-v1",
    GoalUID: summarizeGoal.UID,
    Model:   "openai/gpt-4",
    Weight:  100,
    Messages: []openrouter.Message{
        {Role: "system", Content: "You are a helpful assistant that summarizes text."},
        {Role: "user", Content: "Summarize this text: {{text}}"},
    },
}
```

## Generated Code

The CLI generates a `mango.go` file with:

1. **Type-safe wrapper struct:**
   ```go
   type Mango struct {
       *llmango.LLMangoManager
   }
   ```

2. **Constructor function:**
   ```go
   func CreateMango(or *openrouter.OpenRouter) (*Mango, error)
   ```

3. **Type-safe methods for each goal:**
   ```go
   func (m *Mango) SummarizeText(input *SummarizeInput) (*SummarizeOutput, error)
   func (m *Mango) SummarizeTextRaw(input *SummarizeInput) (*SummarizeOutput, *openrouter.NonStreamingChatResponse, error)
   ```

## Usage in Your Application

```go
package main

import (
    "fmt"
    "log"
    
    "your-app/internal/mango"
    "github.com/llmang/llmango/openrouter"
)

func main() {
    // Initialize OpenRouter
    or := openrouter.NewOpenRouter("your-api-key")
    
    // Create Mango instance
    m, err := mango.CreateMango(or)
    if err != nil {
        log.Fatal(err)
    }
    
    // Use type-safe functions
    input := &mango.SummarizeInput{
        Text: "This is a long text that needs to be summarized...",
    }
    
    output, err := m.SummarizeText(input)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Summary:", output.Summary)
}
```

## Features

- **Hybrid Discovery**: Scans both Go files and YAML/JSON config files
- **Type Safety**: Generates compile-time safe functions
- **Conflict Resolution**: Go definitions take priority over config files
- **A/B Testing**: Support for weighted prompts and canary testing
- **Validation**: Comprehensive validation of goals and prompts
- **SQLC-like DX**: Familiar workflow for Go developers

## Best Practices

1. **Use descriptive UIDs**: Make goal and prompt UIDs clear and unique
2. **Version your prompts**: Use versioned UIDs like `"goal-v1"`, `"goal-v2"`
3. **Validate regularly**: Run `validate` command before generating
4. **Separate concerns**: Keep goals focused on single responsibilities
5. **Document examples**: Provide clear input/output examples

## Troubleshooting

### Common Issues

1. **Parse errors**: Ensure Go files have valid syntax and proper package declarations
2. **Missing types**: Make sure input/output types are defined before goals
3. **UID conflicts**: Use unique UIDs across all goals and prompts
4. **Import errors**: Ensure all required imports are available

### Debug Tips

- Use `--validate` flag to check definitions without generating code
- Check generated `mango.go` for any syntax issues
- Verify that all referenced types are properly imported