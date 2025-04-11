package llmangofrontend

import (
	"net/http"

	"github.com/llmang/llmango/llmango"
	"github.com/llmang/llmango/llmangofrontend/templates_templ"
)

// TemplRouter implements handlers using templ components
type TemplRouter struct {
	*Router        // Embed the main Router to access its fields
	TemplBaseRoute string
}

// RegisterTemplRoutes registers routes that use templ components
func (r *Router) RegisterTemplRoutes(mux *http.ServeMux) {
	tr := &TemplRouter{Router: r, TemplBaseRoute: "/mango/templ"}

	// Register routes with a /templ prefix to avoid conflicts with existing routes
	mux.HandleFunc("/templ", tr.handleTemplHomePage)
	mux.HandleFunc("/templ/home", tr.handleTemplHomePage)
	mux.HandleFunc("/templ/prompt", tr.handleTemplPromptsPage)
	mux.HandleFunc("/templ/prompt/{promptuid}", tr.handleTemplPromptPage)
	mux.HandleFunc("/templ/goal", tr.handleTemplGoalsPage)
	mux.HandleFunc("/templ/goal/{goaluid}", tr.handleTemplGoalDetailPage)
	mux.HandleFunc("/templ/models", tr.handleModelsPage)
	mux.HandleFunc("/templ/logs", tr.handleLogsPage)
}

// handleTemplHomePage serves the home page using templ components
func (tr *TemplRouter) handleTemplHomePage(w http.ResponseWriter, req *http.Request) {
	// Limit prompts and goals to max 3 items
	limitedPrompts := make(map[string]*llmango.Prompt)
	limitedGoals := make(map[string]any)

	count := 0
	for id, prompt := range tr.Prompts {
		if count < 2 {
			limitedPrompts[id] = prompt
			count++
		} else {
			break
		}
	}

	count = 0
	for id, goal := range tr.Goals {
		if count < 2 {
			limitedGoals[id] = goal
			count++
		} else {
			break
		}
	}

	component := templates_templ.HomePage(limitedPrompts, limitedGoals, tr.TemplBaseRoute)
	templates_templ.RenderTempl(w, req, component)
}

// handleTemplPromptsPage serves the prompts page using templ components
func (tr *TemplRouter) handleTemplPromptsPage(w http.ResponseWriter, req *http.Request) {
	templates_templ.RenderTempl(w, req, templates_templ.PromptsPage(tr.Prompts, tr.TemplBaseRoute))
}

func (tr *TemplRouter) handleTemplPromptPage(w http.ResponseWriter, req *http.Request) {
	promptUID := req.PathValue("promptuid")
	if promptUID == "" {
		http.Redirect(w, req, "/templ/prompt", http.StatusFound)
		return
	}

	prompt, ok := tr.Prompts[promptUID]
	if !ok {
		http.Redirect(w, req, "/templ/prompt", http.StatusFound)
		return
	}

	templates_templ.RenderTempl(w, req, templates_templ.PromptDetailPage(promptUID, prompt, tr.TemplBaseRoute))

}

func (tr *TemplRouter) handleModelsPage(w http.ResponseWriter, req *http.Request) {
	templates_templ.RenderTempl(w, req, templates_templ.ModelsPage(tr.TemplBaseRoute))
}

func (tr *TemplRouter) handleLogsPage(w http.ResponseWriter, req *http.Request) {
	// Pass initial goals and prompts to the template
	templates_templ.RenderTempl(w, req, templates_templ.LogPage(
		tr.TemplBaseRoute,
		tr.Goals,
		tr.Prompts,
	))
}

// handleTemplGoalsPage serves the goals page using templ components
func (tr *TemplRouter) handleTemplGoalsPage(w http.ResponseWriter, req *http.Request) {
	component := templates_templ.GoalsPage(tr.Goals, tr.TemplBaseRoute)
	templates_templ.RenderTempl(w, req, component)
}

// handleTemplGoalDetailPage serves the goal detail page using templ components
func (tr *TemplRouter) handleTemplGoalDetailPage(w http.ResponseWriter, req *http.Request) {
	// Extract goal ID from URL path using pathvalue
	goalID := req.PathValue("goaluid")
	if goalID == "" {
		http.Redirect(w, req, "/templ/goals", http.StatusFound)
		return
	}

	// Look up the goal by ID
	goal, exists := tr.Goals[goalID]
	if !exists {
		http.Redirect(w, req, "/templ/goals", http.StatusFound)
		return
	}

	templates_templ.RenderTempl(w, req, templates_templ.GoalDetailPage(goalID, goal, tr.Prompts, tr.TemplBaseRoute))
}
