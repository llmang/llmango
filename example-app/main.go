package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/llmang/llmango/llmango"
	"github.com/llmang/llmango/openrouter"
)

type TestEndpoint struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	GoalUID     string `json:"goalUID"`
	Model       string `json:"model"`
	Description string `json:"description"`
	IsStructured bool  `json:"isStructured"`
}

type TestResponse struct {
	Success  bool   `json:"success"`
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
	Model    string `json:"model"`
	Path     string `json:"path"`
}

var manager *llmango.LLMangoManager
var endpoints []TestEndpoint

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
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

	// Initialize LLMango Manager
	manager, err = llmango.CreateLLMangoManger(openRouter)
	if err != nil {
		log.Fatal("Failed to create LLMango Manager:", err)
	}

	// Setup test goals and prompts
	setupTestData()

	// Setup HTTP routes
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/api/test/", handleTestEndpoint)

	fmt.Println("🚀 LLMango Dual-Path Test Server starting on http://localhost:8080")
	fmt.Println("📝 Make sure to set OPENROUTER_API_KEY in your .env file")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func setupTestData() {
	// 1. Create JSON Goals (Frontend/Dynamic style)
	sentimentInputJSON := json.RawMessage(`{"text": "I love this new product!"}`)
	sentimentOutputJSON := json.RawMessage(`{"sentiment": "positive", "confidence": 0.95, "reasoning": "Contains positive language"}`)
	
	jsonGoal := llmango.NewJSONGoal(
		"sentiment-json",
		"Sentiment Analysis (JSON)",
		"Analyzes sentiment using JSON goal",
		sentimentInputJSON,
		sentimentOutputJSON,
	)

	// 2. Create Typed Goals (Developer style)
	type SummaryInput struct {
		Text string `json:"text"`
	}
	type SummaryOutput struct {
		Summary    string `json:"summary"`
		KeyPoints  []string `json:"key_points"`
		WordCount  int    `json:"word_count"`
	}

	typedGoal := llmango.NewGoal(
		"summary-typed",
		"Text Summary (Typed)",
		"Summarizes text using typed goal",
		SummaryInput{Text: "Long text to summarize..."},
		SummaryOutput{
			Summary: "Brief summary",
			KeyPoints: []string{"Point 1", "Point 2"},
			WordCount: 150,
		},
	)

	// Add goals to manager
	manager.AddGoals(jsonGoal, typedGoal)

	// 3. Create Prompts for Structured Models (OpenAI GPT-4)
	structuredSentimentPrompt := &llmango.Prompt{
		UID:    "sentiment-structured",
		GoalUID: "sentiment-json",
		Model:  "openai/gpt-4",
		Weight: 100,
		Messages: []openrouter.Message{
			{Role: "system", Content: "You are a sentiment analysis expert. Analyze the sentiment of the given text."},
			{Role: "user", Content: "Analyze the sentiment of this text: {{.text}}"},
		},
		Parameters: openrouter.Parameters{
			Temperature: &[]float64{0.3}[0],
		},
	}

	structuredSummaryPrompt := &llmango.Prompt{
		UID:    "summary-structured",
		GoalUID: "summary-typed",
		Model:  "openai/gpt-3.5-turbo",
		Weight: 100,
		Messages: []openrouter.Message{
			{Role: "system", Content: "You are a text summarization expert. Create concise summaries with key points."},
			{Role: "user", Content: "Summarize this text: {{.text}}"},
		},
		Parameters: openrouter.Parameters{
			Temperature: &[]float64{0.5}[0],
		},
	}

	// 4. Create Prompts for Universal Models (Anthropic Claude)
	universalSentimentPrompt := &llmango.Prompt{
		UID:    "sentiment-universal",
		GoalUID: "sentiment-json",
		Model:  "anthropic/claude-3-sonnet",
		Weight: 100,
		Messages: []openrouter.Message{
			{Role: "system", Content: "You are a sentiment analysis expert. Analyze the sentiment of the given text."},
			{Role: "user", Content: "Analyze the sentiment of this text: {{.text}}"},
		},
		Parameters: openrouter.Parameters{
			Temperature: &[]float64{0.3}[0],
		},
	}

	universalSummaryPrompt := &llmango.Prompt{
		UID:    "summary-universal",
		GoalUID: "summary-typed",
		Model:  "meta-llama/llama-3.1-405b-instruct",
		Weight: 100,
		Messages: []openrouter.Message{
			{Role: "system", Content: "You are a text summarization expert. Create concise summaries with key points."},
			{Role: "user", Content: "Summarize this text: {{.text}}"},
		},
		Parameters: openrouter.Parameters{
			Temperature: &[]float64{0.5}[0],
		},
	}

	// Add prompts to manager
	manager.AddPrompts(
		structuredSentimentPrompt,
		structuredSummaryPrompt,
		universalSentimentPrompt,
		universalSummaryPrompt,
	)

	// 5. Setup test endpoints
	endpoints = []TestEndpoint{
		{
			Name:        "Sentiment (Structured - GPT-4)",
			Path:        "sentiment-structured",
			GoalUID:     "sentiment-json",
			Model:       "openai/gpt-4",
			Description: "JSON Goal + Structured Output Model",
			IsStructured: true,
		},
		{
			Name:        "Sentiment (Universal - Claude)",
			Path:        "sentiment-universal",
			GoalUID:     "sentiment-json",
			Model:       "anthropic/claude-3-sonnet",
			Description: "JSON Goal + Universal Compatibility Model",
			IsStructured: false,
		},
		{
			Name:        "Summary (Structured - GPT-3.5)",
			Path:        "summary-structured",
			GoalUID:     "summary-typed",
			Model:       "openai/gpt-3.5-turbo",
			Description: "Typed Goal + Structured Output Model",
			IsStructured: true,
		},
		{
			Name:        "Summary (Universal - Llama)",
			Path:        "summary-universal",
			GoalUID:     "summary-typed",
			Model:       "meta-llama/llama-3.1-405b-instruct",
			Description: "Typed Goal + Universal Compatibility Model",
			IsStructured: false,
		},
	}

	// Update goal prompt UIDs to use specific prompts for each test
	if goal, exists := manager.Goals.Get("sentiment-json"); exists {
		goal.PromptUIDs = []string{"sentiment-structured", "sentiment-universal"}
	}
	if goal, exists := manager.Goals.Get("summary-typed"); exists {
		goal.PromptUIDs = []string{"summary-structured", "summary-universal"}
	}
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>LLMango Dual-Path Test</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; }
        .header { background: #2c3e50; color: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; }
        .endpoint { background: white; margin: 10px 0; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .endpoint h3 { margin: 0 0 10px 0; color: #2c3e50; }
        .endpoint .description { color: #7f8c8d; margin-bottom: 15px; }
        .endpoint .model-info { background: #ecf0f1; padding: 10px; border-radius: 4px; margin-bottom: 15px; font-size: 14px; }
        .structured { border-left: 4px solid #27ae60; }
        .universal { border-left: 4px solid #e74c3c; }
        .input-group { margin-bottom: 15px; }
        .input-group label { display: block; margin-bottom: 5px; font-weight: bold; }
        .input-group input, .input-group textarea { width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px; }
        .input-group textarea { height: 80px; resize: vertical; }
        .test-btn { background: #3498db; color: white; padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; }
        .test-btn:hover { background: #2980b9; }
        .test-btn:disabled { background: #bdc3c7; cursor: not-allowed; }
        .response { margin-top: 15px; padding: 15px; border-radius: 4px; }
        .response.success { background: #d5f4e6; border: 1px solid #27ae60; }
        .response.error { background: #fadbd8; border: 1px solid #e74c3c; }
        .response-header { font-weight: bold; margin-bottom: 10px; }
        .response-content { background: white; padding: 10px; border-radius: 4px; font-family: monospace; white-space: pre-wrap; }
        .loading { color: #f39c12; }
        .badge { display: inline-block; padding: 4px 8px; border-radius: 4px; font-size: 12px; font-weight: bold; }
        .badge.structured { background: #27ae60; color: white; }
        .badge.universal { background: #e74c3c; color: white; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1> 🚀 LLMango Dual-Path Execution Test</h1>
            <p>Test both structured output and universal compatibility execution paths</p>
        </div>

        {{range .}}
        <div class="endpoint {{if .IsStructured}}structured{{else}}universal{{end}}">
            <h3>{{.Name}} <span class="badge {{if .IsStructured}}structured{{else}}universal{{end}}">{{if .IsStructured}}STRUCTURED{{else}}UNIVERSAL{{end}}</span></h3>
            <div class="description">{{.Description}}</div>
            <div class="model-info">
                <strong>Model:</strong> {{.Model}} | 
                <strong>Goal:</strong> {{.GoalUID}} | 
                <strong>Path:</strong> {{if .IsStructured}}Structured Output{{else}}Universal Prompts{{end}}
            </div>
            
            <div class="input-group">
                <label for="input-{{.Path}}">Input Text:</label>
                <textarea id="input-{{.Path}}" placeholder="Enter text to process...">I absolutely love this new AI system! It's incredibly helpful and easy to use.</textarea>
            </div>
            
            <button class="test-btn" onclick="testEndpoint('{{.Path}}', '{{.GoalUID}}')">
                🧪 Test {{.Name}}
            </button>
            
            <div id="response-{{.Path}}" class="response" style="display: none;"></div>
        </div>
        {{end}}
    </div>

    <script>
        async function testEndpoint(path, goalUID) {
            const inputElement = document.getElementById('input-' + path);
            const responseElement = document.getElementById('response-' + path);
            const button = document.querySelector('button[onclick*="' + path + '"]');
            
            const inputText = inputElement.value.trim();
            if (!inputText) {
                alert('Please enter some text to process');
                return;
            }
            
            // Show loading state
            button.disabled = true;
            button.textContent = '⏳ Processing...';
            responseElement.style.display = 'block';
            responseElement.className = 'response loading';
            responseElement.innerHTML = '<div class="response-header">🔄 Processing request...</div>';
            
            try {
                const response = await fetch('/api/test/' + path, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ text: inputText })
                });
                
                const result = await response.json();
                
                if (result.success) {
                    responseElement.className = 'response success';
                    responseElement.innerHTML = 
                        '<div class="response-header">✅ Success - ' + result.model + '</div>' +
                        '<div class="response-content">' + result.response + '</div>';
                } else {
                    responseElement.className = 'response error';
                    responseElement.innerHTML = 
                        '<div class="response-header">❌ Error - ' + result.model + '</div>' +
                        '<div class="response-content">' + (result.error || 'Unknown error') + '</div>';
                }
            } catch (error) {
                responseElement.className = 'response error';
                responseElement.innerHTML = 
                    '<div class="response-header">❌ Network Error</div>' +
                    '<div class="response-content">' + error.message + '</div>';
            }
            
            // Reset button
            button.disabled = false;
            button.textContent = '🧪 Test ' + button.textContent.split('Test ')[1].split('...')[0];
        }
    </script>
</body>
</html>
`

	t, err := template.New("home").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, endpoints)
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

	// Parse request body
	var request struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Prepare input JSON
	inputJSON, err := json.Marshal(map[string]string{"text": request.Text})
	if err != nil {
		writeErrorResponse(w, "Failed to marshal input", endpoint.Model, endpoint.Path)
		return
	}

	// Execute using dual-path system
	log.Printf("Testing %s with model %s (structured: %v)", endpoint.Name, endpoint.Model, endpoint.IsStructured)
	
	// Temporarily override goal's prompt to use specific model for this test
	goal, exists := manager.Goals.Get(endpoint.GoalUID)
	if !exists {
		writeErrorResponse(w, "Goal not found", endpoint.Model, endpoint.Path)
		return
	}

	// Find the prompt for this specific model
	var promptUID string
	for _, pUID := range goal.PromptUIDs {
		if prompt, exists := manager.Prompts.Get(pUID); exists && prompt.Model == endpoint.Model {
			promptUID = pUID
			break
		}
	}
	
	if promptUID == "" {
		writeErrorResponse(w, "No prompt found for model", endpoint.Model, endpoint.Path)
		return
	}

	// Temporarily set goal to use only this prompt
	originalPromptUIDs := goal.PromptUIDs
	goal.PromptUIDs = []string{promptUID}
	defer func() { goal.PromptUIDs = originalPromptUIDs }()

	// Execute with dual-path system
	result, err := manager.ExecuteGoalWithDualPath(endpoint.GoalUID, inputJSON)
	if err != nil {
		writeErrorResponse(w, err.Error(), endpoint.Model, endpoint.Path)
		return
	}

	// Write success response
	response := TestResponse{
		Success:  true,
		Response: string(result),
		Model:    endpoint.Model,
		Path:     endpoint.Path,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func writeErrorResponse(w http.ResponseWriter, errorMsg, model, path string) {
	response := TestResponse{
		Success: false,
		Error:   errorMsg,
		Model:   model,
		Path:    path,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(response)
}