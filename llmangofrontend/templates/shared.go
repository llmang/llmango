package templates

// SharedTemplates contains the shared templates like styles, header, and footer
const SharedTemplates = `
{{define "styles"}}
<style>
    * {box-sizing:border-box;}
    body { font-family: sans-serif; line-height: 1.6; padding: 20px; max-width: 900px; margin: auto; color: #333; }
    h1, h2, h3 { color: #444; }
    .card-container { display: flex; gap: 15px; flex-wrap: wrap; margin-bottom: 20px; }
    .card { border: 1px solid #ddd; border-radius: 5px; padding: 15px; min-width: 150px; background-color: #f9f9f9; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
    .card h3 { margin-top: 0; }
    ul { list-style: none; padding: 0; }
    li { background-color: #eee; margin-bottom: 5px; padding: 8px 12px; border-radius: 3px; }
    hr { border: 0; height: 1px; background: #ddd; margin: 30px 0; }
    code { background-color: #f0f0f0; padding: 2px 5px; border-radius: 3px; }
    ol li { background-color: transparent; margin-bottom: 10px; padding: 0; }
    nav { margin-bottom: 25px; border-bottom: 1px solid #eee; padding-bottom: 10px; }
    nav a { margin-right: 15px; text-decoration: none; color: #007bff; }
    nav a:hover { text-decoration: underline; }
    a{ text-decoration:none;}
    pre { 
        background: #f5f5f5; 
        padding: 10px; 
        border-radius: 4px; 
        white-space: pre-wrap;
        word-wrap: break-word;
        word-break: break-word;
        overflow-wrap: break-word;
        max-width: 100%;
    }
    .json-key { color: #0057b7; }
    .json-string { color: #008000; }
    .json-number { color: #a31515; }
    .json-boolean { color: #0000ff; }
    .json-null { color: #808080; }
    
    /* Log viewer styles */
    .filters-section {
        background-color: #f8f8f8;
        border-radius: 5px;
        padding: 15px;
        margin-bottom: 20px;
    }
    
    .filters-section h3 {
        margin-top: 0;
        margin-bottom: 15px;
    }
    
    .filter-controls {
        display: flex;
        flex-wrap: wrap;
        gap: 15px;
        align-items: flex-end;
    }
    
    .filter-group {
        display: flex;
        flex-direction: column;
        min-width: 150px;
    }
    
    .filter-group label {
        margin-bottom: 5px;
        font-size: 0.9rem;
    }
    
    .log-viewer {
        margin-top: 20px;
    }
    
    .logs-container {
        margin-bottom: 20px;
    }
    
    .log-table {
        width: 100%;
        border-collapse: collapse;
    }
    
    .log-header-row {
        background-color: #f5f5f5;
        font-weight: bold;
        display: flex;
    }
    
    .log-row {
        display: flex;
        border-bottom: 1px solid #eee;
    }
    
    .log-cell {
        padding: 10px;
        flex: 1;
        min-width: 0;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }
    
    .log-details {
        padding: 15px;
        background-color: #f9f9f9;
        border-bottom: 1px solid #eee;
    }
    
    .details-row {
        display: flex;
        margin-bottom: 15px;
    }
    
    .details-cell {
        flex: 1;
        padding-right: 15px;
    }
    
    .log-section {
        margin-bottom: 15px;
    }
    
    .log-section h5 {
        margin-top: 0;
        margin-bottom: 5px;
    }
    
    .token-count {
        font-family: monospace;
    }
    
    .error {
        color: #d32f2f;
    }
    
    .pagination {
        display: flex;
        justify-content: center;
        align-items: center;
        gap: 15px;
        margin-top: 20px;
    }
    
    .loading {
        text-align: center;
        padding: 20px;
        font-style: italic;
        color: #666;
    }
    
    .no-logs {
        text-align: center;
        padding: 20px;
        color: #666;
        font-style: italic;
    }
    
    /* JSON formatter styles */
    .json-container {
        font-family: monospace;
        background: #f5f5f5;
        padding: 10px;
        border-radius: 4px;
        white-space: pre-wrap;
    }
    
    /* Modal styles */
    .modal-overlay {
        position: fixed;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        background: rgba(0,0,0,0.5);
        display: flex;
        justify-content: center;
        align-items: center;
        z-index: 1000;
    }
    
    .modal-container {
        background: white;
        padding: 20px;
        border-radius: 5px;
        width: 90%;
        max-width: 600px;
        max-height: 90vh;
        overflow-y: auto;
        box-shadow: 0 4px 8px rgba(0,0,0,0.1);
    }
    
    .modal-header {
        margin-top: 0;
        padding-bottom: 10px;
        border-bottom: 1px solid #eee;
        margin-bottom: 15px;
    }
    
    .form-group {
        margin-bottom: 15px;
    }
    
    .form-label {
        display: block;
        margin-bottom: 5px;
        font-weight: 500;
    }
    
    .form-control {
        width: 100%;
        padding: 8px;
        border: 1px solid #ddd;
        border-radius: 4px;
    }
    
    .form-check {
        display: flex;
        align-items: center;
    }
    
    .form-check-label {
        margin-left: 5px;
    }
    
    .modal-footer {
        display: flex;
        justify-content: flex-end;
        gap: 10px;
        margin-top: 20px;
        padding-top: 15px;
        border-top: 1px solid #eee;
    }
    
    .btn {
        padding: 8px 15px;
        border-radius: 4px;
        cursor: pointer;
    }
    
    .btn-secondary {
        border: 1px solid #ddd;
        background: #f5f5f5;
    }
    
    .btn-primary {
        border: none;
        background: #007bff;
        color: white;
    }

    /* Variables box styles */
    .variables-box {
        background-color: #f8f8ff;
        border: 1px solid #e0e0ff;
        border-radius: 4px;
        padding: 10px;
        margin-bottom: 15px;
    }
    
    .variables-help {
        margin-top: 0;
        margin-bottom: 10px;
        font-size: 0.9rem;
        color: #555;
    }
    
    .available-variables code {
        background-color: #eef;
        padding: 2px 4px;
        border-radius: 3px;
        font-family: monospace;
    }
    
    #inputVariablesPreview {
        font-family: monospace;
        font-size: 0.85rem;
        max-height: 150px;
        overflow: auto;
    }
    

    .card:hover {
        transform: translateY(-2px);
        box-shadow: 0 4px 8px rgba(0,0,0,0.1);
    }
    
    .prompt-title {
        margin-top: 0;
        margin-bottom: 10px;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
        font-size: 1.1rem;
    }
    
    .prompt-meta {
        display: flex;
        flex-direction: column;
        gap: 8px;
    }
    
    .model-badge.small {
        font-size: 0.75rem;
        padding: 2px 6px;
        display: inline-block;
    }
    
    .message-count {
        font-size: 0.8rem;
        color: #666;
    }
    
    .message-preview {
        margin-top: 10px;
        border-top: 1px solid #eee;
        padding-top: 10px;
    }
    
    .preview-label {
        font-size: 0.75rem;
        color: #666;
        margin-bottom: 3px;
    }
    
    .preview-content {
        font-size: 0.8rem;
        color: #333;
        display: -webkit-box;
        -webkit-line-clamp: 3;
        -webkit-box-orient: vertical;
        overflow: hidden;
        text-overflow: ellipsis;
        max-height: 4.5em;
        font-family: monospace;
        background-color: #f8f8f8;
        padding: 5px;
        border-radius: 3px;
        line-height: 1.5;
        word-break: break-word;
    }
    
    .clickable {
        cursor: pointer;
    }
    
    .plus-icon {
        font-size: 24px;
        margin-bottom: 10px;
    }
    
    .clickable {
        transition: transform 0.2s, box-shadow 0.2s;
    }
    
    .clickable:hover {
        transform: translateY(-2px);
        box-shadow: 0 4px 8px rgba(0,0,0,0.1);
    }
    
    .model-badge {
        display: inline-block;
        background: #e6f3ff;
        padding: 4px 8px;
        border-radius: 4px;
        font-size: 0.9rem;
        margin-bottom: 15px;
    }
</style>

<script>
    // JSON syntax highlighter
    function formatJSON(json) {
        const formatted = JSON.stringify(json, null, 2);
        return formatted.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, 
            function (match) {
                let cls = 'json-number';
                if (/^"/.test(match)) {
                    if (/:$/.test(match)) {
                        cls = 'json-key';
                    } else {
                        cls = 'json-string';
                    }
                } else if (/true|false/.test(match)) {
                    cls = 'json-boolean';
                } else if (/null/.test(match)) {
                    cls = 'json-null';
                }
                return '<span class="' + cls + '">' + match + '</span>';
            });
    }

    function initJSONFormatters() {
        document.querySelectorAll('.json-formatter').forEach(el => {
            try {
                const jsonData = JSON.parse(el.textContent);
                el.innerHTML = formatJSON(jsonData);
            } catch (e) {
                console.error("JSON parsing error:", e);
            }
        });
    }
    
    // Initialize when document is loaded
    document.addEventListener('DOMContentLoaded', function() {
        initJSONFormatters();
    });
</script>
{{end}}

{{define "header"}}
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{if .Title}}{{.Title}}{{else}}LLMango{{end}}</title>
    <script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.13.0/dist/cdn.min.js"></script>
    <script>
        // OpenRouter Model Fetcher
        document.addEventListener('alpine:init', () => {
            Alpine.store('modelStore', {
                models: [],
                loading: false,
                error: null,
                lastFetched: null,
                hasModels: false,

                async fetchModels(force = false) {
                    // Check if we already have models and not forcing refresh
                    if (this.models.length > 0 && !force) {
                        return;
                    }

                    this.loading = true;
                    this.error = null;

                    try {
                        const response = await fetch('https://openrouter.ai/api/v1/models');
                        if (!response.ok) {
                            throw new Error('Failed to fetch models: ' + response.status);
                        }
                        
                        const data = await response.json();
                        this.models = data.data || [];
                        this.hasModels = this.models.length > 0;
                        this.lastFetched = new Date().toISOString();
                        
                        // Save to localStorage for caching
                        localStorage.setItem('openrouter_models', JSON.stringify({
                            models: this.models,
                            lastFetched: this.lastFetched
                        }));
                    } catch (err) {
                        this.error = err.message;
                        console.error('Error fetching models:', err);
                    } finally {
                        this.loading = false;
                    }
                },

                init() {
                    // Try to load from cache first
                    const cached = localStorage.getItem('openrouter_models');
                    if (cached) {
                        try {
                            const data = JSON.parse(cached);
                            this.models = data.models || [];
                            this.lastFetched = data.lastFetched;
                            this.hasModels = this.models.length > 0;
                        } catch (err) {
                            console.error('Error parsing cached models:', err);
                        }
                    }

                    // Fetch fresh data if cache is empty or stale (older than 24 hours)
                    if (!this.models.length || this.isCacheStale()) {
                        this.fetchModels();
                    }
                },

                isCacheStale() {
                    if (!this.lastFetched) return true;
                    
                    const lastFetchedDate = new Date(this.lastFetched);
                    const now = new Date();
                    // Check if last fetch was more than 24 hours ago
                    return (now - lastFetchedDate) > (24 * 60 * 60 * 1000);
                },

                filteredModels(query = '') {
                    if (!query) return this.models.slice().sort((a, b) => b.created - a.created);
                    
                    const lowerQuery = query.toLowerCase();
                    return this.models
                        .filter(model => 
                            model.id.toLowerCase().includes(lowerQuery) || 
                            model.name.toLowerCase().includes(lowerQuery)
                        )
                        .sort((a, b) => b.created - a.created);
                }
            });
        });
    </script>
    {{template "styles"}}
</head>
<body>
    <div style="position: relative;">
        <img src="https://public.llmang.com/logos/llmango.png" alt="LLMango Logo" style="position: absolute; top: 0; right: 0; height: 3rem;">
        <h1>LLMango</h1>
        <nav>
            <a href="{{.BaseRoute}}/">Home</a>
            <a href="{{.BaseRoute}}/prompts">Prompts</a>
            <a href="{{.BaseRoute}}/goals">Goals</a>
            <a href="{{.BaseRoute}}/models">Models</a>
            <a href="{{.BaseRoute}}/logs">Logs</a>
            <a>|</a>
            <a>Chat</a>
            <a>Evaluate</a>
            <a></a>
            <span x-data>
                <span 
                    x-show="$store.modelStore.hasModels" 
                    title="OpenRouter models loaded"
                    style="color: green; cursor: default;">models loaded ✓</span>
                <span 
                    @click="if(confirm('Refresh OpenRouter models?')) $store.modelStore.fetchModels(true)" 
                    title="Refresh OpenRouter models"
                    style="cursor: pointer; margin-left: 5px;">🔄</span>
            </span>
        </nav>
    </div>
{{end}}

{{define "footer"}}
</body>
</html>
{{end}}

{{define "prompt-form-template"}}
<div class="form-group">
    <label class="form-label">Prompt UID</label>
    <input type="text" x-model="newPrompt.uid" class="form-control" placeholder="Enter a unique identifier">
</div>

<div class="form-group">
    <label class="form-label">Model</label>
    <div x-data="{ showModelSearch: false, modelSearch: '' }">
        <div style="position: relative;">
            <div style="display: flex; gap: 10px; align-items: center;">
                <select x-model="newPrompt.model" class="form-control">
                    <option value="">-- Select a model --</option>
                    <template x-if="$store.modelStore.models.length === 0">
                        <!-- Fallback options if models haven't been fetched -->
                        <template>
                            <option value="anthropic/claude-3-sonnet-20240229">Claude 3 Sonnet</option>
                            <option value="anthropic/claude-3-opus-20240229">Claude 3 Opus</option>
                            <option value="anthropic/claude-3-haiku-20240307">Claude 3 Haiku</option>
                            <option value="openai/gpt-4-turbo">GPT-4 Turbo</option>
                            <option value="openai/gpt-3.5-turbo">GPT-3.5 Turbo</option>
                            <option value="google/gemini-1.5-pro">Gemini 1.5 Pro</option>
                        </template>
                    </template>
                    <template x-if="$store.modelStore.models.length > 0">
                        <template x-for="model in $store.modelStore.models" :key="model.id">
                            <option x-bind:value="model.id" x-text="model.name + ' (' + model.id + ')'"></option>
                        </template>
                    </template>
                </select>
                <button 
                    type="button" 
                    class="btn btn-secondary" 
                    @click="showModelSearch = !showModelSearch"
                    x-text="showModelSearch ? 'Hide Search' : 'Search'"
                ></button>
            </div>
            
            <div x-show="showModelSearch" style="position: absolute; top: 100%; left: 0; width: 100%; background: white; border: 1px solid #ddd; border-radius: 4px; z-index: 100; margin-top: 5px; max-height: 300px; overflow-y: auto; padding: 10px; box-shadow: 0 4px 8px rgba(0,0,0,0.1);">
                <div class="form-group" style="margin-bottom: 10px;">
                    <input 
                        type="text" 
                        x-model="modelSearch" 
                        placeholder="Search models..." 
                        class="form-control"
                        style="width: 100%;"
                    >
                </div>
                
                <div x-show="$store.modelStore.models.length === 0">
                    <p>No models available. <a href="{{.BaseRoute}}/models" target="_blank">Fetch models</a> first.</p>
                </div>
                
                <div x-show="$store.modelStore.models.length > 0">
                    <div style="display: flex; flex-direction: column; gap: 5px;">
                        <template x-for="model in $store.modelStore.filteredModels(modelSearch)" :key="model.id">
                            <div 
                                @click="newPrompt.model = model.id; showModelSearch = false" 
                                style="cursor: pointer; padding: 5px; border-radius: 3px;"
                                x-bind:class="{ 'bg-light': newPrompt.model === model.id }"
                                class="hover:bg-light"
                            >
                                <div style="display: flex; justify-content: space-between; align-items: center;">
                                    <div style="font-weight: bold;" x-text="model.name"></div>
                                    <small x-text="new Date(model.created * 1000).toLocaleDateString()"></small>
                                </div>
                                <div style="font-size: 0.8rem;">
                                    <code x-text="model.id"></code>
                                </div>
                                <div style="font-size: 0.8rem;">
                                    Context: <span x-text="model.context_length.toLocaleString()"></span> tokens
                                </div>
                                <div style="font-size: 0.8rem; display: flex; gap: 5px;">
                                    <span x-text="'$' + parseFloat(model.pricing.prompt).toFixed(7) + '/token'"></span>
                                    <span x-show="model.architecture && model.architecture.input_modalities">
                                        | Input: <span x-text="model.architecture.input_modalities.join(', ')"></span>
                                    </span>
                                </div>
                            </div>
                        </template>
                    </div>
                </div>
            </div>
        </div>
    </div>
</div>

<!-- Input Variables Help Box - Only shown when goal input type information is available -->
{{if .GoalInfo}}
<div class="form-group available-variables">
    <label class="form-label">Available Input Variables</label>
    <div class="variables-box">
        {{if .GoalInfo.ExampleInput}}
        <p class="variables-help">You can use these variables in your prompt messages with the syntax <code>{{"{{"}}variable_name{{"}}"}}</code></p>
        <div class="json-preview" id="inputVariablesPreview">{{.GoalInfo.ExampleInput}}</div>
        {{else}}
        <p>No input variables information available</p>
        {{end}}
    </div>
</div>
{{end}}

<div class="form-group">
    <label class="form-label">Messages</label>
    <div x-data="{ messages: [] }" x-init="messages = newPrompt.messages || []">
        <template x-for="(message, index) in messages" :key="index">
            <div style="margin-bottom: 10px; border: 1px solid #eee; padding: 10px; border-radius: 4px;">
                <div class="form-group">
                    <label class="form-label">Role</label>
                    <select x-model="message.role" class="form-control">
                        <option value="user">User</option>
                        <option value="assistant">Assistant</option>
                        <option value="system">System</option>
                    </select>
                </div>
                <div class="form-group">
                    <label class="form-label">Content</label>
                    <textarea x-model="message.content" class="form-control" rows="3"></textarea>
                </div>
                <button type="button" @click="messages.splice(index, 1); newPrompt.messages = messages" class="btn btn-secondary">Remove</button>
            </div>
        </template>
        <button 
            type="button" 
            @click="messages.push({role: 'user', content: ''}); newPrompt.messages = messages" 
            class="btn btn-secondary"
        >
            Add Message
        </button>
    </div>
</div>

<div class="form-group">
    <label class="form-label">Parameters</label>
    <div x-data="{ parameters: {} }" x-init="parameters = newPrompt.parameters || {}">
        <div class="form-group">
            <label class="form-label">Temperature</label>
            <input type="number" x-model="parameters.temperature" min="0" max="2" step="0.1" class="form-control" placeholder="0.7">
            <small class="form-text text-muted">Range: [0, 2]</small>
        </div>
        <div class="form-group">
            <label class="form-label">Max Tokens</label>
            <input type="number" x-model="parameters.max_tokens" min="1" max="10000" class="form-control" placeholder="1000">
        </div>
        <div class="form-group">
            <label class="form-label">Top P</label>
            <input type="number" x-model="parameters.top_p" min="0" max="1" step="0.05" class="form-control" placeholder="0.9">
            <small class="form-text text-muted">Range: (0, 1]</small>
        </div>
        <div class="form-group">
            <label class="form-label">Frequency Penalty</label>
            <input type="number" x-model="parameters.frequency_penalty" min="-2" max="2" step="0.1" class="form-control" placeholder="0">
            <small class="form-text text-muted">Range: [-2, 2]</small>
        </div>
        <div class="form-group">
            <label class="form-label">Presence Penalty</label>
            <input type="number" x-model="parameters.presence_penalty" min="-2" max="2" step="0.1" class="form-control" placeholder="0">
            <small class="form-text text-muted">Range: [-2, 2]</small>
        </div>
        <div x-init="newPrompt.parameters = parameters"></div>
    </div>
</div>
{{end}}

{{define "card-styles"}}
<style>
    .card-container {
        display: flex;
        flex-wrap: wrap;
        gap: 15px;
        margin-bottom: 20px;
    }
    
    .new-card {
        width: 200px;
        height: 150px;
        display: flex;
        justify-content: center;
        align-items: center;
        cursor: pointer;
        background: #f9f9f9;
        border: 1px dashed #aaa;
    }
    
    .new-prompt-content {
        text-align: center;
    }
    
    .plus-icon {
        font-size: 24px;
        margin-bottom: 10px;
    }
    
    .clickable {
        transition: transform 0.2s, box-shadow 0.2s;
    }
    
    .clickable:hover {
        transform: translateY(-2px);
        box-shadow: 0 4px 8px rgba(0,0,0,0.1);
    }
    
    .model-badge {
        display: inline-block;
        background: #e6f3ff;
        padding: 4px 8px;
        border-radius: 4px;
        font-size: 0.9rem;
        margin-bottom: 15px;
    }
    
    .message-card {
        border: 1px solid #eee;
        border-radius: 5px;
        padding: 10px;
        margin-bottom: 10px;
    }
    
    .message-role {
        font-weight: bold;
        margin-top: 0;
        color: #666;
        text-transform: uppercase;
        font-size: 0.8rem;
    }
    
    .message-content {
        margin: 0;
        white-space: pre-wrap;
        font-family: monospace;
        padding: 5px;
        background: #f9f9f9;
        border-radius: 3px;
    }
</style>
{{end}}

{{define "json-formatter"}}
<script>
    // This function formats JSON with syntax highlighting
    function formatJSON(obj) {
        return syntaxHighlight(JSON.stringify(obj, null, 2));
    }
    
    // Add syntax highlighting
    function syntaxHighlight(json) {
        json = json.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
        return json.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, 
            function (match) {
                let cls = 'number';
                if (/^"/.test(match)) {
                    if (/:$/.test(match)) {
                        cls = 'key';
                        match = match.replace(':', '');
                    } else {
                        cls = 'string';
                    }
                } else if (/true|false/.test(match)) {
                    cls = 'boolean';
                } else if (/null/.test(match)) {
                    cls = 'null';
                }
                
                if (cls === 'key') {
                    return '<span class="' + cls + '">' + match + '</span>: ';
                }
                return '<span class="' + cls + '">' + match + '</span>';
            }
        );
    }
    
    document.addEventListener('DOMContentLoaded', function() {
        // Format all elements with class json-preview
        document.querySelectorAll('.json-preview').forEach(el => {
            try {
                const jsonData = JSON.parse(el.textContent);
                el.innerHTML = formatJSON(jsonData);
            } catch(e) {
                console.error("JSON parsing error:", e);
            }
        });
    });
</script>
{{end}}
`
