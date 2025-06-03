# Internal Packages

Core implementation packages for LLMango CLI and code generation.

## Packages

### [`cli/`](cli/)
CLI command implementations for init, generate, and validate operations.

**Key Components:**
- [`generate.go`](cli/generate.go) - Code generation logic
- [`init.go`](cli/init.go) - Project initialization
- [`validate.go`](cli/validate.go) - Definition validation

### [`generator/`](generator/)
Template-based code generation engine for type-safe LLM functions.

**Key Components:**
- [`generator.go`](generator/generator.go) - Main generation logic
- [`templates.go`](generator/templates.go) - Go code templates
- JSON schema generation for validation

### [`parser/`](parser/)
AST parsing and configuration handling for Go files and YAML/JSON configs.

**Key Components:**
- [`ast.go`](parser/ast.go) - Go AST parsing for goals/prompts
- [`config.go`](parser/config.go) - YAML/JSON configuration parsing
- [`types.go`](parser/types.go) - Type definitions

## Features

- **Hybrid Discovery**: Scans both Go source and config files
- **Conflict Resolution**: Go definitions take priority over configs
- **Type Safety**: Generates compile-time safe wrapper functions
- **Comprehensive Testing**: 100+ test cases covering edge cases

## Status: ✅ Complete

All internal packages implemented with full test coverage.