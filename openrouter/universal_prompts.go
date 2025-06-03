package openrouter

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// GenerateUniversalSystemPrompt creates a system prompt that guides ANY LLM to produce
// valid JSON output matching the provided schema, even without structured output support
func GenerateUniversalSystemPrompt(schema map[string]interface{}, inputExample, outputExample json.RawMessage) string {
	schemaStr := FormatSchemaForPrompt(schema)
	
	return fmt.Sprintf(`You are a precise JSON response generator. You must respond with valid JSON that exactly matches the provided schema.

SCHEMA REQUIREMENTS:
%s

INPUT EXAMPLE:
%s

EXPECTED OUTPUT EXAMPLE:
%s

CRITICAL INSTRUCTIONS:
1. Your response must be valid JSON only - no explanations, no markdown, no code blocks
2. Follow the schema exactly - all required fields must be present
3. Use the correct data types as specified in the schema
4. Do not add extra fields not defined in the schema
5. If uncertain about a value, use reasonable defaults that match the expected type
6. Do not include any text before or after the JSON response

Respond with JSON only:`, schemaStr, string(inputExample), string(outputExample))
}

// MergeSystemPrompts implements the collision strategy for combining existing system prompts
// with universal JSON schema instructions. The original prompt comes first, then our instructions.
func MergeSystemPrompts(existingSystemPrompt, universalPrompt string) string {
	// If no existing prompt, just return the universal prompt
	if strings.TrimSpace(existingSystemPrompt) == "" {
		return universalPrompt
	}
	
	// If no universal prompt, just return the existing prompt
	if strings.TrimSpace(universalPrompt) == "" {
		return existingSystemPrompt
	}
	
	// Merge strategy: Original prompt first, then universal instructions
	// Add clear separation between the two parts
	return fmt.Sprintf(`%s

---

IMPORTANT: In addition to the above instructions, you must now follow these JSON response requirements:

%s`, strings.TrimSpace(existingSystemPrompt), strings.TrimSpace(universalPrompt))
}

// CreateUniversalCompatibilityPrompt is the main function that creates a complete
// system prompt for universal LLM compatibility, merging existing prompts with
// JSON schema instructions
func CreateUniversalCompatibilityPrompt(existingSystemPrompt string, schema map[string]interface{}, inputExample, outputExample json.RawMessage) string {
	// Generate the universal JSON schema prompt
	universalPrompt := GenerateUniversalSystemPrompt(schema, inputExample, outputExample)
	
	// Merge with existing system prompt using our collision strategy
	return MergeSystemPrompts(existingSystemPrompt, universalPrompt)
}

// FormatSchemaForPrompt converts a JSON schema into human-readable format
// that's easy for LLMs to understand and follow
func FormatSchemaForPrompt(schema map[string]interface{}) string {
	var result strings.Builder
	
	if schemaType, ok := schema["type"].(string); ok && schemaType == "object" {
		result.WriteString("Object with the following properties:\n")
		
		if properties, ok := schema["properties"].(map[string]interface{}); ok {
			required := make(map[string]bool)
			
			// Handle required fields - support both []interface{} and []string
			if reqArray, ok := schema["required"].([]interface{}); ok {
				for _, req := range reqArray {
					if reqStr, ok := req.(string); ok {
						required[reqStr] = true
					}
				}
			} else if reqArray, ok := schema["required"].([]string); ok {
				// Handle []string directly (from our test data)
				for _, reqStr := range reqArray {
					required[reqStr] = true
				}
			}
			
			// Sort property names for consistent output
			var propNames []string
			for name := range properties {
				propNames = append(propNames, name)
			}
			sort.Strings(propNames)
			
			for _, name := range propNames {
				prop := properties[name]
				requiredStr := ""
				if required[name] {
					requiredStr = " (REQUIRED)"
				}
				
				result.WriteString(fmt.Sprintf("- %s%s: %s\n", name, requiredStr, formatPropertyForPrompt(prop)))
			}
		}
	} else {
		// Handle non-object schemas
		result.WriteString(fmt.Sprintf("Type: %s\n", schemaType))
	}
	
	return result.String()
}

// formatPropertyForPrompt formats individual property definitions for human readability
func formatPropertyForPrompt(prop interface{}) string {
	propMap, ok := prop.(map[string]interface{})
	if !ok {
		return "unknown type"
	}
	
	propType, ok := propMap["type"].(string)
	if !ok {
		return "unknown type"
	}
	
	var details []string
	
	switch propType {
	case "string":
		// Check for enum values - handle both []interface{} and []string
		var enumStrs []string
		
		// Try []interface{} first (standard JSON schema format)
		if enum, ok := propMap["enum"].([]interface{}); ok {
			for _, e := range enum {
				if eStr, ok := e.(string); ok {
					enumStrs = append(enumStrs, fmt.Sprintf(`"%s"`, eStr))
				}
			}
		} else if enum, ok := propMap["enum"].([]string); ok {
			// Handle []string directly (from our test data)
			for _, eStr := range enum {
				enumStrs = append(enumStrs, fmt.Sprintf(`"%s"`, eStr))
			}
		}
		
		if len(enumStrs) > 0 {
			return fmt.Sprintf("string (Must be one of: %s)", strings.Join(enumStrs, ", "))
		}
		// If no enum, just return string
		return "string"
		
	case "number", "integer":
		// Handle minimum - could be int or float64
		if min, ok := propMap["minimum"].(float64); ok {
			details = append(details, fmt.Sprintf("minimum: %.1f", min))
		} else if minInt, ok := propMap["minimum"].(int); ok {
			details = append(details, fmt.Sprintf("minimum: %d", minInt))
		}
		
		// Handle maximum - could be int or float64
		if max, ok := propMap["maximum"].(float64); ok {
			details = append(details, fmt.Sprintf("maximum: %.1f", max))
		} else if maxInt, ok := propMap["maximum"].(int); ok {
			details = append(details, fmt.Sprintf("maximum: %d", maxInt))
		}
		
		result := propType
		if len(details) > 0 {
			result += " (" + strings.Join(details, ", ") + ")"
		}
		return result
		
	case "array":
		if items, ok := propMap["items"].(map[string]interface{}); ok {
			itemType := formatPropertyForPrompt(items)
			return fmt.Sprintf("array (Array of %s)", itemType)
		} else {
			return "array"
		}
		
	case "object":
		if properties, ok := propMap["properties"].(map[string]interface{}); ok {
			var subProps []string
			for name, subProp := range properties {
				subPropType := formatPropertyForPrompt(subProp)
				subProps = append(subProps, fmt.Sprintf("%s: %s", name, subPropType))
			}
			if len(subProps) > 0 {
				sort.Strings(subProps)
				return fmt.Sprintf("object (Object with properties: %s)", strings.Join(subProps, ", "))
			}
		}
		return "object"
		
	case "boolean":
		return "boolean (true or false)"
	}
	
	result := propType
	if len(details) > 0 {
		result += " (" + strings.Join(details, ", ") + ")"
	}
	
	return result
}