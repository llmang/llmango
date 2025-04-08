package llmangofrontend

import (
	"encoding/json"
	"net/http"

	"github.com/llmang/llmango/llmango"
)

type Router struct {
	*llmango.LLMangoManager
	BaseRoute string
}

func CreateLLMMangRouter(l *llmango.LLMangoManager, baseRoute *string) http.Handler {
	router := Router{
		LLMangoManager: l,
	}

	// If baseRoute is nil (no value provided), use default "/mango"
	// If baseRoute is &"" (empty string provided), use no base route
	// Otherwise use the provided value
	if baseRoute == nil {
		router.BaseRoute = "/mango"
	} else {
		router.BaseRoute = *baseRoute
	}

	mux := http.NewServeMux()

	// Register page handlers with specific functions for each route
	mux.HandleFunc("/", router.handleHomePage)
	mux.HandleFunc("GET /home", router.handleHomePage)
	mux.HandleFunc("GET /prompts", router.handlePromptsPage)
	mux.HandleFunc("GET /goals", router.handleGoalsPage)
	mux.HandleFunc("GET /models", router.handleModelsPage)
	mux.HandleFunc("GET /logs", router.handleLogsPage)
	mux.HandleFunc("GET /prompt/{promptID}", router.handlePromptDetailPage)
	mux.HandleFunc("GET /goal/{goalID}/newprompt", router.handleNewPromptForGoalPage)
	mux.HandleFunc("GET /goal/{goalID}", router.handleGoalDetailPage)

	// Register API routes
	mux.HandleFunc("DELETE /api/prompts/{promptID}", router.handleDeletePrompt)

	router.RegisterAPIRoutes(mux)

	// Apply the middlewares to the mux
	return router.apiKeyMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If path is empty, set it to "/"
			if r.URL.Path == "" {
				r.URL.Path = "/"
			}
			mux.ServeHTTP(w, r)
		}),
	)
}

// handleDeletePrompt handles the deletion of a prompt
func (r *Router) handleDeletePrompt(w http.ResponseWriter, req *http.Request) {
	promptID := req.PathValue("promptID")
	if promptID == "" {
		http.Error(w, "Prompt ID is required", http.StatusBadRequest)
		return
	}

	// Delete the prompt from the manager
	err := r.LLMangoManager.DeletePrompt(promptID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}
