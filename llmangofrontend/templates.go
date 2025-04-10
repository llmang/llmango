package llmangofrontend

import (
	"encoding/json"
	"html/template"
	"net/http"
	"reflect"

	"github.com/llmang/llmango/llmango"
	"github.com/llmang/llmango/llmangofrontend/templates"
	"github.com/llmang/llmango/openrouter"
)

// Templates holds all the parsed HTML templates
var Templates *template.Template

func init() {
	// Create a function map with helper functions
	funcMap := template.FuncMap{
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, nil
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, nil
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
		"toJSON": func(v interface{}) template.JS {
			a, err := json.MarshalIndent(v, "", "  ")
			if err != nil {
				return template.JS("{}")
			}
			return template.JS(a)
		},
		"inc": func(i int) int {
			return i + 1
		},
		"getGoalInfo": func(goalAny interface{}) map[string]interface{} {
			// Try to get GoalInfo using type assertion first
			if goal, ok := goalAny.(interface{ GetGoalInfo() *llmango.GoalInfo }); ok {
				info := goal.GetGoalInfo()
				result := map[string]interface{}{
					"Title":       info.Title,
					"Description": info.Description,
					"UID":         info.UID,
					"Solutions":   info.Solutions,
				}
				return result
			}

			// Try direct marshaling to see if that works better with generics
			if goalJSON, err := json.Marshal(goalAny); err == nil {
				var result map[string]interface{}
				if err := json.Unmarshal(goalJSON, &result); err == nil {
					// Make sure it has key fields expected by templates
					if _, hasTitle := result["title"]; hasTitle {
						result["Title"] = result["title"]
					}
					if _, hasDesc := result["description"]; hasDesc {
						result["Description"] = result["description"]
					}
					if _, hasUID := result["UID"]; hasUID {
						// It's already capitalized, good
					} else if _, hasUID := result["uid"]; hasUID {
						result["UID"] = result["uid"]
					}
					if _, hasSolutions := result["solutions"]; hasSolutions {
						result["Solutions"] = result["solutions"]
					}
					return result
				}
			}

			// If marshaling fails, fall back to reflection
			v := reflect.ValueOf(goalAny)

			// If it's a pointer, get the value it points to
			if v.Kind() == reflect.Ptr {
				v = v.Elem()
			}

			// Must be a struct to proceed
			if v.Kind() != reflect.Struct {
				return map[string]interface{}{
					"Title":       "Unknown Goal",
					"Description": "Goal information unavailable (not a struct)",
					"Solutions":   make(map[string]*llmango.Solution),
				}
			}

			// Build a map with the goal info
			result := map[string]interface{}{}

			// First, try to extract the embedded GoalInfo fields
			goalInfoField := v.FieldByName("GoalInfo")
			if goalInfoField.IsValid() {
				// Extract fields from embedded GoalInfo
				title := goalInfoField.FieldByName("Title")
				if title.IsValid() {
					result["Title"] = title.String()
				}

				desc := goalInfoField.FieldByName("Description")
				if desc.IsValid() {
					result["Description"] = desc.String()
				}

				uid := goalInfoField.FieldByName("UID")
				if uid.IsValid() {
					result["UID"] = uid.String()
				}

				solutions := goalInfoField.FieldByName("Solutions")
				if solutions.IsValid() {
					result["Solutions"] = solutions.Interface()
				} else {
					result["Solutions"] = make(map[string]*llmango.Solution)
				}
			} else {
				// Try to access fields directly on the struct (they might not be in an embedded GoalInfo)
				title := v.FieldByName("Title")
				if title.IsValid() {
					result["Title"] = title.String()
				}

				desc := v.FieldByName("Description")
				if desc.IsValid() {
					result["Description"] = desc.String()
				}

				uid := v.FieldByName("UID")
				if uid.IsValid() {
					result["UID"] = uid.String()
				}

				solutions := v.FieldByName("Solutions")
				if solutions.IsValid() {
					result["Solutions"] = solutions.Interface()
				}
			}

			// Try to extract example input and output fields
			exampleInput := v.FieldByName("ExampleInput")
			if exampleInput.IsValid() {
				inputJSON, err := json.Marshal(exampleInput.Interface())
				if err == nil {
					result["ExampleInput"] = string(inputJSON)
				}
			}

			exampleOutput := v.FieldByName("ExampleOutput")
			if exampleOutput.IsValid() {
				outputJSON, err := json.Marshal(exampleOutput.Interface())
				if err == nil {
					result["ExampleOutput"] = string(outputJSON)
				}
			}

			// If we couldn't extract title and description, use defaults
			if _, hasTitle := result["Title"]; !hasTitle {
				result["Title"] = "Unknown Goal"
			}

			if _, hasDesc := result["Description"]; !hasDesc {
				result["Description"] = "Goal information unavailable"
			}

			if _, hasSolutions := result["Solutions"]; !hasSolutions {
				result["Solutions"] = make(map[string]*llmango.Solution)
			}

			return result
		},
		"getMessageContent": func(message openrouter.Message) string {
			return message.Content
		},
		"solutionStatus": func(solution *llmango.Solution) string {
			if solution.IsCanary {
				if solution.TotalRuns >= solution.MaxRuns {
					return "COMPLETE"
				}
				return "RUNNING"
			}

			if solution.Weight > 0 {
				return "ON"
			}
			return "OFF"
		},
	}

	// Parse all template files
	Templates = template.New("").Funcs(funcMap)

	// Parse templates from shared components and page templates
	template.Must(Templates.Parse(templates.SharedTemplates))
	template.Must(Templates.Parse(templates.ComponentTemplates))
	template.Must(Templates.Parse(templates.HomeTemplate))
	template.Must(Templates.Parse(templates.PromptsTemplates))
	template.Must(Templates.Parse(templates.GoalsTemplates))
	template.Must(Templates.Parse(templates.ModelsTemplate))
	template.Must(Templates.Parse(templates.LogsTemplates))
	template.Must(Templates.Parse(templates.LogsPageTemplate))
}

// Page data structures
type HomePageData struct {
	Prompts map[string]*llmango.Prompt
	Goals   map[string]any
}

type PromptsPageData struct {
	Prompts map[string]*llmango.Prompt
}

type GoalsPageData struct {
	Goals map[string]any
}

type ModelsPageData struct {
	// No data needed as models are fetched client-side
}

type LogsPageData struct {
	Prompts       map[string]*llmango.Prompt
	Goals         map[string]any
	FilterOptions map[string]interface{}
}

type PromptDetailData struct {
	Prompt    *llmango.Prompt
	PromptUID string
}

type GoalDetailData struct {
	Goal    any
	GoalID  string
	Prompts map[string]*llmango.Prompt
}

// Render templates with data
func (r *Router) renderTemplate(w http.ResponseWriter, tmplName string, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Setup template data with BaseRoute if it's not already a map
	var templateData map[string]interface{}

	// Convert existing data to a map if needed
	if existingMap, ok := data.(map[string]interface{}); ok {
		templateData = existingMap
	} else {
		templateData = make(map[string]interface{})

		// Add the original data with a meaningful key
		if data != nil {
			templateData["Data"] = data

			// Get the data fields and add them directly to the template data
			v := reflect.ValueOf(data)
			if v.Kind() == reflect.Ptr {
				v = v.Elem()
			}
			if v.Kind() == reflect.Struct {
				t := v.Type()
				for i := 0; i < v.NumField(); i++ {
					field := t.Field(i)
					value := v.Field(i)
					templateData[field.Name] = value.Interface()
				}
			}
		}
	}

	// Add the BaseRoute
	templateData["BaseRoute"] = r.BaseRoute

	// Add title based on template name
	switch tmplName {
	case "home":
		templateData["Title"] = "LLMango View"
	case "prompts":
		templateData["Title"] = "LLMango View - Prompts"
	case "prompt-detail":
		templateData["Title"] = "LLMango View - Prompt Detail"
	case "goals":
		templateData["Title"] = "LLMango View - Goals"
	case "goal-detail":
		templateData["Title"] = "LLMango View - Goal Detail"
	default:
		templateData["Title"] = "LLMango View"
	}

	// Execute template with the enhanced data
	err := Templates.ExecuteTemplate(w, tmplName, templateData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
