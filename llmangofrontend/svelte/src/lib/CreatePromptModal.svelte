<script lang="ts">
    import Modal from './Modal.svelte';
    import type { Prompt, Goal } from './classes/llmangoAPI.svelte';
    import { llmangoAPI, PromptParameters } from './classes/llmangoAPI.svelte';
    import { openrouter, type OpenRouterModel } from './classes/openrouter.svelte';
    import { onMount } from 'svelte';
    import PromptMessageFormatter from './PromptMessageFormatter.svelte';

    let { 
        isOpen, 
        mode = 'create',
        onClose,
        prompt = null,
        promptUID = '',
        goalUID = '',
        onSave = null
    } = $props<{
        isOpen: boolean;
        mode?: 'create' | 'edit';
        onClose: () => void;
        prompt?: Prompt | null;
        promptUID?: string;
        goalUID?: string;
        onSave?: ((updatedPrompt: Prompt) => void) | null;
    }>();

    // Form states
    let editModel = $state('');
    let editParameters = $state<PromptParameters>(new PromptParameters());
    let error = $state<string | null>(null);
    let isSubmitting = $state(false);
    let modelList = $state<OpenRouterModel[]>([]);
    let modelsLoading = $state(true);
    
    // Goal state
    let goal = $state<Goal | null>(null);
    let goalLoading = $state(false);
    
    // Dynamic messages
    type MessageType = 'system' | 'user' | 'assistant';
    interface PromptMessage {
        role: MessageType;
        content: string;
    }
    
    let messages = $state<PromptMessage[]>([
        { role: 'user', content: '' }
    ]);

    // Load OpenRouter models and goal data on mount
    onMount(async () => {
        try {
            modelsLoading = true;
            modelList = await openrouter.load();
            
            // If we have a prompt to edit, fill the form
            if (prompt && mode === 'edit') {
                fillEditForm();
            }
            
            // If we have a goalUID, fetch the goal
            if (goalUID) {
                await fetchGoal();
            }
        } catch (e) {
            console.error("Failed to load initial data:", e);
            error = e instanceof Error ? e.message : 'Failed to load initial data';
        } finally {
            modelsLoading = false;
        }
    });
    
    // Fetch goal by ID
    async function fetchGoal() {
        if (!goalUID) return;
        
        try {
            goalLoading = true;
            goal = await llmangoAPI.getGoal(goalUID);
        } catch (e) {
            console.error("Failed to load goal:", e);
            error = e instanceof Error ? e.message : 'Failed to load goal';
        } finally {
            goalLoading = false;
        }
    }

    // Fill the edit form with current prompt data
    function fillEditForm() {
        if (!prompt) return;
        
        editModel = prompt.model || '';
        editParameters = PromptParameters.fromObject(prompt.parameters);

        // Extract messages
        if (prompt.messages && prompt.messages.length > 0) {
            messages = [...prompt.messages];
        } else {
            // Default to one user message if no messages
            messages = [{ role: 'user', content: '' }];
        }
    }

    // Validate the form before submission
    function validateForm() {
        if (!editModel) {
            error = 'Please select a model';
            return false;
        }
        if (messages.length === 0) {
            const errMessage = 'At least one message is required';
            error = errMessage;
            setTimeout(() => {
                    error = null;
            }, 500);
            return false;
        }
        
        for (const message of messages) {
            if (!message.content.trim()) {
                error = `${message.role.charAt(0).toUpperCase() + message.role.slice(1)} message cannot be empty`;
                return false;
            }
        }
        
        return true;
    }
    
    // Add a new message
    function addMessage(type: MessageType) {
        messages = [...messages, { role: type, content: '' }];
    }
    
    // Remove a message at a specific index
    function removeMessage(index: number) {
        if (messages.length <= 1) {
            error = 'At least one message is required';
            return;
        }
        
        messages = messages.filter((_, i) => i !== index);
    }
    
    // Move message up in the order
    function moveMessageUp(index: number) {
        if (index <= 0) return;
        
        const newMessages = [...messages];
        const temp = newMessages[index];
        newMessages[index] = newMessages[index - 1];
        newMessages[index - 1] = temp;
        
        messages = newMessages;
    }
    
    // Move message down in the order
    function moveMessageDown(index: number) {
        if (index >= messages.length - 1) return;
        
        const newMessages = [...messages];
        const temp = newMessages[index];
        newMessages[index] = newMessages[index + 1];
        newMessages[index + 1] = temp;
        
        messages = newMessages;
    }

    // Handle form submission
    async function handleSubmit() {
        if (!validateForm()) return;
        
        isSubmitting = true;
        error = null;
        
        try {
            if (mode === 'create') {
                const newPrompt: Prompt = {
                    UID: '',  // Will be assigned by the API
                    model: editModel,
                    parameters: editParameters,
                    messages,
                    goalUID: goalUID || ''
                };
                
                await llmangoAPI.createPrompt(newPrompt);
                
                if (onSave) {
                    onSave(newPrompt);
                }
            } else if (mode === 'edit' && promptUID) {
                const updatedPrompt: Prompt = {
                    UID: promptUID,
                    model: editModel,
                    parameters: editParameters,
                    messages,
                    goalUID: prompt?.goalUID || goalUID || ''
                };
                
                await llmangoAPI.updatePrompt(promptUID, updatedPrompt);
                
                if (onSave) {
                    onSave(updatedPrompt);
                }
            }
            
            // Close the modal after successful operation
            onClose();
        } catch (e) {
            error = e instanceof Error ? e.message : 'An error occurred';
        } finally {
            isSubmitting = false;
        }
    }
</script>

<Modal isOpen={isOpen} title={mode === 'create' ? 'Create Prompt' : 'Edit Prompt'} onClose={onClose}>
    <div class="create-prompt">
    {#if mode === 'edit'}
        <div class="warning-section">
            <h4>⚠️ Warning: Data Consistency Risk</h4>
            <p>Editing this prompt may affect existing solutions and goals that use it.</p>
        </div>
    {/if}
    
    {#if error}
        <div class="error-message">{error}</div>
    {/if}
    
    <form>
        <div class="form-group">
            <label for="promptModel">Model</label>
            {#if modelsLoading}
                <div class="form-control placeholder">Loading models...</div>
            {:else}
                <select 
                    id="promptModel" 
                    class="form-control" 
                    bind:value={editModel}
                    required
                >
                    <option value="">Select a model</option>
                    {#each modelList as model}
                        <option value={model.id}>{model.name}</option>
                    {/each}
                </select>
            {/if}
        </div>
        
        <div class="form-group">
            <label>Parameters</label>
            <div class="parameters-section">
                <div class="parameter-item primary-param">
                    <label for="promptTemperature">Temperature</label>
                    <input 
                        type="number" 
                        id="promptTemperature" 
                        class="form-control" 
                        step="0.1" 
                        min="0" 
                        max="2"
                        bind:value={editParameters.temperature}
                    />
                </div>
                
                <details class="advanced-params">
                    <summary>More options</summary>
                    <div class="parameters-grid">
                        <div class="parameter-item">
                            <label for="promptMaxTokens">Max Tokens</label>
                            <input 
                                type="number" 
                                id="promptMaxTokens" 
                                class="form-control" 
                                min="1"
                                bind:value={editParameters.max_tokens}
                            />
                        </div>
                        <div class="parameter-item">
                            <label for="promptTopP">Top P</label>
                            <input 
                                type="number" 
                                id="promptTopP" 
                                class="form-control" 
                                step="0.1" 
                                min="0" 
                                max="1"
                                bind:value={editParameters.top_p}
                            />
                        </div>
                        <div class="parameter-item">
                            <label for="promptFrequencyPenalty">Frequency</label>
                            <input 
                                type="number" 
                                id="promptFrequencyPenalty" 
                                class="form-control" 
                                step="0.1" 
                                min="-2" 
                                max="2"
                                bind:value={editParameters.frequency_penalty}
                            />
                        </div>
                        <div class="parameter-item">
                            <label for="promptPresencePenalty">Presence</label>
                            <input 
                                type="number" 
                                id="promptPresencePenalty" 
                                class="form-control" 
                                step="0.1" 
                                min="-2" 
                                max="2"
                                bind:value={editParameters.presence_penalty}
                            />
                        </div>
                    </div>
                </details>
            </div>
        </div>
        
        {#if goalUID}
            <div class="form-group">
                <label>Goal Variables</label>
                {#if goalLoading}
                    <div class="placeholder">Loading goal data...</div>
                {:else if goal && goal.exampleInput}
                    <div class="goal-variables">
                        <h4>Example Input for {goal.title || 'Goal'}</h4>
                        <pre class="example-input">{JSON.stringify(goal.exampleInput, null, 2)}</pre>
                        <p class="variables-help">You can use these variables in your messages with the syntax <code>{"{{variable.path}}"}</code></p>
                    </div>
                {:else}
                    <div class="placeholder">No example input available for this goal</div>
                {/if}
            </div>
        {/if}
                 
        {#if !goalUID}
        <div class="variables-info">
            <span>Available variable: <code>{"{{.Input}}"}</code> - The goal's input data</span>
        </div>
        {/if}
        <div class="form-group messages-section">
            <div class="messages-header">
                <label>Messages</label>
                <div class="message-actions">
                    <button type="button" class="btn btn-sm" onclick={() => addMessage('system')}>
                        + System
                    </button>
                    <button type="button" class="btn btn-sm" onclick={() => addMessage('user')}>
                        + User
                    </button>
                    <button type="button" class="btn btn-sm" onclick={() => addMessage('assistant')}>
                        + Assistant
                    </button>
                </div>
            </div>
   
            
            {#each messages as message, index}
                <label class="message-card" for={`message-textarea-${index}`}>
                    <div class="message-type-header">
                        <span class="message-type {message.role}">
                            {message.role.charAt(0).toUpperCase() + message.role.slice(1)}
                        </span>
                    </div>
                    
                    <textarea 
                        id={`message-textarea-${index}`}
                        class="message-textarea" 
                        rows="3" 
                        placeholder={`Enter ${message.role} message here...`}
                        bind:value={message.content}
                        required
                    ></textarea>
                    
                    <div class="preview">
                        <span class="preview-label">Preview</span>
                        <div class="preview-content">
                            <PromptMessageFormatter message={message.content} goal={goal || undefined} />
                        </div>
                    </div>
                    
                    <div class="message-controls-bar">
                        <button 
                            type="button" 
                            class="btn-icon" 
                            onclick={(e) => { e.preventDefault(); moveMessageUp(index); }}
                            disabled={index === 0}
                            title="Move up"
                        >
                            ↑
                        </button>
                        <button 
                            type="button" 
                            class="btn-icon" 
                            onclick={(e) => { e.preventDefault(); moveMessageDown(index); }}
                            disabled={index === messages.length - 1}
                            title="Move down"
                        >
                            ↓
                        </button>
                        <button 
                            type="button" 
                            class="btn-icon btn-danger" 
                            onclick={(e) => { e.preventDefault(); removeMessage(index); }}
                            title="Remove"
                        >
                            ×
                        </button>
                    </div>
                </label>
            {/each}
        </div>
        
        <div class="modal-footer">
            <button
                type="button"
                class="btn btn-primary"
                onclick={handleSubmit}
                disabled={isSubmitting}
            >
                {isSubmitting ? 'Saving...' : (mode === 'create' ? 'Create' : 'Save Changes')}
            </button>
            <button
                type="button"
                class="btn btn-secondary"
                onclick={onClose}
                disabled={isSubmitting}
            >
                Cancel
            </button>
        </div>
    </form>
</div>
</Modal>

<style>
    .create-prompt{
        max-width: 100%;
        width: 40rem;
    }
    .form-group {
        margin-bottom: 1rem;
    }
    
    .parameters-section {
        margin-top: 0.5rem;
    }
    
    .primary-param {
        margin-bottom: 0.5rem;
    }
    
    .primary-param input {
        width: 8rem;
    }
    
    .advanced-params {
        font-size: 0.85rem;
    }
    
    .advanced-params summary {
        cursor: pointer;
        color: #007bff;
        display: inline-block;
        user-select: none;
    }
    
    .advanced-params summary:hover {
        text-decoration: underline;
    }
    
    .parameters-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
        gap: 0.5rem;
    }
    
    .parameter-item {
        display: flex;
        flex-direction: column;
        gap: 0.25rem;
    }
    
    .parameter-item label {
        font-size: 0.75rem;
        color: #555;
        font-weight: 500;
    }
    
    .parameter-item input {
        height: 30px;
        padding: 0.25rem 0.4rem;
        border: 1px solid #ccc;
        border-radius: 3px;
    }
    
    .warning-section {
        background-color: #fff3cd;
        border-radius: 0.25rem;
        padding: 0.75rem;
        margin-bottom: 1rem;
    }
    
    .warning-section h4 {
        color: #856404;
        margin-top: 0;
        margin-bottom: 0.25rem;
        font-size: 0.9rem;
    }
    
    .warning-section p {
        color: #856404;
        margin-bottom: 0;
        font-size: 0.8rem;
    }
    
    .error-message {
        background-color: #f8d7da;
        color: #dc3545;
        padding: 0.5rem 0.75rem;
        border-radius: 0.25rem;
        margin-bottom: 1rem;
        font-size: 0.85rem;
    }
    
    .modal-footer {
        display: flex;
        justify-content: flex-end;
        gap: 0.5rem;
        padding-top: 1rem;
    }
    
    .variables-info {
        background-color: #f5f5f5;
        padding: 0.5rem;
        border-radius: 3px;
        margin-bottom: 1rem;
        font-size: 0.75rem;
    }
    
    .variables-info code {
        background-color: #eef;
        padding: 2px 4px;
        border-radius: 3px;
        font-family: monospace;
    }
    
    .placeholder {
        color: #777;
        padding: 0.5rem;
        background-color: #f5f5f5;
        font-size: 0.85rem;
        border-radius: 3px;
    }
    
    .messages-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 1rem;
    }
    
    .message-actions {
        display: flex;
        gap: 0.25rem;
    }
    
    .btn-sm {
        font-size: 0.8rem;
        padding: 0.2rem 0.5rem;
        background-color: #f5f5f5;
        border: 1px solid #ddd;
        border-radius: 3px;
        cursor: pointer;
    }
    
    .btn-sm:hover {
        background-color: #e9e9e9;
    }
    
    .messages-section {
        margin-top: 2rem;
    }
    
    .message-card {
        display: block;
        margin-bottom: 1.25rem;
        background-color: #fcfcfc;
        border: 1px solid #e0e0e0;
        border-radius: 4px;
        padding: 1rem;
        position: relative;
        cursor: text;
        transition: all 0.2s ease;
    }
    
    .message-card:focus-within {
        border-color: #80bdff;
        box-shadow: 0 0 0 0.2rem rgba(0,123,255,.25);
        outline: 0;
    }
    
    .message-type-header {
        margin-bottom: 0.75rem;
    }
    
    .message-type {
        font-weight: 500;
        padding: 0.15rem 0.35rem;
        border-radius: 3px;
        font-size: 0.8rem;
        display: inline-block;
    }
    
    .message-type.system {
        background-color: #e2e3e5;
        color: #383d41;
    }
    
    .message-type.user {
        background-color: #d1ecf1;
        color: #0c5460;
    }
    
    .message-type.assistant {
        background-color: #d4edda;
        color: #155724;
    }
    
    .message-textarea {
        width: 100%;
        border: none;
        padding: 0.35rem;
        min-height: 70px;
        font-family: monospace;
        font-size: 0.85rem;
        resize: vertical;
        margin-bottom: 0.75rem;
        background-color: transparent;
    }
    
    .message-textarea:focus {
        outline: none;
    }
    
    .preview {
        background-color: #f8f9fa;
        border: 1px solid #e0e0e0;
        border-radius: 3px;
        padding: 0.5rem;
        font-size: 0.85rem;
        margin-bottom: 0.75rem;
    }
    
    .preview-label {
        font-weight: 500;
        color: #555;
        font-size: 0.75rem;
        display: block;
        margin-bottom: 0.25rem;
    }
    
    .preview-content {
        min-height: 1.5rem;
        font-size: 0.85rem;
    }
    
    .message-controls-bar {
        display: flex;
        gap: 0.25rem;
        justify-content: flex-end;
    }
    
    .btn-icon {
        width: 22px;
        height: 22px;
        padding: 0;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 0.85rem;
        background: none;
        border: 1px solid #ddd;
        border-radius: 3px;
        cursor: pointer;
        z-index: 2;
    }
    
    .btn-icon:hover {
        background-color: #f5f5f5;
    }
    
    .btn-danger {
        color: #dc3545;
        border-color: #dc3545;
    }
    
    .btn-danger:hover {
        background-color: #f8d7da;
    }
    
    .goal-variables {
        background-color: #f8f9fa;
        border-radius: 3px;
        padding: 0.75rem;
    }
    
    .goal-variables h4 {
        margin-top: 0;
        margin-bottom: 0.5rem;
        font-size: 0.9rem;
    }
    
    .example-input {
        background-color: #f0f0f0;
        padding: 0.5rem;
        border-radius: 3px;
        font-family: monospace;
        font-size: 0.75rem;
        overflow: auto;
        max-height: 150px;
    }
    
    .variables-help {
        font-size: 0.75rem;
        color: #555;
        margin: 0.5rem 0 0;
    }
</style>
