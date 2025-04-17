package llmangofrontend

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

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

	// Get all prompts from the SyncedMap using GetAll
	allPromptsMap := r.Prompts.GetAll()

	// Convert map to slice for sorting
	prompts := make([]*llmango.Prompt, 0, len(allPromptsMap))
	for _, prompt := range allPromptsMap { // Iterate over the returned map
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

	// Use Exists() and Get() methods on r.Prompts
	if !r.Prompts.Exists(promptUID) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Prompt not found"))
		return
	}
	prompt := r.Prompts.Get(promptUID)

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

	if r.LLMangoManager == nil {
		ServerError(w, fmt.Errorf("LLMangoManager not initialized"))
		return
	}

	// Check if prompt exists using Exists() on the manager's map
	if !r.LLMangoManager.Prompts.Exists(promptUID) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Prompt not found"))
		return
	}

	// Get the prompt before deleting to find its GoalUID
	promptToDelete := r.LLMangoManager.Prompts.Get(promptUID)
	goalUID := promptToDelete.GoalUID

	// Delete the prompt using Delete() on the manager's map
	r.LLMangoManager.Prompts.Delete(promptUID)

	// Remove the prompt UID from the corresponding goal's PromptUIDs list in the manager's map
	if goalUID != "" && r.LLMangoManager.Goals.Exists(goalUID) {
		goal := r.LLMangoManager.Goals.Get(goalUID)
		newPromptUIDs := slices.DeleteFunc(goal.PromptUIDs, func(uid string) bool {
			return uid == promptUID
		})
		if len(newPromptUIDs) < len(goal.PromptUIDs) { // Check if deletion happened
			goal.PromptUIDs = newPromptUIDs
			r.LLMangoManager.Goals.Set(goalUID, goal) // Update the goal in the SyncedMap
		}
	}

	// Save state if SaveState function is set
	if r.LLMangoManager.SaveState != nil {
		if err := r.LLMangoManager.SaveState(); err != nil {
			log.Printf("WARN: SaveState failed after deleting prompt %s: %v", promptUID, err)
			// ServerError(w, err) // Decide if this should be a fatal error for the request
			// We choose to continue and report success on delete, as the core operation succeeded.
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
		BadRequest(w, "Invalid request body: "+err.Error())
		return
	}

	if prompt.GoalUID == "" {
		BadRequest(w, "GoalUID is required for a new prompt")
		return
	}

	if r.LLMangoManager == nil {
		ServerError(w, fmt.Errorf("LLMangoManager not initialized"))
		return
	}

	// Ensure the target goal exists
	if !r.LLMangoManager.Goals.Exists(prompt.GoalUID) {
		BadRequest(w, fmt.Sprintf("Goal with UID %s not found", prompt.GoalUID))
		return
	}

	if prompt.UID == "" {
		prompt.UID = generateUID() // Assuming generateUID() exists
	}

	// Sanitize UID
	prompt.UID = strings.TrimSpace(prompt.UID)
	prompt.UID = strings.ReplaceAll(prompt.UID, " ", "_")
	prompt.UID = url.PathEscape(prompt.UID)

	// Check if UID already exists
	if r.LLMangoManager.Prompts.Exists(prompt.UID) {
		BadRequest(w, fmt.Sprintf("Prompt with UID %s already exists", prompt.UID))
		return
	}

	now := int(time.Now().Unix())
	if prompt.CreatedAt == 0 {
		prompt.CreatedAt = now
	}
	prompt.UpdatedAt = now // Always set UpdatedAt on create/update

	// Add prompt to the manager's map using Set()
	r.LLMangoManager.Prompts.Set(prompt.UID, prompt)

	// Add the prompt UID to the corresponding goal's PromptUIDs list
	goal := r.LLMangoManager.Goals.Get(prompt.GoalUID)
	if !slices.Contains(goal.PromptUIDs, prompt.UID) {
		goal.PromptUIDs = append(goal.PromptUIDs, prompt.UID)
		r.LLMangoManager.Goals.Set(goal.UID, goal) // Update the goal
	}

	// Save state after creating the prompt and updating the goal
	if r.LLMangoManager.SaveState != nil {
		if err := r.LLMangoManager.SaveState(); err != nil {
			log.Printf("WARN: SaveState failed after creating prompt %s: %v", prompt.UID, err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
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

	if r.LLMangoManager == nil {
		ServerError(w, fmt.Errorf("LLMangoManager not initialized"))
		return
	}

	// Check existence and get the prompt from the manager's map
	if !r.LLMangoManager.Prompts.Exists(promptUID) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Prompt not found"))
		return
	}
	prompt := r.LLMangoManager.Prompts.Get(promptUID)

	// Parse request body - use pointers to check for field presence
	var updateReq struct {
		Model      *string            `json:"model,omitempty"`
		Parameters *map[string]any    `json:"parameters,omitempty"`
		Messages   *[]json.RawMessage `json:"messages,omitempty"`
		Weight     *int               `json:"weight,omitempty"`
		IsCanary   *bool              `json:"isCanary,omitempty"`
		MaxRuns    *int               `json:"maxRuns,omitempty"`
	}

	if err := json.NewDecoder(req.Body).Decode(&updateReq); err != nil {
		BadRequest(w, "Invalid request body: "+err.Error())
		return
	}

	updated := false

	// Update fields only if they are provided in the request
	if updateReq.Model != nil {
		prompt.Model = *updateReq.Model
		updated = true
	}
	if updateReq.Weight != nil {
		prompt.Weight = *updateReq.Weight
		updated = true
	}
	if updateReq.IsCanary != nil {
		prompt.IsCanary = *updateReq.IsCanary
		updated = true
	}
	if updateReq.MaxRuns != nil {
		prompt.MaxRuns = *updateReq.MaxRuns
		updated = true
	}

	// Handle parameters update
	if updateReq.Parameters != nil {
		// Reset parameters entirely if the field is present
		prompt.Parameters = openrouter.Parameters{}
		params := *updateReq.Parameters
		if temp, ok := params["temperature"].(float64); ok {
			prompt.Parameters.Temperature = &temp
		}
		if maxTok, ok := params["max_tokens"].(float64); ok {
			maxTokInt := int(maxTok)
			prompt.Parameters.MaxTokens = &maxTokInt
		}
		if topP, ok := params["top_p"].(float64); ok {
			prompt.Parameters.TopP = &topP
		}
		if freqPen, ok := params["frequency_penalty"].(float64); ok {
			prompt.Parameters.FrequencyPenalty = &freqPen
		}
		if presPen, ok := params["presence_penalty"].(float64); ok {
			prompt.Parameters.PresencePenalty = &presPen
		}
		updated = true
	}

	// Handle messages update
	if updateReq.Messages != nil {
		messagesData := *updateReq.Messages
		// Replace messages entirely if the field is present, even if empty
		newMessages := make([]openrouter.Message, len(messagesData))
		for i, msgData := range messagesData {
			if err := json.Unmarshal(msgData, &newMessages[i]); err != nil {
				BadRequest(w, fmt.Sprintf("Invalid message format at index %d: %v", i, err))
				return
			}
		}
		prompt.Messages = newMessages
		updated = true
	}

	if updated {
		prompt.UpdatedAt = int(time.Now().Unix())
		// Save the updated prompt back to the manager's SyncedMap
		r.LLMangoManager.Prompts.Set(promptUID, prompt)

		// Save state after updating the prompt
		if r.LLMangoManager.SaveState != nil {
			if err := r.LLMangoManager.SaveState(); err != nil {
				log.Printf("WARN: SaveState failed after updating prompt %s: %v", promptUID, err)
			}
		}
	}

	// Return the potentially updated prompt as JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prompt)
}
