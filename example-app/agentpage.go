package main

import (
	"html/template"
	"net/http"
)

func handleAgentPage(w http.ResponseWriter, r *http.Request) {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>LLMango Agent System Test</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; }
        .header { background: #9b59b6; color: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; }
        .nav { background: white; padding: 15px; border-radius: 8px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .nav a { color: #3498db; text-decoration: none; margin-right: 20px; font-weight: bold; }
        .nav a:hover { text-decoration: underline; }
        .agent-section { background: white; margin: 20px 0; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); border-left: 4px solid #9b59b6; }
        .agent-section h3 { margin: 0 0 10px 0; color: #2c3e50; }
        .agent-section .description { color: #7f8c8d; margin-bottom: 15px; }
        .agent-info { background: #f8f9fa; padding: 15px; border-radius: 4px; margin-bottom: 15px; font-size: 14px; }
        .input-group { margin-bottom: 15px; }
        .input-group label { display: block; margin-bottom: 5px; font-weight: bold; }
        .input-group textarea { width: 100%; padding: 12px; border: 1px solid #ddd; border-radius: 4px; height: 120px; resize: vertical; font-family: monospace; font-size: 14px; }
        .input-group select { width: 100%; padding: 12px; border: 1px solid #ddd; border-radius: 4px; font-size: 14px; }
        .test-btn { background: #9b59b6; color: white; padding: 12px 24px; border: none; border-radius: 4px; cursor: pointer; font-size: 16px; font-weight: bold; }
        .test-btn:hover { background: #8e44ad; }
        .test-btn:disabled { background: #bdc3c7; cursor: not-allowed; }
        .response { margin-top: 15px; padding: 15px; border-radius: 4px; }
        .response.success { background: #d5f4e6; border: 1px solid #27ae60; }
        .response.error { background: #fadbd8; border: 1px solid #e74c3c; }
        .response-header { font-weight: bold; margin-bottom: 10px; }
        .response-content { background: white; padding: 15px; border-radius: 4px; font-family: monospace; white-space: pre-wrap; max-height: 400px; overflow-y: auto; }
        .loading { color: #f39c12; }
        .badge { display: inline-block; padding: 4px 8px; border-radius: 4px; font-size: 12px; font-weight: bold; background: #9b59b6; color: white; }
        .workflow-info { background: #e8f4fd; padding: 10px; border-radius: 4px; margin-bottom: 15px; border-left: 3px solid #3498db; }
        .examples { margin-top: 10px; }
        .example-btn { background: #ecf0f1; color: #2c3e50; padding: 6px 12px; border: 1px solid #bdc3c7; border-radius: 4px; cursor: pointer; margin-right: 10px; margin-bottom: 5px; font-size: 12px; }
        .example-btn:hover { background: #d5dbdb; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🤖 LLMango Agent System</h1>
            <p>Test the integrated agent system with workflows and AI agents</p>
            <p><strong>Powered by:</strong> llmangoagents + OpenRouter + LLMango</p>
        </div>

        <div class="nav">
            <a href="/">← Back to CLI Tests</a>
            <a href="/agent">🤖 Agent System</a>
        </div>

        <div class="agent-section">
            <h3>🚀 Agent Workflow Tester <span class="badge">AGENT-POWERED</span></h3>
            <div class="description">Send messages to AI agents through configured workflows. The agent system handles routing, execution, and response generation.</div>
            
            <div class="workflow-info">
                <strong>Active Workflow:</strong> text_classifier_workflow<br>
                <strong>Agent:</strong> example_agent (Claude 3 Sonnet)<br>
                <strong>Capabilities:</strong> Text analysis, conversation, reasoning, and general assistance
            </div>
            
            <div class="input-group">
                <label for="agent-input">Your Message:</label>
                <textarea id="agent-input" placeholder="Type your message to the agent...">Hello! Can you help me analyze the sentiment of this text: 'I absolutely love this new AI system! It makes everything so much easier and more efficient.'</textarea>
            </div>

            <div class="input-group">
                <label for="workflow-select">Workflow:</label>
                <select id="workflow-select">
                    <option value="text_classifier_workflow">text_classifier_workflow (Default Agent Flow)</option>
                </select>
            </div>

            <div class="examples">
                <strong>Example prompts:</strong><br>
                <button class="example-btn" onclick="setExample('sentiment')">Sentiment Analysis</button>
                <button class="example-btn" onclick="setExample('summary')">Text Summary</button>
                <button class="example-btn" onclick="setExample('creative')">Creative Writing</button>
                <button class="example-btn" onclick="setExample('analysis')">Data Analysis</button>
                <button class="example-btn" onclick="setExample('conversation')">General Chat</button>
            </div>
            
            <button class="test-btn" onclick="testAgent()">
                🤖 Send to Agent
            </button>
            
            <div id="agent-response" class="response" style="display: none;"></div>
        </div>

        <div class="agent-info">
            <h4>🔧 How it works:</h4>
            <ol>
                <li><strong>Input Processing:</strong> Your message is sent to the agent system</li>
                <li><strong>Workflow Execution:</strong> The system routes through the configured workflow</li>
                <li><strong>Agent Processing:</strong> The AI agent processes your request using the specified model</li>
                <li><strong>Response Generation:</strong> The agent generates and returns a response</li>
            </ol>
            <p><strong>API Endpoint:</strong> POST /agents</p>
            <p><strong>Request Format:</strong> {"input": "your message", "workflowUID": "text_classifier_workflow"}</p>
        </div>
    </div>

    <script>
        const examples = {
            sentiment: "Can you analyze the sentiment of this text: 'I'm really frustrated with this software. It keeps crashing and losing my work!'",
            summary: "Please summarize this article: 'Artificial intelligence is rapidly transforming industries worldwide. From healthcare to finance, AI systems are automating complex tasks, improving efficiency, and enabling new capabilities. Machine learning algorithms can process vast amounts of data to identify patterns and make predictions. However, the rapid advancement also raises questions about ethics, job displacement, and regulation.'",
            creative: "Write a short story about a robot who discovers emotions for the first time.",
            analysis: "Help me analyze the pros and cons of remote work vs office work in the post-pandemic era.",
            conversation: "Hi there! I'm having a tough day and could use some encouragement. Can you help cheer me up?"
        };

        function setExample(type) {
            document.getElementById('agent-input').value = examples[type];
        }

        async function testAgent() {
            const inputElement = document.getElementById('agent-input');
            const workflowElement = document.getElementById('workflow-select');
            const responseElement = document.getElementById('agent-response');
            const button = document.querySelector('.test-btn');
            
            const inputText = inputElement.value.trim();
            const workflowUID = workflowElement.value;
            
            if (!inputText) {
                alert('Please enter a message for the agent');
                return;
            }
            
            // Show loading state
            button.disabled = true;
            button.textContent = '🔄 Processing...';
            responseElement.style.display = 'block';
            responseElement.className = 'response loading';
            responseElement.innerHTML = '<div class="response-header">🤖 Agent is thinking...</div><div class="response-content">Sending your message through the workflow system...</div>';
            
            try {
                const response = await fetch('/agents', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        input: inputText,
                        workflowUID: workflowUID
                    })
                });
                
                const result = await response.json();
                
                if (result.status === 'completed' || result.status === 'success') {
                    responseElement.className = 'response success';
                    responseElement.innerHTML =
                        '<div class="response-header">✅ Agent Response (Status: ' + result.status + ')</div>' +
                        '<div class="response-content">' + (result.output || 'No output received') + '</div>';
                } else {
                    responseElement.className = 'response error';
                    responseElement.innerHTML =
                        '<div class="response-header">❌ Agent Error (Status: ' + result.status + ')</div>' +
                        '<div class="response-content">' + (result.error || result.output || 'Unknown error occurred') + '</div>';
                }
            } catch (error) {
                responseElement.className = 'response error';
                responseElement.innerHTML =
                    '<div class="response-header">❌ Network Error</div>' +
                    '<div class="response-content">Failed to communicate with agent system: ' + error.message + '</div>';
            }
            
            // Reset button
            button.disabled = false;
            button.textContent = '🤖 Send to Agent';
        }
        
        // Allow Enter key to submit (with Shift+Enter for new line)
        document.getElementById('agent-input').addEventListener('keydown', function(e) {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                testAgent();
            }
        });
    </script>
</body>
</html>
`

	t, err := template.New("agent").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, nil)
}
