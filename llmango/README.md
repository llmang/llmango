# LLMango Core

Core goal-driven LLM framework with dual-mode architecture for maximum flexibility.

## Features

### Dual-Mode Goal System ✅
Create goals using either strongly-typed Go structs or pure JSON:

```go
// Typed Goals (Developer Mode)
goal := llmango.NewGoal("sentiment", "Sentiment Analysis", "...", 
    SentimentInput{Text: "example"}, 
    SentimentOutput{Sentiment: "positive", Confidence: 0.95})

// JSON Goals (Frontend Mode)  
goal := llmango.NewJSONGoal("sentiment", "Sentiment Analysis", "...",
    json.RawMessage(`{"text": "example"}`),
    json.RawMessage(`{"sentiment": "positive", "confidence": 0.95}`))
```

### Message Parsing System ✅
Advanced templating with variable replacement, conditional blocks, and message insertion:

```go
// Variable replacement: {{variableName}}
// Conditional blocks: {{#if variable}}...{{/if}}
// Message insertion: {{insertMessages}}
```

### Execution Router ✅
Intelligent routing between execution paths based on model capabilities.

## Key Components

- [`llmango.go`](llmango.go) - Core manager and goal execution
- [`messageparser.go`](messageparser.go) - Advanced message templating
- [`execution_router.go`](execution_router.go) - Dual-path execution routing
- [`new_goal_system.go`](new_goal_system.go) - Dual-mode goal creation

## Status: ✅ Complete

Core framework with dual-mode goals, advanced message parsing, and execution routing fully implemented.