package llmango

import (
	"fmt"
	"time"

	"github.com/llmang/llmango/openrouter"
)

var ErrMaxRateLimitRetries = fmt.Errorf("failed to get a valid response after retrying %v times with exponential backoff: %w", MAX_BACKOFF_ATTEMPTS, openrouter.ErrRateLimited)
var MAX_BACKOFF_ATTEMPTS = 10
var BASE_BACKOFF_DELAY = 100 * time.Millisecond

type LLMangoManager struct {
	SAFTEYSHUTOFF  bool
	RetryRateLimit bool
	OpenRouter     *openrouter.OpenRouter
	Prompts        map[string]*Prompt
	Goals          map[string]any
	SaveState      func() error
	*Logging
}

func CreateLLMangoManger(o *openrouter.OpenRouter) (*LLMangoManager, error) {
	// defaultFileName := "llmango.json"
	return &LLMangoManager{
		OpenRouter: o,
		Prompts:    make(map[string]*Prompt),
		Goals:      make(map[string]any),
	}, nil
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

type GoalInfo struct {
	UID         string `json:"UID"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CreatedAt   int    `json:"createdAt"`
	UpdatedAt   int    `json:"updatedAt"`
}

// Do we want to add the ability to
type Goal[Input any, Output any] struct {
	GoalInfo
	Validator     func(*Output) bool `json:"-"`
	ExampleInput  Input              `json:"exampleInput"`
	ExampleOutput Output             `json:"exampleOutput"`
	PromptUIDs    []string           `json:"promptUIDs"`
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
