# Raw Method Generation Implementation Plan

## Overview
Implement optional Raw method generation using an extensible `generateOptions` wrapper in YAML config files and comment detection for Go files.

## YAML Structure Design

```yaml
# Extensible generation options wrapper
generateOptions:
  rawGoalFunctions:
    - "email-classification"
    - "sentiment-analysis"
  # Future extensibility:
  # debugMode: true
  # customValidators: ["goal-uid"]
  # asyncMethods: ["goal-uid"]

goals:
  - uid: "email-classification"
    title: "Email Classification"
    # ... rest of config

prompts:
  # ... existing prompts
```

## Go File Comment Detection

```go
//llmango:raw
var sentimentGoal = llmango.NewGoal(...)
```

## Implementation Steps

### Step 1: Update Config Types (`internal/parser/types.go`)

Add new struct for extensible generation options:

```go
// ConfigGenerateOptions represents generation options in config
type ConfigGenerateOptions struct {
    RawGoalFunctions []string `json:"rawGoalFunctions,omitempty" yaml:"rawGoalFunctions,omitempty"`
    // Future options can be added here:
    // DebugMode bool `json:"debugMode,omitempty" yaml:"debugMode,omitempty"`
    // CustomValidators []string `json:"customValidators,omitempty" yaml:"customValidators,omitempty"`
}

// Update Config struct to include generateOptions
type Config struct {
    Goals           []ConfigGoal           `json:"goals" yaml:"goals"`
    Prompts         []ConfigPrompt         `json:"prompts" yaml:"prompts"`
    GenerateOptions *ConfigGenerateOptions `json:"generateOptions,omitempty" yaml:"generateOptions,omitempty"`
}
```

### Step 2: Update Config Parsing (`internal/parser/config.go`)

Modify `parseConfigFile()` function to:
1. Parse the optional `generateOptions` section
2. Create a map of raw goal functions for quick lookup
3. Pass this information to the template generation process

Key changes:
- Extract `config.GenerateOptions.RawGoalFunctions` after unmarshaling
- Store as `map[string]bool` for O(1) lookup during generation
- Pass to template generation context

### Step 3: Update AST Comment Detection (`internal/parser/ast.go`)

Enhance AST parsing to detect `//llmango:raw` comments:
1. Check comment groups above `NewGoal()` calls
2. Look for exact match: `//llmango:raw`
3. Add detected goal UIDs to raw functions list

### Step 4: Update Template Generation (`internal/generator/templates.go`)

Add Raw method template generation:

```go
// Add to mangoFileTemplate
{{range .Goals}}
// {{.MethodName}} executes the {{.Title}} goal
func (m *Mango) {{.MethodName}}(input *{{.InputType}}) (*{{.OutputType}}, error) {
{{- if .IsPointer}}
    return llmango.Run[{{.InputType}}, {{.OutputType}}](m.LLMangoManager, {{.VarName}}, input)
{{- else}}
    return llmango.Run[{{.InputType}}, {{.OutputType}}](m.LLMangoManager, &{{.VarName}}, input)
{{- end}}
}

{{- if .ShouldGenerateRaw}}
// {{.MethodName}}Raw executes the {{.Title}} goal and returns raw response
func (m *Mango) {{.MethodName}}Raw(input *{{.InputType}}) (string, error) {
{{- if .IsPointer}}
    return llmango.RunRaw[{{.InputType}}](m.LLMangoManager, {{.VarName}}, input)
{{- else}}
    return llmango.RunRaw[{{.InputType}}](m.LLMangoManager, &{{.VarName}}, input)
{{- end}}
}
{{- end}}

{{end}}
```

### Step 5: Update Generator Logic (`internal/generator/generator.go`)

Modify the generation process to:
1. Collect raw goal functions from both config and AST parsing
2. Create a combined map of goal UIDs that need Raw methods
3. Pass this information to template data structure
4. Set `ShouldGenerateRaw` field for each goal during template execution

### Step 6: Template Data Structure Enhancement

Add field to template data:
```go
type TemplateGoal struct {
    // ... existing fields
    ShouldGenerateRaw bool // New field for Raw method generation
}
```

## Generated Method Examples

### Regular Method (always generated)
```go
func (m *Mango) SentimentAnalysis(input *SentimentInput) (*SentimentOutput, error) {
    return llmango.Run[SentimentInput, SentimentOutput](m.LLMangoManager, sentimentGoal, input)
}
```

### Raw Method (only when specified)
```go
func (m *Mango) SentimentAnalysisRaw(input *SentimentInput) (string, error) {
    return llmango.RunRaw[SentimentInput](m.LLMangoManager, sentimentGoal, input)
}
```

## Usage Examples

### YAML Config Usage
```yaml
generateOptions:
  rawGoalFunctions:
    - "email-classification"
    - "language-detection"

goals:
  - uid: "email-classification"
    title: "Email Classification"
    # ... rest of config
```

### Go File Usage
```go
//llmango:raw
var sentimentGoal = llmango.NewGoal(
    "sentiment-analysis",
    "Sentiment Analysis",
    // ... rest of goal definition
)
```

## Benefits

1. **Extensible**: `generateOptions` wrapper allows future options without breaking changes
2. **Opt-in**: Raw methods only generated when explicitly requested
3. **Flexible**: Works with both config-defined and Go-defined goals
4. **Clean**: No struct changes to existing types, just parsing logic updates
5. **Familiar**: Uses comment directives similar to other Go tools

## Files to Modify

1. `internal/parser/types.go` - Add ConfigGenerateOptions struct
2. `internal/parser/config.go` - Parse generateOptions section
3. `internal/parser/ast.go` - Detect //llmango:raw comments
4. `internal/generator/templates.go` - Add Raw method templates
5. `internal/generator/generator.go` - Pass raw function info to templates

## Testing Plan

1. Update `example-app/llmango.yaml` with generateOptions section
2. Add `//llmango:raw` comment to one goal in `example-app/internal/mango/example.go`
3. Regenerate and verify both regular and Raw methods are created
4. Test that Raw methods return string responses correctly

## Future Extensibility

The `generateOptions` structure allows for easy addition of new features:
- `debugMode: true` - Generate debug logging
- `customValidators: ["goal-uid"]` - Generate custom validation methods
- `asyncMethods: ["goal-uid"]` - Generate async/await style methods
- `batchMethods: ["goal-uid"]` - Generate batch processing methods