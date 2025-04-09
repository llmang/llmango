# LLMang
[llmang.com](https://llmang.com)

Hardcode Goals, Not Prompts

Goal-driven LLM framework for Go. By [Carson](https://carsho.dev).

## What It Does

Organizes LLM queries into Goals, Solutions, and Prompts. Focus on outcomes, not prompts.

- **Goals**: Define the result.
- **Solutions**: Ways to get there, with canary testing.
- **Prompts**: Model-specific inputs.

## Features

- Type-safe Go structs.
- Canary testing in production.
- Handles logging, retries, rate limits.
- Optional frontend UI.
- Dynamic message templating with conditional blocks and variable replacement.

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

### Conditional Blocks

Use `{{#if variableName}}...{{/if}}` to include content only when a variable exists and is not empty:

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
- If statements are processed first, then variable replacements

## Install

~~go get github.com/llmang/llmango~~
