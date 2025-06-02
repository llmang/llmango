# Dual-Mode Goal System Architecture

## Overview

This document describes LLMango's **Dual-Mode Goal System**, which enables both strongly-typed (developer-friendly) and JSON-based (frontend/dynamic) goal creation. This system allows external goals and prompts to be created either through Go structs or pure JSON.

**Note**: This is separate from the "Dual-Path Execution System" which handles LLM execution strategies.

## System Purpose

Enable goal creation in two modes:
1. **Typed Goals** (Developer Mode): Using Go structs for compile-time safety
2. **JSON Goals** (Frontend Mode): Using pure JSON for runtime flexibility

## Architectural Inspiration

This approach is inspired by:
- **GraphQL**: Typed schema definitions vs dynamic schema introspection
- **Protocol Buffers**: Compiled `.proto` definitions vs dynamic message parsing
- **gRPC**: Strongly-typed interfaces vs generic clients using reflection

## Problem Statement

Original LLMango used only strongly-typed generics:

```go
type Goal struct {
    InputOutput any `json:"inputOutput"` // InputOutput[ConcreteType, ConcreteType]
}

type InputOutput[input any, output any] struct {
    InputExample    input             `json:"inputExample"`
    InputValidator  func(input) bool  `json:"-"`
    OutputExample   output            `json:"outputExample"`
    OutputValidator func(output) bool `json:"-"`
}
```

**Limitations:**
1. **Not Serializable**: Validator functions cannot be serialized to JSON
2. **Frontend Complexity**: Frontend cannot create goals without Go type definitions
3. **Runtime Inflexibility**: Cannot create goals dynamically from user input
4. **External Integration**: Third-party systems cannot easily create goals

## Solution: Dual-Mode Goal Creation

### Core Concept

Support two goal creation patterns:

1. **Typed Goals (Developer Mode)**: `InputOutput[ConcreteType, ConcreteType]` for compile-time safety
2. **JSON Goals (Frontend Mode)**: `InputOutput[json.RawMessage, json.RawMessage]` for runtime flexibility

### Type Detection System

```go
func isJSONRawMessageInputOutput(inputOutput any) bool {
    // Uses reflection to detect InputOutput[json.RawMessage, json.RawMessage]
    typeName := reflect.TypeOf(inputOutput).Name()
    return strings.HasPrefix(typeName, "InputOutput[") && 
           bothFieldsAreJSONRawMessage(inputOutput)
}
```

**Key Insight**: Generic types in Go include full type parameters in their name:
- Typed: `InputOutput[mypackage.MyInput,mypackage.MyOutput]`
- JSON: `InputOutput[encoding/json.RawMessage,encoding/json.RawMessage]`

### Schema Generation Pipeline

```go
// From JSON examples, generate schemas for validation
func GenerateSchemaFromJSONExample(example json.RawMessage) (*Definition, error) {
    var parsed interface{}
    json.Unmarshal(example, &parsed)
    return generateSchemaFromInterface(parsed)
}
```

### Factory Functions

```go
// For developers (compile-time safety)
func NewTypedGoal[I, O any](uid, title, description string, inputExample I, outputExample O) *Goal

// For frontend/dynamic use (runtime flexibility)
func NewJSONGoal(uid, title, description string, inputExample, outputExample json.RawMessage) *Goal

// Migration utility
func ConvertTypedGoalToJSON(typedGoal *Goal) (*Goal, error)
```

## Implementation Status

✅ **COMPLETED**
- Type detection system
- Schema generation from JSON examples
- JSON validation against schemas
- Factory functions for both modes
- Comprehensive testing (41 test suites passing)
- Goal conversion utilities

## Benefits

### For Developers
- **Backwards Compatibility**: Existing typed goals continue to work unchanged
- **Type Safety**: Compile-time checking for typed goals
- **Migration Path**: Easy conversion from typed to JSON goals

### For Frontend/External Systems
- **No Go Dependencies**: Create goals using pure JSON
- **Runtime Flexibility**: Generate goals from user input
- **External Integration**: Third-party systems can create goals via JSON

### For System Architecture
- **Serializable**: All goal data can be stored/transmitted as JSON
- **Extensible**: Easy to add new validation rules or schema features
- **Performant**: Minimal overhead for type detection and schema generation

## Usage Examples

### Typed Goals (Developer Mode)
```go
type SentimentInput struct {
    Text string `json:"text"`
}

type SentimentOutput struct {
    Sentiment  string  `json:"sentiment"`
    Confidence float64 `json:"confidence"`
}

goal := llmango.NewGoal(
    "sentiment-classifier",
    "Sentiment Analysis",
    "Analyzes text sentiment",
    SentimentInput{Text: "example"},
    SentimentOutput{Sentiment: "positive", Confidence: 0.95},
    validator,
)
```

### JSON Goals (Frontend Mode)
```go
inputJSON := json.RawMessage(`{"text": "example query"}`)
outputJSON := json.RawMessage(`{"result": "example result", "confidence": 0.95}`)

goal := llmango.NewJSONGoal(
    "user-goal-1",
    "User Created Goal",
    "Created from frontend",
    inputJSON,
    outputJSON,
)
```

### Frontend Integration
```typescript
// Frontend can now create goals directly
const goal = {
  UID: "user-goal-1",
  title: "User Created Goal",
  description: "Created from frontend",
  inputOutput: {
    inputExample: { "query": "example query" },
    outputExample: { "result": "example result", "confidence": 0.95 }
  }
}
```

## Testing Strategy

Comprehensive test coverage includes:
1. **Type Detection Tests**: Verify correct identification of JSON vs typed goals
2. **Schema Generation Tests**: Validate schema creation from various JSON structures
3. **Validation Tests**: Ensure proper validation of JSON against schemas
4. **Factory Function Tests**: Test goal creation in both modes
5. **Integration Tests**: End-to-end testing of dual-mode execution

## Migration Guide

### For Existing Users
1. **No Action Required**: Existing typed goals continue to work
2. **Optional Migration**: Use `ConvertTypedGoalToJSON()` to convert goals
3. **New Goals**: Choose typed (developer) or JSON (frontend) based on use case

## Conclusion

The Dual-Mode Goal System successfully enables both developer productivity and runtime flexibility. External systems can now create goals and prompts using pure JSON, while developers retain the benefits of strongly-typed goals when appropriate.

This system is the foundation that enables the separate "Dual-Path Execution System" to work with any LLM, regardless of structured output support.