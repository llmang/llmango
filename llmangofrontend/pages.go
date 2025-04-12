package llmangofrontend

import (
	"fmt"
	"net/http"
)

func (r *Router) handleTestsPage(w http.ResponseWriter, req *http.Request) {
	data := GoalsPageData{
		Goals: r.Goals,
	}
	r.renderTemplate(w, "tests", data)
}

// handleHomePage serves the home page
func (r *Router) handleHomePage(w http.ResponseWriter, req *http.Request) {
	data := HomePageData{
		Prompts: r.Prompts,
		Goals:   r.Goals,
	}
	r.renderTemplate(w, "home", data)
}

// handlePromptsPage serves the prompts list page
func (r *Router) handlePromptsPage(w http.ResponseWriter, req *http.Request) {
	data := PromptsPageData{
		Prompts: r.Prompts,
	}
	r.renderTemplate(w, "prompts", data)
}

// handleGoalsPage serves the goals list page
func (r *Router) handleGoalsPage(w http.ResponseWriter, req *http.Request) {
	data := GoalsPageData{
		Goals: r.Goals,
	}
	r.renderTemplate(w, "goals", data)
}

// handleModelsPage serves the models list page
func (r *Router) handleModelsPage(w http.ResponseWriter, req *http.Request) {
	// No data needed as models are fetched client-side
	r.renderTemplate(w, "models", nil)
}

// handleLogsPage serves the logs view page
func (r *Router) handleLogsPage(w http.ResponseWriter, req *http.Request) {
	// Initialize default filter options
	filterOptions := map[string]interface{}{
		"goalId":    nil,
		"promptUID": nil,
		"perPage":   10,
	}

	data := LogsPageData{
		Prompts:       r.Prompts,
		Goals:         r.Goals,
		FilterOptions: filterOptions,
	}
	r.renderTemplate(w, "logs", data)
}

// handlePromptDetailPage serves a specific prompt's detail page
func (r *Router) handlePromptDetailPage(w http.ResponseWriter, req *http.Request) {
	// Extract prompt ID from URL pattern
	promptID := req.PathValue("promptID")
	if promptID == "" {
		http.NotFound(w, req)
		return
	}

	prompt, ok := r.Prompts[promptID]
	if !ok {
		http.NotFound(w, req)
		return
	}

	data := map[string]interface{}{
		"Prompt":    prompt,
		"PromptUID": promptID,
	}
	r.renderTemplate(w, "prompt-detail", data)
}

// handleGoalDetailPage serves a specific goal's detail page
func (r *Router) handleGoalDetailPage(w http.ResponseWriter, req *http.Request) {
	// Extract goal ID from URL pattern
	goalID := req.PathValue("goalID")
	if goalID == "" {
		http.NotFound(w, req)
		return
	}

	goal, ok := r.Goals[goalID]
	if !ok {
		http.NotFound(w, req)
		return
	}

	data := map[string]interface{}{
		"Goal":    goal,
		"GoalID":  goalID,
		"Prompts": r.Prompts,
	}
	r.renderTemplate(w, "goal-detail", data)
}

// handleNewPromptForGoalPage serves the page to create a new prompt for a goal
func (r *Router) handleNewPromptForGoalPage(w http.ResponseWriter, req *http.Request) {
	// Extract goal ID from URL pattern
	goalID := req.PathValue("goalID")
	if goalID == "" {
		http.NotFound(w, req)
		return
	}

	_, ok := r.Goals[goalID]
	if !ok {
		http.NotFound(w, req)
		return
	}

	// For now, just redirect to goal detail page
	http.Redirect(w, req, fmt.Sprintf("%s/goal/%s", r.BaseRoute, goalID), http.StatusSeeOther)
}
