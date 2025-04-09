package llmango

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/llmang/llmango/openrouter"
)

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

	// Regex pattern to match {{#if varName}}...{{/if}} blocks
	ifPattern := regexp.MustCompile(`\{\{#if\s+([^{}]+)\}\}([\s\S]*?)\{\{\/if\}\}`)

	// Process each copied message
	for i, msg := range copiedMessages {
		// Replace if blocks in message content
		copiedMessages[i].Content = ifPattern.ReplaceAllStringFunc(msg.Content, func(match string) string {
			// Extract variable name and block content
			submatch := ifPattern.FindStringSubmatch(match)
			if len(submatch) < 3 {
				return match // Return original if pattern doesn't match expected format
			}

			varName := strings.TrimSpace(submatch[1])
			blockContent := submatch[2]

			// Check if variable exists and is not empty
			if val, ok := inputMap[varName]; ok {
				// Check if value is empty string or nil
				isEmpty := false
				if val == nil {
					isEmpty = true
				} else {
					switch v := val.(type) {
					case string:
						isEmpty = v == ""
					case *string:
						isEmpty = v == nil || *v == ""
					default:
						// For other types, consider them non-empty
					}
				}

				if !isEmpty {
					// Keep block content but remove the if tags
					return blockContent
				}
			}

			// Variable doesn't exist or is empty, remove entire block
			return ""
		})
	}

	return copiedMessages, nil
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

	// Convert input to map of field names to values
	inputMap := make(map[string]string)
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
		// Convert value to string
		inputMap[tag] = fmt.Sprintf("%v", value.Interface())
	}

	// Regex pattern to match {{variable}} but not /{{variable}}
	pattern := regexp.MustCompile(`\{\{([^{}]+)\}\}`)

	// Process each copied message
	for i, msg := range copiedMessages {
		// Replace matches in message content
		copiedMessages[i].Content = pattern.ReplaceAllStringFunc(msg.Content, func(match string) string {
			// Extract variable name
			varName := match[2 : len(match)-2]
			if val, ok := inputMap[varName]; ok {
				return val
			}
			return match // Return original if not found
		})
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

	// Then substitute variables
	finalMessages, err := InsertVariableValuesIntoPromptMessagesCopy(input, processedMessages)
	if err != nil {
		return nil, fmt.Errorf("error substituting variables: %w", err)
	}

	return finalMessages, nil
}
