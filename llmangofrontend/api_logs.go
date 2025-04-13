package llmangofrontend

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/llmang/llmango/llmango"
)

// LogFilter represents the filter parameters for log queries
type LogFilter struct {
	GoalUID   string `json:"goalUID,omitempty"`
	PromptUID string `json:"promptUID,omitempty"`
	Level     string `json:"level,omitempty"`
	Page      int    `json:"page"`
	PerPage   int    `json:"perPage"`
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
	// Check if logging is enabled
	if r.Logging == nil || r.Logging.GetLogs == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode("Logging is not enabled in this LLMango implementation")
		return
	}

	// Parse request body
	var filterReq struct {
		GoalUID    *string `json:"goalUID"`
		PromptUID  *string `json:"promptUID"`
		IncludeRaw bool    `json:"includeRaw"`
		Limit      int     `json:"limit"`
		Offset     int     `json:"offset"`
	}

	if err := json.NewDecoder(req.Body).Decode(&filterReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Invalid request body: " + err.Error())
		return
	}

	// Set defaults if not provided
	if filterReq.Limit <= 0 {
		filterReq.Limit = 10
	}

	// Convert our filter to llmango's filter type
	filter := &llmango.LLmangoLogFilter{
		GoalUID:    filterReq.GoalUID,
		PromptUID:  filterReq.PromptUID,
		IncludeRaw: filterReq.IncludeRaw,
		Limit:      filterReq.Limit,
		Offset:     filterReq.Offset,
	}

	// Get logs
	logs, total, err := r.Logging.GetLogs(filter)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("Failed to get logs: " + err.Error())
		return
	}

	// Calculate pagination
	page := (filterReq.Offset / filterReq.Limit) + 1
	totalPages := (total + filterReq.Limit - 1) / filterReq.Limit
	if totalPages == 0 {
		totalPages = 1
	}

	response := LogResponse{
		Logs: logs,
		Pagination: PaginationResponse{
			Total:      total,
			Page:       page,
			PerPage:    filterReq.Limit,
			TotalPages: totalPages,
		},
	}

	json.NewEncoder(w).Encode(response)
}

// handleGetGoalLogs handles log queries for a specific goal
func (r *APIRouter) handleGetGoalLogs(w http.ResponseWriter, req *http.Request) {

	goalUID := req.PathValue("goaluid")
	if goalUID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Missing goal ID")
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
		GoalUID: &goalUID,
		Limit:   perPage,
		Offset:  (page - 1) * perPage,
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
