package llmangofrontend

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/llmang/llmango/llmango"
	"github.com/llmang/llmango/openrouter"
)

// handleGetPrompts handles getting all prompts with pagination and optional goal filtering
func (r *APIRouter) handleGetPrompts(w http.ResponseWriter, req *http.Request) {
	limit := -1 // default limit
	if limitStr := req.Header.Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Convert map to slice for sorting
	prompts := make([]*llmango.Prompt, 0, len(r.Prompts))
	for _, prompt := range r.Prompts {
		prompts = append(prompts, prompt)
	}

	// Sort prompts by UpdatedAt (desc), CreatedAt (desc), UID (asc)
	sort.Slice(prompts, func(i, j int) bool {
		p1 := prompts[i]
		p2 := prompts[j]
		if p1.UpdatedAt != p2.UpdatedAt {
			return p1.UpdatedAt > p2.UpdatedAt // Most recent UpdatedAt first
		}
		if p1.CreatedAt != p2.CreatedAt {
			return p1.CreatedAt > p2.CreatedAt // Most recent CreatedAt first
		}
		return p1.UID < p2.UID // Alphabetical UID for ties
	})

	// Apply limit if specified
	if limit > 0 && limit < len(prompts) {
		prompts = prompts[:limit]
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
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Prompt deleted successfully",
	})
}

// handleCreatePrompt creates a new prompt
func (r *APIRouter) handleCreatePrompt(w http.ResponseWriter, req *http.Request) {
	// Parse request body
	var prompt *llmango.Prompt

	if err := json.NewDecoder(req.Body).Decode(&prompt); err != nil {
		BadRequest(w, "Invalid request body")
		return
	}
	if prompt.UID == "" {
		prompt.UID = generateUID()
	}

	// Sanitize UID to be URL-safe and readable
	prompt.UID = strings.TrimSpace(prompt.UID)
	prompt.UID = strings.ReplaceAll(prompt.UID, " ", "_") // Replace spaces with underscores
	prompt.UID = url.PathEscape(prompt.UID)

	// Add prompt to the map
	r.Prompts[prompt.UID] = prompt

	// Add the prompt UID to the corresponding goal's PromptUIDs list
	if err := r.AddPromptToGoal(prompt.GoalUID, prompt.UID); err != nil {
		log.Printf("Error adding prompt %s to goal %s: %v", prompt.UID, prompt.GoalUID, err)
	}

	// Save state after creating the prompt and updating the goal
	if r.SaveState != nil {
		if err := r.SaveState(); err != nil {
			ServerError(w, err)
			return
		}
	}

	json.NewEncoder(w).Encode(map[string]string{
		"promptUID": prompt.UID,
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
		Weight     int               `json:"weight"`
		IsCanary   bool              `json:"isCanary"`
		MaxRuns    int               `json:"maxRuns"`
	}

	if err := json.NewDecoder(req.Body).Decode(&updateReq); err != nil {
		BadRequest(w, "Invalid request body")
		return
	}

	// Update prompt
	prompt.Model = updateReq.Model
	prompt.Weight = updateReq.Weight
	prompt.IsCanary = updateReq.IsCanary
	prompt.MaxRuns = updateReq.MaxRuns

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
