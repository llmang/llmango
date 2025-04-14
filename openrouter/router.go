package openrouter

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type OpenRouterRouter struct {
	OpenRouter *OpenRouter
}

// emptyPathMiddleware converts empty request paths to "/" to ensure proper routing
func emptyPathMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "" {
			r.URL.Path = "/"
		}
		next.ServeHTTP(w, r)
	})
}

func CreateOpenRouterRouter(openRouter *OpenRouter) http.Handler {
	// Create the OpenRouter router
	router := &OpenRouterRouter{
		OpenRouter: openRouter,
	}

	// Create a new ServeMux
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("/chat", router.openRouterChat)
	mux.HandleFunc("/generation-stats", router.getGenerationStats)

	// Add the chat UI route referring to the function defined in frontend.go
	mux.HandleFunc("/", ServeChatUI)

	// Wrap the mux with the empty path middleware
	return emptyPathMiddleware(mux)
}

func (or *OpenRouterRouter) openRouterChat(w http.ResponseWriter, r *http.Request) {
	// Read and parse request body
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "could not read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var request OpenRouterRequest
	err = json.Unmarshal(bytes, &request)
	if err != nil {
		http.Error(w, "request body was not in correct format", http.StatusBadRequest)
		return
	}

	// Check if streaming is requested
	if request.Stream != nil && *request.Stream {
		// Handle streaming response
		or.handleStreamingChat(w, r, &request)
		return
	}

	// Generate non-streaming response
	response, err := or.OpenRouter.GenerateNonStreamingChatResponse(&request)
	if err != nil {
		log.Printf("OPENROUTER:CHAT: failed to generate response with error message: %v", err)
		http.Error(w, "error generating response", http.StatusInternalServerError)
		return
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (or *OpenRouterRouter) handleStreamingChat(w http.ResponseWriter, r *http.Request, request *OpenRouterRequest) {
	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	// Create a channel for streaming responses
	streamChan, err := or.OpenRouter.GenerateStreamingChatResponse(r.Context(), request)
	if err != nil {
		// If we can't start streaming, write an error event
		fmt.Fprintf(w, "data: {\"error\": \"%s\"}\n\n", err.Error())
		return
	}

	// Set up a flusher if the ResponseWriter supports it
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Process streaming responses
	for chunk := range streamChan {
		// Convert to JSON
		jsonData, err := json.Marshal(chunk)
		if err != nil {
			fmt.Fprintf(w, "data: {\"error\": \"Error encoding chunk: %s\"}\n\n", err.Error())
			flusher.Flush()
			continue
		}

		// Write the event
		fmt.Fprintf(w, "data: %s\n\n", string(jsonData))
		flusher.Flush()
	}

	// Signal end of stream
	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
}

func (or *OpenRouterRouter) getGenerationStats(w http.ResponseWriter, r *http.Request) {
	// Get the generation ID from query parameters
	generationID := r.URL.Query().Get("id")
	if generationID == "" {
		http.Error(w, "generation ID is required", http.StatusBadRequest)
		return
	}

	// Retrieve generation stats
	stats, err := or.OpenRouter.GetGenerationStats(generationID)
	if err != nil {
		http.Error(w, fmt.Sprintf("error retrieving generation stats: %v", err), http.StatusInternalServerError)
		return
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(stats)
}
