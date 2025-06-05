package main

import (
	"html/template"
	"net/http"
)

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

        <div style="background: white; padding: 15px; border-radius: 8px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
            <a href="/agent" style="color: #9b59b6; text-decoration: none; font-weight: bold; font-size: 16px;">
                🤖 Try the New Agent System →
            </a>
            <span style="color: #7f8c8d; margin-left: 10px;">Test AI agents with workflows and natural language</span>
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
