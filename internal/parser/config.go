package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseConfigFiles scans for and parses llmango configuration files
func ParseConfigFiles(dir string) (*ParseResult, error) {
	result := &ParseResult{
		Goals:            []DiscoveredGoal{},
		Prompts:          []DiscoveredPrompt{},
		Errors:           []ParseError{},
		RawGoalFunctions: make(map[string]bool),
	}

	// Look for configuration files
	configFiles := []string{
		filepath.Join(dir, "llmango.yaml"),
		filepath.Join(dir, "llmango.yml"),
		filepath.Join(dir, "llmango.json"),
	}

	for _, configFile := range configFiles {
		if _, err := os.Stat(configFile); err == nil {
			if err := parseConfigFile(configFile, result); err != nil {
				result.Errors = append(result.Errors, ParseError{
					File:    configFile,
					Message: fmt.Sprintf("Failed to parse config file: %v", err),
					Type:    "error",
				})
			}
		}
	}

	return result, nil
}

// parseConfigFile parses a single configuration file
func parseConfigFile(filename string, result *ParseResult) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var config Config
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("failed to parse YAML: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("failed to parse JSON: %w", err)
		}
	default:
		return fmt.Errorf("unsupported file extension: %s", ext)
	}

	// Process generateOptions if present
	if config.GenerateOptions != nil && len(config.GenerateOptions.RawGoalFunctions) > 0 {
		for _, goalUID := range config.GenerateOptions.RawGoalFunctions {
			result.RawGoalFunctions[goalUID] = true
		}
	}

	// Convert config goals to discovered goals
	for _, configGoal := range config.Goals {
		goal := DiscoveredGoal{
			UID:         configGoal.UID,
			Title:       configGoal.Title,
			Description: configGoal.Description,
			InputType:   configGoal.InputType,
			OutputType:  configGoal.OutputType,
			SourceFile:  filename,
			SourceType:  "config",
			VarName:     generateVarName(configGoal.UID, "Goal"),
			IsPointer:   false, // Config goals are generated as values, we'll take their address
		}

		// Convert input example to JSON string and validate
		if configGoal.InputExample != nil {
			inputJSON, err := json.Marshal(configGoal.InputExample)
			if err != nil {
				result.Errors = append(result.Errors, ParseError{
					File:    filename,
					Message: fmt.Sprintf("Goal '%s' has invalid input_example: %v", goal.UID, err),
					Type:    "error",
				})
				continue
			}
			
			// Validate that the example has at least one field
			if !hasAtLeastOneField(inputJSON) {
				result.Errors = append(result.Errors, ParseError{
					File:    filename,
					Message: fmt.Sprintf("Goal '%s' input_example is empty '{}' - structured output requires at least one field (e.g., input_example: {\"text\": \"example\"})", goal.UID),
					Type:    "error",
				})
				continue
			}
			
			goal.InputExampleJSON = string(inputJSON)
		} else {
			result.Errors = append(result.Errors, ParseError{
				File:    filename,
				Message: fmt.Sprintf("Goal '%s' missing required input_example field - add 'input_example: {\"field\": \"value\"}' to your goal definition", goal.UID),
				Type:    "error",
			})
			continue
		}

		// Convert output example to JSON string and validate
		if configGoal.OutputExample != nil {
			outputJSON, err := json.Marshal(configGoal.OutputExample)
			if err != nil {
				result.Errors = append(result.Errors, ParseError{
					File:    filename,
					Message: fmt.Sprintf("Goal '%s' has invalid output_example: %v", goal.UID, err),
					Type:    "error",
				})
				continue
			}
			
			// Validate that the example has at least one field
			if !hasAtLeastOneField(outputJSON) {
				result.Errors = append(result.Errors, ParseError{
					File:    filename,
					Message: fmt.Sprintf("Goal '%s' output_example is empty '{}' - structured output requires at least one field (e.g., output_example: {\"result\": \"example\"})", goal.UID),
					Type:    "error",
				})
				continue
			}
			
			goal.OutputExampleJSON = string(outputJSON)
		} else {
			result.Errors = append(result.Errors, ParseError{
				File:    filename,
				Message: fmt.Sprintf("Goal '%s' missing required output_example field - add 'output_example: {\"field\": \"value\"}' to your goal definition", goal.UID),
				Type:    "error",
			})
			continue
		}

		// Validate required fields
		if goal.UID == "" {
			result.Errors = append(result.Errors, ParseError{
				File:    filename,
				Message: "Goal missing required UID field",
				Type:    "error",
			})
			continue
		}
		if goal.InputType == "" {
			result.Errors = append(result.Errors, ParseError{
				File:    filename,
				Message: fmt.Sprintf("Goal '%s' missing required input_type field", goal.UID),
				Type:    "error",
			})
			continue
		}
		if goal.OutputType == "" {
			result.Errors = append(result.Errors, ParseError{
				File:    filename,
				Message: fmt.Sprintf("Goal '%s' missing required output_type field", goal.UID),
				Type:    "error",
			})
			continue
		}

		result.Goals = append(result.Goals, goal)
	}

	// Convert config prompts to discovered prompts
	for _, configPrompt := range config.Prompts {
		prompt := DiscoveredPrompt{
			UID:        configPrompt.UID,
			GoalUID:    configPrompt.GoalUID,
			Model:      configPrompt.Model,
			Parameters: configPrompt.Parameters,
			Messages:   configPrompt.Messages,
			Weight:     configPrompt.Weight,
			IsCanary:   configPrompt.IsCanary,
			MaxRuns:    configPrompt.MaxRuns,
			SourceFile: filename,
			SourceType: "config",
			VarName:    generateVarName(configPrompt.UID, "Prompt"),
		}

		// Set defaults
		if prompt.Weight == 0 {
			prompt.Weight = 100
		}

		// Validate required fields
		if prompt.UID == "" {
			result.Errors = append(result.Errors, ParseError{
				File:    filename,
				Message: "Prompt missing required UID field",
				Type:    "error",
			})
			continue
		}
		if prompt.GoalUID == "" {
			result.Errors = append(result.Errors, ParseError{
				File:    filename,
				Message: fmt.Sprintf("Prompt '%s' missing required goal_uid field", prompt.UID),
				Type:    "error",
			})
			continue
		}
		if prompt.Model == "" {
			result.Errors = append(result.Errors, ParseError{
				File:    filename,
				Message: fmt.Sprintf("Prompt '%s' missing required model field", prompt.UID),
				Type:    "error",
			})
			continue
		}
		if len(prompt.Messages) == 0 {
			result.Errors = append(result.Errors, ParseError{
				File:    filename,
				Message: fmt.Sprintf("Prompt '%s' missing required messages field", prompt.UID),
				Type:    "error",
			})
			continue
		}

		result.Prompts = append(result.Prompts, prompt)
	}

	return nil
}

// generateVarName creates a valid Go variable name from a UID
func generateVarName(uid, suffix string) string {
	// Convert UID to camelCase and add suffix
	parts := strings.FieldsFunc(uid, func(r rune) bool {
		return r == '-' || r == '_' || r == ' '
	})

	if len(parts) == 0 {
		return "unnamed" + suffix
	}

	var result strings.Builder
	result.WriteString(strings.ToLower(parts[0]))

	for i := 1; i < len(parts); i++ {
		if parts[i] != "" {
			result.WriteString(strings.Title(strings.ToLower(parts[i])))
		}
	}

	result.WriteString(suffix)
	return result.String()
}

// MergeResults combines results from Go files and config files, handling conflicts
func MergeResults(goResult, configResult *ParseResult) *ParseResult {
	merged := &ParseResult{
		Goals:            make([]DiscoveredGoal, 0),
		Prompts:          make([]DiscoveredPrompt, 0),
		Errors:           make([]ParseError, 0),
		RawGoalFunctions: make(map[string]bool),
	}

	// Copy all errors
	merged.Errors = append(merged.Errors, goResult.Errors...)
	merged.Errors = append(merged.Errors, configResult.Errors...)

	// Merge goals - Go definitions take priority
	goalMap := make(map[string]DiscoveredGoal)

	// Add config goals first
	for _, goal := range configResult.Goals {
		goalMap[goal.UID] = goal
	}

	// Add Go goals, overriding config goals and warning about conflicts
	for _, goal := range goResult.Goals {
		if existing, exists := goalMap[goal.UID]; exists && existing.SourceType == "config" {
			merged.Errors = append(merged.Errors, ParseError{
				File:    goal.SourceFile,
				Message: fmt.Sprintf("Goal '%s' defined in both Go code and config file. Using Go definition.", goal.UID),
				Type:    "warning",
			})
		}
		goalMap[goal.UID] = goal
	}

	// Convert map back to slice
	for _, goal := range goalMap {
		merged.Goals = append(merged.Goals, goal)
	}

	// Merge prompts - Go definitions take priority
	promptMap := make(map[string]DiscoveredPrompt)

	// Add config prompts first
	for _, prompt := range configResult.Prompts {
		promptMap[prompt.UID] = prompt
	}

	// Add Go prompts, overriding config prompts and warning about conflicts
	for _, prompt := range goResult.Prompts {
		if existing, exists := promptMap[prompt.UID]; exists && existing.SourceType == "config" {
			merged.Errors = append(merged.Errors, ParseError{
				File:    prompt.SourceFile,
				Message: fmt.Sprintf("Prompt '%s' defined in both Go code and config file. Using Go definition.", prompt.UID),
				Type:    "warning",
			})
		}
		promptMap[prompt.UID] = prompt
	}

	// Convert map back to slice
	for _, prompt := range promptMap {
		merged.Prompts = append(merged.Prompts, prompt)
	}

	// Merge RawGoalFunctions maps - combine both config and Go-detected raw functions
	for goalUID := range configResult.RawGoalFunctions {
		merged.RawGoalFunctions[goalUID] = true
	}
	for goalUID := range goResult.RawGoalFunctions {
		merged.RawGoalFunctions[goalUID] = true
	}

	return merged
}

// hasAtLeastOneField checks if a JSON object has at least one field
func hasAtLeastOneField(jsonData []byte) bool {
	var obj map[string]interface{}
	if err := json.Unmarshal(jsonData, &obj); err != nil {
		return false // Invalid JSON
	}
	return len(obj) > 0
}