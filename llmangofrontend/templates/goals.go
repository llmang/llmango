package templates

// GoalsTemplates contains the goals and goal-detail templates
const GoalsTemplates = `
{{define "goals"}}
{{template "header" .}}
    <h2>Goals</h2>
    <div class="card-container">
        {{range $id, $goalAny := .Goals}}
            {{$goal := getGoalInfo $goalAny}}
            {{template "goal-card" dict "ID" $id "Goal" $goal "BaseRoute" $.BaseRoute}}
        {{else}}
            <p>No goals available</p>
        {{end}}
    </div>
    
    {{template "card-styles"}}
    {{template "json-formatter"}}
{{template "footer"}}
{{end}}

{{define "goal-detail"}}
{{template "header" .}}
    {{$goal := getGoalInfo .Goal}}
    <h2>Goal: {{$goal.Title}}</h2>
    
    <div class="card">
        <h3>Details</h3>
        <p><strong>Description:</strong> {{$goal.Description}}</p>
        
        <div class="examples-container">
            {{if $goal.ExampleInput}}
            <div class="example-section">
                <h3>Example Input</h3>
                <div class="json-viewer" id="exampleInput">{{$goal.ExampleInput}}</div>
            </div>
            {{end}}
            
            {{if $goal.ExampleOutput}}
            <div class="example-section">
                <h3>Example Output</h3>
                <div class="json-viewer" id="exampleOutput">{{$goal.ExampleOutput}}</div>
            </div>
            {{end}}
        </div>
        
        <h3>Solutions</h3>
        <div 
            x-data="{ 
                showNewSolutionModal: false,
                showNewPromptModal: false,
                showEditSolutionModal: false,
                currentSolutionID: '',
                newSolution: { 
                    promptUid: '', 
                    weight: 1, 
                    isCanary: false, 
                    maxRuns: 10 
                },
                editSolution: {
                    promptUid: '',
                    weight: 1,
                    isCanary: false,
                    maxRuns: 10
                },
                newPrompt: {
                    uid: '',
                    model: '',
                    parameters: {},
                    messages: []
                },
                showPromptSearch: false,
                promptSearch: '',
                showEditPromptSearch: false,
                editPromptSearch: '',
                openEditModal(solutionID, solution) {
                    this.currentSolutionID = solutionID;
                    this.editSolution.promptUid = solution.promptUID;
                    this.editSolution.weight = solution.weight;
                    this.editSolution.isCanary = solution.isCanary;
                    this.editSolution.maxRuns = solution.maxRuns;
                    this.showEditSolutionModal = true;
                }
            }"
        >
            <div class="solutions-container" style="display: flex; flex-wrap: wrap; gap: 15px; margin-bottom: 20px;">
                {{range $solutionID, $solution := $goal.Solutions}}
                <div 
                    class="solution-card" 
                    style="
                        width: 180px; 
                        padding: 15px; 
                        border: 1px solid #ddd; 
                        border-radius: 5px; 
                        position: relative;
                        cursor: pointer;
                        {{if eq (solutionStatus $solution) "ON"}}
                            background: #e6ffe6;
                        {{else if eq (solutionStatus $solution) "OFF"}}
                            background: #f5f5f5;
                        {{else if eq (solutionStatus $solution) "COMPLETE"}}
                            background: #f0f0ff;
                        {{else if eq (solutionStatus $solution) "RUNNING"}}
                            background: #fffde6;
                        {{end}}
                    "
                    @click="openEditModal('{{$solutionID}}', {{toJSON $solution}})"
                >
                    <div 
                        class="status-indicator" 
                        style="
                            position: absolute;
                            top: 10px;
                            right: 10px;
                            width: 12px;
                            height: 12px;
                            border-radius: 50%;
                            {{if eq (solutionStatus $solution) "ON"}}
                                background: #00cc00;
                            {{else if eq (solutionStatus $solution) "OFF"}}
                                background: #999999;
                            {{else if eq (solutionStatus $solution) "COMPLETE"}}
                                background: #3333cc;
                            {{else if eq (solutionStatus $solution) "RUNNING"}}
                                background: #cccc00;
                            {{end}}
                        "
                    ></div>
                    <h4 style="margin-top: 0; margin-bottom: 10px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;" title="{{$solution.PromptUID}}">
                        {{if $solution.PromptUID}}
                            {{$solution.PromptUID}}
                        {{else}}
                            No Prompt
                        {{end}}
                    </h4>
                    <div style="font-size: 14px;">
                        <div><strong>Status:</strong> {{solutionStatus $solution}}</div>
                        <div><strong>Weight:</strong> {{$solution.Weight}}</div>
                        {{if $solution.IsCanary}}
                            <div><strong>Runs:</strong> {{$solution.TotalRuns}}/{{$solution.MaxRuns}}</div>
                        {{end}}
                    </div>
                </div>
                {{end}}
                
                <!-- Add New Solution Button -->
                <div 
                    class="solution-card add-new" 
                    style="
                        width: 180px; 
                        height: 140px;
                        padding: 15px; 
                        border: 1px dashed #aaa; 
                        border-radius: 5px;
                        display: flex;
                        justify-content: center;
                        align-items: center;
                        cursor: pointer;
                        background: #f9f9f9;
                    "
                    @click="showNewSolutionModal = true"
                >
                    <div style="text-align: center;">
                        <div style="font-size: 24px; margin-bottom: 10px;">+</div>
                        <div>Add new solution</div>
                    </div>
                </div>
            </div>
            
            <!-- Add New Solution Modal -->
            <div 
                x-show="showNewSolutionModal" 
                class="modal-overlay"
            >
                <div 
                    class="modal-container"
                    @click.outside="showNewSolutionModal = false"
                >
                    <h3 class="modal-header">Add New Solution</h3>
                    <form>
                        <div class="form-group">
                            <label class="form-label">Prompt</label>
                            <div style="position: relative;">
                                <div style="display: flex; gap: 10px; align-items: center;">
                                    <select 
                                        x-model="newSolution.promptUid" 
                                        class="form-control"
                                    >
                                        <option value="">-- Select a prompt --</option>
                                        {{range $promptID, $prompt := $.Prompts}}
                                        <option value="{{$promptID}}">{{if $prompt.UID}}{{$prompt.UID}}{{else}}{{$promptID}}{{end}}</option>
                                        {{end}}
                                    </select>
                                    <button 
                                        type="button" 
                                        class="btn btn-secondary" 
                                        @click="showPromptSearch = !showPromptSearch"
                                        x-text="showPromptSearch ? 'Hide Search' : 'Search'"
                                    ></button>
                                </div>
                                
                                <div x-show="showPromptSearch" style="position: absolute; top: 100%; left: 0; width: 100%; background: white; border: 1px solid #ddd; border-radius: 4px; z-index: 100; margin-top: 5px; max-height: 300px; overflow-y: auto; padding: 10px; box-shadow: 0 4px 8px rgba(0,0,0,0.1);">
                                    <div class="form-group" style="margin-bottom: 10px;">
                                        <input 
                                            type="text" 
                                            x-model="promptSearch" 
                                            placeholder="Search prompts..." 
                                            class="form-control"
                                            style="width: 100%;"
                                        >
                                    </div>
                                    
                                    <div style="display: flex; flex-direction: column; gap: 5px;">
                                        {{range $promptID, $prompt := $.Prompts}}
                                        <div 
                                            x-show="!promptSearch || '{{$promptID}}'.toLowerCase().includes(promptSearch.toLowerCase()) || '{{$prompt.UID}}'.toLowerCase().includes(promptSearch.toLowerCase()) || {{if $prompt.Messages}}{{if gt (len $prompt.Messages) 0}}'{{index $prompt.Messages 0 | getMessageContent}}'.toLowerCase().includes(promptSearch.toLowerCase()){{else}}false{{end}}{{else}}false{{end}}"
                                            @click="newSolution.promptUid = '{{$promptID}}'; showPromptSearch = false" 
                                            style="cursor: pointer; padding: 5px; border-radius: 3px;"
                                            x-bind:class="{ 'bg-light': newSolution.promptUid === '{{$promptID}}' }"
                                            class="hover:bg-light"
                                        >
                                            <div style="display: flex; justify-content: space-between; align-items: center;">
                                                <div style="font-weight: bold;">{{if $prompt.UID}}{{$prompt.UID}}{{else}}{{$promptID}}{{end}}</div>
                                                <small>{{if $prompt.Model}}{{$prompt.Model}}{{end}}</small>
                                            </div>
                                            <div style="font-size: 0.8rem;">
                                                <code>{{$promptID}}</code>
                                            </div>
                                            {{if $prompt.Messages}}
                                            <div style="font-size: 0.8rem; margin-top: 5px; max-height: 3.6em; overflow: hidden; text-overflow: ellipsis; display: -webkit-box; -webkit-line-clamp: 3; -webkit-box-orient: vertical; background: #f5f5f5; padding: 5px; border-radius: 3px;">
                                                {{if gt (len $prompt.Messages) 0}}
                                                    {{index $prompt.Messages 0 | getMessageContent}}
                                                {{else}}
                                                    No messages
                                                {{end}}
                                            </div>
                                            {{end}}
                                        </div>
                                        {{end}}
                                    </div>
                                </div>
                            </div>
                            <button 
                                type="button" 
                                @click="showNewPromptModal = true; showNewSolutionModal = false" 
                                class="btn btn-secondary"
                                style="margin-top: 5px; font-size: 12px;"
                            >
                                Create new prompt
                            </button>
                        </div>
                        
                        <div class="form-group">
                            <label class="form-label">Weight</label>
                            <input 
                                type="number" 
                                x-model="newSolution.weight" 
                                min="0" 
                                class="form-control"
                            >
                        </div>
                        
                        <div class="form-group">
                            <label class="form-check">
                                <input type="checkbox" x-model="newSolution.isCanary">
                                <span class="form-check-label">Is Canary Test</span>
                            </label>
                        </div>
                        
                        <div x-show="newSolution.isCanary" class="form-group">
                            <label class="form-label">Max Runs</label>
                            <input 
                                type="number" 
                                x-model="newSolution.maxRuns" 
                                min="1" 
                                class="form-control"
                            >
                        </div>
                        
                        <div class="modal-footer">
                            <button 
                                type="button" 
                                @click="showNewSolutionModal = false" 
                                class="btn btn-secondary"
                            >
                                Cancel
                            </button>
                            <button 
                                type="button"
                                class="btn btn-primary"
                                @click="
                                    // Call API to create solution
                                    fetch('{{$.BaseRoute}}/api/solutions/new', {
                                        method: 'POST',
                                        headers: { 'Content-Type': 'application/json' },
                                        body: JSON.stringify({
                                            goalId: '{{.GoalID}}',
                                            promptUid: newSolution.promptUid,
                                            weight: parseInt(newSolution.weight) || 0,
                                            isCanary: newSolution.isCanary,
                                            maxRuns: parseInt(newSolution.maxRuns) || 10
                                        })
                                    })
                                    .then(response => response.json())
                                    .then(data => {
                                        if (data.success) {
                                            // Reload page on success
                                            window.location.reload();
                                        } else {
                                            alert('Error: ' + data.error);
                                        }
                                    })
                                    .catch(error => {
                                        alert('Error: ' + error);
                                    });
                                "
                            >
                                Create Solution
                            </button>
                        </div>
                    </form>
                </div>
            </div>
            
            <!-- Add New Prompt Modal -->
            <div 
                x-show="showNewPromptModal" 
                class="modal-overlay"
            >
                <div 
                    class="modal-container"
                    @click.outside="showNewPromptModal = false"
                >
                    <h3 class="modal-header">Create New Prompt</h3>
                    <form>
                        {{template "prompt-form-template" dict "GoalInfo" $goal}}
                        
                        <div class="modal-footer">
                            <button 
                                type="button" 
                                @click="showNewPromptModal = false; showNewSolutionModal = true" 
                                class="btn btn-secondary"
                            >
                                Back to Solution
                            </button>
                            <button 
                                type="button"
                                class="btn btn-primary"
                                @click="
                                    // Call API to create prompt
                                    fetch('{{$.BaseRoute}}/api/prompts/new', {
                                        method: 'POST',
                                        headers: { 'Content-Type': 'application/json' },
                                        body: JSON.stringify({
                                            uid: newPrompt.uid,
                                            model: newPrompt.model,
                                            parameters: newPrompt.parameters || {},
                                            messages: newPrompt.messages || []
                                        })
                                    })
                                    .then(response => response.json())
                                    .then(data => {
                                        if (data.success) {
                                            // Set the new prompt as selected in solution form
                                            newSolution.promptUid = data.data.promptId;
                                            // Go back to solution form
                                            showNewPromptModal = false;
                                            showNewSolutionModal = true;
                                        } else {
                                            alert('Error: ' + data.error);
                                        }
                                    })
                                    .catch(error => {
                                        alert('Error: ' + error);
                                    });
                                "
                            >
                                Create Prompt
                            </button>
                        </div>
                    </form>
                </div>
            </div>

            <!-- Edit Solution Modal -->
            <div 
                x-show="showEditSolutionModal" 
                class="modal-overlay"
            >
                <div 
                    class="modal-container"
                    @click.outside="showEditSolutionModal = false"
                >
                    <h3 class="modal-header">Edit Solution</h3>
                    <form>
                        <div class="form-group">
                            <label class="form-label">Prompt</label>
                            <div style="position: relative;">
                                <div style="display: flex; gap: 10px; align-items: center;">
                                    <select 
                                        x-model="editSolution.promptUid" 
                                        class="form-control"
                                    >
                                        <option value="">-- Select a prompt --</option>
                                        {{range $promptID, $prompt := $.Prompts}}
                                        <option value="{{$promptID}}">{{if $prompt.UID}}{{$prompt.UID}}{{else}}{{$promptID}}{{end}}</option>
                                        {{end}}
                                    </select>
                                    <button 
                                        type="button" 
                                        class="btn btn-secondary" 
                                        @click="showEditPromptSearch = !showEditPromptSearch"
                                        x-text="showEditPromptSearch ? 'Hide Search' : 'Search'"
                                    ></button>
                                </div>
                                
                                <div x-show="showEditPromptSearch" style="position: absolute; top: 100%; left: 0; width: 100%; background: white; border: 1px solid #ddd; border-radius: 4px; z-index: 100; margin-top: 5px; max-height: 300px; overflow-y: auto; padding: 10px; box-shadow: 0 4px 8px rgba(0,0,0,0.1);">
                                    <div class="form-group" style="margin-bottom: 10px;">
                                        <input 
                                            type="text" 
                                            x-model="editPromptSearch" 
                                            placeholder="Search prompts..." 
                                            class="form-control"
                                            style="width: 100%;"
                                        >
                                    </div>
                                    
                                    <div style="display: flex; flex-direction: column; gap: 5px;">
                                        {{range $promptID, $prompt := $.Prompts}}
                                        <div 
                                            x-show="!editPromptSearch || '{{$promptID}}'.toLowerCase().includes(editPromptSearch.toLowerCase()) || '{{$prompt.UID}}'.toLowerCase().includes(editPromptSearch.toLowerCase()) || {{if $prompt.Messages}}{{if gt (len $prompt.Messages) 0}}'{{index $prompt.Messages 0 | getMessageContent}}'.toLowerCase().includes(editPromptSearch.toLowerCase()){{else}}false{{end}}{{else}}false{{end}}"
                                            @click="editSolution.promptUid = '{{$promptID}}'; showEditPromptSearch = false" 
                                            style="cursor: pointer; padding: 5px; border-radius: 3px;"
                                            x-bind:class="{ 'bg-light': editSolution.promptUid === '{{$promptID}}' }"
                                            class="hover:bg-light"
                                        >
                                            <div style="display: flex; justify-content: space-between; align-items: center;">
                                                <div style="font-weight: bold;">{{if $prompt.UID}}{{$prompt.UID}}{{else}}{{$promptID}}{{end}}</div>
                                                <small>{{if $prompt.Model}}{{$prompt.Model}}{{end}}</small>
                                            </div>
                                            <div style="font-size: 0.8rem;">
                                                <code>{{$promptID}}</code>
                                            </div>
                                            {{if $prompt.Messages}}
                                            <div style="font-size: 0.8rem; margin-top: 5px; max-height: 3.6em; overflow: hidden; text-overflow: ellipsis; display: -webkit-box; -webkit-line-clamp: 3; -webkit-box-orient: vertical; background: #f5f5f5; padding: 5px; border-radius: 3px;">
                                                {{if gt (len $prompt.Messages) 0}}
                                                    {{index $prompt.Messages 0 | getMessageContent}}
                                                {{else}}
                                                    No messages
                                                {{end}}
                                            </div>
                                            {{end}}
                                        </div>
                                        {{end}}
                                    </div>
                                </div>
                            </div>
                        </div>
                        
                        <div class="form-group">
                            <label class="form-label">Weight</label>
                            <input 
                                type="number" 
                                x-model="editSolution.weight" 
                                min="0" 
                                class="form-control"
                            >
                        </div>
                        
                        <div class="form-group">
                            <label class="form-check">
                                <input type="checkbox" x-model="editSolution.isCanary">
                                <span class="form-check-label">Is Canary Test</span>
                            </label>
                        </div>
                        
                        <div x-show="editSolution.isCanary" class="form-group">
                            <label class="form-label">Max Runs</label>
                            <input 
                                type="number" 
                                x-model="editSolution.maxRuns" 
                                min="1" 
                                class="form-control"
                            >
                        </div>
                        
                        <div class="modal-footer">
                            <button 
                                type="button" 
                                @click="showEditSolutionModal = false" 
                                class="btn btn-secondary"
                            >
                                Cancel
                            </button>
                            <button 
                                type="button"
                                class="btn btn-danger"
                                style="margin-right: auto;"
                                @click="
                                    if (confirm('Are you sure you want to delete this solution?')) {
                                        fetch('{{$.BaseRoute}}/api/solutions/' + currentSolutionID + '/delete', {
                                            method: 'POST',
                                            headers: { 'Content-Type': 'application/json' },
                                            body: JSON.stringify({
                                                goalId: '{{.GoalID}}'
                                            })
                                        })
                                        .then(response => response.json())
                                        .then(data => {
                                            if (data.success) {
                                                window.location.reload();
                                            } else {
                                                alert('Error: ' + data.error);
                                            }
                                        })
                                        .catch(error => {
                                            alert('Error: ' + error);
                                        });
                                    }
                                "
                            >
                                Delete
                            </button>
                            <button 
                                type="button"
                                class="btn btn-primary"
                                @click="
                                    // Call API to update solution
                                    fetch('{{$.BaseRoute}}/api/solutions/' + currentSolutionID + '/update', {
                                        method: 'POST',
                                        headers: { 'Content-Type': 'application/json' },
                                        body: JSON.stringify({
                                            goalId: '{{.GoalID}}',
                                            promptUid: editSolution.promptUid,
                                            weight: parseInt(editSolution.weight) || 0,
                                            isCanary: editSolution.isCanary,
                                            maxRuns: parseInt(editSolution.maxRuns) || 10
                                        })
                                    })
                                    .then(response => response.json())
                                    .then(data => {
                                        if (data.success) {
                                            // Reload page on success
                                            window.location.reload();
                                        } else {
                                            alert('Error: ' + data.error);
                                        }
                                    })
                                    .catch(error => {
                                        alert('Error: ' + error);
                                    });
                                "
                            >
                                Update Solution
                            </button>
                        </div>
                    </form>
                </div>
            </div>
        </div>
    </div>

    <style>
    .json-viewer {
        background-color: #f8f8f8;
        border-radius: 4px;
        padding: 15px;
        margin-bottom: 20px;
        border: 1px solid #ddd;
        max-height: 300px;
        overflow: auto;
        font-family: monospace;
        position: relative;
    }
    
    .examples-container {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 20px;
        margin: 20px 0 30px 0;
    }
    
    @media (max-width: 768px) {
        .examples-container {
            grid-template-columns: 1fr;
        }
    }
    
    .example-section {
        margin-top: 0;
    }
    
    .example-section h3 {
        margin-top: 0;
        margin-bottom: 10px;
        font-size: 1.1rem;
    }
    
    .goal-debug {
        margin-top: 30px;
        border-top: 1px solid #eee;
        padding-top: 0.5rem;
    }
    
    .goal-debug summary {
        cursor: pointer;
        color: #666;
        font-size: 0.9rem;
    }
    
    .goal-debug pre {
        background: #f5f5f5;
        padding: 10px;
        border-radius: 4px;
        overflow: auto;
        font-size: 0.8rem;
    }
    
    /* JSON syntax highlighting */
    .string { color: #008000; }
    .number { color: #0000ff; }
    .boolean { color: #b22222; }
    .null { color: #808080; }
    .key { color: #a52a2a; font-weight: bold; }
    </style>

    <!-- Add JSON formatter script -->
    <script>
    document.addEventListener('DOMContentLoaded', function() {
        // Format example input JSON if available
        {{if $goal.ExampleInput}}
        try {
            const inputData = JSON.parse({{$goal.ExampleInput}});
            document.getElementById('exampleInput').innerHTML = 
                formatJSON(inputData);
        } catch(e) {
            document.getElementById('exampleInput').innerHTML = 
                "<pre>" + {{$goal.ExampleInput}} + "</pre>";
        }
        {{end}}
        
        // Format example output JSON if available
        {{if $goal.ExampleOutput}}
        try {
            const outputData = JSON.parse({{$goal.ExampleOutput}});
            document.getElementById('exampleOutput').innerHTML = 
                formatJSON(outputData);
        } catch(e) {
            document.getElementById('exampleOutput').innerHTML = 
                "<pre>" + {{$goal.ExampleOutput}} + "</pre>";
        }
        {{end}}
        
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
{{template "footer"}}
{{end}}
`
