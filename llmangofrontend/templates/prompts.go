package templates

// PromptsTemplates contains the prompts and prompt-detail templates
const PromptsTemplates = `
{{define "prompts"}}
{{template "header" .}}
    <h2>Prompts</h2>
    
    <!-- Modal container at the root level -->
    <div
        x-data="{ 
            showNewPromptModal: false,
            newPrompt: {
                uid: '',
                model: '',
                parameters: {
                    temperature: 0.7,
                    max_tokens: 1000,
                    top_p: 1.0,
                    frequency_penalty: 0.0,
                    presence_penalty: 0.0
                },
                messages: []
            },
            addMessage() {
                this.newPrompt.messages.push({ role: 'user', content: '' });
            },
            removeMessage(index) {
                this.newPrompt.messages.splice(index, 1);
            },
            submitPrompt() {
                // Format parameters properly before sending
                const formattedPrompt = { ...this.newPrompt };
                
                // Convert string values to appropriate types
                if (formattedPrompt.parameters) {
                    if (formattedPrompt.parameters.temperature) 
                        formattedPrompt.parameters.temperature = parseFloat(formattedPrompt.parameters.temperature);
                    if (formattedPrompt.parameters.max_tokens) 
                        formattedPrompt.parameters.max_tokens = parseInt(formattedPrompt.parameters.max_tokens);
                    if (formattedPrompt.parameters.top_p) 
                        formattedPrompt.parameters.top_p = parseFloat(formattedPrompt.parameters.top_p);
                    if (formattedPrompt.parameters.frequency_penalty) 
                        formattedPrompt.parameters.frequency_penalty = parseFloat(formattedPrompt.parameters.frequency_penalty);
                    if (formattedPrompt.parameters.presence_penalty) 
                        formattedPrompt.parameters.presence_penalty = parseFloat(formattedPrompt.parameters.presence_penalty);
                }
                
                // Call API to create prompt
                fetch('{{.BaseRoute}}/api/prompts/new', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(formattedPrompt)
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
            }
        }"
    >
        <div class="card-container">
            <!-- Add New Prompt Button Card -->
            <div class="card new-card clickable"
                @click="showNewPromptModal = true">
                <div class="new-prompt-content">
                    <div class="plus-icon">+</div>
                    <div>New Prompt</div>
                </div>
            </div>
            
            <!-- Existing Prompts -->
            {{range $id, $prompt := .Prompts}}
                {{template "card" dict "ID" $id "Prompt" $prompt "BaseRoute" $.BaseRoute}}
            {{else}}
                <p>No prompts available</p>
            {{end}}
        </div>

        <!-- New Prompt Modal -->
        <div 
            x-show="showNewPromptModal" 
            class="modal-overlay"
        >
            <div class="modal-container"
                @click.outside="showNewPromptModal = false"
            >
                <h3 class="modal-header">Create New Prompt</h3>
                <form>
                    <div class="form-group">
                        <label for="model">Model</label>
                        <input 
                            type="text" 
                            id="model" 
                            x-model="newPrompt.model" 
                            class="form-control"
                            required
                        >
                    </div>

                    <div class="form-group">
                        <label>Parameters</label>
                        <div class="parameters-grid">
                            <div class="parameter-item">
                                <label for="temperature">Temperature</label>
                                <input 
                                    type="number" 
                                    id="temperature" 
                                    x-model="newPrompt.parameters.temperature" 
                                    step="0.1" 
                                    min="0" 
                                    max="2"
                                    class="form-control"
                                >
                            </div>
                            <div class="parameter-item">
                                <label for="max_tokens">Max Tokens</label>
                                <input 
                                    type="number" 
                                    id="max_tokens" 
                                    x-model="newPrompt.parameters.max_tokens" 
                                    min="1"
                                    class="form-control"
                                >
                            </div>
                            <div class="parameter-item">
                                <label for="top_p">Top P</label>
                                <input 
                                    type="number" 
                                    id="top_p" 
                                    x-model="newPrompt.parameters.top_p" 
                                    step="0.1" 
                                    min="0" 
                                    max="1"
                                    class="form-control"
                                >
                            </div>
                            <div class="parameter-item">
                                <label for="frequency_penalty">Frequency Penalty</label>
                                <input 
                                    type="number" 
                                    id="frequency_penalty" 
                                    x-model="newPrompt.parameters.frequency_penalty" 
                                    step="0.1" 
                                    min="-2" 
                                    max="2"
                                    class="form-control"
                                >
                            </div>
                            <div class="parameter-item">
                                <label for="presence_penalty">Presence Penalty</label>
                                <input 
                                    type="number" 
                                    id="presence_penalty" 
                                    x-model="newPrompt.parameters.presence_penalty" 
                                    step="0.1" 
                                    min="-2" 
                                    max="2"
                                    class="form-control"
                                >
                            </div>
                        </div>
                    </div>

                    <div class="form-group">
                        <label>Messages</label>
                        <div class="messages-container">
                            <template x-for="(message, index) in newPrompt.messages" :key="index">
                                <div class="message-item">
                                    <select 
                                        x-model="message.role" 
                                        class="form-control"
                                    >
                                        <option value="system">System</option>
                                        <option value="user">User</option>
                                        <option value="assistant">Assistant</option>
                                    </select>
                                    <textarea 
                                        x-model="message.content" 
                                        class="form-control"
                                        rows="3"
                                    ></textarea>
                                    <button 
                                        type="button" 
                                        @click="removeMessage(index)"
                                        class="btn btn-danger"
                                    >
                                        Remove
                                    </button>
                                </div>
                            </template>
                            <button 
                                type="button" 
                                @click="addMessage()"
                                class="btn btn-secondary"
                            >
                                Add Message
                            </button>
                        </div>
                    </div>
                    
                    <div class="modal-footer">
                        <button 
                            type="button" 
                            @click="showNewPromptModal = false" 
                            class="btn btn-secondary"
                        >
                            Cancel
                        </button>
                        <button 
                            type="button"
                            class="btn btn-primary"
                            @click="submitPrompt()"
                        >
                            Create Prompt
                        </button>
                    </div>
                </form>
            </div>
        </div>
    </div>

    {{template "card-styles"}}
    {{template "json-formatter"}}
{{template "footer"}}
{{end}}

{{define "card"}}
<div class="card card clickable" onclick="window.location.href='{{.BaseRoute}}/prompt/{{.ID}}'">
    <h3 class="prompt-title">{{.Prompt.UID}}</h3>
    <div class="prompt-meta">
        <span class="model-badge small">{{.Prompt.Model}}</span>
        <div class="message-count">
            <span>{{len .Prompt.Messages}} messages</span>
        </div>
    </div>
    {{if gt (len .Prompt.Messages) 0}}
    <div class="message-preview">
        <div class="preview-label">First message:</div>
        <div class="preview-content">{{(index .Prompt.Messages 0).Content}}</div>
    </div>
    {{end}}
</div>
{{end}}

{{define "prompt-detail"}}
{{template "header" .}}
    <!-- Add JavaScript functions -->
    <script>
    function openModal(modalId) {
        const modal = document.getElementById(modalId);
        if (modal) {
            modal.style.display = 'flex';
        } else {
            console.error('Modal not found:', modalId);
        }
    }

    function closeModal(modalId) {
        const modal = document.getElementById(modalId);
        if (modal) {
            modal.style.display = 'none';
        }
    }

    function editPrompt(promptUID) {
        openModal('warningConfirmModal');
    }

    function proceedWithEdit() {
        closeModal('warningConfirmModal');
        
        // Set form values from the parsed JSON
        document.getElementById('edit-model').value = promptData.model || '';
        
        // Set parameter values if they exist
        if (promptData.parameters) {
            document.getElementById('edit-temperature').value = promptData.parameters.temperature || '';
            document.getElementById('edit-max_tokens').value = promptData.parameters.max_tokens || '';
            document.getElementById('edit-top_p').value = promptData.parameters.top_p || '';
            document.getElementById('edit-frequency_penalty').value = promptData.parameters.frequency_penalty || '';
            document.getElementById('edit-presence_penalty').value = promptData.parameters.presence_penalty || '';
        }
        
        openModal('editPromptModal');
    }

    function confirmUpdate(promptUID) {
        if (confirm('⚠️ Are you sure you want to update this prompt? This action cannot be undone and may affect existing solutions.')) {
            submitEditPrompt(promptUID);
        }
    }

    function submitEditPrompt(promptUID) {
        // Get form values with safe null handling
        const model = document.getElementById('edit-model').value;
        
        // Create parameters object with safe null handling
        const parameters = {};
        
        // Helper function to safely get number values
        const getNumberValue = (id) => {
            const value = document.getElementById(id).value;
            return value === '' ? null : parseFloat(value);
        };

        // Get parameter values
        const temperature = getNumberValue('edit-temperature');
        const maxTokens = getNumberValue('edit-max_tokens');
        const topP = getNumberValue('edit-top_p');
        const frequencyPenalty = getNumberValue('edit-frequency_penalty');
        const presencePenalty = getNumberValue('edit-presence_penalty');

        // Only add parameters if they have values
        if (temperature !== null) parameters.temperature = temperature;
        if (maxTokens !== null) parameters.max_tokens = maxTokens;
        if (topP !== null) parameters.top_p = topP;
        if (frequencyPenalty !== null) parameters.frequency_penalty = frequencyPenalty;
        if (presencePenalty !== null) parameters.presence_penalty = presencePenalty;

        // Create prompt object
        const prompt = {
            model: model,
            parameters: parameters
        };


        // Call API to update prompt
        fetch('{{.BaseRoute}}/api/prompts/' + promptUID + '/update', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(prompt)
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
    }

    // Store the prompt data as a JSON string in a data attribute
    const promptData = JSON.parse('{{toJSON .Prompt}}');
    </script>

    <h2>Prompt: {{.Prompt.UID}}</h2>

    <div class="card">
        <div class="card-header">
            <h3>Details</h3>
        </div>
        <div class="model-badge">{{.Prompt.Model}}</div>
        <h4>Parameters:</h4>
        <div class="json-preview parameters-json">{{toJSON .Prompt.Parameters}}</div>
        
        <h4>Messages:</h4>
        {{range $index, $message := .Prompt.Messages}}
        <div class="message-card">
            <p class="message-role">{{$message.Role}}</p>
            <pre class="message-content">{{$message.Content}}</pre>
        </div>
        {{end}}
        
        <!-- Debug info -->
        <details class="prompt-debug">
            <summary>Debug Info</summary>
            <pre>{{printf "%#v" .Prompt}}</pre>
        </details>

        <!-- Unsafe Actions -->
        <details class="unsafe-actions">
            <summary>⚠️ Unsafe Actions</summary>
            <div class="unsafe-actions-content">
                <p class="warning-text">These actions may affect data consistency and cannot be undone. Use with caution.</p>
                <div class="unsafe-buttons">
                    <button 
                        class="btn btn-warning" 
                        onclick="editPrompt('{{.PromptUID}}')"
                    >
                        ⚠️ Edit Prompt
                    </button>
                    <button 
                        class="btn btn-danger" 
                        onclick="if(confirm('⚠️ Are you sure you want to delete this prompt? This action cannot be undone and may affect existing solutions.')) { deletePrompt('{{.PromptUID}}') }"
                    >
                        ⚠️ Delete Prompt
                    </button>
                </div>
            </div>
        </details>

        <h3>Recent Logs</h3>
        {{template "log-viewer" dict "BaseRoute" .BaseRoute "FilterOptions" (dict "promptUID" .Prompt.UID)}}
    </div>

    <!-- Edit Prompt Modal -->
    <div 
        id="editPromptModal"
        class="modal-overlay"
        style="display: none;"
    >
        <div class="modal-container">
            <div class="modal-header">
                <h3>⚠️ Edit Prompt</h3>
                <button 
                    class="close-button"
                    onclick="closeModal('editPromptModal')"
                >
                    ×
                </button>
            </div>
            
            <div class="warning-section">
                <h4>⚠️ Warning: Data Consistency Risk</h4>
                <p>Editing this prompt may affect existing solutions and goals that use it. Please ensure you understand the implications before proceeding.</p>
            </div>

            <div class="form-group">
                <label for="edit-model">Model</label>
                <input 
                    type="text" 
                    id="edit-model" 
                    class="form-control"
                >
            </div>

            <div class="form-group">
                <label>Parameters</label>
                <div class="parameters-grid">
                    <div class="parameter-item">
                        <label for="edit-temperature">Temperature</label>
                        <input 
                            type="number" 
                            id="edit-temperature" 
                            class="form-control"
                            step="0.1" 
                            min="0" 
                            max="2"
                        >
                    </div>
                    <div class="parameter-item">
                        <label for="edit-max_tokens">Max Tokens</label>
                        <input 
                            type="number" 
                            id="edit-max_tokens" 
                            class="form-control"
                            min="1"
                        >
                    </div>
                    <div class="parameter-item">
                        <label for="edit-top_p">Top P</label>
                        <input 
                            type="number" 
                            id="edit-top_p" 
                            class="form-control"
                            step="0.1" 
                            min="0" 
                            max="1"
                        >
                    </div>
                    <div class="parameter-item">
                        <label for="edit-frequency_penalty">Frequency Penalty</label>
                        <input 
                            type="number" 
                            id="edit-frequency_penalty" 
                            class="form-control"
                            step="0.1" 
                            min="-2" 
                            max="2"
                        >
                    </div>
                    <div class="parameter-item">
                        <label for="edit-presence_penalty">Presence Penalty</label>
                        <input 
                            type="number" 
                            id="edit-presence_penalty" 
                            class="form-control"
                            step="0.1" 
                            min="-2" 
                            max="2"
                        >
                    </div>
                </div>
            </div>

            <div class="modal-footer">
                <button 
                    class="btn btn-secondary"
                    onclick="closeModal('editPromptModal')"
                >
                    Cancel
                </button>
                <button 
                    class="btn btn-warning"
                    onclick="confirmUpdate('{{.PromptUID}}')"
                >
                    ⚠️ Update Prompt
                </button>
            </div>
        </div>
    </div>

    <!-- Warning Confirmation Modal -->
    <div 
        id="warningConfirmModal"
        class="modal-overlay"
        style="display: none;"
    >
        <div class="modal-container">
            <div class="modal-header">
                <h3>⚠️ Warning</h3>
                <button 
                    class="close-button"
                    onclick="closeModal('warningConfirmModal')"
                >
                    ×
                </button>
            </div>
            
            <div class="warning-section">
                <h4>⚠️ Data Consistency Risk</h4>
                <p>Editing this prompt may affect existing solutions and goals that use it. This action cannot be undone.</p>
                <p>Are you sure you want to proceed?</p>
            </div>

            <div class="modal-footer">
                <button 
                    class="btn btn-secondary"
                    onclick="closeModal('warningConfirmModal')"
                >
                    Cancel
                </button>
                <button 
                    class="btn btn-warning"
                    onclick="proceedWithEdit()"
                >
                    ⚠️ Proceed
                </button>
            </div>
        </div>
    </div>

    <style>
    .card {
        padding: 1.5rem;
        border-radius: 0.5rem;
        border: 1px solid #ddd;
        background-color: #fff;
        box-shadow: 0 2px 4px rgba(0,0,0,0.05);
        margin-bottom: 2rem;
    }

    .card-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 1rem;
    }

    .card-actions {
        display: flex;
        gap: 0.5rem;
    }
    
    .model-badge {
        display: inline-block;
        background-color: #f0f0f0;
        border-radius: 4px;
        padding: 0.3rem 0.6rem;
        font-size: 0.8rem;
        font-weight: 600;
        color: #555;
        margin-bottom: 1rem;
    }
    
    .message-card {
        margin-bottom: 15px;
        padding: 1rem;
        border: 1px solid #ddd;
        border-radius: 0.5rem;
    }
    
    .message-role {
        font-weight: 600;
        margin-top: 0;
        margin-bottom: 0.5rem;
        color: #555;
    }
    
    .message-content {
        margin: 0;
        white-space: pre-wrap;
        overflow-x: auto;
        font-family: monospace;
        font-size: 0.9rem;
        background-color: #f8f8f8;
        padding: 0.75rem;
        border-radius: 0.25rem;
    }
    
    .json-preview {
        background-color: #f5f5f5;
        border-radius: 0.25rem;
        padding: 0.75rem;
        font-family: monospace;
        font-size: 0.8rem;
        max-height: 200px;
        overflow: auto;
        white-space: pre-wrap;
        margin-bottom: 1rem;
        position: relative;
        
        /* Custom syntax highlighting */
        color: #333;
    }
    
    .prompt-debug {
        margin-top: 2rem;
        border-top: 1px solid #eee;
        padding-top: 0.5rem;
    }
    
    .prompt-debug summary {
        cursor: pointer;
        color: #666;
        font-size: 0.8rem;
    }
    
    .prompt-debug pre {
        background: #f5f5f5;
        padding: 0.5rem;
        border-radius: 0.25rem;
        overflow: auto;
    }

    .warning-section {
        background-color: #fff3cd;
        border: 1px solid #ffeeba;
        border-radius: 0.5rem;
        padding: 1.5rem;
        margin-bottom: 1rem;
    }

    .warning-icon {
        font-size: 2rem;
        margin-bottom: 1rem;
    }

    .warning-section h4 {
        color: #856404;
        margin-top: 0;
    }

    .warning-section p {
        color: #856404;
        margin-bottom: 1rem;
    }

    .warning-actions {
        display: flex;
        gap: 0.5rem;
        justify-content: flex-end;
    }

    .form-group {
        margin-bottom: 1rem;
    }

    .form-control {
        width: 100%;
        padding: 0.5rem;
        border: 1px solid #ddd;
        border-radius: 0.25rem;
        font-size: 1rem;
    }

    .parameters-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
        gap: 1rem;
    }

    .parameter-item {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
    }

    .messages-container {
        display: flex;
        flex-direction: column;
        gap: 1rem;
    }

    .message-item {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
        padding: 1rem;
        border: 1px solid #ddd;
        border-radius: 0.25rem;
    }

    .btn {
        padding: 0.5rem 1rem;
        border-radius: 0.25rem;
        border: none;
        cursor: pointer;
        font-size: 1rem;
    }

    .btn-primary {
        background-color: #007bff;
        color: white;
    }

    .btn-secondary {
        background-color: #6c757d;
        color: white;
    }

    .btn-danger {
        background-color: #dc3545;
        color: white;
    }

    .btn-warning {
        background-color: #ffc107;
        color: #000;
    }

    .btn-warning:hover {
        background-color: #e0a800;
    }

    .modal-overlay {
        position: fixed;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        background-color: rgba(0, 0, 0, 0.5);
        display: flex;
        justify-content: center;
        align-items: center;
        z-index: 2000;
    }

    .modal-container {
        background-color: white;
        padding: 20px;
        border-radius: 5px;
        width: 600px;
        max-width: 90%;
        max-height: 90vh;
        overflow-y: auto;
    }

    .modal-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 1rem;
    }

    .close-button {
        background: none;
        border: none;
        font-size: 1.5rem;
        cursor: pointer;
        color: #666;
        padding: 0;
        line-height: 1;
    }

    .close-button:hover {
        color: #333;
    }

    .unsafe-actions {
        margin-top: 1rem;
        border-top: 1px solid #eee;
        padding-top: 0.5rem;
    }
    
    .unsafe-actions summary {
        cursor: pointer;
        color: #dc3545;
        font-size: 0.8rem;
        font-weight: 600;
    }
    
    .unsafe-actions-content {
        padding: 1rem;
        background-color: #fff8f8;
        border: 1px solid #ffd6d6;
        border-radius: 0.25rem;
        margin-top: 0.5rem;
    }
    
    .warning-text {
        color: #dc3545;
        margin-bottom: 1rem;
        font-size: 0.9rem;
    }
    
    .unsafe-buttons {
        display: flex;
        gap: 0.5rem;
    }
    
    .btn-warning {
        background-color: #ffc107;
        color: #000;
    }
    
    .btn-warning:hover {
        background-color: #e0a800;
    }
    
    .btn-danger {
        background-color: #dc3545;
        color: white;
    }
    
    .btn-danger:hover {
        background-color: #c82333;
    }

    .log-table {
        width: 100%;
        border-collapse: collapse;
    }

    .log-header-row {
        display: flex;
        padding: 0.5rem;
        background-color: #f5f5f5;
        font-weight: bold;
        border-bottom: 2px solid #ddd;
    }

    .log-row-container {
        display: flex;
        flex-direction: column;
        border-bottom: 1px solid #eee;
    }

    .log-row {
        display: flex;
        padding: 0.5rem;
        align-items: center;
        font-size: 0.85rem;
    }

    .log-cell {
        flex: 1;
        padding: 0 0.5rem;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }

    .log-cell[title]:hover {
        cursor: help;
    }

    .token-count {
        display: inline-block;
        min-width: 2.5rem;
        text-align: right;
    }

    .log-details {
        width: 100%;
        padding: 1rem;
        background-color: #f8f8f8;
        border-top: 1px solid #ddd;
        font-size: 1rem;
    }

    .details-row {
        display: flex;
        gap: 1rem;
        margin-bottom: 1rem;
    }

    .details-cell {
        flex: 1;
    }

    .details-cell h5 {
        margin: 0 0 0.5rem 0;
        color: #666;
    }

    .details-cell pre {
        background-color: #fff;
        padding: 0.5rem;
        border-radius: 0.25rem;
        overflow-x: auto;
        margin: 0;
        white-space: pre-wrap;
        word-wrap: break-word;
    }

    .log-section {
        margin-bottom: 1rem;
    }

    .log-section:last-child {
        margin-bottom: 0;
    }

    .log-section h5 {
        margin: 0 0 0.5rem 0;
        color: #666;
    }

    .log-section pre {
        background-color: #fff;
        padding: 0.5rem;
        border-radius: 0.25rem;
        overflow-x: auto;
        margin: 0;
        white-space: pre-wrap;
        word-wrap: break-word;
    }

    .log-section.error {
        color: #dc3545;
    }

    .btn-sm {
        padding: 0.25rem 0.5rem;
        font-size: 0.875rem;
    }

    .pagination {
        display: flex;
        justify-content: center;
        align-items: center;
        gap: 1rem;
        margin-top: 1rem;
    }

    .loading {
        text-align: center;
        padding: 1rem;
        color: #666;
    }
    </style>
{{template "footer"}}
{{end}}
`
