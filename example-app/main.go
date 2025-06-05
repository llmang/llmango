package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/llmang/llmango/llmangologger"
	"github.com/llmang/llmango/openrouter"

	"github.com/llmang/llmango/example-app/internal/agent"
	"github.com/llmang/llmango/example-app/internal/mango"
)

type TestEndpoint struct {
	Name         string `json:"name"`
	Path         string `json:"path"`
	Method       string `json:"method"`
	Description  string `json:"description"`
	InputExample string `json:"inputExample"`
}

type TestResponse struct {
	Success  bool   `json:"success"`
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
	Method   string `json:"method"`
}

var mangoClient *mango.Mango
var agentSystem *agent.AgentSystem
var endpoints []TestEndpoint

func main() {
	// Load environment variables
	if err := godotenv.Overload(".env"); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENROUTER_API_KEY environment variable is required")
	}
	// Initialize OpenRouter
	openRouter, err := openrouter.CreateOpenRouter(apiKey)
	if err != nil {
		log.Fatal("Failed to create OpenRouter:", err)
	}

	// Initialize CLI-generated Mango client with logging
	mangoClient, err = mango.CreateMango(openRouter)
	if err != nil {
		log.Fatal("Failed to create Mango client:", err)
	}

	// Enable logging with print logger (logs input/output objects only)
	// For full request/response logging, use: llmangologger.CreatePrintLogger(true)
	mangoClient.WithLogging(llmangologger.CreatePrintLogger(true))

	// Initialize agent system (separate from mango)
	agentSystem, err = agent.SetupAgentSystem(openRouter)
	if err != nil {
		log.Fatal("Failed to initialize agent system:", err)
	}
	agentSystem.SetDebug(true) // Enable debug logging for agents

	// Setup test endpoints using CLI-generated methods
	setupTestEndpoints()

	// Setup HTTP routes
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/agent", handleAgentPage)
	http.HandleFunc("/api/test/", handleTestEndpoint)
	http.HandleFunc("/agents", agentSystem.HandleAgentRequest)

	fmt.Println("🚀 LLMango CLI-Generated Test Server starting on http://localhost:8080")
	fmt.Println("📝 Make sure to set OPENROUTER_API_KEY in your .env file in the same dir as the example-app")
	fmt.Println("🔧 Testing CLI-generated goals and methods")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func setupTestEndpoints() {
	endpoints = []TestEndpoint{
		{
			Name:         "📝 Text Summary",
			Path:         "text-summary",
			Method:       "TextSummary",
			Description:  "CLI-generated text summarization using TextSummary() method",
			InputExample: `{"text": "Artificial intelligence (AI) is transforming industries worldwide. From healthcare to finance, AI systems are automating complex tasks, improving efficiency, and enabling new capabilities. Machine learning algorithms can now process vast amounts of data to identify patterns and make predictions. However, the rapid advancement of AI also raises important questions about ethics, job displacement, and the need for proper regulation. As we move forward, it's crucial to balance innovation with responsible development to ensure AI benefits society as a whole."}`,
		},
		{
			Name:         "💭 Sentiment Analysis",
			Path:         "sentiment-analysis",
			Method:       "SentimentAnalysis",
			Description:  "CLI-generated sentiment analysis using SentimentAnalysis() method",
			InputExample: `{"text": "I absolutely love this new AI system! It's incredibly helpful and makes my work so much easier."}`,
		},
		{
			Name:         "📧 Email Classification",
			Path:         "email-classification",
			Method:       "EmailClassification",
			Description:  "CLI-generated email classification using EmailClassification() method",
			InputExample: `{"subject": "Limited Time Offer - 50% Off Everything!", "body": "Don't miss out on our biggest sale of the year! Click here to shop now.", "sender": "sales@example.com"}`,
		},
		{
			Name:         "🌍 Language Detection",
			Path:         "language-detection",
			Method:       "LanguageDetection",
			Description:  "CLI-generated language detection using LanguageDetection() method",
			InputExample: `{"text": "Bonjour, comment allez-vous aujourd'hui? J'espère que vous passez une excellente journée!"}`,
		},
	}
}

func handleTestEndpoint(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract endpoint path
	path := r.URL.Path[len("/api/test/"):]

	// Find matching endpoint
	var endpoint *TestEndpoint
	for _, ep := range endpoints {
		if ep.Path == path {
			endpoint = &ep
			break
		}
	}

	if endpoint == nil {
		http.Error(w, "Endpoint not found", http.StatusNotFound)
		return
	}

	// Parse request body as raw JSON
	var inputJSON json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&inputJSON); err != nil {
		writeErrorResponse(w, "Invalid JSON: "+err.Error(), endpoint.Method)
		return
	}

	log.Printf("Testing CLI-generated method: %s", endpoint.Method)

	// Call the appropriate CLI-generated method
	var result interface{}
	var err error

	switch endpoint.Path {
	case "text-summary":
		var input mango.SummaryInput
		if err := json.Unmarshal(inputJSON, &input); err != nil {
			writeErrorResponse(w, "Invalid input for TextSummary: "+err.Error(), endpoint.Method)
			return
		}
		result, err = mangoClient.TextSummary(&input)

	case "sentiment-analysis":
		var input mango.SentimentInput
		if err := json.Unmarshal(inputJSON, &input); err != nil {
			writeErrorResponse(w, "Invalid input for SentimentAnalysis: "+err.Error(), endpoint.Method)
			return
		}
		result, err = mangoClient.SentimentAnalysis(&input)

	case "email-classification":
		var input mango.EmailInput
		if err := json.Unmarshal(inputJSON, &input); err != nil {
			writeErrorResponse(w, "Invalid input for EmailClassification: "+err.Error(), endpoint.Method)
			return
		}
		result, err = mangoClient.EmailClassification(&input)

	case "language-detection":
		var input mango.LanguageInput
		if err := json.Unmarshal(inputJSON, &input); err != nil {
			writeErrorResponse(w, "Invalid input for LanguageDetection: "+err.Error(), endpoint.Method)
			return
		}
		result, err = mangoClient.LanguageDetection(&input)

	default:
		writeErrorResponse(w, "Unknown endpoint", endpoint.Method)
		return
	}

	if err != nil {
		writeErrorResponse(w, err.Error(), endpoint.Method)
		return
	}

	// Marshal result to JSON
	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		writeErrorResponse(w, "Failed to marshal result: "+err.Error(), endpoint.Method)
		return
	}

	// Write success response
	response := TestResponse{
		Success:  true,
		Response: string(resultJSON),
		Method:   endpoint.Method,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func writeErrorResponse(w http.ResponseWriter, errorMsg, method string) {
	response := TestResponse{
		Success: false,
		Error:   errorMsg,
		Method:  method,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(response)
}
