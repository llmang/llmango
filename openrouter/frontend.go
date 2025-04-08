package openrouter

import (
	"html/template"
	"net/http"
)

// ChatTemplate holds the HTML template for the chat interface
const ChatTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>OpenRouter Chat</title>
    <script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.13.0/dist/cdn.min.js"></script>
    <!-- Add Showdown.js for Markdown support -->
    <script src="https://cdnjs.cloudflare.com/ajax/libs/showdown/2.1.0/showdown.min.js"></script>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif; }
        
        .app-container {
            display: flex;
            height: 100vh;
            width: 100vw;
        }
        
        .sidebar {
            width: 280px;
            background-color: #f5f5f5;
            border-right: 1px solid #ddd;
            display: flex;
            flex-direction: column;
            overflow: hidden;
        }
        
        .model-selector {
            padding: 15px;
            border-bottom: 1px solid #ddd;
        }
        
        .model-selector input {
            width: 100%;
            padding: 8px;
            border-radius: 4px;
            border: 1px solid #ccc;
            margin-bottom: 8px;
        }
        
        .model-list {
            max-height: 300px;
            overflow-y: auto;
            margin-bottom: 10px;
        }
        
        .model-card {
            padding: 10px;
            margin-bottom: 8px;
            border-radius: 4px;
            background-color: white;
            border: 1px solid #ddd;
            cursor: pointer;
            transition: background-color 0.2s;
        }
        
        .model-card:hover {
            background-color: #f0f0f0;
        }
        
        .model-card.selected {
            background-color: #e6f3ff;
            border-color: #007bff;
        }
        
        .model-name {
            font-weight: 500;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }
        
        .model-date {
            font-size: 0.7rem;
            color: #666;
            margin-top: 2px;
        }
        
        .model-description {
            font-size: 0.8rem;
            margin-top: 5px;
            color: #333;
            overflow: hidden;
            text-overflow: ellipsis;
            display: -webkit-box;
            -webkit-line-clamp: 2;
            -webkit-box-orient: vertical;
        }
        
        .expanded-model-details {
            margin-top: 5px;
            padding-top: 5px;
            border-top: 1px solid #eee;
            font-size: 0.8rem;
        }
        
        .model-meta {
            margin-bottom: 5px;
        }
        
        .model-expand-toggle {
            display: block;
            text-align: center;
            font-size: 0.7rem;
            color: #007bff;
            margin-top: 5px;
            cursor: pointer;
        }
        
        .chat-history {
            flex: 1;
            overflow-y: auto;
            padding: 10px;
        }
        
        .chat-item {
            padding: 10px;
            margin-bottom: 8px;
            border-radius: 4px;
            cursor: pointer;
            transition: background-color 0.2s;
        }
        
        .chat-item:hover {
            background-color: #e9e9e9;
        }
        
        .chat-item.active {
            background-color: #e0e0e0;
        }
        
        .chat-title {
            font-weight: 500;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }
        
        .chat-timestamp {
            font-size: 0.8rem;
            color: #666;
            margin-top: 4px;
        }
        
        .new-chat-btn {
            padding: 10px 15px;
            background-color: #007bff;
            color: white;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            margin: 15px;
            text-align: center;
        }
        
        .chat-main {
            flex: 1;
            display: flex;
            flex-direction: column;
            overflow: hidden;
        }
        
        .chat-header {
            padding: 15px;
            border-bottom: 1px solid #ddd;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        
        .model-info {
            display: flex;
            align-items: center;
        }
        
        .model-badge {
            display: inline-block;
            background: #e6f3ff;
            padding: 4px 8px;
            border-radius: 4px;
            font-size: 0.9rem;
        }
        
        .messages-container {
            flex: 1;
            padding: 20px;
            overflow-y: auto;
            overflow-x: hidden; /* Fix horizontal scrolling issue */
            background-color: #fff;
        }
        
        .message {
            display: flex;
            margin-bottom: 20px;
            position: relative; /* For absolute positioning of copy button */
			flex-direction:column;
        }
        
        .message.user {
            align-items: flex-end;
        }
        
        .message-content {
            max-width: 80%;
            padding: 12px 16px;
            border-radius: 8px;
            position: relative;
            word-wrap: break-word; /* Ensure long words don't cause overflow */
            overflow-wrap: break-word;
        }
        
        .user .message-content {
            background-color: #007bff;
            color: white;
            border-radius: 18px 18px 0 18px;
        }
        
        .assistant .message-content {
            background-color: #f0f0f0;
            color: #333;
            border-radius: 18px 18px 18px 0;
        }
        
        .message-model {
            font-size: 0.7rem;
            color: #666;
            margin-top: 4px;
            align-self: flex-start;
        }
        
        .copy-btn {
            position: absolute;
            bottom: 5px;
            right: 5px;
            background-color: rgba(255, 255, 255, 0.7);
            border: none;
            border-radius: 3px;
            padding: 2px 5px;
            font-size: 0.7rem;
            cursor: pointer;
            opacity: 0;
            transition: opacity 0.2s;
            z-index: 10;
        }
        
        .message:hover .copy-btn {
            opacity: 1;
        }
        
        .user .copy-btn {
            background-color: rgba(255, 255, 255, 0.5);
            color: #fff;
        }
        
        .input-container {
            padding: 15px;
            border-top: 1px solid #ddd;
            display: flex;
        }
        
        .message-input {
            flex: 1;
            padding: 12px;
            border: 1px solid #ddd;
            border-radius: 4px;
            resize: none;
            font-family: inherit;
            font-size: 14px;
        }
        
        .send-btn {
            margin-left: 10px;
            padding: 0 20px;
            background-color: #007bff;
            color: white;
            border: none;
            border-radius: 4px;
            cursor: pointer;
        }
        
        .send-btn:disabled {
            background-color: #cccccc;
            cursor: not-allowed;
        }
        
        .refresh-models-btn {
            padding: 6px 10px;
            background-color: #f0f0f0;
            border: 1px solid #ddd;
            border-radius: 4px;
            font-size: 0.8rem;
            cursor: pointer;
            margin-bottom: 8px;
        }
        
        .refresh-models-btn:hover {
            background-color: #e0e0e0;
        }
        
        .api-url-container {
            margin-top: 8px;
        }
        
        /* Loading indicator */
        .typing-indicator {
            display: flex;
            padding: 12px 16px;
            background-color: #f0f0f0;
            border-radius: 18px 18px 18px 0;
            margin-bottom: 20px;
        }
        
        .typing-indicator span {
            height: 8px;
            width: 8px;
            background-color: #666;
            border-radius: 50%;
            display: inline-block;
            margin: 0 2px;
            animation: bounce 1.3s linear infinite;
        }
        
        .typing-indicator span:nth-child(2) {
            animation-delay: 0.2s;
        }
        
        .typing-indicator span:nth-child(3) {
            animation-delay: 0.4s;
        }
        
        @keyframes bounce {
            0%, 60%, 100% { transform: translateY(0); }
            30% { transform: translateY(-5px); }
        }
        
        /* Code blocks styling */
        pre {
            background: #f5f5f5;
            padding: 10px;
            border-radius: 4px;
            overflow-x: auto;
            font-family: monospace;
            margin: 10px 0;
        }
        
        code {
            font-family: monospace;
            background-color: #f0f0f0;
            padding: 2px 4px;
            border-radius: 3px;
        }
        
        /* Mobile responsiveness */
        @media (max-width: 768px) {
            .app-container {
                flex-direction: column;
            }
            
            .sidebar {
                width: 100%;
                height: 50vh;
                border-right: none;
                border-bottom: 1px solid #ddd;
            }
            
            .chat-main {
                height: 50vh;
            }
        }
    </style>
</head>
<body>
    <div 
        x-data="chatApp()" 
        x-init="initApp()"
        class="app-container">
        
        <!-- Sidebar -->
        <div class="sidebar">
            <div class="model-selector">
                <input 
                    type="text" 
                    x-model="modelSearchQuery" 
                    placeholder="Search models..."
                    @focus="isSearchActive = true"
                    @blur="setTimeout(() => { if (!isModelDetailsActive) isSearchActive = false; }, 200)"
                />
                
                <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px;">
                    <span x-show="isModelsLoading">Loading models...</span>
                    <button 
                        class="refresh-models-btn"
                        @click="loadModels(true)" 
                        x-bind:disabled="isModelsLoading"
                    >
                        Refresh Models
                    </button>
                </div>
                
                <div x-show="modelsError" style="color: red; font-size: 0.8rem; margin-bottom: 8px;" x-text="modelsError"></div>
                
                <!-- Show model list only when actively searching and query exists -->
                <div class="model-list" x-show="isSearchActive && modelSearchQuery.trim().length > 0">
                    <template x-for="model in filteredModels" :key="model.id">
                        <div 
                            class="model-card" 
                            :class="{'selected': selectedModelId === model.id}"
                            @click="selectModel(model.id); isSearchActive = false; modelSearchQuery = '';"
                        >
                            <div class="model-name" x-text="model.name"></div>
                            <div class="model-date" x-text="formatModelDate(model.created)"></div>
                            <div class="model-description" x-text="model.description || 'No description available'"></div>
                            
                            <span 
                                class="model-expand-toggle" 
                                @click.stop="toggleModelExpand(model.id); isModelDetailsActive = true;"
                                x-text="expandedModelId === model.id ? 'Hide Details' : 'Show Details'"
                            ></span>
                            
                            <div class="expanded-model-details" 
                                x-show="expandedModelId === model.id" 
                                @click.stop 
                                @mouseenter="isModelDetailsActive = true" 
                                @mouseleave="isModelDetailsActive = false">
                                <div class="model-meta">
                                    <div><strong>ID:</strong> <span x-text="model.id"></span></div>
                                    <div><strong>Context:</strong> <span x-text="model.context_length.toLocaleString()"></span> tokens</div>
                                </div>
                                
                                <div class="model-pricing">
                                    <div><strong>Pricing:</strong></div>
                                    <div style="margin-left: 8px;">
                                        Prompt: $<span x-text="formatPrice(model.pricing.prompt)"></span> / token<br>
                                        Completion: $<span x-text="formatPrice(model.pricing.completion)"></span> / token
                                    </div>
                                </div>
                            </div>
                        </div>
                    </template>
                </div>
                
                <!-- Show selected model info when not searching -->
                <div x-show="!isSearchActive || !modelSearchQuery.trim()" class="selected-model-info" style="margin-top: 10px; margin-bottom: 10px;">
                    <div x-show="selectedModelId" style="padding: 10px; background-color: #e6f3ff; border-radius: 4px;">
                        <div style="font-weight: bold;" x-text="getCurrentModelName()"></div>
                        <div style="font-size: 0.8rem;" x-text="'ID: ' + selectedModelId"></div>
                    </div>
                    <div x-show="!selectedModelId" style="padding: 10px; background-color: #fff4e5; border-radius: 4px; font-size: 0.8rem;">
                        No model selected. Search and select a model to start chatting.
                    </div>
                </div>
                
                <div class="api-url-container">
                    <input 
                        type="text" 
                        x-model="apiUrl" 
                        @change="saveApiUrl()"
                        placeholder="API URL (default: /chat)" 
                    />
                </div>
            </div>
            
            <div class="chat-history">
                <template x-for="(chat, index) in chatHistory" :key="index">
                    <div 
                        class="chat-item" 
                        :class="{'active': activeChatIndex === index}"
                        @click="loadChat(index)">
                        <div class="chat-title" x-text="chat.title || 'New chat'"></div>
                        <div class="chat-timestamp" x-text="formatDate(chat.timestamp)"></div>
                    </div>
                </template>
            </div>
            
            <div class="new-chat-btn" @click="newChat()">
                New Chat
            </div>
        </div>
        
        <!-- Main Chat Area -->
        <div class="chat-main">
            <div class="chat-header">
                <div class="model-info">
                    <div class="model-badge" x-text="getCurrentModelName()"></div>
                </div>
                <div>
                    <button @click="clearMessages()" style="padding: 5px 10px; border: none; background: #f0f0f0; border-radius: 4px; cursor: pointer;">
                        Clear Chat
                    </button>
                </div>
            </div>
            
            <div class="messages-container" x-ref="messagesContainer">
                <template x-for="(message, index) in activeChat.messages" :key="index">
                    <div class="message" :class="message.role">
						<div class="message-content" x-html="message.role === 'assistant' ? formatMessage(message.content) : message.content"></div>
						<button 
							class="copy-btn" 
							@click="copyMessageContent($event)" 
							title="Copy message">
							Copy
						</button>
						<div x-show="message.role === 'assistant' && message.model" class="message-model">
							<span x-text="message.model"></span>
						</div>
                    </div>
                </template>
                
                <div x-show="isLoading" class="message assistant">
                    <div class="typing-indicator">
                        <span></span>
                        <span></span>
                        <span></span>
                    </div>
                </div>
            </div>
            
            <div class="input-container">
                <textarea 
                    x-model="userInput" 
                    @keydown="handleKeyDown($event)"
                    class="message-input" 
                    placeholder="Type your message here... (Shift+Enter for new line)"
                    rows="2"></textarea>
                <button 
                    class="send-btn" 
                    :disabled="isLoading || !userInput.trim() || !selectedModelId"
                    @click="sendMessage()">
                    Send
                </button>
            </div>
        </div>
    </div>
    
    <script>
        function chatApp() {
            return {
                // Models
                models: [],
                isModelsLoading: true,
                modelsError: null,
                selectedModelId: '',
                modelSearchQuery: '',
                expandedModelId: null,
                isSearchActive: false,
                isModelDetailsActive: false,
                apiUrl: '',
                
                // Chat state
                chatHistory: [],
                activeChatIndex: 0,
                activeChat: { messages: [], timestamp: Date.now() },
                userInput: '',
                isLoading: false,
                
                // Markdown converter
                converter: null,
                
                // Get filtered models based on search
                get filteredModels() {
                    // Only search when there's a query
                    if (!this.modelSearchQuery.trim()) {
                        return [];
                    }
                    
                    const query = this.modelSearchQuery.toLowerCase();
                    return this.models
                        .filter(model => 
                            model.name.toLowerCase().includes(query) || 
                            model.id.toLowerCase().includes(query)
                            // Removed description from search criteria
                        )
                        .sort((a, b) => b.created - a.created); // Sort by created date (newest first)
                },
                
                // Initialize the app
                initApp() {
                    this.loadModels();
                    this.loadChatHistory();
                    this.selectedModelId = localStorage.getItem('openrouter_selected_model') || '';
                    this.apiUrl = localStorage.getItem('openrouter_api_url') || '/chat';
                    
                    // Initialize Showdown converter
                    this.converter = new showdown.Converter({
                        tables: true,
                        simplifiedAutoLink: true,
                        strikethrough: true,
                        tasklists: true,
                        ghCodeBlocks: true
                    });
                },
                
                // Format price to a readable format
                formatPrice(price) {
                    return parseFloat(price).toFixed(7);
                },
                
                // Format model date
                formatModelDate(timestamp) {
                    return new Date(timestamp * 1000).toLocaleDateString();
                },
                
                // Toggle expanded model details
                toggleModelExpand(modelId) {
                    if (this.expandedModelId === modelId) {
                        this.expandedModelId = null;
                    } else {
                        this.expandedModelId = modelId;
                    }
                },
                
                // Select a model
                selectModel(modelId) {
                    this.selectedModelId = modelId;
                    this.saveSelectedModel();
                },
                
                // Load models from OpenRouter API
                async loadModels(forceRefresh = false) {
                    this.isModelsLoading = true;
                    this.modelsError = null;
                    
                    try {
                        // Try to load from localStorage first (unless force refresh)
                        if (!forceRefresh) {
                            const cachedModels = localStorage.getItem('openrouter_models');
                            if (cachedModels) {
                                const parsed = JSON.parse(cachedModels);
                                if (parsed.models && Array.isArray(parsed.models)) {
                                    this.models = parsed.models;
                                    // Only use cache if it's less than 24 hours old
                                    const cacheTime = new Date(parsed.lastFetched || 0);
                                    const now = new Date();
                                    if ((now - cacheTime) < 24 * 60 * 60 * 1000) {
                                        this.isModelsLoading = false;
                                        return;
                                    }
                                }
                            }
                        }
                        
                        // Fetch from API
                        const response = await fetch('https://openrouter.ai/api/v1/models');
                        if (!response.ok) {
                            throw new Error('Failed to fetch models: ' + response.status);
                        }
                        
                        const data = await response.json();
                        this.models = data.data || [];
                        
                        // Save to localStorage
                        localStorage.setItem('openrouter_models', JSON.stringify({
                            models: this.models,
                            lastFetched: new Date().toISOString()
                        }));
                    } catch (err) {
                        this.modelsError = err.message;
                        console.error('Error loading models:', err);
                    } finally {
                        this.isModelsLoading = false;
                    }
                },
                
                // Save selected model to localStorage
                saveSelectedModel() {
                    localStorage.setItem('openrouter_selected_model', this.selectedModelId);
                },
                
                // Save API URL to localStorage
                saveApiUrl() {
                    localStorage.setItem('openrouter_api_url', this.apiUrl);
                },
                
                // Get current model name
                getCurrentModelName() {
                    if (!this.selectedModelId) return 'No model selected';
                    const model = this.models.find(m => m.id === this.selectedModelId);
                    return model ? model.name : this.selectedModelId;
                },
                
                // Load chat history from localStorage
                loadChatHistory() {
                    const history = localStorage.getItem('openrouter_chat_history');
                    if (history) {
                        try {
                            this.chatHistory = JSON.parse(history);
                            if (this.chatHistory.length > 0) {
                                this.activeChat = this.chatHistory[0];
                            } else {
                                this.chatHistory = [{ messages: [], timestamp: Date.now() }];
                                this.activeChat = this.chatHistory[0];
                            }
                        } catch (err) {
                            console.error('Error parsing chat history:', err);
                            this.chatHistory = [{ messages: [], timestamp: Date.now() }];
                            this.activeChat = this.chatHistory[0];
                        }
                    } else {
                        this.chatHistory = [{ messages: [], timestamp: Date.now() }];
                        this.activeChat = this.chatHistory[0];
                    }
                },
                
                // Save chat history to localStorage
                saveChatHistory() {
                    localStorage.setItem('openrouter_chat_history', JSON.stringify(this.chatHistory));
                },
                
                // Create a new chat
                newChat() {
                    const newChat = {
                        messages: [],
                        timestamp: Date.now()
                    };
                    this.chatHistory.unshift(newChat);
                    this.activeChatIndex = 0;
                    this.activeChat = newChat;
                    this.saveChatHistory();
                },
                
                // Load a specific chat
                loadChat(index) {
                    this.activeChatIndex = index;
                    this.activeChat = this.chatHistory[index];
                },
                
                // Handle keydown events in the textarea
                handleKeyDown(event) {
                    // If Shift+Enter is pressed, insert a newline instead of sending
                    if (event.key === 'Enter' && !event.shiftKey) {
                        event.preventDefault();
                        this.sendMessage();
                    }
                    // Otherwise, let the default behavior happen (which will insert a newline)
                },
                
                // Copy message content to clipboard
                copyMessageContent(event) {
                    const button = event.target;
                    const messageDiv = button.closest('.message');
                    const content = messageDiv.querySelector('.message-content').textContent;
                    
                    // Create a temporary text area to copy from
                    const textArea = document.createElement('textarea');
                    textArea.value = content;
                    document.body.appendChild(textArea);
                    textArea.select();
                    document.execCommand('copy');
                    document.body.removeChild(textArea);
                    
                    // Change button text to "Copied"
                    const originalText = button.textContent;
                    button.textContent = "Copied!";
                    
                    // Reset button text after 2 seconds
                    setTimeout(() => {
                        button.textContent = originalText;
                    }, 2000);
                },
                
                // Send a message
                async sendMessage() {
                    if (!this.userInput.trim() || !this.selectedModelId) return;
                    
                    // Add user message
                    const userMessage = {
                        role: 'user',
                        content: this.userInput.trim()
                    };
                    this.activeChat.messages.push(userMessage);
                    
                    // Set the first message as the title
                    if (this.activeChat.messages.length === 1) {
                        this.activeChat.title = this.userInput.trim().substring(0, 30) + (this.userInput.length > 30 ? '...' : '');
                    }
                    
                    // Clear input and save
                    this.userInput = '';
                    this.saveChatHistory();
                    
                    // Auto scroll to bottom
                    this.$nextTick(() => {
                        this.scrollToBottom();
                    });
                    
                    // Get AI response
                    this.isLoading = true;
                    try {
                        const messages = this.activeChat.messages.map(m => ({
                            role: m.role,
                            content: m.content
                        }));
                        
                        const endpoint = this.apiUrl || '/chat';
                        const response = await fetch(endpoint, {
                            method: 'POST',
                            headers: {
                                'Content-Type': 'application/json',
                            },
                            body: JSON.stringify({
                                model: this.selectedModelId,
                                messages: messages
                            })
                        });
                        
                        if (!response.ok) {
                            throw new Error('Error: ' + response.status);
                        }
                        
                        const data = await response.json();
                        const assistantMessage = {
                            role: 'assistant',
                            content: data.choices[0].message.content,
                            model: data.model // Save the model used for this response
                        };
                        
                        this.activeChat.messages.push(assistantMessage);
                        this.saveChatHistory();
                    } catch (err) {
                        // Add error message
                        this.activeChat.messages.push({
                            role: 'assistant',
                            content: 'Error: ' + err.message,
                            model: 'error'
                        });
                        console.error('Error getting response:', err);
                    } finally {
                        this.isLoading = false;
                        this.$nextTick(() => {
                            this.scrollToBottom();
                        });
                    }
                },
                
                // Format date for display
                formatDate(timestamp) {
                    return new Date(timestamp).toLocaleString();
                },
                
                // Format message content using Showdown for Markdown
                formatMessage(content) {
                    if (!content) return '';
                    
                    // Use Showdown to convert Markdown to HTML
                    return this.converter.makeHtml(content);
                },
                
                // Clear messages
                clearMessages() {
                    if (confirm('Are you sure you want to clear this chat?')) {
                        this.activeChat.messages = [];
                        this.saveChatHistory();
                    }
                },
                
                // Scroll to bottom of messages
                scrollToBottom() {
                    const container = this.$refs.messagesContainer;
                    container.scrollTop = container.scrollHeight;
                }
            };
        }
    </script>
</body>
</html>
`

// ServeChatUI handles the request to show the chat UI
func ServeChatUI(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("chat").Parse(ChatTemplate)
	if err != nil {
		http.Error(w, "Failed to parse template", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}
