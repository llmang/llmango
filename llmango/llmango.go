package llmango

import (
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/carsongh/strongmap/concurrentmap"
	"github.com/llmang/llmango/openrouter"
)

var ErrMaxRateLimitRetries = fmt.Errorf("failed to get a valid response after retrying %v times with exponential backoff: %w", MAX_BACKOFF_ATTEMPTS, openrouter.ErrRateLimited)
var MAX_BACKOFF_ATTEMPTS = 10
var BASE_BACKOFF_DELAY = 100 * time.Millisecond

type LLMangoManager struct {
	RetryRateLimit bool
	OpenRouter     *openrouter.OpenRouter
	Goals          concurrentmap.SyncedMap[string, *Goal]
	Prompts        concurrentmap.SyncedMap[string, *Prompt]
	SaveState      func() error
	Logging        *Logging
}

func CreateLLMangoManger(o *openrouter.OpenRouter) (*LLMangoManager, error) {
	return &LLMangoManager{
		OpenRouter: o,
		Prompts:    concurrentmap.SyncedMap[string, *Prompt]{},
		Goals:      concurrentmap.SyncedMap[string, *Goal]{},
	}, nil
}

// WithLogging sets up logging for the LLMangoManager using the fluent interface pattern
func (m *LLMangoManager) WithLogging(logger *Logging) *LLMangoManager {
	m.Logging = logger
	return m
}

type Prompt struct {
	UID        string                `json:"UID"`
	GoalUID    string                `json:"goalUID"`
	Model      string                `json:"model"`
	Parameters openrouter.Parameters `json:"parameters"`
	Messages   []openrouter.Message  `json:"messages"`

	CreatedAt int `json:"createdAt"`
	UpdatedAt int `json:"updatedAt"`

	Weight    int  `json:"weight"`
	IsCanary  bool `json:"isCanary"`
	MaxRuns   int  `json:"maxRuns"`
	TotalRuns int  `json:"totalRuns"`
}

type Goal struct {
	UID         string   `json:"UID"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	CreatedAt   int      `json:"createdAt"`
	UpdatedAt   int      `json:"updatedAt"`
	PromptUIDs  []string `json:"promptUIDs" savestate:"-"` //built during runtime so make sure to not save it in json or database

	// Flag: false = typed goal, true = JSON object goal
	IsSchemaValidated bool `json:"isSchemaValidated"`

	// Unified serializable storage
	InputExample  json.RawMessage `json:"inputExample"`
	OutputExample json.RawMessage `json:"outputExample"`

	// Runtime validators (reconstructed on startup)
	InputValidator  func(json.RawMessage) error `json:"-"`
	OutputValidator func(json.RawMessage) error `json:"-"`
}

// GoalValidator interface for typed goals
type GoalValidator[I, O any] interface {
	ValidateInput(input I) error
	ValidateOutput(output O) error
}

type Result[T any] struct {
	Result T            `json:"result"`
	Error  *ResultError `json:"error,omitempty"`
}

type ResultError struct {
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

func (re *ResultError) Error() string {
	return fmt.Sprintf("Mango error occured: Reason:%v Message: %v", re.Reason, re.Message)
}

// AddOrUpdateGoals adds or updates goals in the LLMangoManager.
// It updates the Title, Description, CreatedAt, and UpdatedAt fields of existing goals.
func (m *LLMangoManager) AddOrUpdateGoals(goals ...*Goal) {
	now := int(time.Now().Unix())
	for _, goal := range goals {
		if goal != nil && goal.UID != "" {
			if goal.CreatedAt == 0 {
				goal.CreatedAt = now
			}
			if goal.UpdatedAt == 0 {
				goal.UpdatedAt = now
			}
			if existingGoal, ok := m.Goals.Get(goal.UID); ok {
				existingGoal.Title = goal.Title
				existingGoal.Description = goal.Description
				existingGoal.CreatedAt = goal.CreatedAt // Keep original CreatedAt? No, instruction implies updating based on input goal.
				existingGoal.UpdatedAt = goal.UpdatedAt // Update UpdatedAt based on input goal.
				m.Goals.Set(goal.UID, existingGoal)
			} else {
				goal.PromptUIDs = []string{}
				m.Goals.Set(goal.UID, goal)
				// Iterate over a snapshot for thread safety
				for _, prompt := range m.Prompts.Snapshot() {
					if prompt != nil && prompt.GoalUID == goal.UID {
						goal.PromptUIDs = append(goal.PromptUIDs, prompt.UID)
					}
				}
				m.Goals.Set(goal.UID, goal)
			}
		}
	}
}

// AddGoals adds or updates goals in the LLMangoManager.
// It overwrites the entire goal object if a goal with the same UID already exists.
func (m *LLMangoManager) AddGoals(goals ...*Goal) {
	now := int(time.Now().Unix())
	for _, goal := range goals {
		if goal != nil && goal.UID != "" {
			if goal.CreatedAt == 0 {
				goal.CreatedAt = now
			}
			if goal.UpdatedAt == 0 {
				goal.UpdatedAt = now
			}
			goal.PromptUIDs = []string{}
			m.Goals.Set(goal.UID, goal)
			// Iterate over a snapshot for thread safety
			for _, prompt := range m.Prompts.Snapshot() {
				if prompt != nil && prompt.GoalUID == goal.UID {
					goal.PromptUIDs = append(goal.PromptUIDs, prompt.UID)
				}
			}
			m.Goals.Set(goal.UID, goal)
		}
	}
}

// AddPrompts adds or updates prompts in the LLMangoManager.
// It always overwrites the entire prompt object if a prompt with the same UID already exists.
func (m *LLMangoManager) AddPrompts(prompts ...*Prompt) {
	now := int(time.Now().Unix())
	for _, prompt := range prompts {
		if prompt != nil && prompt.UID != "" {
			if prompt.CreatedAt == 0 {
				prompt.CreatedAt = now
			}
			if prompt.UpdatedAt == 0 {
				prompt.UpdatedAt = now
			}
			m.Prompts.Set(prompt.UID, prompt)
			if prompt.GoalUID != "" {
				// Get now returns item, ok
				goal, ok := m.Goals.Get(prompt.GoalUID)
				if ok { // Check if the goal exists
					found := slices.Contains(goal.PromptUIDs, prompt.UID)
					if !found {
						goal.PromptUIDs = append(goal.PromptUIDs, prompt.UID)
						m.Goals.Set(goal.UID, goal) // Update the goal
					}
				}
			}
		}
	}
}

// Goal Validation and Utility Functions

// Validate checks for inconsistent goal states and returns warnings
func (g *Goal) Validate() []string {
	var warnings []string

	// For typed goals, warn if no validators are provided
	if !g.IsSchemaValidated && g.InputValidator == nil && g.OutputValidator == nil {
		warnings = append(warnings, "Typed goal has no validators - consider adding validation")
	}

	// Note: JSON goals are expected to have function validators (generated from schemas)
	// so we don't warn about that case

	return warnings
}

// TypedValidator provides typed validation functions for NewGoal
type TypedValidator[I, O any] struct {
	ValidateInput  func(I) error
	ValidateOutput func(O) error
}

// Factory Functions for Goal Creation

// NewGoal creates a typed goal with optional validators (standard way for developers)
func NewGoal[I, O any](uid, title, description string, inputExample I, outputExample O, validator ...TypedValidator[I, O]) *Goal {
	// Marshal examples to JSON for unified storage
	inputJSON, err := json.Marshal(inputExample)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal input example: %v", err))
	}

	outputJSON, err := json.Marshal(outputExample)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal output example: %v", err))
	}

	goal := &Goal{
		UID:               uid,
		Title:             title,
		Description:       description,
		CreatedAt:         int(time.Now().Unix()),
		UpdatedAt:         int(time.Now().Unix()),
		PromptUIDs:        []string{},
		IsSchemaValidated: false, // Typed goal
		InputExample:      inputJSON,
		OutputExample:     outputJSON,
	}

	// Create wrapper validators if provided
	if len(validator) > 0 {
		v := validator[0]

		// Wrap input validator
		if v.ValidateInput != nil {
			goal.InputValidator = func(jsonInput json.RawMessage) error {
				var input I
				if err := json.Unmarshal(jsonInput, &input); err != nil {
					return fmt.Errorf("invalid input JSON: %w", err)
				}
				return v.ValidateInput(input)
			}
		}

		// Wrap output validator
		if v.ValidateOutput != nil {
			goal.OutputValidator = func(jsonOutput json.RawMessage) error {
				var output O
				if err := json.Unmarshal(jsonOutput, &output); err != nil {
					return fmt.Errorf("invalid output JSON: %w", err)
				}
				return v.ValidateOutput(output)
			}
		}
	}

	return goal
}

// NewJSONGoal creates a JSON object goal (standard way for frontend/dynamic use)
func NewJSONGoal(uid, title, description string, inputExample, outputExample json.RawMessage) *Goal {
	goal := &Goal{
		UID:               uid,
		Title:             title,
		Description:       description,
		CreatedAt:         int(time.Now().Unix()),
		UpdatedAt:         int(time.Now().Unix()),
		PromptUIDs:        []string{},
		IsSchemaValidated: true, // JSON object goal
		InputExample:      inputExample,
		OutputExample:     outputExample,
	}

	// Generate schema validators from JSON examples
	if err := goal.generateSchemaValidators(); err != nil {
		panic(fmt.Sprintf("failed to generate schema validators: %v", err))
	}

	return goal
}

// generateSchemaValidators creates JSON schema validators from examples
func (g *Goal) generateSchemaValidators() error {
	// Generate input schema validator
	if len(g.InputExample) > 0 {
		inputSchema, err := openrouter.GenerateSchemaFromJSONExample(g.InputExample)
		if err != nil {
			return fmt.Errorf("failed to generate input schema: %w", err)
		}

		g.InputValidator = func(jsonInput json.RawMessage) error {
			return openrouter.ValidateJSONAgainstSchema(jsonInput, inputSchema)
		}
	}

	// Generate output schema validator
	if len(g.OutputExample) > 0 {
		outputSchema, err := openrouter.GenerateSchemaFromJSONExample(g.OutputExample)
		if err != nil {
			return fmt.Errorf("failed to generate output schema: %w", err)
		}

		g.OutputValidator = func(jsonOutput json.RawMessage) error {
			return openrouter.ValidateJSONAgainstSchema(jsonOutput, outputSchema)
		}
	}

	return nil
}