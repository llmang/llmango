package llmangoagents

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

//serializable at its core. This is both how frontend users and golang devs will build agents. By unifying the system we allow for standardized reusable logic

// These are the objects that Go developers work with directly
// Following llmango pattern: developers work with these, then they get compiled to runtime objects later
//Developers use these structs, systems or frontend users use configs.

type GlobalSettings struct {
	KeyBank             map[string]string `json:"keyBank"`             // Environment variable mapping
	CompatibilityCutoff int               `json:"compatibilityCutoff"` // Unix timestamp
	DefaultLimits       WorkflowLimits    `json:"defaultLimits"`
}

type WorkflowLimits struct {
	MaxTime  int `json:"maxTime"`  // seconds
	MaxSteps int `json:"maxSteps"` // number of steps
	MaxSpend int `json:"maxSpend"` // cost units
}

type HTTPToolBuilderConfig struct {
	UID         string `json:"uid"`
	Type        string `json:"type"` // "builtin" | "http" | "function"
	Name        string `json:"name"`
	Description string `json:"description"`

	Endpoint        string `json:"endpoint"`        // POST endpoint for the tool
	ExtraHeaders    string `json:"extraHeaders"`    // Headers as JSON or key:value\nkey:value
	RequiredSecrets string `json:"requiredSecrets"` // comma-separated list of required secrets

	InputSchema  json.RawMessage `json:"inputSchema"`  // JSON schema for validation
	OutputSchema json.RawMessage `json:"outputSchema"` // JSON schema for validation
}

// you will only get the secrets from required secrets this is to provide transparency and saftey regarding secret sharing
type Tool struct {
	Uid             string
	Name            string
	Description     string
	Function        func(map[string]string, json.RawMessage) (json.RawMessage, error)
	RequiredSecrets string
	InputSchema     string
	OutputSchema    string
}

// parseHeaders parses header string (key:value\nkey:value) and replaces ${var} with secrets from secretsMap
func parseHeaders(headerStr string, secretsMap map[string]string) (map[string]string, error) {
	headers := make(map[string]string)
	lines := strings.Split(headerStr, "\n")
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		val = re.ReplaceAllStringFunc(val, func(match string) string {
			varName := re.FindStringSubmatch(match)
			if len(varName) == 2 {
				if v, ok := secretsMap[varName[1]]; ok {
					return v
				}
			}
			return match
		})
		headers[key] = val
	}
	return headers, nil
}

// CreateHTTPTool builds a Tool from HTTPToolBuilderConfig
// This is useful as you can setup cloudflare edge functions as tools! Remember to protect all endpoints with proper authentication
func CreateHTTPTool(input HTTPToolBuilderConfig) (*Tool, error) {
	if input.UID == "" || input.Name == "" {
		return nil, fmt.Errorf("UID and Name are required")
	}
	// The function expects: secrets map, input json.RawMessage
	toolFunc := func(secrets map[string]string, payload json.RawMessage) (json.RawMessage, error) {
		// Prepare headers
		headers := map[string]string{
			"Content-Type": "application/json",
		}
		if input.ExtraHeaders != "" {
			parsedHeaders, err := parseHeaders(input.ExtraHeaders, secrets)
			if err != nil {
				return nil, fmt.Errorf("failed to parse headers: %w", err)
			}
			for k, v := range parsedHeaders {
				headers[k] = v
			}
		}
		// Prepare request
		req, err := http.NewRequest("POST", input.Endpoint, bytes.NewReader(payload))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("http request failed: %w", err)
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}
		return json.RawMessage(body), nil
	}
	return &Tool{
		Uid:             input.UID,
		Name:            input.Name,
		Description:     input.Description,
		Function:        toolFunc,
		RequiredSecrets: input.RequiredSecrets,
		InputSchema:     string(input.InputSchema),
		OutputSchema:    string(input.OutputSchema),
	}, nil
}

// customToolBuilder builds a Tool from direct Tool params and validates them
func createTool(uid, name, description, inputSchema, outputSchema string, requiredSecrets []string, fn func(map[string]string, json.RawMessage) (json.RawMessage, error)) (*Tool, error) {
	if uid == "" || name == "" {
		return nil, fmt.Errorf("uid and name are required")
	}
	if fn == nil {
		return nil, fmt.Errorf("function is required")
	}

	// Clean required secrets to remove commas and join them
	cleanedSecrets := make([]string, len(requiredSecrets))
	for i, secret := range requiredSecrets {
		cleanedSecrets[i] = strings.ReplaceAll(strings.TrimSpace(secret), ",", "")
	}
	secretsString := strings.Join(cleanedSecrets, ",")

	return &Tool{
		Uid:             uid,
		Name:            name,
		Description:     description,
		Function:        fn,
		RequiredSecrets: secretsString,
		InputSchema:     inputSchema,
		OutputSchema:    outputSchema,
	}, nil
}

type Agent struct {
	Name          string
	SystemMessage string
	Model         string
	Parameters    string
	Tools         []string //abilities for the agent.
	PreProcessors []string //in order of their usage
	SubAgents     []string
	SubWorkflows  []string
}

// createAgent constructs an Agent from instantiated PreProcessors, Tools, sub-agents, and sub-workflows.
// For correctness, deduplicate each list using a map before converting back to a string array.
func createAgent(
	name string,
	systemMessage string,
	model string,
	parameters string,
	tools []*Tool,
	preProcessors []Preprocess,
	subAgents []*Agent,
	subWorkflows []*Workflow,
) *Agent {
	toolMap := make(map[string]struct{})
	for _, t := range tools {
		toolMap[t.Name] = struct{}{}
	}
	toolNames := make([]string, 0, len(toolMap))
	for name := range toolMap {
		toolNames = append(toolNames, name)
	}

	// PreProcessors are function types, so we still use placeholder names, but deduplicate by index
	preProcessorMap := make(map[string]struct{})
	for i := range preProcessors {
		key := fmt.Sprintf("preprocessor_%d", i)
		preProcessorMap[key] = struct{}{}
	}
	preProcessorNames := make([]string, 0, len(preProcessorMap))
	for name := range preProcessorMap {
		preProcessorNames = append(preProcessorNames, name)
	}

	subAgentMap := make(map[string]struct{})
	for _, a := range subAgents {
		subAgentMap[a.Name] = struct{}{}
	}
	subAgentNames := make([]string, 0, len(subAgentMap))
	for name := range subAgentMap {
		subAgentNames = append(subAgentNames, name)
	}

	subWorkflowMap := make(map[string]struct{})
	for _, w := range subWorkflows {
		subWorkflowMap[w.Name] = struct{}{}
	}
	subWorkflowNames := make([]string, 0, len(subWorkflowMap))
	for name := range subWorkflowMap {
		subWorkflowNames = append(subWorkflowNames, name)
	}

	return &Agent{
		Name:          name,
		SystemMessage: systemMessage,
		Model:         model,
		Parameters:    parameters,
		Tools:         toolNames,
		PreProcessors: preProcessorNames,
		SubAgents:     subAgentNames,
		SubWorkflows:  subWorkflowNames,
	}
}

// createWorkflow constructs a Workflow from the provided parameters
func createWorkflow(
	uid string,
	name string,
	description string,
	options WorkflowLimits,
	steps []*WorkflowStep,
) (*Workflow, error) {
	if uid == "" || name == "" {
		return nil, fmt.Errorf("uid and name are required")
	}

	if len(steps) == 0 {
		return nil, fmt.Errorf("workflow must have at least one step")
	}

	// Validate step UIDs are unique
	stepUIDs := make(map[string]struct{})
	for _, step := range steps {
		if step.UID == "" {
			return nil, fmt.Errorf("all workflow steps must have a UID")
		}
		if _, exists := stepUIDs[step.UID]; exists {
			return nil, fmt.Errorf("duplicate step UID: %s", step.UID)
		}
		stepUIDs[step.UID] = struct{}{}
	}

	return &Workflow{
		UID:         uid,
		Name:        name,
		Description: description,
		Options:     options,
		Steps:       steps,
	}, nil
}

// Workflow - what developers create, gets compiled to runtime Workflow
type Workflow struct {
	UID         string          `json:"uid"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Options     WorkflowLimits  `json:"options"`
	Steps       []*WorkflowStep `json:"steps"`
}

type WorkflowStep struct {
	UID           string   `json:"uid"`
	Agent         string   `json:"agent"`         // Reference to agent UID
	SubAgents     []string `json:"subAgents"`     // References to agent UIDs
	AllowHandoffs bool     `json:"allowHandoffs"` // allows helper agents to pass to one another
	ExitBehavior  string   `json:"exitBehavior"`  // "default", "return|s4", "user"
}
