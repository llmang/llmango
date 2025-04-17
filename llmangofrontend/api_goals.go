package llmangofrontend

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"sort"
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

	// Get all goals and convert to a slice for sorting
	// We need to use reflection to access GoalInfo fields from the map[string]any
	goalsSlice := make([]any, 0, len(r.Goals))
	for _, goal := range r.Goals {
		goalsSlice = append(goalsSlice, goal)
	}

	// Sort the goals slice using reflection
	sort.Slice(goalsSlice, func(i, j int) bool {
		goalI := goalsSlice[i]
		goalJ := goalsSlice[j]

		valI := reflect.ValueOf(goalI)
		valJ := reflect.ValueOf(goalJ)

		// Ensure they are pointers and get the element
		if valI.Kind() == reflect.Ptr {
			valI = valI.Elem()
		}
		if valJ.Kind() == reflect.Ptr {
			valJ = valJ.Elem()
		}

		// Access GoalInfo struct (assuming it's the first embedded field or named GoalInfo)
		// This assumes a consistent structure for Goal types
		goalInfoIField := valI.FieldByName("GoalInfo")
		goalInfoJField := valJ.FieldByName("GoalInfo")

		// Fallback if GoalInfo field isn't found directly (might be embedded anonymously)
		// This part is tricky and might need adjustment based on exact Goal struct definition
		if !goalInfoIField.IsValid() || !goalInfoJField.IsValid() {
			// Attempt to find embedded GoalInfo fields indirectly - this is less robust
			// Or handle error/log warning
			log.Printf("Warning: Could not find GoalInfo field directly for sorting goal. Check Goal struct definition.")
			// As a basic fallback, sort by UID if possible, else keep original order
			uidIField := valI.FieldByName("UID")
			uidJField := valJ.FieldByName("UID")
			if uidIField.IsValid() && uidJField.IsValid() && uidIField.Kind() == reflect.String && uidJField.Kind() == reflect.String {
				return uidIField.String() < uidJField.String() // Basic UID sort if timestamps fail
			}
			return false // Keep original order if fields inaccessible
		}

		updatedAtI := goalInfoIField.FieldByName("UpdatedAt").Int()
		updatedAtJ := goalInfoJField.FieldByName("UpdatedAt").Int()
		if updatedAtI != updatedAtJ {
			return updatedAtI > updatedAtJ // Descending UpdatedAt
		}

		createdAtI := goalInfoIField.FieldByName("CreatedAt").Int()
		createdAtJ := goalInfoJField.FieldByName("CreatedAt").Int()
		if createdAtI != createdAtJ {
			return createdAtI > createdAtJ // Descending CreatedAt
		}

		titleI := goalInfoIField.FieldByName("Title").String()
		titleJ := goalInfoJField.FieldByName("Title").String()
		return titleI < titleJ // Ascending Title for ties
	})

	// Apply limit if specified
	finalGoals := goalsSlice
	if limit > 0 && limit < len(goalsSlice) {
		finalGoals = goalsSlice[:limit]
	}

	json.NewEncoder(w).Encode(finalGoals)
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
