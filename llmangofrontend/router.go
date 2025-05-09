package llmangofrontend

import (
	"embed"
	"net/http"

	"github.com/carsongh/go-sveltespa"
	"github.com/llmang/llmango/llmango"
)

//go:embed all:svelte/build/**
var embeddedSvelteBuild embed.FS

type Router struct {
	*llmango.LLMangoManager
	BaseRoute string
}

func CreateLLMMangRouter(l *llmango.LLMangoManager, baseRoute *string) http.Handler {
	router := Router{
		LLMangoManager: l,
	}

	mux := http.NewServeMux()

	// Register page handlers with specific functions for each route

	// Register API routes
	apiRouter := router.CreateAPIRouter()
	mux.Handle("/api/", http.StripPrefix("/api", apiRouter))

	svelteRouter := sveltespa.EmbeddedRouter(embeddedSvelteBuild, "svelte/build", "index.html")
	mux.HandleFunc("/", svelteRouter)

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
