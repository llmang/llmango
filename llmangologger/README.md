# LLMango Logger

Comprehensive logging system for LLM interactions with SQLite storage and structured data capture.

## Features

### Structured Logging ✅
Captures detailed execution data for analysis and debugging:

```go
type LogEntry struct {
    Timestamp    time.Time
    GoalUID      string
    PromptUID    string
    Model        string
    Input        json.RawMessage
    Output       json.RawMessage
    Success      bool
    ErrorMessage string
    Duration     time.Duration
    TokensUsed   int
}
```

### SQLite Storage ✅
Persistent storage with efficient querying and indexing:

- Automatic table creation and schema management
- Indexed queries for fast filtering and search
- Compact storage with JSON compression

### Query Interface ✅
Flexible querying for analytics and monitoring:

```go
// Filter by goal, model, time range, success status
logs := logger.GetLogs(filter.GoalUID("sentiment").Model("gpt-4").Since(yesterday))

// Aggregate statistics
stats := logger.GetStats(filter.GoalUID("sentiment").LastWeek())
```

## Key Components

- [`llmangologger.go`](llmangologger.go) - Core logging interface and entry management
- [`sqlite.go`](sqlite.go) - SQLite storage implementation with optimized queries

## Usage

```go
// Initialize logger
logger := llmangologger.NewSQLiteLogger("logs.db")

// Log execution
logger.LogExecution(goalUID, promptUID, model, input, output, duration, err)

// Query logs
logs := logger.GetLogsSince(time.Now().Add(-24*time.Hour))
```

## Status: ✅ Complete

Full logging system with SQLite storage, structured data capture, and flexible querying capabilities.