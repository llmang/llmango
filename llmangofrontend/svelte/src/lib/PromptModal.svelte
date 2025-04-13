<script lang="ts">
    import Modal from './Modal.svelte';
    import type { Prompt, Goal } from './classes/llmangoAPI.svelte';
    import { llmangoAPI, PromptParameters } from './classes/llmangoAPI.svelte';
    import { openrouter, type OpenRouterModel } from './classes/openrouter.svelte';
    import { onMount } from 'svelte';
    import PromptMessageFormatter from './PromptMessageFormatter.svelte';

    // Props definition using Svelte 5 syntax
    let { 
        isOpen, 
        goalUID, // This is the goalUID string
        prompt = null, 
        onClose = () => {},
        onSave = null
    } = $props<{
        isOpen: boolean;
        goalUID: string; // goalUID
        prompt?: Prompt | null; // null for new prompt, Prompt object for editing
        onClose: () => void;
        onSave?: ((prompt: Prompt) => void) | null;
    }>();

    // Determine if we're creating a new prompt or editing an existing one
    const isNewPrompt = $derived(!prompt);
    let goal = $derived(llmangoAPI.goals[prompt?.goalUID || goalUID] || null);
    
    // Form state
    let formData = $state<Prompt>({
        UID: '',
        goalUID: goalUID || "",
        model: '',
        parameters: new PromptParameters(),
        messages: [{ role: 'user', content: '' }],
        weight: 1,
        isCanary: false,
        maxRuns: 0,
        totalRuns: 0
    });
    
    // UI state
    let error = $state<string | null>(null);
    let isSubmitting = $state(false);
    let modelList = $state<OpenRouterModel[]>([]);
    let modelsLoading = $state(true);
    let uidError = $state<string | null>(null);
    
    // Goal state
    let goalData = $state<Goal | null>(null);
    let goalLoading = $state(false);

    // Initialize the form with existing prompt data if editing
    $effect(() => {
        if (prompt && formData.UID!=prompt.UID) {
            formData = { ...prompt };
            // Ensure we have at least one message
            if (!formData.messages || formData.messages.length === 0) {
                formData.messages = [{ role: 'user', content: '' }];
            }
        } else {
            // Set the goalUID from prop for new prompts
            formData.goalUID = goal.UID;
        }
    });

    // Fetch available models and goal data when component mounts
    onMount(async () => {
        try {
            modelsLoading = true;
            modelList = await openrouter.load();
        } catch (error) {
            console.error('Failed to fetch models:', error);
        } finally {
            modelsLoading = false;
        }
    });
    

    // Validate UID format (URL-safe characters)
    function validateUID(uid: string): boolean {
        // Check for URL-safe characters only (alphanumeric, dash, underscore)
        const urlSafePattern = /^[a-zA-Z0-9-_]+$/;
        if (!urlSafePattern.test(uid)) {
            uidError = "UID can only contain letters, numbers, dashes, and underscores (no spaces)";
            return false;
        }
        uidError = null;
        return true;
    }

    // Add a new message to the prompt
    function addMessage(role: string) {
        formData.messages = [...formData.messages, { role, content: '' }];
    }

    // Remove a message at the specified index
    function removeMessage(index: number) {
        if (formData.messages.length <= 1) {
            error = 'At least one message is required';
            return;
        }
        formData.messages = formData.messages.filter((_, i) => i !== index);
    }

    // Move a message up in the sequence
    function moveMessageUp(index: number) {
        if (index <= 0) return;
        
        const newMessages = [...formData.messages];
        [newMessages[index-1], newMessages[index]] = [newMessages[index], newMessages[index-1]];
        formData.messages = newMessages;
    }

    // Move a message down in the sequence
    function moveMessageDown(index: number) {
        if (index >= formData.messages.length - 1) return;
        
        const newMessages = [...formData.messages];
        [newMessages[index], newMessages[index+1]] = [newMessages[index+1], newMessages[index]];
        formData.messages = newMessages;
    }

    // Validate form before submission
    function validateForm(): boolean {
        // Clear previous errors
        error = null;
        uidError = null;
        
        if (isNewPrompt && !formData.UID.trim()) {
            uidError = 'UID is required';
            return false;
        }
        
        if (isNewPrompt && !validateUID(formData.UID)) {
            return false;
        }

        if (!formData.model) {
            error = 'Please select a default model';
            return false;
        }

        if (!formData.messages || formData.messages.length === 0) {
            error = 'At least one message is required';
            return false;
        }

        for (const message of formData.messages) {
            if (!message.content.trim()) {
                error = `${message.role} message cannot be empty`;
                return false;
            }
        }

        return true;
    }

    // Handle form submission
    async function handleSubmit() {
        if (!validateForm()) return;
        
        isSubmitting = true;
        error = null;
        
        try {
            if (isNewPrompt) {
                // Create new prompt
                await llmangoAPI.createPrompt(formData);
            } else {
                // Update existing prompt
                await llmangoAPI.updatePrompt(formData.UID, formData);
            }
            
            // Call onSave callback if provided
            if (onSave) {
                onSave(formData);
            }
            
            // Close the modal
            onClose();
        } catch (e) {
            error = e instanceof Error ? e.message : 'An error occurred';
        } finally {
            isSubmitting = false;
        }
    }
</script>

<Modal isOpen={isOpen} title={isNewPrompt ? 'Create Prompt' : 'Edit Prompt'} onClose={onClose}>
    <div class="prompt-modal">
        {#if error}
            <div class="error-message">{error}</div>
        {/if}

        <form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }}>
            <!-- Canary Options Section (now first) -->
            <div class="canary-section">
                <h3>Canary Testing</h3>
                <div class="form-group">
                    <label class="checkbox-label">
                        <input 
                            type="checkbox" 
                            bind:checked={formData.isCanary} 
                            disabled={!isNewPrompt}
                        />
                        Use as canary test
                    </label>
                    <small>Canary prompts are used for testing and monitoring</small>
                </div>

                {#if formData.isCanary}
                    <div class="form-group">
                        <label for="maxRuns">Max Runs</label>
                        <input 
                            type="number" 
                            id="maxRuns" 
                            bind:value={formData.maxRuns} 
                            min="0"
                            class="form-control"
                        />
                    </div>
                {/if}
            </div>

            <!-- Only show these sections if in new prompt mode -->
            {#if isNewPrompt}
                <!-- UID Field -->
                <div class="form-group">
                    <label for="uid">UID</label>
                    <input 
                        type="text" 
                        id="uid" 
                        bind:value={formData.UID} 
                        class="form-control"
                        required
                    />
                    {#if uidError}
                        <small class="error-text">{uidError}</small>
                    {:else}
                        <small>Unique identifier (URL-safe characters only)</small>
                    {/if}
                </div>

                <!-- Model Selection -->
                <div class="form-group">
                    <label for="model">Default Model</label>
                    {#if modelsLoading}
                        <div class="form-control loading-placeholder">Loading models...</div>
                    {:else}
                        <select 
                            id="model" 
                            bind:value={formData.model} 
                            class="form-control"
                            required
                        >
                            <option value="">Select a model</option>
                            {#each modelList as model}
                                <option value={model.id}>{model.name}</option>
                            {/each}
                        </select>
                    {/if}
                </div>

                <!-- Weight -->
                <div class="form-group">
                    <label for="weight">Weight</label>
                    <input 
                        type="number" 
                        id="weight" 
                        bind:value={formData.weight} 
                        min="0"
                        step="0.1"
                        class="form-control"
                    />
                    <small>Higher weight means this prompt will be used more often</small>
                </div>

                <!-- Parameters -->
                <div class="form-group parameters-section">
                    <h3>Parameters</h3>
                    <details class="advanced-params">
                        <summary>Show parameters</summary>
                        <div class="parameter-grid">
                            <div class="parameter-item">
                                <label for="temperature">Temperature</label>
                                <input 
                                    type="number" 
                                    id="temperature" 
                                    bind:value={formData.parameters.temperature} 
                                    min="0" 
                                    max="2" 
                                    step="0.1" 
                                    class="form-control"
                                />
                            </div>
                            
                            <div class="parameter-item">
                                <label for="max_tokens">Max Tokens</label>
                                <input 
                                    type="number" 
                                    id="max_tokens" 
                                    bind:value={formData.parameters.max_tokens} 
                                    min="1" 
                                    class="form-control"
                                />
                            </div>
                            
                            <div class="parameter-item">
                                <label for="top_p">Top P</label>
                                <input 
                                    type="number" 
                                    id="top_p" 
                                    bind:value={formData.parameters.top_p} 
                                    min="0" 
                                    max="1" 
                                    step="0.01" 
                                    class="form-control"
                                />
                            </div>
                            
                            <div class="parameter-item">
                                <label for="frequency_penalty">Frequency Penalty</label>
                                <input 
                                    type="number" 
                                    id="frequency_penalty" 
                                    bind:value={formData.parameters.frequency_penalty} 
                                    min="-2" 
                                    max="2" 
                                    step="0.1" 
                                    class="form-control"
                                />
                            </div>
                            
                            <div class="parameter-item">
                                <label for="presence_penalty">Presence Penalty</label>
                                <input 
                                    type="number" 
                                    id="presence_penalty" 
                                    bind:value={formData.parameters.presence_penalty} 
                                    min="-2" 
                                    max="2" 
                                    step="0.1" 
                                    class="form-control"
                                />
                            </div>
                        </div>
                    </details>
                </div>
                
                <!-- Goal Variables -->
                {#if goal}
                    <div class="form-group">
                        <label>Goal Variables</label>
                        {#if goalLoading}
                            <div class="loading-placeholder">Loading goal data...</div>
                        {:else if goalData && goalData.exampleInput}
                            <div class="goal-variables">
                                <h4>Example Input for {goalData.title || 'Goal'}</h4>
                                <pre class="example-input">{JSON.stringify(goalData.exampleInput, null, 2)}</pre>
                                <p class="variables-help">You can use these variables in your messages with the syntax <code>{"{{variable.path}}"}</code></p>
                            </div>
                        {:else}
                            <div class="loading-placeholder">No example input available for this goal</div>
                        {/if}
                    </div>
                {/if}
            {/if}

            <!-- Messages Section -->
            <div class="form-group messages-section">
                <div class="messages-header">
                    <h3>Messages</h3>
                    {#if isNewPrompt}
                        <div class="message-controls">
                            <button type="button" class="btn btn-sm" onclick={() => addMessage('system')}>Add System</button>
                            <button type="button" class="btn btn-sm" onclick={() => addMessage('user')}>Add User</button>
                            <button type="button" class="btn btn-sm" onclick={() => addMessage('assistant')}>Add Assistant</button>
                        </div>
                    {/if}
                </div>
                
                <div class="messages-list">
                    {#each formData.messages as message, i}
                        <label class="message-card" for={`message-textarea-${i}`}>
                            <div class="message-type-header">
                                <span class="message-type {message.role}">
                                    {message.role.charAt(0).toUpperCase() + message.role.slice(1)}
                                </span>
                            </div>
                            
                            {#if isNewPrompt}
                                <textarea 
                                    id={`message-textarea-${i}`}
                                    class="message-textarea" 
                                    placeholder={`Enter ${message.role} message...`}
                                    bind:value={message.content}
                                    rows="4"
                                    required
                                ></textarea>
                            {/if}
                            
                            <div class="preview">
                                <span class="preview-label">Preview</span>
                                <div class="preview-content">
                                    <PromptMessageFormatter message={message.content} goal={goalData} />
                                </div>
                            </div>
                            
                            {#if isNewPrompt}
                                <div class="message-controls-bar">
                                    <button 
                                        type="button" 
                                        class="btn-icon" 
                                        onclick={(e) => { e.preventDefault(); moveMessageUp(i); }}
                                        disabled={i === 0}
                                        title="Move up"
                                    >
                                        ↑
                                    </button>
                                    <button 
                                        type="button" 
                                        class="btn-icon" 
                                        onclick={(e) => { e.preventDefault(); moveMessageDown(i); }}
                                        disabled={i === formData.messages.length - 1}
                                        title="Move down"
                                    >
                                        ↓
                                    </button>
                                    <button 
                                        type="button" 
                                        class="btn-icon btn-danger" 
                                        onclick={(e) => { e.preventDefault(); removeMessage(i); }}
                                        title="Remove"
                                    >
                                        ×
                                    </button>
                                </div>
                            {/if}
                        </label>
                    {/each}
                </div>
            </div>

            <div class="form-actions">
                <button type="button" class="btn btn-secondary" onclick={onClose}>Cancel</button>
                <button type="submit" class="btn btn-primary" disabled={isSubmitting}>
                    {isSubmitting ? 'Saving...' : isNewPrompt ? 'Create' : 'Update'}
                </button>
            </div>
        </form>
    </div>
</Modal>

<style>
    .prompt-modal {
       width: 60rem;
       max-width: 100%;
    }
    
    .error-message {
        background-color: #f8d7da;
        color: #721c24;
        padding: 0.75rem;
        margin-bottom: 1rem;
        border-radius: 4px;
    }
    
    .form-group {
        margin-bottom: 1rem;
    }
    
    .form-control {
        display: block;
        width: 100%;
        padding: 0.5rem;
        font-size: 1rem;
        border: 1px solid #ced4da;
        border-radius: 0.25rem;
    }
    
    .form-control:disabled {
        background-color: #e9ecef;
        cursor: not-allowed;
    }
    
    .loading-placeholder {
        color: #6c757d;
        font-style: italic;
        padding: 0.5rem;
        background-color: #f5f5f5;
        border-radius: 0.25rem;
    }
    
    .checkbox-label {
        display: flex;
        align-items: center;
        gap: 0.5rem;
    }
    
    small {
        display: block;
        color: #6c757d;
        margin-top: 0.25rem;
    }
    
    .error-text {
        color: #dc3545;
    }
    
    .parameters-section, .messages-section {
        margin-top: 1.5rem;
    }
    
    .advanced-params summary {
        cursor: pointer;
        color: #007bff;
        display: inline-block;
        user-select: none;
        margin-bottom: 0.5rem;
    }
    
    .advanced-params summary:hover {
        text-decoration: underline;
    }
    
    .parameter-grid {
        display: grid;
        grid-template-columns: repeat(2, 1fr);
        gap: 1rem;
        margin-top: 0.5rem;
    }
    
    .messages-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 1rem;
    }
    
    .message-controls {
        display: flex;
        gap: 0.5rem;
    }
    
    .messages-list {
        display: flex;
        flex-direction: column;
        gap: 1rem;
    }
    
    .message-card {
        display: block;
        margin-bottom: 0.5rem;
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
    
    .btn {
        display: inline-block;
        font-weight: 500;
        text-align: center;
        white-space: nowrap;
        vertical-align: middle;
        user-select: none;
        border: 1px solid transparent;
        padding: 0.375rem 0.75rem;
        font-size: 1rem;
        line-height: 1.5;
        border-radius: 0.25rem;
        cursor: pointer;
    }
    
    .btn-sm {
        padding: 0.25rem 0.5rem;
        font-size: 0.875rem;
        border-radius: 0.2rem;
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
    
    .btn-primary {
        color: #fff;
        background-color: #0070f3;
        border-color: #0070f3;
    }
    
    .btn-secondary {
        color: #212529;
        background-color: #f8f9fa;
        border-color: #ced4da;
    }
    
    .form-actions {
        display: flex;
        justify-content: flex-end;
        gap: 1rem;
        margin-top: 2rem;
    }
    
    button:disabled {
        opacity: 0.65;
        cursor: not-allowed;
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
    
    .variables-info {
        background-color: #f5f5f5;
        padding: 0.5rem;
        border-radius: 3px;
        margin-bottom: 1rem;
        font-size: 0.75rem;
    }
    


    /* Add this new style for the canary section */
    .canary-section {
        background: #f8f9fa;
        padding: 1rem;
        border-radius: 4px;
        margin-bottom: 1.5rem;
        border: 1px solid #e9ecef;
    }
</style> 