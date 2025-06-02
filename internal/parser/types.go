package parser

import "github.com/llmang/llmango/openrouter"

// DiscoveredGoal represents a goal found during parsing
type DiscoveredGoal struct {
	UID         string `json:"uid" yaml:"uid"`
	Title       string `json:"title" yaml:"title"`
	Description string `json:"description" yaml:"description"`
	InputType   string `json:"input_type" yaml:"input_type"`
	OutputType  string `json:"output_type" yaml:"output_type"`
	SourceFile  string `json:"source_file" yaml:"source_file"`
	SourceType  string `json:"source_type" yaml:"source_type"` // "go" or "config"
	VarName     string `json:"var_name" yaml:"var_name"`       // Variable name in Go code
}

// DiscoveredPrompt represents a prompt found during parsing
type DiscoveredPrompt struct {
	UID        string                `json:"uid" yaml:"uid"`
	GoalUID    string                `json:"goal_uid" yaml:"goal_uid"`
	Model      string                `json:"model" yaml:"model"`
	Parameters openrouter.Parameters `json:"parameters" yaml:"parameters"`
	Messages   []openrouter.Message  `json:"messages" yaml:"messages"`
	Weight     int                   `json:"weight" yaml:"weight"`
	IsCanary   bool                  `json:"is_canary" yaml:"is_canary"`
	MaxRuns    int                   `json:"max_runs" yaml:"max_runs"`
	SourceFile string                `json:"source_file" yaml:"source_file"`
	SourceType string                `json:"source_type" yaml:"source_type"` // "go" or "config"
	VarName    string                `json:"var_name" yaml:"var_name"`       // Variable name in Go code
}

// ParseResult contains all discovered goals and prompts
type ParseResult struct {
	Goals   []DiscoveredGoal   `json:"goals" yaml:"goals"`
	Prompts []DiscoveredPrompt `json:"prompts" yaml:"prompts"`
	Errors  []ParseError       `json:"errors,omitempty" yaml:"errors,omitempty"`
}

// ParseError represents an error encountered during parsing
type ParseError struct {
	File    string `json:"file" yaml:"file"`
	Line    int    `json:"line,omitempty" yaml:"line,omitempty"`
	Column  int    `json:"column,omitempty" yaml:"column,omitempty"`
	Message string `json:"message" yaml:"message"`
	Type    string `json:"type" yaml:"type"` // "warning" or "error"
}

// Config represents the structure of llmango.yaml/json files
type Config struct {
	Goals   []ConfigGoal   `json:"goals" yaml:"goals"`
	Prompts []ConfigPrompt `json:"prompts" yaml:"prompts"`
}

// ConfigGoal represents a goal defined in configuration
type ConfigGoal struct {
	UID          string      `json:"uid" yaml:"uid"`
	Title        string      `json:"title" yaml:"title"`
	Description  string      `json:"description" yaml:"description"`
	InputType    string      `json:"input_type" yaml:"input_type"`
	OutputType   string      `json:"output_type" yaml:"output_type"`
	InputExample interface{} `json:"input_example,omitempty" yaml:"input_example,omitempty"`
	OutputExample interface{} `json:"output_example,omitempty" yaml:"output_example,omitempty"`
}

// ConfigPrompt represents a prompt defined in configuration
type ConfigPrompt struct {
	UID        string                `json:"uid" yaml:"uid"`
	GoalUID    string                `json:"goal_uid" yaml:"goal_uid"`
	Model      string                `json:"model" yaml:"model"`
	Parameters openrouter.Parameters `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Messages   []openrouter.Message  `json:"messages" yaml:"messages"`
	Weight     int                   `json:"weight,omitempty" yaml:"weight,omitempty"`
	IsCanary   bool                  `json:"is_canary,omitempty" yaml:"is_canary,omitempty"`
	MaxRuns    int                   `json:"max_runs,omitempty" yaml:"max_runs,omitempty"`
}

// GenerateOptions contains options for code generation
type GenerateOptions struct {
	InputDir    string
	OutputFile  string
	ConfigFile  string
	PackageName string
	Validate    bool
}