package agent

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/llmang/llmango/llmangoagents"
	"github.com/llmang/llmango/openrouter"
)

// AgentSystem holds the agent system manager
type AgentSystem struct {
	Manager *llmangoagents.AgentSystemManager
	Debug   bool
}

// AgentRequest represents the JSON request structure for agent endpoints
type AgentRequest struct {
	Input       string `json:"input"`
	WorkflowUID string `json:"workflowUID"`
}

// AgentResponse represents the JSON response structure for agent endpoints
type AgentResponse struct {
	Output string `json:"output"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// SetupAgentSystem initializes the agent system with the given OpenRouter instance
func SetupAgentSystem(openRouter *openrouter.OpenRouter) (*AgentSystem, error) {
	// Load JSON config from agents.json
	config, err := (&llmangoagents.AgentSystemManager{}).LoadJSONConfig("agents.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load agent config: %w", err)
	}
	for i, a := range config.Agents {
		fmt.Printf("Agent %d: %v \n\n", i, a)
	}
	// Create system input list
	systemInputs := llmangoagents.SystemInputList{
		Tools:             []llmangoagents.Tool{}, // No custom tools for now
		CustomToolConfigs: config.CustomToolConfigs,
		Agents:            config.Agents,
		Workflows:         config.Workflows,
	}

	// Validate and create agent system manager
	agentSystemManager, err := llmangoagents.ValidateSystemWithDependencies(systemInputs)
	if err != nil {
		return nil, fmt.Errorf("failed to validate agent system: %w", err)
	}

	// Set the OpenRouter instance
	agentSystemManager.Openrouter = openRouter

	// Initialize global key bank if needed
	if agentSystemManager.GlobalKeyBank == nil {
		agentSystemManager.GlobalKeyBank = make(map[string]string)
	}

	agentSystem := &AgentSystem{
		Manager: agentSystemManager,
		Debug:   false,
	}

	fmt.Printf("✅ Agent system initialized with %d agents and %d workflows\n",
		len(agentSystemManager.Agents), len(agentSystemManager.Workflows))

	return agentSystem, nil
}

// HandleAgentRequest handles HTTP requests to the agent endpoint
func (as *AgentSystem) HandleAgentRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req AgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAgentErrorResponse(w, "Invalid JSON request: "+err.Error())
		return
	}

	// Validate request
	if req.Input == "" {
		writeAgentErrorResponse(w, "Input is required")
		return
	}
	if req.WorkflowUID == "" {
		writeAgentErrorResponse(w, "WorkflowUID is required")
		return
	}

	// Check if agent system is initialized
	if as.Manager == nil {
		writeAgentErrorResponse(w, "Agent system not initialized")
		return
	}

	if as.Debug {
		fmt.Printf("=== Agent Request ===\n")
		fmt.Printf("Input: %s\n", req.Input)
		fmt.Printf("WorkflowUID: %s\n", req.WorkflowUID)
	}

	// Execute workflow
	workflowInstance, err := as.Manager.StartNewWorkflowInstance(req.WorkflowUID, 1, req.Input)
	if err != nil {
		if as.Debug {
			fmt.Printf("Workflow execution failed: %v\n", err)
		}
		writeAgentErrorResponse(w, "Workflow execution failed: "+err.Error())
		return
	}

	// Get the result from the workflow context
	result := ""
	if workflowInstance.Context != nil && workflowInstance.Context.GlobalKeyBank != nil {
		if finalResult, exists := workflowInstance.Context.GlobalKeyBank["final_result"]; exists {
			result = finalResult
		}
	}

	// If no result found, use a default message
	if result == "" {
		result = "Workflow completed successfully but no output was generated"
	}

	if as.Debug {
		fmt.Printf("Workflow completed with status: %s\n", workflowInstance.Status)
		fmt.Printf("Result: %s\n", result)
		fmt.Printf("=== Agent Request Complete ===\n")
	}

	// Write success response
	response := AgentResponse{
		Output: result,
		Status: workflowInstance.Status,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SetDebug enables or disables debug logging
func (as *AgentSystem) SetDebug(enabled bool) {
	as.Debug = enabled
}

// writeAgentErrorResponse writes an error response for agent requests
func writeAgentErrorResponse(w http.ResponseWriter, errorMsg string) {
	response := AgentResponse{
		Output: "",
		Status: "error",
		Error:  errorMsg,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(response)
}
