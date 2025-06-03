# LLMango CLI

Command-line interface for generating type-safe Go functions from LLM goals and prompts.

## Features

- **SQLC-like workflow**: Generate type-safe functions from goals and prompts
- **Hybrid discovery**: Scans both Go files and YAML/JSON config files  
- **Type safety**: Compile-time safe functions with proper error handling
- **A/B testing**: Support for weighted prompts and canary testing

## Commands

- [`llmango init`](llmango/main.go:init) - Initialize new project with examples
- [`llmango generate`](llmango/main.go:generate) - Generate type-safe functions
- [`llmango validate`](llmango/main.go:validate) - Validate definitions

## Usage

```bash
# Initialize project
llmango-cli init --package myapp

# Generate functions  
llmango-cli generate --output internal/mango/mango.go

# Validate definitions
llmango-cli validate
```

## Status: ✅ Complete

All CLI functionality implemented and tested.