package llmangofrontend

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// APIRouter handles API endpoints for the LLMango frontend
type APIRouter struct {
	*Router // Embed the main Router to access its fields
}

// APIRouter returns a subrouter with all API routes configured
func (router *Router) CreateAPIRouter() *http.ServeMux {
	r := &APIRouter{Router: router}
	apiMux := http.NewServeMux()

	// Key management
	apiMux.HandleFunc("POST /update-key", r.handleUpdateAPIKey)

	// Goal endpoints
	apiMux.HandleFunc("GET /goals", r.handleGetGoals)
	apiMux.HandleFunc("GET /goal/{goaluid}", r.handleGetGoal)
	apiMux.HandleFunc("POST /goal/{goaluid}/update", r.handleUpdateGoal)

	// Solution endpoints
	apiMux.HandleFunc("POST /solution/create", r.handleCreateSolution)
	apiMux.HandleFunc("POST /solutions/{solutionuid}/update", r.handleUpdateSolution)
	apiMux.HandleFunc("POST /solutions/{solutionuid}/delete", r.handleDeleteSolution)

	// Prompt endpoints
	apiMux.HandleFunc("GET /prompts", r.handleGetPrompts)
	apiMux.HandleFunc("GET /prompts/{promptuid}", r.handleGetPrompt)
	apiMux.HandleFunc("POST /prompt/create", r.handleCreatePrompt)
	apiMux.HandleFunc("POST /prompts/{promptuid}/update", r.handleUpdatePrompt)

	// Logging endpoints
	apiMux.HandleFunc("POST /logs", r.handleGetLogs)
	apiMux.HandleFunc("POST /logs/goal/{goaluid}", r.handleGetGoalLogs)
	apiMux.HandleFunc("POST /logs/prompt/{promptuid}", r.handleGetPromptLogs)

	// Register API routes
	apiMux.HandleFunc("POST /prompt/delete", r.handleDeletePrompt)

	return apiMux
}

// handleUpdateAPIKey updates the OpenRouter API key
func (r *APIRouter) handleUpdateAPIKey(w http.ResponseWriter, req *http.Request) {
	var updateReq struct {
		ApiKey string `json:"apiKey"`
	}

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&updateReq)
	if err != nil {
		BadRequest(w, "Invalid request body")
		return
	}
	defer req.Body.Close()

	if updateReq.ApiKey == "" {
		BadRequest(w, "API key cannot be empty")
		return
	}

	// Ensure OpenRouter config exists
	if r.LLMangoManager.OpenRouter == nil {
		ServerError(w, fmt.Errorf("OpenRouter configuration not initialized"))
		return
	}

	r.LLMangoManager.OpenRouter.ApiKey = updateReq.ApiKey
	// Optionally: Persist the key somewhere if needed beyond the struct's lifetime

	w.Write([]byte("API Key updated successfully"))
}
