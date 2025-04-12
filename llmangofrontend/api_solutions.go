package llmangofrontend

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/llmang/llmango/llmango"
)

// handleCreateSolution creates a new solution for a goal
func (r *APIRouter) handleCreateSolution(w http.ResponseWriter, req *http.Request) {
	// Parse request body
	var createReq struct {
		GoalID    string `json:"goalId"`
		PromptUID string `json:"promptUid"`
		Weight    int    `json:"weight"`
		IsCanary  bool   `json:"isCanary"`
		MaxRuns   int    `json:"maxRuns"`
	}

	if err := json.NewDecoder(req.Body).Decode(&createReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Invalid request body")
		return
	}

	// Validate request
	goalAny, exists := r.Goals[createReq.GoalID]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode("Goal not found")
		return
	}

	if createReq.PromptUID != "" {
		if _, exists := r.Prompts[createReq.PromptUID]; !exists {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode("Prompt not found")
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
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("Could not find Solutions field in goal object")
		return
	}

	// Set the solution in the map
	solutionsMap.SetMapIndex(reflect.ValueOf(solutionID), reflect.ValueOf(solution))

	// Save state after creating the solution
	if r.SaveState != nil {
		if err := r.SaveState(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode("Failed to save state: " + err.Error())
			return
		}
	}

	json.NewEncoder(w).Encode(map[string]string{
		"solutionId": solutionID,
	})
}

// handleUpdateSolution updates an existing solution
func (r *APIRouter) handleUpdateSolution(w http.ResponseWriter, req *http.Request) {
	// Extract solution ID from URL pattern
	solutionID := req.PathValue("solutionuid")
	if solutionID == "" {
		BadRequest(w, "Missing solution ID")
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
		BadRequest(w, "Invalid request body")
		return
	}

	// Validate goal exists
	goalAny, exists := r.Goals[updateReq.GoalID]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Goal not found"))
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
		ServerError(w, fmt.Errorf("could not find Solutions field in goal object"))
		return
	}

	// Get the solution from the map
	solutionValue := solutionsMap.MapIndex(reflect.ValueOf(solutionID))
	if !solutionValue.IsValid() {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Solution not found"))
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
			ServerError(w, err)
			return
		}
	}

	w.Write([]byte("Solution updated successfully"))
}

// handleDeleteSolution deletes a solution
func (r *APIRouter) handleDeleteSolution(w http.ResponseWriter, req *http.Request) {
	// Extract solution ID from URL pattern
	solutionID := req.PathValue("solutionuid")
	if solutionID == "" {
		BadRequest(w, "Missing solution ID")
		return
	}

	// Parse request body to get the goal ID
	var deleteReq struct {
		GoalID string `json:"goalId"`
	}

	if err := json.NewDecoder(req.Body).Decode(&deleteReq); err != nil {
		BadRequest(w, "Invalid request body")
		return
	}

	// Validate goal exists
	goalAny, exists := r.Goals[deleteReq.GoalID]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Goal not found"))
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
		ServerError(w, fmt.Errorf("could not find Solutions field in goal object"))
		return
	}

	// Check if solution exists
	if !solutionsMap.MapIndex(reflect.ValueOf(solutionID)).IsValid() {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Solution not found"))
		return
	}

	// Delete the solution by setting its map entry to zero value
	solutionsMap.SetMapIndex(reflect.ValueOf(solutionID), reflect.Value{})

	// Save state after deleting the solution
	if r.SaveState != nil {
		if err := r.SaveState(); err != nil {
			ServerError(w, err)
			return
		}
	}

	w.Write([]byte("Solution deleted successfully"))
}
