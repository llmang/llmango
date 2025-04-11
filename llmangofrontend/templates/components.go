package templates

// ComponentTemplates contains reusable template components for displaying goals and prompts
const ComponentTemplates = `
{{define "goal-card"}}
    <a href="{{$.BaseRoute}}/goal/{{.ID}}" class="card clickable" style="display: block; text-decoration: none; color: inherit; transition: transform 0.2s, box-shadow 0.2s;">
        <h3>{{if .Goal.Title}}{{.Goal.Title}}{{else}}Unnamed Goal{{end}}</h3>
        <p style="display: -webkit-box; -webkit-line-clamp: 3; -webkit-box-orient: vertical; overflow: hidden; text-overflow: ellipsis;">{{.Goal.Description}}</p>
        
        <div class="examples-grid">
            {{if .Goal.ExampleInput}}
            <div class="example-section">
                <h4 class="example-title">Input</h4>
                <div class="json-preview">{{.Goal.ExampleInput}}</div>
            </div>
            {{end}}
            
            {{if .Goal.ExampleOutput}}
            <div class="example-section">
                <h4 class="example-title">Output</h4>
                <div class="json-preview">{{.Goal.ExampleOutput}}</div>
            </div>
            {{end}}
        </div>
        
        <p style="font-size: 0.8rem; margin-top: auto; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; color: #666;">ID: {{.ID}}</p>
        
        <!-- Debug info -->
        <details>
            <summary>Debug Info</summary>
            <pre>{{printf "%#v" .Goal}}</pre>
        </details>
    </a>
{{end}}

{{define "card"}}
    <a href="{{$.BaseRoute}}/prompt/{{.ID}}" class="card clickable" style="display: block; text-decoration: none; color: inherit; transition: transform 0.2s, box-shadow 0.2s;">
        <div class="model-badge">{{.Prompt.Model}}</div>
        {{if gt (len .Prompt.Messages) 0}}
            <p style="display: -webkit-box; -webkit-line-clamp: 3; -webkit-box-orient: vertical; overflow: hidden; text-overflow: ellipsis; margin-bottom: 1rem;">{{index .Prompt.Messages 0 | getMessageContent}}</p>
        {{else}}
            <p style="margin-bottom: 1rem;">No messages</p>
        {{end}}
        <div style="display: flex; justify-content: space-between; margin-top: auto;">
            <p style="font-size: 0.8rem; color: #666;">Messages: {{len .Prompt.Messages}}</p>
            <p style="font-size: 0.8rem; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; color: #666;">ID: {{.ID}}</p>
        </div>
    </a>
{{end}}

{{define "card-styles"}}
<style>
    .card-container {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
        gap: 1.5rem;
        margin-bottom: 2rem;
    }
    
    .card {
        display: flex;
        flex-direction: column;
        height: 100%;
        width: 100%;
        padding: 1.25rem;
        border-radius: 0.5rem;
        border: 1px solid #ddd;
        background-color: #fff;
        box-shadow: 0 2px 4px rgba(0,0,0,0.05);
    }
    
    .card.clickable {
        cursor: pointer;
    }
    
    .card.clickable:hover {
        transform: translateY(-3px);
        box-shadow: 0 4px 8px rgba(0,0,0,0.1);
    }
    
    .card h3 {
        margin-top: 0;
        margin-bottom: 1rem;
    }
    
    .card p {
        margin-bottom: 1rem;
    }
    
    .card p:last-of-type {
        margin-top: auto;
        margin-bottom: 0;
    }
    
    .model-badge {
        display: inline-block;
        background-color: #f0f0f0;
        border-radius: 4px;
        padding: 0.3rem 0.6rem;
        font-size: 0.75rem;
        font-weight: 600;
        color: #555;
        margin-bottom: 0.75rem;
    }
    
    .examples-grid {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 0.5rem;
        margin: 0.75rem 0 1rem 0;
    }
    
    .example-section {
        margin: 0;
        font-size: 0.85rem;
    }
    
    .example-title {
        margin: 0 0 0.3rem 0;
        font-size: 0.8rem;
        color: #555;
    }
    
    .json-preview {
        background-color: #f5f5f5;
        border-radius: 0.25rem;
        padding: 0.4rem;
        font-family: monospace;
        font-size: 0.7rem;
        height: 80px;
        overflow: auto;
        white-space: pre-wrap;
        margin-bottom: 0;
        position: relative;
        
        /* Custom syntax highlighting */
        color: #333;
    }
    
    .new-card {
        border: 1px dashed #aaa;
        background-color: #f9f9f9;
        display: flex;
        justify-content: center;
        align-items: center;
    }
    
    .new-prompt-content {
        text-align: center;
    }
    
    .plus-icon {
        font-size: 24px;
        margin-bottom: 10px;
    }
    
    details {
        margin-top: 1rem;
        border-top: 1px solid #eee;
        padding-top: 0.5rem;
    }
    
    details pre {
        background: #f5f5f5;
        padding: 0.5rem;
        border-radius: 0.25rem;
        overflow: auto;
        font-size: 0.8rem;
    }
    
    summary {
        cursor: pointer;
        color: #666;
        font-size: 0.8rem;
    }
</style>
{{end}}

{{define "json-formatter"}}
<script>
    document.addEventListener('DOMContentLoaded', function() {
        // Format all JSON preview elements
        document.querySelectorAll('.json-preview').forEach(function(element) {
            try {
                const jsonText = element.textContent;
                const jsonData = JSON.parse(jsonText);
                element.innerHTML = formatJSON(jsonData);
            } catch(e) {
                // If parsing fails, keep the original content
                console.error('JSON parse error:', e);
            }
        });
        
        // Helper function to format JSON with syntax highlighting
        function formatJSON(obj) {
            return syntaxHighlight(JSON.stringify(obj, replacer, 2));
        }
        
        // Replace empty values with type indicators
        function replacer(key, value) {
            if (value === null) return null;
            if (value === "") return "[string]";
            if (value === 0 && key !== "") return "[number]";  // Keep 0 if it's meaningful
            if (Array.isArray(value) && value.length === 0) return "[array]";
            if (typeof value === 'object' && Object.keys(value).length === 0) return "[object]";
            return value;
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
    });
</script>
{{end}}
`
