<script lang="ts">
    import Modal from './Modal.svelte';
    import type { Prompt } from './classes/llmangoAPI.svelte';
    import { llmangoAPI, PromptParameters } from './classes/llmangoAPI.svelte';
    import { openrouter, type OpenRouterModel } from './classes/openrouter.svelte';
    import { onMount } from 'svelte';
    import PromptMessageFormatter from './PromptMessageFormatter.svelte';
    import InfoTooltip from './InfoTooltip.svelte';

    // Props definition using Svelte 5 syntax
    let { 
        isOpen, 
        goalUID, // This is the goalUID string
        prompt = null, 
        prefillData = null, // Add new prop for prefilling data
        onClose = () => {},
        onSave = null
    } = $props<{
        isOpen: boolean;
        goalUID: string; // goalUID
        prompt?: Prompt | null; // null for new prompt, Prompt object for editing
        prefillData?: Prompt | null; // Optional data to prefill the form
        onClose: () => void;
        onSave?: ((prompt: Prompt) => void) | null;
    }>();

    // Determine if we're creating a new prompt or editing an existing one
    const isNewPrompt = $derived(!prompt || !!prefillData);
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
    let goalLoading = $state(false);

    // Initialize the form with existing prompt data if editing, or prefill data if provided
    $effect(() => {
        const dataToUse = prompt || prefillData;
        if (dataToUse) {
            // Update properties individually instead of replacing the object
            formData.UID = prefillData ? '' : dataToUse.UID || '';
            formData.goalUID = dataToUse.goalUID || goalUID; // Ensure goalUID is set
            formData.model = dataToUse.model || '';
            // Ensure parameters is a class instance, deep copy messages
            formData.parameters = new PromptParameters(dataToUse.parameters); 
            formData.messages = (dataToUse.messages && dataToUse.messages.length > 0)
                ? JSON.parse(JSON.stringify(dataToUse.messages)) // Deep copy needed
                : [{ role: 'user', content: '' }];
            formData.weight = dataToUse.weight ?? 1; // Use nullish coalescing for default
            formData.isCanary = dataToUse.isCanary ?? false; // Use nullish coalescing for default
            formData.maxRuns = dataToUse.maxRuns ?? 0; // Use nullish coalescing for default
            formData.totalRuns = prefillData ? 0 : dataToUse.totalRuns || 0;
        } else {
            // Reset to defaults for a completely new prompt (no initial data)
            formData.UID = '';
            formData.goalUID = goal?.UID || goalUID; // Use optional chaining and fallback
            formData.model = '';
            formData.parameters = new PromptParameters();
            formData.messages = [{ role: 'user', content: '' }];
            formData.weight = 1;
            formData.isCanary = false; // Explicitly set to false
            formData.maxRuns = 0;
            formData.totalRuns = 0;
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
            console.error("Form submission error:", e);
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

        <!-- UID Field -->
        {#if isNewPrompt}
            <label for="uid" class="section-label">
                <span class="titles-secondary">UID</span>
                {#if uidError}
                    <small class="label-info error-text">{uidError}</small>
                {:else}
                    <small class="label-info">Unique identifier (URL-safe characters only){#if prefillData} - must be different from original{/if}</small>
                {/if}
            </label>
            <input 
                type="text" 
                id="uid" 
                bind:value={formData.UID} 
                class="form-control styled-input"
                required
            />
        {/if}

        <form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }}>
            <!-- Weight Field -->
            <label for="weight" class="section-label">
                 <span class="titles-secondary">Weight</span>
                 <small class="label-info">Higher weight means this prompt will be used more often</small>
            </label>
            <input 
                 type="number" 
                 id="weight" 
                 bind:value={formData.weight} 
                 min="0"
                 step="0.1"
                 class="form-control styled-input"
            />
             
            <div class="section-title">
                  <span class="titles-secondary">Canary Testing</span>
             </div>
            <div class="section-content"> 
                <div class="canary-options"> 
                    <label class="checkbox-label">
                        <input 
                            type="checkbox" 
                            bind:checked={formData.isCanary} 
                        />
                        <span class="titles-secondary">Use as canary test</span>
                        <small class="label-info">(Used for testing and monitoring)</small>
                    </label>
                     
                     {#if formData.isCanary}
                        <label for="maxRuns" class="section-label inner-label">Max Runs</label>
                        <input 
                            type="number" 
                            id="maxRuns" 
                            bind:value={formData.maxRuns} 
                            min="0"
                            class="form-control styled-input inner-input"
                        />
                     {/if}
                </div>
            </div>
            <hr/>

            {#if isNewPrompt}
                <!-- Model Selection -->
                <label for="model" class="section-label">
                    <span class="titles-secondary">Default Model</span>
                </label>
                <div class="select-wrapper"> 
                    {#if modelsLoading}
                        <div class="styled-input loading-placeholder">Loading models...</div>
                    {:else}
                        <select 
                            id="model" 
                            bind:value={formData.model} 
                            class="form-control styled-input"
                            required
                        >
                            <option value="">Select a model</option>
                            {#each modelList as model}
                                <option value={model.id}>{model.name}</option>
                            {/each}
                        </select>
                    {/if}
                </div>

                <!-- Parameters -->
                <div class="section-title">
                        <span class="titles-secondary">Parameters</span>
                </div>
                <div class="section-content"> 
                    <details class="advanced-params">
                        <summary><span class="titles-secondary">Show parameters</span></summary>
                        <div class="parameter-grid">
                            <div class="parameter-item">
                                <label for="temperature" class="inner-label">
                                    <span class="titles-secondary">Temperature</span> <span class="label-info">(Default: 1.0, Range: 0.0-2.0)</span>
                                    <InfoTooltip text="Influences the variety in responses. Lower values are more predictable, higher values more diverse. 0 is deterministic." />
                                </label>
                                <input 
                                    type="number" 
                                    id="temperature" 
                                    bind:value={formData.parameters.temperature} 
                                    min="0" 
                                    max="2" 
                                    step="0.1" 
                                    class="form-control styled-input inner-input"
                                />
                            </div>
                            
                            <div class="parameter-item">
                                <label for="max_tokens" class="inner-label">
                                    <span class="titles-secondary">Max Tokens</span> <span class="label-info">(Default: N/A, Range: 1+)</span>
                                    <InfoTooltip text="Sets the maximum number of tokens the model can generate. Max value is context length minus prompt length." />
                                </label>
                                <input 
                                    type="number" 
                                    id="max_tokens" 
                                    bind:value={formData.parameters.max_tokens} 
                                    min="1" 
                                    class="form-control styled-input inner-input"
                                />
                            </div>
                            
                            <div class="parameter-item">
                                <label for="top_p" class="inner-label">
                                    <span class="titles-secondary">Top P</span> <span class="label-info">(Default: 1.0, Range: 0.0-1.0)</span>
                                    <InfoTooltip text="Limits choices to a percentage of likely tokens. Lower values make responses more predictable." />
                                </label>
                                <input 
                                    type="number" 
                                    id="top_p" 
                                    bind:value={formData.parameters.top_p} 
                                    min="0" 
                                    max="1" 
                                    step="0.01" 
                                    class="form-control styled-input inner-input"
                                />
                            </div>
                            
                            <div class="parameter-item">
                                <label for="frequency_penalty" class="inner-label">
                                    <span class="titles-secondary">Frequency Penalty</span> <span class="label-info">(Default: 0.0, Range: -2.0-2.0)</span>
                                    <InfoTooltip text="Controls token repetition based on input frequency. Higher values reduce repetition of frequent tokens. Negative values encourage reuse." />
                                </label>
                                <input 
                                    type="number" 
                                    id="frequency_penalty" 
                                    bind:value={formData.parameters.frequency_penalty} 
                                    min="-2" 
                                    max="2" 
                                    step="0.1" 
                                    class="form-control styled-input inner-input"
                                />
                            </div>
                            
                            <div class="parameter-item">
                                <label for="presence_penalty" class="inner-label">
                                    <span class="titles-secondary">Presence Penalty</span> <span class="label-info">(Default: 0.0, Range: -2.0-2.0)</span>
                                    <InfoTooltip text="Adjusts repetition of tokens already used in input. Higher values make repetition less likely. Negative values encourage reuse." />
                                </label>
                                <input 
                                    type="number" 
                                    id="presence_penalty" 
                                    bind:value={formData.parameters.presence_penalty} 
                                    min="-2" 
                                    max="2" 
                                    step="0.1" 
                                    class="form-control styled-input inner-input"
                                />
                            </div>
                        </div>
                    </details>
                </div>
            <hr/>
            {/if}
            <!-- Goal Variables -->
            {#if goal}
                <div class="section-title">
                        <span class="titles-secondary">Goal Variables</span>
                        <small class="label-info">You can use these variables in your messages with the syntax <code>{"{{variableName}}}}"}</code></small>
                </div>
                <div class="section-content"> 
                    {#if goalLoading}
                        <div class="loading-placeholder">Loading...</div>
                    {:else if goal && goal.inputOutput.inputExample && Object.keys(goal.inputOutput.inputExample).length > 0}
                        <div class="variables-list">
                            {#each Object.keys(goal.inputOutput.inputExample) as key}
                                <code class="variable-key">{`{{${key}}}`}</code>
                            {/each}
                        </div>
                    {:else}
                            <div class="loading-placeholder">No input variables defined for this goal.</div>
                    {/if}
                </div>
            {/if}
            <!-- Messages Section -->
            <div class="section-title">
                 <span class="titles-secondary">Messages</span>
             </div>
             
             <div class="section-content messages-list"> 
                 {#if isNewPrompt}
                     <div class="message-controls top-controls">
                         <button type="button" class="btn btn-sm" onclick={() => addMessage('system')}>Add System</button>
                         <button type="button" class="btn btn-sm" onclick={() => addMessage('user')}>Add User</button>
                         <button type="button" class="btn btn-sm" onclick={() => addMessage('assistant')}>Add Assistant</button>
                     </div>
                 {/if}
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
                                 <PromptMessageFormatter message={message.content} {goal} />
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

    .variable-key{
        background: rgba(0, 0, 255, 0.201)

    }
    .prompt-modal {
       width: 45rem;
       max-width: 100%;
    }
    
    .error-message {
        background-color: #f8d7da;
        color: #721c24;
        padding: 0.75rem;
        margin-bottom: 1rem;
        border-radius: 4px;
    }
    
    .form-control,
    .styled-input,
    .section-content {
        display: block;
        width: 100%;
        font-size: 1rem;
        border: none;
        border-radius: 0.25rem;
        background-color: rgba(128, 128, 128, 0.075);
        margin-bottom: 2rem;
        transition: border-color .15s ease-in-out,box-shadow .15s ease-in-out;
    }
    .section-content{
        padding:1rem;
    }
    
    .form-control:focus,
    .styled-input:focus,
    .message-textarea:focus {
        outline: 0;
        box-shadow: 0 0 0 0.2rem rgba(0,123,255,.25);
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
    
    .label-info {
        color: #6c757d;
        font-size: 0.8rem;
        margin-left: 0.5rem;
        font-weight: normal;
        margin-top: 0;
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
    
    
    .goal-variables-section {
        margin-top: 1.5rem;
    }
    
    .section-title,
    .section-label {
        margin-bottom: 0.5rem;
        margin-top: 0;
        display: block;
        font-weight: 600;
    }
    
    .variables-list {
        display: flex;
        flex-wrap: wrap;
        gap: 0.5rem;
        font-size: 0.85rem;
    }
    
    .section-content .loading-placeholder {
        background-color: #f8f9fa;
        padding: 0.75rem;
        border-radius: 3px;
    }
    
    .example-input {
        display: flex;
        gap: 1rem;
        background-color: #f0f0f0;
        padding: 0.5rem;
        border-radius: 3px;
        font-family: monospace;
        font-size: 0.75rem;
        overflow: auto;
        max-height: 150px;
    }
    
    .variables-info {
        background-color: #f5f5f5;
        padding: 0.5rem;
        border-radius: 3px;
        margin-bottom: 1rem;
        font-size: 0.75rem;
    }

    .canary-section {
        border-radius: 4px;
    }

    .section-content .styled-input.inner-input {
        margin-bottom: 0;
    }
    .inner-label {
        margin-bottom: 0.25rem;
        display: block;
    }

    .top-controls {
        margin-bottom: 1rem; 
    }

</style> 