package llmangofrontend

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"reflect"
	"strconv"

	"github.com/llmang/llmango/llmango"
	"github.com/llmang/llmango/openrouter"
)

// APIRouter handles API endpoints for the LLMango frontend
type APIRouter struct {
	*Router // Embed the main Router to access its fields
}

// RegisterAPIRoutes registers all API routes
func (router *Router) RegisterAPIRoutes(mux *http.ServeMux) {

	r := &APIRouter{Router: router}
	// Create API sub-mux to handle all API routes
	apiMux := http.NewServeMux()

	// Key management
	apiMux.HandleFunc("POST /update-key", r.handleUpdateAPIKey)

	// Goal endpoints
	apiMux.HandleFunc("POST /goals/{goalID}/update", r.handleUpdateGoal)

	// Solution endpoints
	apiMux.HandleFunc("POST /solutions/new", r.handleCreateSolution)
	apiMux.HandleFunc("POST /solutions/{solutionID}/update", r.handleUpdateSolution)
	apiMux.HandleFunc("POST /solutions/{solutionID}/delete", r.handleDeleteSolution)

	// Prompt endpoints
	apiMux.HandleFunc("POST /prompts/new", r.handleCreatePrompt)
	apiMux.HandleFunc("POST /prompts/{promptID}/update", r.handleUpdatePrompt)

	// Logging endpoints
	apiMux.HandleFunc("GET /logs", r.handleGetLogs)
	apiMux.HandleFunc("GET /logs/goal/{goalID}", r.handleGetGoalLogs)
	apiMux.HandleFunc("GET /logs/prompt/{promptID}", r.handleGetPromptLogs)

	// Mount the API router at /api path with StripPrefix
	mux.Handle("/api/", http.StripPrefix("/api", apiMux))
}

// Response is a generic API response
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// sendJSONResponse sends a JSON response
func sendJSONResponse(w http.ResponseWriter, status int, resp Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

// handleUpdateGoal updates a goal's title and description
func (r *APIRouter) handleUpdateGoal(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract goal ID from URL pattern
	goalID := req.PathValue("goalID")
	if goalID == "" {
		sendJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Missing goal ID",
		})
		return
	}

	goalAny, exists := r.Goals[goalID]
	if !exists {
		sendJSONResponse(w, http.StatusNotFound, Response{
			Success: false,
			Error:   "Goal not found",
		})
		return
	}

	// Parse request body
	var updateReq struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(req.Body).Decode(&updateReq); err != nil {
		sendJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	// Get GoalInfo from the any type
	if goal, ok := goalAny.(interface{ GetGoalInfo() *llmango.GoalInfo }); ok {
		goalInfo := goal.GetGoalInfo()
		// Update goal
		goalInfo.Title = updateReq.Title
		goalInfo.Description = updateReq.Description

		// Save state after updating the goal
		if r.SaveState != nil {
			if err := r.SaveState(); err != nil {
				sendJSONResponse(w, http.StatusInternalServerError, Response{
					Success: false,
					Error:   "Failed to save state: " + err.Error(),
				})
				return
			}
		}

		sendJSONResponse(w, http.StatusOK, Response{
			Success: true,
			Message: "Goal updated successfully",
		})
	} else {
		sendJSONResponse(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   "Goal type assertion failed",
		})
	}
}

// handleCreateSolution creates a new solution for a goal
func (r *APIRouter) handleCreateSolution(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var createReq struct {
		GoalID    string `json:"goalId"`
		PromptUID string `json:"promptUid"`
		Weight    int    `json:"weight"`
		IsCanary  bool   `json:"isCanary"`
		MaxRuns   int    `json:"maxRuns"`
	}

	if err := json.NewDecoder(req.Body).Decode(&createReq); err != nil {
		sendJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	// Validate request
	goalAny, exists := r.Goals[createReq.GoalID]
	if !exists {
		sendJSONResponse(w, http.StatusNotFound, Response{
			Success: false,
			Error:   "Goal not found",
		})
		return
	}

	if createReq.PromptUID != "" {
		if _, exists := r.Prompts[createReq.PromptUID]; !exists {
			sendJSONResponse(w, http.StatusBadRequest, Response{
				Success: false,
				Error:   "Prompt not found",
			})
			return
		}
	}

	// Create the solution
	solutionID := generateUID()
	solution := &llmango.Solution{
		PromptUID: createReq.PromptUID,
		Weight:    createReq.Weight,
		IsCanary:  createReq.IsCanary,
		MaxRuns:   createReq.MaxRuns,
		TotalRuns: 0,
	}

	// Use reflection to access and modify the Solutions map in the goal
	v := reflect.ValueOf(goalAny)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// First try direct GoalInfo field
	var solutionsMap reflect.Value
	if goalInfo := v.FieldByName("GoalInfo"); goalInfo.IsValid() {
		solutionsMap = goalInfo.FieldByName("Solutions")
	} else {
		// If no GoalInfo field, try direct Solutions field
		solutionsMap = v.FieldByName("Solutions")
	}

	if !solutionsMap.IsValid() {
		sendJSONResponse(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   "Could not find Solutions field in goal object",
		})
		return
	}

	// Set the solution in the map
	solutionsMap.SetMapIndex(reflect.ValueOf(solutionID), reflect.ValueOf(solution))

	// Save state after creating the solution
	if r.SaveState != nil {
		if err := r.SaveState(); err != nil {
			sendJSONResponse(w, http.StatusInternalServerError, Response{
				Success: false,
				Error:   "Failed to save state: " + err.Error(),
			})
			return
		}
	}

	sendJSONResponse(w, http.StatusCreated, Response{
		Success: true,
		Message: "Solution created successfully",
		Data: map[string]string{
			"solutionId": solutionID,
		},
	})
}

// handleUpdateSolution updates an existing solution
func (r *APIRouter) handleUpdateSolution(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract solution ID from URL pattern
	solutionID := req.PathValue("solutionID")
	if solutionID == "" {
		sendJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Missing solution ID",
		})
		return
	}

	// Parse request body
	var updateReq struct {
		GoalID    string `json:"goalId"`
		PromptUID string `json:"promptUid"`
		Weight    int    `json:"weight"`
		IsCanary  bool   `json:"isCanary"`
		MaxRuns   int    `json:"maxRuns"`
	}

	if err := json.NewDecoder(req.Body).Decode(&updateReq); err != nil {
		sendJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	// Validate goal exists
	goalAny, exists := r.Goals[updateReq.GoalID]
	if !exists {
		sendJSONResponse(w, http.StatusNotFound, Response{
			Success: false,
			Error:   "Goal not found",
		})
		return
	}

	// Use reflection to access the Solutions map
	v := reflect.ValueOf(goalAny)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// First try direct GoalInfo field
	var solutionsMap reflect.Value
	if goalInfo := v.FieldByName("GoalInfo"); goalInfo.IsValid() {
		solutionsMap = goalInfo.FieldByName("Solutions")
	} else {
		// If no GoalInfo field, try direct Solutions field
		solutionsMap = v.FieldByName("Solutions")
	}

	if !solutionsMap.IsValid() {
		sendJSONResponse(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   "Could not find Solutions field in goal object",
		})
		return
	}

	// Get the solution from the map
	solutionValue := solutionsMap.MapIndex(reflect.ValueOf(solutionID))
	if !solutionValue.IsValid() {
		sendJSONResponse(w, http.StatusNotFound, Response{
			Success: false,
			Error:   "Solution not found",
		})
		return
	}

	// Get actual solution
	solution := solutionValue.Interface().(*llmango.Solution)

	// Update the solution
	solution.PromptUID = updateReq.PromptUID
	solution.Weight = updateReq.Weight
	solution.IsCanary = updateReq.IsCanary
	solution.MaxRuns = updateReq.MaxRuns

	// Save state after updating the solution
	if r.SaveState != nil {
		if err := r.SaveState(); err != nil {
			sendJSONResponse(w, http.StatusInternalServerError, Response{
				Success: false,
				Error:   "Failed to save state: " + err.Error(),
			})
			return
		}
	}

	sendJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Solution updated successfully",
	})
}

// handleDeleteSolution deletes a solution
func (r *APIRouter) handleDeleteSolution(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract solution ID from URL pattern
	solutionID := req.PathValue("solutionID")
	if solutionID == "" {
		sendJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Missing solution ID",
		})
		return
	}

	// Parse request body to get the goal ID
	var deleteReq struct {
		GoalID string `json:"goalId"`
	}

	if err := json.NewDecoder(req.Body).Decode(&deleteReq); err != nil {
		sendJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	// Validate goal exists
	goalAny, exists := r.Goals[deleteReq.GoalID]
	if !exists {
		sendJSONResponse(w, http.StatusNotFound, Response{
			Success: false,
			Error:   "Goal not found",
		})
		return
	}

	// Use reflection to access the Solutions map
	v := reflect.ValueOf(goalAny)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// First try direct GoalInfo field
	var solutionsMap reflect.Value
	if goalInfo := v.FieldByName("GoalInfo"); goalInfo.IsValid() {
		solutionsMap = goalInfo.FieldByName("Solutions")
	} else {
		// If no GoalInfo field, try direct Solutions field
		solutionsMap = v.FieldByName("Solutions")
	}

	if !solutionsMap.IsValid() {
		sendJSONResponse(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   "Could not find Solutions field in goal object",
		})
		return
	}

	// Check if solution exists
	if !solutionsMap.MapIndex(reflect.ValueOf(solutionID)).IsValid() {
		sendJSONResponse(w, http.StatusNotFound, Response{
			Success: false,
			Error:   "Solution not found",
		})
		return
	}

	// Delete the solution by setting its map entry to zero value
	solutionsMap.SetMapIndex(reflect.ValueOf(solutionID), reflect.Value{})

	// Save state after deleting the solution
	if r.SaveState != nil {
		if err := r.SaveState(); err != nil {
			sendJSONResponse(w, http.StatusInternalServerError, Response{
				Success: false,
				Error:   "Failed to save state: " + err.Error(),
			})
			return
		}
	}

	sendJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Solution deleted successfully",
	})
}

// handleCreatePrompt creates a new prompt
func (r *APIRouter) handleCreatePrompt(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var createReq struct {
		UID        string            `json:"uid"`
		Model      string            `json:"model"`
		Parameters map[string]any    `json:"parameters"`
		Messages   []json.RawMessage `json:"messages"`
	}

	if err := json.NewDecoder(req.Body).Decode(&createReq); err != nil {
		sendJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	// Create prompt ID
	promptID := generateUID()
	if createReq.UID != "" {
		promptID = createReq.UID
	}

	// Create prompt
	prompt := &llmango.Prompt{
		UID:        promptID,
		Model:      createReq.Model,
		Parameters: openrouter.Parameters{},
		Messages:   []openrouter.Message{},
	}

	// Handle parameters
	if createReq.Parameters != nil {
		if temperature, ok := createReq.Parameters["temperature"].(float64); ok {
			prompt.Parameters.Temperature = &temperature
		}
		if maxTokens, ok := createReq.Parameters["max_tokens"].(float64); ok {
			maxTokensInt := int(maxTokens)
			prompt.Parameters.MaxTokens = &maxTokensInt
		}
		if topP, ok := createReq.Parameters["top_p"].(float64); ok {
			prompt.Parameters.TopP = &topP
		}
		if frequencyPenalty, ok := createReq.Parameters["frequency_penalty"].(float64); ok {
			prompt.Parameters.FrequencyPenalty = &frequencyPenalty
		}
		if presencePenalty, ok := createReq.Parameters["presence_penalty"].(float64); ok {
			prompt.Parameters.PresencePenalty = &presencePenalty
		}
	}

	if len(createReq.Messages) > 0 {
		prompt.Messages = make([]openrouter.Message, len(createReq.Messages))
		for i, msgData := range createReq.Messages {
			if err := json.Unmarshal(msgData, &prompt.Messages[i]); err != nil {
				sendJSONResponse(w, http.StatusBadRequest, Response{
					Success: false,
					Error:   "Invalid message format",
				})
				return
			}
		}
	}

	// Add prompt to the map
	r.Prompts[promptID] = prompt

	// Save state after creating the prompt
	if r.SaveState != nil {
		if err := r.SaveState(); err != nil {
			sendJSONResponse(w, http.StatusInternalServerError, Response{
				Success: false,
				Error:   "Failed to save state: " + err.Error(),
			})
			return
		}
	}

	sendJSONResponse(w, http.StatusCreated, Response{
		Success: true,
		Message: "Prompt created successfully",
		Data: map[string]string{
			"promptId": promptID,
		},
	})
}

// handleUpdatePrompt updates an existing prompt
func (r *APIRouter) handleUpdatePrompt(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract prompt ID from URL pattern
	promptID := req.PathValue("promptID")
	if promptID == "" {
		sendJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Missing prompt ID",
		})
		return
	}

	prompt, exists := r.Prompts[promptID]
	if !exists {
		sendJSONResponse(w, http.StatusNotFound, Response{
			Success: false,
			Error:   "Prompt not found",
		})
		return
	}

	// Parse request body
	var updateReq struct {
		Model      string            `json:"model"`
		Parameters map[string]any    `json:"parameters"`
		Messages   []json.RawMessage `json:"messages"`
	}

	if err := json.NewDecoder(req.Body).Decode(&updateReq); err != nil {
		sendJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	// Update prompt
	prompt.Model = updateReq.Model

	// Handle parameters
	if updateReq.Parameters != nil {
		// Reset parameters to avoid lingering values
		prompt.Parameters = openrouter.Parameters{}

		if temperature, ok := updateReq.Parameters["temperature"].(float64); ok {
			prompt.Parameters.Temperature = &temperature
		}
		if maxTokens, ok := updateReq.Parameters["max_tokens"].(float64); ok {
			maxTokensInt := int(maxTokens)
			prompt.Parameters.MaxTokens = &maxTokensInt
		}
		if topP, ok := updateReq.Parameters["top_p"].(float64); ok {
			prompt.Parameters.TopP = &topP
		}
		if frequencyPenalty, ok := updateReq.Parameters["frequency_penalty"].(float64); ok {
			prompt.Parameters.FrequencyPenalty = &frequencyPenalty
		}
		if presencePenalty, ok := updateReq.Parameters["presence_penalty"].(float64); ok {
			prompt.Parameters.PresencePenalty = &presencePenalty
		}
	}

	if len(updateReq.Messages) > 0 {
		prompt.Messages = make([]openrouter.Message, len(updateReq.Messages))
		for i, msgData := range updateReq.Messages {
			if err := json.Unmarshal(msgData, &prompt.Messages[i]); err != nil {
				sendJSONResponse(w, http.StatusBadRequest, Response{
					Success: false,
					Error:   "Invalid message format",
				})
				return
			}
		}
	}

	// Save state after updating the prompt
	if r.SaveState != nil {
		if err := r.SaveState(); err != nil {
			sendJSONResponse(w, http.StatusInternalServerError, Response{
				Success: false,
				Error:   "Failed to save state: " + err.Error(),
			})
			return
		}
	}

	sendJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Prompt updated successfully",
	})
}

// handleUpdateAPIKey updates the OpenRouter API key
func (r *APIRouter) handleUpdateAPIKey(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var updateReq struct {
		ApiKey string `json:"apiKey"`
	}

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&updateReq)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	if updateReq.ApiKey == "" {
		http.Error(w, "API key cannot be empty", http.StatusBadRequest)
		return
	}

	// Ensure OpenRouter config exists
	if r.LLMangoManager.OpenRouter == nil {
		http.Error(w, "OpenRouter configuration not initialized", http.StatusInternalServerError)
		return
	}

	r.LLMangoManager.OpenRouter.ApiKey = updateReq.ApiKey
	// Optionally: Persist the key somewhere if needed beyond the struct's lifetime

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "API Key updated successfully.")
}

// generateUID generates a unique ID
func generateUID() string {
	// Simple implementation for now - in production use a more robust method
	return "uid_" + randomString(8)
}

// randomString generates a random string of the specified length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[randomInt(len(charset))]
	}
	return string(b)
}

// randomInt returns a random int in range [0, max)
func randomInt(max int) int {
	// Simple implementation for now
	return rand.Intn(max)
}

// LogFilter represents the filter parameters for log queries
type LogFilter struct {
	GoalID   string `json:"goalId,omitempty"`
	PromptID string `json:"promptId,omitempty"`
	Level    string `json:"level,omitempty"`
	Page     int    `json:"page"`
	PerPage  int    `json:"perPage"`
}

// PaginationResponse represents pagination information for API responses
type PaginationResponse struct {
	Total      int `json:"total"`
	Page       int `json:"page"`
	PerPage    int `json:"perPage"`
	TotalPages int `json:"totalPages"`
}

// LogResponse represents the response structure for log queries
type LogResponse struct {
	Logs       []llmango.LLMangoLog `json:"logs"`
	Pagination PaginationResponse   `json:"pagination"`
}

// handleGetLogs handles general log queries with filters
func (r *APIRouter) handleGetLogs(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if logging is enabled
	if r.Logging == nil || r.Logging.GetLogs == nil {
		sendJSONResponse(w, http.StatusServiceUnavailable, Response{
			Success: false,
			Error:   "Logging is not enabled in this LLMango implementation",
		})
		return
	}

	// Parse query parameters
	page := 1
	perPage := 10
	if pageStr := req.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if perPageStr := req.URL.Query().Get("perPage"); perPageStr != "" {
		if p, err := strconv.Atoi(perPageStr); err == nil && p > 0 {
			perPage = p
		}
	}

	// Convert our filter to llmango's filter type
	filter := &llmango.LLmangoLogFilter{
		Limit:  perPage,
		Offset: (page - 1) * perPage,
	}

	if goalID := req.URL.Query().Get("goalId"); goalID != "" {
		filter.GoalUID = &goalID
	}
	if promptID := req.URL.Query().Get("promptId"); promptID != "" {
		filter.PromptUID = &promptID
	}

	// Get logs
	logs, total, err := r.Logging.GetLogs(filter)
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   "Failed to get logs: " + err.Error(),
		})
		return
	}

	// Calculate pagination using the returned total count
	totalPages := (total + perPage - 1) / perPage
	if totalPages == 0 {
		totalPages = 1
	}

	response := LogResponse{
		Logs: logs,
		Pagination: PaginationResponse{
			Total:      total,
			Page:       page,
			PerPage:    perPage,
			TotalPages: totalPages,
		},
	}

	sendJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Data:    response,
	})
}

// handleGetGoalLogs handles log queries for a specific goal
func (r *APIRouter) handleGetGoalLogs(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	goalID := req.PathValue("goalID")
	if goalID == "" {
		sendJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Missing goal ID",
		})
		return
	}

	// Check if logging is enabled
	if r.Logging == nil || r.Logging.GetLogs == nil {
		sendJSONResponse(w, http.StatusServiceUnavailable, Response{
			Success: false,
			Error:   "Logging is not enabled in this LLMango implementation",
		})
		return
	}

	// Parse pagination parameters
	page := 1
	perPage := 10
	if pageStr := req.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if perPageStr := req.URL.Query().Get("perPage"); perPageStr != "" {
		if p, err := strconv.Atoi(perPageStr); err == nil && p > 0 {
			perPage = p
		}
	}

	filter := &llmango.LLmangoLogFilter{
		GoalUID: &goalID,
		Limit:   perPage,
		Offset:  (page - 1) * perPage,
	}

	// Get logs
	logs, total, err := r.Logging.GetLogs(filter)
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   "Failed to get logs: " + err.Error(),
		})
		return
	}

	// Calculate pagination using the returned total count
	totalPages := (total + perPage - 1) / perPage
	if totalPages == 0 {
		totalPages = 1
	}

	response := LogResponse{
		Logs: logs,
		Pagination: PaginationResponse{
			Total:      total,
			Page:       page,
			PerPage:    perPage,
			TotalPages: totalPages,
		},
	}

	sendJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Data:    response,
	})
}

// handleGetPromptLogs handles log queries for a specific prompt
func (r *APIRouter) handleGetPromptLogs(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	promptID := req.PathValue("promptID")
	if promptID == "" {
		sendJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Missing prompt ID",
		})
		return
	}

	// Check if logging is enabled
	if r.Logging == nil || r.Logging.GetLogs == nil {
		sendJSONResponse(w, http.StatusServiceUnavailable, Response{
			Success: false,
			Error:   "Logging is not enabled in this LLMango implementation",
		})
		return
	}

	// Parse pagination parameters
	page := 1
	perPage := 10
	if pageStr := req.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if perPageStr := req.URL.Query().Get("perPage"); perPageStr != "" {
		if p, err := strconv.Atoi(perPageStr); err == nil && p > 0 {
			perPage = p
		}
	}

	filter := &llmango.LLmangoLogFilter{
		PromptUID: &promptID,
		Limit:     perPage,
		Offset:    (page - 1) * perPage,
	}

	// Get logs
	logs, total, err := r.Logging.GetLogs(filter)
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   "Failed to get logs: " + err.Error(),
		})
		return
	}

	// Calculate pagination using the returned total count
	totalPages := (total + perPage - 1) / perPage
	if totalPages == 0 {
		totalPages = 1
	}

	response := LogResponse{
		Logs: logs,
		Pagination: PaginationResponse{
			Total:      total,
			Page:       page,
			PerPage:    perPage,
			TotalPages: totalPages,
		},
	}

	sendJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Data:    response,
	})
}
