package llmangofrontend

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/llmang/llmango/llmango"
)

// handleUpdateGoal updates a goal's title and description using the helper function
func (r *APIRouter) handleUpdateGoal(w http.ResponseWriter, req *http.Request) {
	goalUID := req.PathValue("goaluid")
	if goalUID == "" {
		BadRequest(w, "Missing goal ID")
		return
	}

	goalAny, exists := r.Goals[goalUID]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Goal not found"))
		return
	}

	var updateReq struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(req.Body).Decode(&updateReq); err != nil {
		BadRequest(w, "Invalid request body")
		return
	}

	// Call the helper function from the llmango package
	// goalAny must be a pointer for the changes to persist in the map
	err := llmango.UpdateGoalTitleDescription(goalAny, updateReq.Title, updateReq.Description)
	if err != nil {
		// Log the specific reflection error
		log.Printf("Error updating goal '%s' via reflection: %v", goalUID, err)
		// Provide a slightly more generic error to the client
		ServerError(w, fmt.Errorf("failed to update goal fields: %w", err))
		return
	}

	// Save state after updating the goal
	if r.SaveState != nil {
		if err := r.SaveState(); err != nil {
			log.Printf("SaveState failed with error: %v\n", err)
			ServerError(w, err)
			return
		}
	}

	// Return the updated goal as JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(goalAny) // Encode the original (now modified) object
}

// handleGetGoals handles getting all goals with pagination
func (r *APIRouter) handleGetGoals(w http.ResponseWriter, req *http.Request) {
	// Get limit from header
	limit := 0 // default limit
	if limitStr := req.Header.Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Get all goals
	goals := make(map[string]interface{})
	for uid, goal := range r.Goals {
		goals[uid] = goal
		if limit > 0 && len(goals) >= limit {
			break
		}
	}

	json.NewEncoder(w).Encode(goals)
}

// handleGetGoal handles getting a single goal by UID
func (r *APIRouter) handleGetGoal(w http.ResponseWriter, req *http.Request) {
	goalUID := req.PathValue("goaluid")
	if goalUID == "" {
		BadRequest(w, "Missing goal UID")
		return
	}

	goal, exists := r.Goals[goalUID]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Goal not found"))
		return
	}

	json.NewEncoder(w).Encode(goal)
}
