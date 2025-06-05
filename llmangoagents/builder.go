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

// createTool builds a Tool from direct Tool params and validates them
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

// createAgent constructs an Agent from instantiated PreProcessors, Tools, sub-agents, and sub-workflows.
// For correctness, deduplicate each list using a map before converting back to a string array.
func createAgent(
	uid string,
	name string,
	systemMessage string,
	model string,
	parameters string,
	tools []*Tool,
	preProcessors []string,
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
		subAgentMap[a.UID] = struct{}{}
	}
	subAgentNames := make([]string, 0, len(subAgentMap))
	for uid := range subAgentMap {
		subAgentNames = append(subAgentNames, uid)
	}

	subWorkflowMap := make(map[string]struct{})
	for _, w := range subWorkflows {
		subWorkflowMap[w.UID] = struct{}{}
	}
	subWorkflowNames := make([]string, 0, len(subWorkflowMap))
	for uid := range subWorkflowMap {
		subWorkflowNames = append(subWorkflowNames, uid)
	}

	return &Agent{
		UID:           uid,
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

// CreateAgentSystemManager creates and validates an agent system manager
func CreateAgentSystemManager(
	inputs SystemInputList,
) (*AgentSystemManager, error) {
	return ValidateSystemWithDependencies(inputs)
}