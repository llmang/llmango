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

type SpendResponse struct {
	Spend float64 `json:"spend"`
	Count int     `json:"count"`
}

// handleGetLogs handles general log queries with filters
func (r *APIRouter) handleGetLogs(w http.ResponseWriter, req *http.Request) {
	// Check if logging is enabled
	if r.Logging == nil || r.Logging.GetLogs == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode("Logging is not enabled in this LLMango implementation")
		return
	}
	var filter llmango.LLmangoLogFilter

	if err := json.NewDecoder(req.Body).Decode(&filter); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Invalid request body format: " + err.Error())
		return
	}

	defer req.Body.Close()

	// --- Apply Defaults ---
	defaultLimit := 10
	if filter.Limit == nil || *filter.Limit <= 0 {
		filter.Limit = &defaultLimit
	}

	defaultOffset := 0
	if filter.Offset == nil || *filter.Offset < 0 {
		filter.Offset = &defaultOffset
	}
	// --- End Apply Defaults ---

	// Get logs using the populated filter struct (passed by address)
	logs, total, err := r.Logging.GetLogs(&filter)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("Failed to get logs: " + err.Error())
		return
	}

	// Calculate pagination using dereferenced pointers after ensuring they are not nil
	limit := *filter.Limit   // Use the value after default is applied
	offset := *filter.Offset // Use the value after default is applied
	page := (offset / limit) + 1
	totalPages := (total + limit - 1) / limit
	if totalPages == 0 {
		totalPages = 1
	}

	response := LogResponse{
		Logs: logs,
		Pagination: PaginationResponse{
			Total:      total,
			Page:       page,
			PerPage:    limit, // Use the effective limit
			TotalPages: totalPages,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// get the spend for a prompt or goal
func (r *APIRouter) handleGetSpend(w http.ResponseWriter, req *http.Request) {
	// Check if logging is enabled
	if r.Logging == nil || r.Logging.GetLogs == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode("Logging is not enabled in this LLMango implementation")
		return
	}

	// Parse request body
	var filter *llmango.LLmangoLogFilter

	if err := json.NewDecoder(req.Body).Decode(&filter); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Invalid request body: " + err.Error())
		return
	}

	limit := 0
	offset := 0
	filter.Limit = &limit
	filter.Offset = &offset

	// Get logs
	logs, total, err := r.Logging.GetLogs(filter)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("Failed to get logs: " + err.Error())
		return
	}
	var totalSpend float64
	for _, log := range logs {
		totalSpend += log.Cost
	}

	response := SpendResponse{
		Spend: totalSpend,
		Count: total,
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

	limit := perPage
	offset := (page - 1) * perPage
	filter := &llmango.LLmangoLogFilter{
		GoalUID: &goalUID,
		Limit:   &limit,
		Offset:  &offset,
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

	limit := perPage
	offset := (page - 1) * perPage
	filter := &llmango.LLmangoLogFilter{
		PromptUID: &promptUID,
		Limit:     &limit,
		Offset:    &offset,
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
