USE AT OWN RISK!!!!! VALIDATE ALL CODE BEFORE USAGE!!!
# LLMango
[llmang.com](https://llmang.com)

Hardcode Goals, Not Prompts

Goal-driven LLM framework for Go with universal model compatibility. By [Carson](https://carsho.dev).


## CLI TOOL install
go get -tool github.com/llmang/llmango/cmd/llmango


## What It Does

Organizes LLM queries into Goals and Prompts. Focus on outcomes, not prompts.

- **Goals**: Define the desired result and structure
- **Prompts**: Model-specific inputs with A/B testing support
- **Universal Compatibility**: Works with any LLM, structured output or not

## Architecture

LLMango consists of several integrated packages:

### Core Packages ✅
- [`llmango/`](llmango/) - Core goal-driven framework with dual-mode architecture
- [`openrouter/`](openrouter/) - Universal LLM integration with dual-path execution
- [`cmd/`](cmd/) - CLI tool for generating type-safe functions

### Supporting Packages ✅
- [`llmangofrontend/`](llmangofrontend/) - Web UI for goal and prompt management
- [`llmangologger/`](llmangologger/) - Comprehensive logging with SQLite storage
- [`llmangosavestate/`](llmangosavestate/) - Persistent state management
- [`internal/`](internal/) - CLI implementation and code generation

### In Development 🚧
- [`llmangoagents/`](llmangoagents/) - Multi-agent system with tool support

## Key Features

- **Dual-Mode Goals**: Create goals using Go structs OR pure JSON
- **Universal LLM Support**: Automatic routing for structured/non-structured models
- **Type-Safe Generation**: CLI generates compile-time safe functions
- **Advanced Templating**: Variable replacement, conditionals, message insertion
- **Web Interface**: Optional frontend for goal/prompt management
- **Comprehensive Logging**: SQLite-based execution tracking

## Message Parsing

LLMango includes a powerful message parsing system that supports:

### Variable Replacements

Use `{{variableName}}` syntax to insert values from your input structs:

```go
type MyInput struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}

// This will replace {{name}} with "John" and {{age}} with "30"
prompt := "Hello {{name}}, you are {{age}} years old."
```

Variable names are matched against the JSON tags of your input struct fields.

### Message Insertion

Use `{{insertMessages}}` as the ***EXACT*** content of a message to insert multiple messages from your input:

```go
type MyInput struct {
    InsertMessages string `json:"insertMessages"`
}

// In your code
input := MyInput{
    InsertMessages: `[{"role":"user","content":"Hello there"},{"role":"assistant","content":"Hi, how can I help?"}]`,
}

// Original message list
messages := []openrouter.Message{
    {Role: "user", Content: "First message"},
    {Role: "assistant", Content: "{{insertMessages}}"},
    {Role: "user", Content: "Last message"},
}

// Result after parsing will be:
// [
//   {Role: "user", Content: "First message"},
//   {Role: "user", Content: "Hello there"},
//   {Role: "assistant", Content: "Hi, how can I help?"},
//   {Role: "user", Content: "Last message"},
// ]
```

#### Rules and Limitations for Message Insertion

- The message content must be exactly `{{insertMessages}}` (no other content)
- The `insertMessages` field in your input struct must contain a valid JSON array of message objects
- Each message must have a `role` that is either "user" or "assistant"
- Each message must have non-empty `content`
- If the JSON is invalid or any message is invalid, the original `{{insertMessages}}` will be replaced with the raw string value
- If `{{insertMessages}}` appears as part of other text, it will be processed as a regular variable replacement
- Message insertion is processed after conditional blocks but during variable replacement

### Conditional Blocks

Use `{{#if variableName}}...{{/if}}` to include content only when a variable exists and is not empty. You can also use `{{:else}}` for an else block.

**Note:** While `{{:else}}` blocks are functional, they haven't been extensively tested.

```go
// This section will only appear if previousItems is not empty
prompt := `
  {{#if previousItems}}
  Previous items:
  {{previousItems}}
  {{/if}}
  Create a list of {{count}} items.
`
```

### Rules and Limitations

- Variables must match the JSON tags of your input struct fields
- Non-existent or empty variables in if statements will remove the entire block
- Empty strings and nil values are considered "empty"
- Zero numeric values (0) are considered "non-empty"
- Nested if statements are not supported (will be treated as text)
- Malformed if statements will remain unchanged in the output
- Unmatched variable names will remain as `{{variableName}}` in the output
- If statements are processed first, then variable replacements and message insertions
- Else statements (`{{:else}}`) are functional but not fully tested.

## Optional Frontend UI Setup

LLMango includes an optional frontend UI for managing Goals and Prompts. To use it:

1.  **Create the Router**: Use `llmangofrontend.CreateLLMMangRouter` to get an `http.Handler`.
2.  **Mount at `/mango`**: Mount the handler strictly at the `/mango` path prefix. Use `http.StripPrefix` to ensure correct routing.
3.  **Add Authentication**: **Crucially**, protect this endpoint with your application's authentication middleware (e.g., admin-only access). Exposing this endpoint publicly could grant unauthorized access to your configured LLM providers (like OpenRouter).
4.  **Disable Caching**: Consider disabling caching for the `/mango` routes to prevent potential information leakage or stale data issues, depending on your authentication and use case.

Example setup using `net/http` compatible router (like `chi` or standard library `http.ServeMux`):

```go
import (
	"net/http"
	"yourapp/middleware" // Replace with your actual middleware import
	"github.com/llmang/llmango/llmangofrontend"
)

// Assuming 'app.Mango.LLMangoManager' is your initialized LLMango manager
// Assuming 'app.AdminDevEnv' is your authentication middleware
mangoRouter := llmangofrontend.CreateLLMMangRouter(app.Mango.LLMangoManager, nil)

// Ensure router handles paths with and without trailing slash correctly
router.Handle("/mango", http.StripPrefix("/mango", middleware.AdminDevEnv(mangoRouter)))
router.Handle("/mango/", http.StripPrefix("/mango", middleware.AdminDevEnv(mangoRouter)))

// Remember to configure cache control headers via middleware if needed
// e.g., w.Header().Set("Cache-Control", "no-store")
```

## Install
USE AT OWN RISK!!!!!
go get github.com/llmang/llmango
