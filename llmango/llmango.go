package llmango

import (
	"fmt"
	"slices"
	"time"

	"github.com/llmang/llmango/openrouter"
)

var ErrMaxRateLimitRetries = fmt.Errorf("failed to get a valid response after retrying %v times with exponential backoff: %w", MAX_BACKOFF_ATTEMPTS, openrouter.ErrRateLimited)
var MAX_BACKOFF_ATTEMPTS = 10
var BASE_BACKOFF_DELAY = 100 * time.Millisecond

type LLMangoManager struct {
	RetryRateLimit bool
	OpenRouter     *openrouter.OpenRouter
	Goals          SyncedMap[string, *Goal]
	Prompts        SyncedMap[string, *Prompt]
	SaveState      func() error
	Logging        *Logging
}

func CreateLLMangoManger(o *openrouter.OpenRouter) (*LLMangoManager, error) {
	// defaultFileName := "llmango.json"
	return &LLMangoManager{
		OpenRouter: o,
		Prompts:    SyncedMap[string, *Prompt]{},
		Goals:      SyncedMap[string, *Goal]{},
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

type Goal struct {
	UID         string   `json:"UID"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	CreatedAt   int      `json:"createdAt"`
	UpdatedAt   int      `json:"updatedAt"`
	PromptUIDs  []string `json:"promptUIDs" savestate:"-"` //built during runtime so make sure to not save it in json or database
	InputOutput any      `json:"inputOutput"`              //use InputOutput here
}

type InputOutput[input any, output any] struct {
	InputExample    input             `json:"inputExample"`
	InputValidator  func(input) bool  `json:"-"`
	OutputExample   output            `json:"outputExample"`
	OutputValidator func(output) bool `json:"-"`
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
// It does NOT overwrite the InputOutput field of existing goals.
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
			if m.Goals.Exists(goal.UID) {
				existingGoal := m.Goals.Get(goal.UID)
				existingGoal.Title = goal.Title
				existingGoal.Description = goal.Description
				existingGoal.CreatedAt = goal.CreatedAt
				existingGoal.UpdatedAt = goal.UpdatedAt
				m.Goals.Set(goal.UID, existingGoal)
			} else {
				goal.PromptUIDs = []string{}
				m.Goals.Set(goal.UID, goal)
				for _, prompt := range m.Prompts.m {
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
			for _, prompt := range m.Prompts.m {
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
			if prompt.GoalUID != "" && m.Goals.Exists(prompt.GoalUID) {
				goal := m.Goals.Get(prompt.GoalUID)
				found := slices.Contains(goal.PromptUIDs, prompt.UID)
				if !found {
					goal.PromptUIDs = append(goal.PromptUIDs, prompt.UID)
					m.Goals.Set(goal.UID, goal)
				}
			}
		}
	}
}
