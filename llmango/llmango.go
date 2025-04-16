package llmango

import (
	"fmt"
	"reflect"
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

func (m *LLMangoManager) AddPromptToGoal(goalUID, promptUID string) error {
	goalAny, ok := m.Goals[goalUID]
	if !ok {
		return fmt.Errorf("goal with UID '%s' not found", goalUID)
	}

	goalValue := reflect.ValueOf(goalAny)

	// Check if goalAny is a pointer, if so, get the element it points to
	if goalValue.Kind() == reflect.Ptr {
		goalValue = goalValue.Elem()
	}

	// Ensure we are working with a struct
	if goalValue.Kind() != reflect.Struct {
		return fmt.Errorf("goal with UID '%s' is not a struct, but %v", goalUID, goalValue.Kind())
	}

	promptUIDsField := goalValue.FieldByName("PromptUIDs")
	if !promptUIDsField.IsValid() {
		return fmt.Errorf("goal struct for UID '%s' does not have a 'PromptUIDs' field", goalUID)
	}

	if promptUIDsField.Kind() != reflect.Slice {
		return fmt.Errorf("'PromptUIDs' field for goal UID '%s' is not a slice, but %v", goalUID, promptUIDsField.Kind())
	}

	// Check if the slice element type is string
	if promptUIDsField.Type().Elem().Kind() != reflect.String {
		return fmt.Errorf("'PromptUIDs' field for goal UID '%s' is not a slice of strings, but slice of %v", goalUID, promptUIDsField.Type().Elem().Kind())
	}

	// Check if the field is addressable and settable
	if !promptUIDsField.CanSet() {
		// This might happen if goalAny was not a pointer originally.
		// We need a pointer to modify the original struct in the map.
		// Let's try getting a pointer to the value if it's addressable.
		if goalValue.CanAddr() {
			goalPtrValue := goalValue.Addr()
			promptUIDsField = goalPtrValue.Elem().FieldByName("PromptUIDs") // Re-fetch the field from the addressable struct
			if !promptUIDsField.CanSet() {
				return fmt.Errorf("cannot set 'PromptUIDs' field for goal UID '%s', ensure the goal in the map is a pointer or the map stores addressable structs", goalUID)
			}
		} else {
			return fmt.Errorf("cannot set 'PromptUIDs' field for goal UID '%s', the goal value is not addressable", goalUID)
		}
	}

	// Append the new promptUID
	newPromptUIDs := reflect.Append(promptUIDsField, reflect.ValueOf(promptUID))
	promptUIDsField.Set(newPromptUIDs)

	// If the original value in the map was not a pointer, update the map with the modified struct
	// This path is less likely if goals are typically stored as pointers.
	// if reflect.ValueOf(goalAny).Kind() != reflect.Ptr && goalValue.CanAddr() {
	// 	m.Goals[goalUID] = goalValue.Interface() // Update map with the modified struct value
	// }
	// No need to update the map explicitly if goalAny was already a pointer, as we modified the pointed-to struct.

	return nil
}

// UpdateGoalTitleDescription updates the Title and Description fields of a goal using reflection.
// The goal parameter must be a pointer to a struct or an addressable struct value.
func UpdateGoalTitleDescription(goal any, title, description string) error {
	val := reflect.ValueOf(goal)

	// Ensure goal is a pointer to a struct
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("goal must be a pointer to a struct, got %s", val.Kind())
	}
	val = val.Elem() // Dereference the pointer to get the struct value

	// Ensure we are dealing with a struct
	if val.Kind() != reflect.Struct {
		// This should technically not happen if the input is pointer-to-struct, but good practice
		return fmt.Errorf("goal element is not a struct, kind: %s", val.Kind())
	}

	// Get the Title field
	titleField := val.FieldByName("Title")
	if !titleField.IsValid() {
		return fmt.Errorf("goal struct does not have a 'Title' field")
	}
	if !titleField.CanSet() {
		return fmt.Errorf("'Title' field cannot be set (is it exported?)")
	}
	if titleField.Kind() != reflect.String {
		return fmt.Errorf("'Title' field is not a string")
	}

	// Get the Description field
	descField := val.FieldByName("Description")
	if !descField.IsValid() {
		return fmt.Errorf("goal struct does not have a 'Description' field")
	}
	if !descField.CanSet() {
		return fmt.Errorf("'Description' field cannot be set (is it exported?)")
	}
	if descField.Kind() != reflect.String {
		return fmt.Errorf("'Description' field is not a string")
	}

	// Set the values
	titleField.SetString(title)
	descField.SetString(description)

	return nil
}
