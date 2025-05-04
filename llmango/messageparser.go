package llmango

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/llmang/llmango/openrouter"
)

// ArbitraryMessage represents a simple message structure used for parsing insertMessages
type ArbitraryMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ParseMessageIfStatements processes conditional blocks in messages.
// It handles {{#if varName}}...{{/if}} blocks, keeping content if varName is not empty.
func ParseMessageIfStatements(input any, messages []openrouter.Message) ([]openrouter.Message, error) {
	// Create a deep copy of messages to avoid modifying original
	copiedMessages := make([]openrouter.Message, len(messages))
	for i, msg := range messages {
		copiedMessages[i] = openrouter.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Convert input to map of field names to values
	inputMap := make(map[string]interface{})
	v := reflect.ValueOf(input)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("input must be a struct or pointer to struct")
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)
		// Get JSON tag name, fall back to field name
		tag := field.Tag.Get("json")
		if tag == "" || tag == "-" {
			tag = field.Name
		} else {
			// Handle json tag options like "omitempty"
			if commaIdx := strings.Index(tag, ","); commaIdx != -1 {
				tag = tag[:commaIdx]
			}
		}
		// Store actual value, not just string representation
		inputMap[tag] = value.Interface()
	}

	// Regex pattern to match {{#if varName}}...({{:else}}...)?{{/if}} blocks
	// Group 1: Variable name
	// Group 2: Content if true
	// Group 3: Content if false (else block) - may be empty if no else
	ifPattern := regexp.MustCompile(`\{\{#if\s+([^{}]+)\}\}([\s\S]*?)(?:\{\{:else\}\}([\s\S]*?))?\{\{\/if\}\}`)

	// Process each copied message
	for i, msg := range copiedMessages {
		// Replace if blocks in message content
		copiedMessages[i].Content = ifPattern.ReplaceAllStringFunc(msg.Content, func(match string) string {
			// Extract variable name and block content
			submatch := ifPattern.FindStringSubmatch(match)
			// Expected submatches:
			// [0]: Full match
			// [1]: Variable name
			// [2]: Content if true
			// [3]: Content if false (else block)
			if len(submatch) < 4 { // Need at least 4 elements (full match + 3 groups)
				return match // Return original if pattern doesn't match expected format
			}

			varName := strings.TrimSpace(submatch[1])
			ifContent := submatch[2]
			elseContent := submatch[3] // Will be empty if no {{:else}} was present

			// Check if variable exists and is considered "truthy"
			isTruthy := false
			if val, ok := inputMap[varName]; ok && val != nil {
				// Check specific types for emptiness
				switch v := val.(type) {
				case string:
					isTruthy = v != ""
				case *string:
					isTruthy = v != nil && *v != ""
				// Add other types if needed, e.g., slices, maps
				// case []any:
				// 	 isTruthy = len(v) > 0
				// case map[any]any:
				//  isTruthy = len(v) > 0
				default:
					// For non-nil types not explicitly checked (like numbers, bools),
					// consider them truthy if they exist.
					// Note: This treats 0 and false as truthy if the field exists.
					// Adjust if different behavior is desired (e.g., check for zero values).
					isTruthy = true
				}
			}

			if isTruthy {
				// Keep the 'if' block content
				return ifContent
			} else {
				// Keep the 'else' block content (which might be empty if no {{:else}})
				return elseContent
			}
		})
	}

	return copiedMessages, nil
}

// InsertVariableValue inserts a single variable value into content
func InsertVariableValue(input any, varName string, content string) (string, error) {
	// Convert input to map of field names to values
	inputMap := make(map[string]string)
	v := reflect.ValueOf(input)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return content, fmt.Errorf("input must be a struct or pointer to struct")
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)
		// Get JSON tag name, fall back to field name
		tag := field.Tag.Get("json")
		if tag == "" || tag == "-" {
			tag = field.Name
		} else {
			// Handle json tag options like "omitempty"
			if commaIdx := strings.Index(tag, ","); commaIdx != -1 {
				tag = tag[:commaIdx]
			}
		}
		// Convert value to string
		inputMap[tag] = fmt.Sprintf("%v", value.Interface())
	}

	// Check if varName exists in the map
	if val, ok := inputMap[varName]; ok {
		return val, nil
	}

	return content, nil
}

// InsertVariableValuesIntoContent replaces {{variable}} with its value in content
func InsertVariableValuesIntoContent(input any, content string) (string, error) {
	// Convert input to map of field names to values
	inputMap := make(map[string]string)
	v := reflect.ValueOf(input)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return content, fmt.Errorf("input must be a struct or pointer to struct")
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)
		// Get JSON tag name, fall back to field name
		tag := field.Tag.Get("json")
		if tag == "" || tag == "-" {
			tag = field.Name
		} else {
			// Handle json tag options like "omitempty"
			if commaIdx := strings.Index(tag, ","); commaIdx != -1 {
				tag = tag[:commaIdx]
			}
		}
		// Convert value to string
		inputMap[tag] = fmt.Sprintf("%v", value.Interface())
	}

	// Regex pattern to match {{variable}} but not /{{variable}}
	pattern := regexp.MustCompile(`\{\{([^{}]+)\}\}`)

	// Replace matches in content
	result := pattern.ReplaceAllStringFunc(content, func(match string) string {
		// Extract variable name
		varName := match[2 : len(match)-2]
		if val, ok := inputMap[varName]; ok {
			return val
		}
		return match // Return original if not found
	})

	return result, nil
}

// TryParseMessageArray attempts to parse input.insertMessages as an array of messages
// Returns the parsed messages and true if successful, otherwise nil and false
func TryParseMessageArray(input any) ([]openrouter.Message, bool) {
	// Get the insertMessages value from input
	v := reflect.ValueOf(input)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, false
	}

	var messagesJSON string

	// Find insertMessages field
	found := false
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		// Get JSON tag name
		tag := field.Tag.Get("json")
		if tag == "" || tag == "-" {
			tag = field.Name
		} else {
			// Handle json tag options like "omitempty"
			if commaIdx := strings.Index(tag, ","); commaIdx != -1 {
				tag = tag[:commaIdx]
			}
		}

		if tag == "insertMessages" {
			found = true
			messagesJSON = fmt.Sprintf("%v", v.Field(i).Interface())
			break
		}
	}

	if !found || messagesJSON == "" {
		return nil, false
	}

	// Try to parse as JSON array of messages
	var arbitraryMessages []ArbitraryMessage
	if err := json.Unmarshal([]byte(messagesJSON), &arbitraryMessages); err != nil {
		return nil, false
	}

	// Validate messages
	if len(arbitraryMessages) == 0 {
		return nil, false
	}

	var validMessages []openrouter.Message

	for _, msg := range arbitraryMessages {
		// Validate role and content
		if msg.Role == "" || msg.Content == "" {
			return nil, false
		}

		// Role must be user or assistant
		if msg.Role != "user" && msg.Role != "assistant" {
			return nil, false
		}

		validMessages = append(validMessages, openrouter.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	if len(validMessages) == 0 {
		return nil, false
	}

	return validMessages, true
}

func InsertVariableValuesIntoPromptMessagesCopy(input any, messages []openrouter.Message) ([]openrouter.Message, error) {
	// Create a deep copy of messages to avoid modifying original
	copiedMessages := make([]openrouter.Message, len(messages))
	for i, msg := range messages {
		copiedMessages[i] = openrouter.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Process each message
	i := 0
	for i < len(copiedMessages) {
		msg := copiedMessages[i]

		// Check if this is a message insertion point
		if msg.Content == "{{insertMessages}}" {
			// Try to parse insertMessages as an array of messages
			if parsedMessages, ok := TryParseMessageArray(input); ok {
				// If successful, replace this message with the parsed messages
				// Remove the current message
				beforeMessages := make([]openrouter.Message, i)
				copy(beforeMessages, copiedMessages[:i])

				afterMessages := make([]openrouter.Message, len(copiedMessages)-i-1)
				if i < len(copiedMessages)-1 {
					copy(afterMessages, copiedMessages[i+1:])
				}

				// Create new message slice with parts assembled in order
				newMessages := make([]openrouter.Message, 0, len(beforeMessages)+len(parsedMessages)+len(afterMessages))
				newMessages = append(newMessages, beforeMessages...)
				newMessages = append(newMessages, parsedMessages...)
				newMessages = append(newMessages, afterMessages...)

				copiedMessages = newMessages

				// Don't increment i so we process the next message at the current position
				continue
			}
		}

		// Otherwise do normal variable substitution
		newContent, err := InsertVariableValuesIntoContent(input, msg.Content)
		if err != nil {
			return nil, err
		}
		copiedMessages[i].Content = newContent

		// Move to next message
		i++
	}

	return copiedMessages, nil
}

// ParseMessages combines conditional block processing and variable substitution
// into a single operation, processing messages in the correct order.
func ParseMessages(input any, messages []openrouter.Message) ([]openrouter.Message, error) {
	// First process conditional blocks
	processedMessages, err := ParseMessageIfStatements(input, messages)
	if err != nil {
		return nil, fmt.Errorf("error processing conditional blocks: %w", err)
	}

	// Then substitute variables and handle message insertions
	finalMessages, err := InsertVariableValuesIntoPromptMessagesCopy(input, processedMessages)
	if err != nil {
		return nil, fmt.Errorf("error substituting variables: %w", err)
	}

	return finalMessages, nil
}
