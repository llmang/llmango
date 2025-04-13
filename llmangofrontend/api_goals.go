package llmangofrontend

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/llmang/llmango/llmango"
)

// handleUpdateGoal updates a goal's title and description
func (r *APIRouter) handleUpdateGoal(w http.ResponseWriter, req *http.Request) {
	// Extract goal ID from URL pattern
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

	// Parse request body
	var updateReq struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(req.Body).Decode(&updateReq); err != nil {
		BadRequest(w, "Invalid request body")
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
			fmt.Printf("SaveState function exists, attempting to save state\n")
			if err := r.SaveState(); err != nil {
				fmt.Printf("SaveState failed with error: %v\n", err)
				ServerError(w, err)
				return
			}
			fmt.Printf("SaveState completed successfully\n")
		} else {
			fmt.Printf("SaveState function is nil, skipping state persistence\n")
		}

		// Return the updated goal as JSON response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(goalAny)
	} else {
		ServerError(w, fmt.Errorf("goal type assertion failed"))
	}
}

// handleGetGoals handles getting all goals with pagination
func (r *APIRouter) handleGetGoals(w http.ResponseWriter, req *http.Request) {
	// Get limit from header
	limit := 10 // default limit
	if limitStr := req.Header.Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Get all goals
	goals := make(map[string]interface{})
	for uid, goal := range r.Goals {
		goals[uid] = goal
		if len(goals) >= limit {
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
