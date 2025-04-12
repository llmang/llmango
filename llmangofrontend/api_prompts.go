package llmangofrontend

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/llmang/llmango/llmango"
	"github.com/llmang/llmango/openrouter"
)

// handleGetPrompts handles getting all prompts with pagination and optional goal filtering
func (r *APIRouter) handleGetPrompts(w http.ResponseWriter, req *http.Request) {
	limit := 10 // default limit
	if limitStr := req.Header.Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Get goal UID from query parameter
	goalUID := req.URL.Query().Get("goaluid")

	// Get all prompts
	prompts := make(map[string]*llmango.Prompt)
	for uid, prompt := range r.Prompts {
		// If goalUID is specified, only include prompts used by that goal
		if goalUID != "" {
			usedByGoal := false
			for _, goal := range r.Goals {
				if goalInfo, ok := goal.(interface{ GetGoalInfo() *llmango.GoalInfo }); ok {
					for _, solution := range goalInfo.GetGoalInfo().Solutions {
						if solution.PromptUID == uid {
							usedByGoal = true
							break
						}
					}
				}
			}
			if !usedByGoal {
				continue
			}
		}

		prompts[uid] = prompt
		if len(prompts) >= limit {
			break
		}
	}

	json.NewEncoder(w).Encode(prompts)
}

// handleGetPrompt handles getting a single prompt by UID
func (r *APIRouter) handleGetPrompt(w http.ResponseWriter, req *http.Request) {
	promptUID := req.PathValue("promptuid")
	if promptUID == "" {
		BadRequest(w, "Missing prompt UID")
		return
	}

	prompt, exists := r.Prompts[promptUID]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Prompt not found"))
		return
	}

	json.NewEncoder(w).Encode(prompt)
}

// handleGetPromptLogs handles log queries for a specific prompt
func (r *APIRouter) handleGetPromptLogs(w http.ResponseWriter, req *http.Request) {
	promptUID := req.PathValue("promptuid")
	if promptUID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Missing prompt ID")
		return
	}

	// Check if logging is enabled
	if r.Logging == nil || r.Logging.GetLogs == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode("Logging is not enabled in this LLMango implementation")
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
		PromptUID: &promptUID,
		Limit:     perPage,
		Offset:    (page - 1) * perPage,
	}

	// Get logs
	logs, total, err := r.Logging.GetLogs(filter)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("Failed to get logs: " + err.Error())
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

	json.NewEncoder(w).Encode(response)
}

// handleDeletePrompt handles the deletion of a prompt
func (r *APIRouter) handleDeletePrompt(w http.ResponseWriter, req *http.Request) {
	var deleteReq struct {
		PromptUID string `json:"promptuid"`
	}

	if err := json.NewDecoder(req.Body).Decode(&deleteReq); err != nil {
		BadRequest(w, "Invalid request body")
		return
	}

	promptUID := deleteReq.PromptUID
	if promptUID == "" {
		BadRequest(w, "Prompt ID is required")
		return
	}

	if r.LLMangoManager.Prompts == nil {
		ServerError(w, fmt.Errorf("prompts map not initialized"))
		return
	}

	// Check if prompt exists
	if _, exists := r.LLMangoManager.Prompts[promptUID]; !exists {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Prompt not found"))
		return
	}

	// Delete the prompt
	delete(r.LLMangoManager.Prompts, promptUID)

	// Save state if SaveState function is set
	if r.LLMangoManager.SaveState != nil {
		if err := r.LLMangoManager.SaveState(); err != nil {
			ServerError(w, err)
			return
		}
	}

	w.Write([]byte("Prompt deleted successfully"))
}

// handleCreatePrompt creates a new prompt
func (r *APIRouter) handleCreatePrompt(w http.ResponseWriter, req *http.Request) {
	// Parse request body
	var createReq struct {
		UID        string            `json:"uid"`
		Model      string            `json:"model"`
		Parameters map[string]any    `json:"parameters"`
		Messages   []json.RawMessage `json:"messages"`
	}

	if err := json.NewDecoder(req.Body).Decode(&createReq); err != nil {
		BadRequest(w, "Invalid request body")
		return
	}

	// Create prompt ID
	promptUID := generateUID()
	if createReq.UID != "" {
		promptUID = createReq.UID
	}

	// Create prompt
	prompt := &llmango.Prompt{
		UID:        promptUID,
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
				BadRequest(w, "Invalid message format")
				return
			}
		}
	}

	// Add prompt to the map
	r.Prompts[promptUID] = prompt

	// Save state after creating the prompt
	if r.SaveState != nil {
		if err := r.SaveState(); err != nil {
			ServerError(w, err)
			return
		}
	}

	json.NewEncoder(w).Encode(map[string]string{
		"promptUID": promptUID,
	})
}

// handleUpdatePrompt updates an existing prompt
func (r *APIRouter) handleUpdatePrompt(w http.ResponseWriter, req *http.Request) {
	// Extract prompt ID from URL pattern
	promptUID := req.PathValue("promptuid")
	if promptUID == "" {
		BadRequest(w, "Missing prompt ID")
		return
	}

	prompt, exists := r.Prompts[promptUID]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Prompt not found"))
		return
	}

	// Parse request body
	var updateReq struct {
		Model      string            `json:"model"`
		Parameters map[string]any    `json:"parameters"`
		Messages   []json.RawMessage `json:"messages"`
	}

	if err := json.NewDecoder(req.Body).Decode(&updateReq); err != nil {
		BadRequest(w, "Invalid request body")
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
				BadRequest(w, "Invalid message format")
				return
			}
		}
	}

	// Save state after updating the prompt
	if r.SaveState != nil {
		if err := r.SaveState(); err != nil {
			ServerError(w, err)
			return
		}
	}

	// Return the updated prompt as JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prompt)
}
