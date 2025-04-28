package llmangofrontend

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/llmang/llmango/llmango"
)

// handleUpdateGoal updates a goal's title and description
func (r *APIRouter) handleUpdateGoal(w http.ResponseWriter, req *http.Request) {
	goalUID := req.PathValue("goaluid")
	if goalUID == "" {
		BadRequest(w, "Missing goal ID")
		return
	}

	if r.LLMangoManager == nil {
		ServerError(w, fmt.Errorf("LLMangoManager not initialized"))
		return
	}

	// Check existence and get the goal from the manager's map
	goal, ok := r.LLMangoManager.Goals.Get(goalUID)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Goal not found"))
		return
	}

	var updateReq struct {
		Title       *string `json:"title,omitempty"` // Use pointers to check presence
		Description *string `json:"description,omitempty"`
	}

	if err := json.NewDecoder(req.Body).Decode(&updateReq); err != nil {
		BadRequest(w, "Invalid request body: "+err.Error())
		return
	}

	updated := false
	if updateReq.Title != nil && *updateReq.Title != goal.Title {
		goal.Title = *updateReq.Title
		updated = true
	}
	if updateReq.Description != nil && *updateReq.Description != goal.Description {
		goal.Description = *updateReq.Description
		updated = true
	}

	// If changes were made, update timestamp and save
	if updated {
		goal.UpdatedAt = int(time.Now().Unix())
		r.LLMangoManager.Goals.Set(goalUID, goal) // Save back to the SyncedMap

		// Save state after updating the goal
		if r.LLMangoManager.SaveState != nil {
			if err := r.LLMangoManager.SaveState(); err != nil {
				// Log the error but do not fail the request
				log.Printf("WARN: SaveState failed after updating goal %s: %v", goalUID, err)
				// ServerError(w, err) // Removed
				// Consider rollback? Revert goal changes?
				// return // Removed
			}
		}
	}

	// Return the potentially updated goal as JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(goal) // Encode the goal object
}

// handleGetGoals handles getting all goals with pagination
func (r *APIRouter) handleGetGoals(w http.ResponseWriter, req *http.Request) {
	// Get limit from header
	limit := -1 // default: no limit
	if limitStr := req.Header.Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if r.LLMangoManager == nil {
		ServerError(w, fmt.Errorf("LLMangoManager not initialized"))
		return
	}

	// Get all goals using Snapshot for safe iteration
	allGoalsMap := r.LLMangoManager.Goals.Snapshot()

	// Convert map to slice for sorting
	goalsSlice := make([]*llmango.Goal, 0, len(allGoalsMap))
	for _, goal := range allGoalsMap {
		goalsSlice = append(goalsSlice, goal)
	}

	// Sort the goals slice directly using Goal fields
	sort.Slice(goalsSlice, func(i, j int) bool {
		goalI := goalsSlice[i]
		goalJ := goalsSlice[j]

		if goalI.UpdatedAt != goalJ.UpdatedAt {
			return goalI.UpdatedAt > goalJ.UpdatedAt // Descending UpdatedAt
		}

		if goalI.CreatedAt != goalJ.CreatedAt {
			return goalI.CreatedAt > goalJ.CreatedAt // Descending CreatedAt
		}

		return goalI.Title < goalJ.Title // Ascending Title for ties
	})

	// Apply limit if specified
	finalGoals := goalsSlice
	if limit > 0 && limit < len(goalsSlice) {
		finalGoals = goalsSlice[:limit]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(finalGoals)
}

// handleGetGoal handles getting a single goal by UID
func (r *APIRouter) handleGetGoal(w http.ResponseWriter, req *http.Request) {
	goalUID := req.PathValue("goaluid")
	if goalUID == "" {
		BadRequest(w, "Missing goal UID")
		return
	}

	if r.LLMangoManager == nil {
		ServerError(w, fmt.Errorf("LLMangoManager not initialized"))
		return
	}

	// Use Exists() and Get() on the manager's map
	goal, ok := r.LLMangoManager.Goals.Get(goalUID)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Goal not found"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(goal)
}
