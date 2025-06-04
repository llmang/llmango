package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/llmang/llmango/llmangologger"
	"github.com/llmang/llmango/openrouter"

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
	fmt.Println(apiKey[:10])

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

	// Setup test endpoints using CLI-generated methods
	setupTestEndpoints()

	// Setup HTTP routes
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/api/test/", handleTestEndpoint)

	fmt.Println("🚀 LLMango CLI-Generated Test Server starting on http://localhost:8080")
	fmt.Println("📝 Make sure to set OPENROUTER_API_KEY in your .env file")
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

func handleHome(w http.ResponseWriter, r *http.Request) {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>LLMango CLI-Generated Test</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; }
        .header { background: #2c3e50; color: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; }
        .endpoint { background: white; margin: 10px 0; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); border-left: 4px solid #3498db; }
        .endpoint h3 { margin: 0 0 10px 0; color: #2c3e50; }
        .endpoint .description { color: #7f8c8d; margin-bottom: 15px; }
        .endpoint .method-info { background: #ecf0f1; padding: 10px; border-radius: 4px; margin-bottom: 15px; font-size: 14px; }
        .input-group { margin-bottom: 15px; }
        .input-group label { display: block; margin-bottom: 5px; font-weight: bold; }
        .input-group textarea { width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px; height: 120px; resize: vertical; font-family: monospace; }
        .test-btn { background: #3498db; color: white; padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; }
        .test-btn:hover { background: #2980b9; }
        .test-btn:disabled { background: #bdc3c7; cursor: not-allowed; }
        .response { margin-top: 15px; padding: 15px; border-radius: 4px; }
        .response.success { background: #d5f4e6; border: 1px solid #27ae60; }
        .response.error { background: #fadbd8; border: 1px solid #e74c3c; }
        .response-header { font-weight: bold; margin-bottom: 10px; }
        .response-content { background: white; padding: 10px; border-radius: 4px; font-family: monospace; white-space: pre-wrap; }
        .loading { color: #f39c12; }
        .badge { display: inline-block; padding: 4px 8px; border-radius: 4px; font-size: 12px; font-weight: bold; background: #3498db; color: white; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🚀 LLMango CLI-Generated Test</h1>
            <p>Testing CLI-generated goals and methods from mango.go</p>
            <p><strong>Generated from:</strong> llmango.yaml + internal/mango/example.go</p>
        </div>

        {{range .}}
        <div class="endpoint">
            <h3>{{.Name}} <span class="badge">CLI-GENERATED</span></h3>
            <div class="description">{{.Description}}</div>
            <div class="method-info">
                <strong>Method:</strong> mango.{{.Method}}() | 
                <strong>Path:</strong> {{.Path}} | 
                <strong>Source:</strong> CLI-generated from config + Go types
            </div>
            
            <div class="input-group">
                <label for="input-{{.Path}}">Input JSON:</label>
                <textarea id="input-{{.Path}}" placeholder="Enter JSON input...">{{.InputExample}}</textarea>
            </div>
            
            <button class="test-btn" onclick="testEndpoint('{{.Path}}', '{{.Method}}')">
                🧪 Test {{.Method}}()
            </button>
            
            <div id="response-{{.Path}}" class="response" style="display: none;"></div>
        </div>
        {{end}}
    </div>

    <script>
        async function testEndpoint(path, method) {
            const inputElement = document.getElementById('input-' + path);
            const responseElement = document.getElementById('response-' + path);
            const button = document.querySelector('button[onclick*="' + path + '"]');
            
            const inputText = inputElement.value.trim();
            if (!inputText) {
                alert('Please enter some JSON input');
                return;
            }
            
            // Validate JSON
            try {
                JSON.parse(inputText);
            } catch (e) {
                alert('Invalid JSON: ' + e.message);
                return;
            }
            
            // Show loading state
            button.disabled = true;
            button.textContent = '⏳ Processing...';
            responseElement.style.display = 'block';
            responseElement.className = 'response loading';
            responseElement.innerHTML = '<div class="response-header">🔄 Calling ' + method + '()...</div>';
            
            try {
                const response = await fetch('/api/test/' + path, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: inputText
                });
                
                const result = await response.json();
                
                if (result.success) {
                    responseElement.className = 'response success';
                    responseElement.innerHTML = 
                        '<div class="response-header">✅ Success - ' + result.method + '()</div>' +
                        '<div class="response-content">' + result.response + '</div>';
                } else {
                    responseElement.className = 'response error';
                    responseElement.innerHTML = 
                        '<div class="response-header">❌ Error - ' + result.method + '()</div>' +
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
            button.textContent = '🧪 Test ' + method + '()';
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
