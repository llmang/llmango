<script lang="ts">
    import { llmangoAPI } from '$lib/classes/llmangoAPI.svelte';
    import type { Prompt, Solution, Goal } from '$lib/classes/llmangoAPI.svelte';
    import Modal from '$lib/Modal.svelte';
    import { onMount } from 'svelte';
    import PromptMessageFormatter from '$lib/PromptMessageFormatter.svelte';
    
    let { isOpen, mode, goalUID, prompts, currentSolution, currentSolutionId, onClose } = $props<{
        isOpen: boolean;
        mode: 'create' | 'edit';
        goalUID: string;
        prompts: Record<string, Prompt>;
        currentSolution: Solution | null;
        currentSolutionId: string;
        onClose: () => void;
    }>();
    
    let selectedPromptUID = $state('');
    let weight = $state(1);
    let isCanary = $state(false);
    let maxRuns = $state(10);
    let searchText = $state('');
    let showSearch = $state(false);
    let loading = $state(false);
    let error = $state<string | null>(null);
    let selectedPrompt = $state<Prompt | null>(null);
    let goal = $state<Goal | null>(null);
    let goalLoading = $state(false);
    
    // Fetch the goal data
    onMount(async () => {
        if (goalUID) {
            await fetchGoal();
        }
    });
    
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
    
    // Initialize form when modal opens
    $effect(() => {
        if (isOpen) {
            if (mode === 'edit' && currentSolution) {
                selectedPromptUID = currentSolution.promptUID;
                weight = currentSolution.weight;
                isCanary = currentSolution.isCanary;
                maxRuns = currentSolution.maxRuns;
                
                // Get the prompt details for preview
                if (selectedPromptUID && prompts[selectedPromptUID]) {
                    selectedPrompt = prompts[selectedPromptUID];
                }
            } else {
                // Reset form for create mode
                selectedPromptUID = '';
                weight = 1;
                isCanary = false;
                maxRuns = 10;
                selectedPrompt = null;
            }
        }
    });
    
    // Update selected prompt when prompt UID changes
    $effect(() => {
        if (selectedPromptUID && prompts[selectedPromptUID]) {
            selectedPrompt = prompts[selectedPromptUID];
        } else {
            selectedPrompt = null;
        }
    });
    
    const toggleSearch = () => {
        showSearch = !showSearch;
    };
    
    const selectPrompt = (promptUID: string) => {
        selectedPromptUID = promptUID;
        showSearch = false;
    };
    let filteredPrompts = $derived.by(() => {
        if (!searchText) return Object.entries(prompts) as [string, Prompt][];
        
        return (Object.entries(prompts) as [string, Prompt][]).filter(([id, prompt]) => {
            const searchString = `${id} ${prompt.UID} ${prompt.model}`.toLowerCase();
            return searchString.includes(searchText.toLowerCase());
        });
    });

    const handleSubmit = async () => {
        if (mode === 'create' && !selectedPromptUID) {
            error = 'Please select a prompt';
            return;
        }
        
        loading = true;
        error = null;
        
        try {
            // In edit mode, we need to ensure we preserve the original promptUID
            // since we're not allowing prompt selection change in the UI
            const promptToUse = mode === 'edit' && currentSolution 
                ? currentSolution.promptUID 
                : selectedPromptUID;
                
            const solution: Solution = {
                promptUID: promptToUse,
                weight: weight,
                isCanary: isCanary,
                maxRuns: maxRuns,
                totalRuns: currentSolution?.totalRuns || 0 // Preserve totalRuns when editing
            };
            
            if (mode === 'create') {
                await llmangoAPI.createSolution(goalUID, solution);
            } else if (mode === 'edit' && currentSolutionId) {
                await llmangoAPI.updateSolution(currentSolutionId, solution, goalUID);
            }
            
            onClose();
            location.href=location.href
        } catch (err) {
            error = err instanceof Error ? err.message : 'An unknown error occurred';
            console.error('Error saving solution:', err);
        } finally {
            loading = false;
        }
    };
    
    const handleDelete = async () => {
        if (!currentSolutionId || mode !== 'edit') return;
        
        if (!confirm('Are you sure you want to delete this solution?')) {
            return;
        }
        
        loading = true;
        error = null;
        
        try {
            await llmangoAPI.deleteSolution(currentSolutionId, goalUID);
            onClose();
            location.href=location.href
        } catch (err) {
            error = err instanceof Error ? err.message : 'An unknown error occurred';
            console.error('Error deleting solution:', err);
        } finally {
            loading = false;
        }
    };
</script>

<Modal isOpen={isOpen} title={mode === 'create' ? 'Add New Solution' : 'Edit Solution'} onClose={onClose}>
    <div class="modal-body">
        {#if error}
            <div class="error">{error}</div>
        {/if}
        
        <div class="form-group">
            <label>Prompt</label>
            <div class="prompt-selector">
                <div class="prompt-select-row">
                    {#if mode === 'edit'}
                        <div class="selected-prompt-display">
                            <strong>Prompt:</strong> {selectedPrompt ? selectedPrompt.UID : selectedPromptUID}
                        </div>
                    {:else}
                        <select 
                            class="form-control"
                            bind:value={selectedPromptUID}
                        >
                            <option value="">-- Select a prompt --</option>
                            {#each Object.entries(prompts) as [promptUID, prompt] (promptUID)}
                                <option value={promptUID}>
                                    {(prompt as Prompt).UID || promptUID}
                                </option>
                            {/each}
                        </select>
                        <button 
                            type="button" 
                            class="btn btn-secondary"
                            onclick={toggleSearch}
                            style="text-wrap:nowrap;"
                        >
                            {showSearch ? 'Hide Search' : 'Search'}
                        </button>
                    {/if}
                </div>
                
                {#if showSearch && mode !== 'edit'}
                    <div class="prompt-search-dropdown">
                        <div class="form-group">
                            <input 
                                type="text" 
                                placeholder="Search prompts..." 
                                class="form-control"
                                bind:value={searchText}
                            />
                        </div>
                        
                        <div class="prompt-search-results">
                            {#each filteredPrompts as [promptUID, prompt] (promptUID)}
                                <div 
                                    class="prompt-search-item {selectedPromptUID === promptUID ? 'active' : ''}"
                                    onclick={() => selectPrompt(promptUID)}
                                >
                                    <div class="prompt-item-header">
                                        <div class="prompt-item-title">
                                            {(prompt as Prompt).UID || promptUID}
                                        </div>
                                        <small class="prompt-item-model">
                                            {(prompt as Prompt).model}
                                        </small>
                                    </div>
                                    <div class="prompt-item-id">
                                        <code>{promptUID}</code>
                                    </div>
                                </div>
                            {/each}
                        </div>
                    </div>
                {/if}
                
                <!-- Prompt Preview Section -->
                {#if selectedPrompt}
                    <div class="prompt-preview">
                        <div class="prompt-preview-header">
                            <div class="preview-title">Prompt Preview</div>
                            <div class="preview-model">{selectedPrompt.model}</div>
                        </div>
                        
                        {#if goalLoading}
                            <div class="loading-goal">Loading goal data for variable formatting...</div>
                        {/if}
                        
                        <div class="messages-container">
                            {#if selectedPrompt.messages && selectedPrompt.messages.length > 0}
                                {#each selectedPrompt.messages as message, index}
                                    <div class="message {message.role}">
                                        <div class="message-header">
                                            <span class="message-role">{message.role}</span>
                                            <span class="message-index">#{index + 1}</span>
                                        </div>
                                        <div class="message-content">
                                            <PromptMessageFormatter message={message.content} goal={goal} />
                                        </div>
                                    </div>
                                {/each}
                            {:else}
                                <div class="no-messages">No messages defined for this prompt</div>
                            {/if}
                        </div>
                    </div>
                {/if}
            </div>
        </div>
        
        <div class="form-group">
            <label>Weight</label>
            <input 
                type="number" 
                min="0" 
                bind:value={weight}
                class="form-control"
            />
        </div>
        
        <div class="form-group">
            <label class="form-check">
                <input 
                    type="checkbox" 
                    bind:checked={isCanary}
                />
                <span class="form-check-label">Is Canary Test</span>
            </label>
        </div>
        
        {#if isCanary}
            <div class="form-group">
                <label>Max Runs</label>
                <input 
                    type="number" 
                    min="1" 
                    bind:value={maxRuns}
                    class="form-control"
                />
            </div>
        {/if}
    </div>
    
    <div class="modal-footer">
        {#if mode === 'edit'}
            <button 
                type="button"
                class="btn btn-danger"
                onclick={handleDelete}
                disabled={loading}
            >
                Delete
            </button>
        {/if}
        
        <div class="modal-actions">
            <button 
                type="button"
                class="btn btn-secondary"
                onclick={onClose}
                disabled={loading}
            >
                Cancel
            </button>
            <button 
                type="button"
                class="btn btn-primary"
                onclick={handleSubmit}
                disabled={loading}
            >
                {loading ? 'Saving...' : mode === 'create' ? 'Create' : 'Update'}
            </button>
        </div>
    </div>
</Modal>

<style>
    .modal-body {
        padding: 0;
        max-width: 60rem;
        width: 40rem;
        margin:0 auto;
    }
    
    .modal-footer {
        padding: 1rem 0;
        border-top: 1px solid #eee;
        display: flex;
        justify-content: space-between;
        align-items: center;
    }
    
    .modal-actions {
        display: flex;
        gap: 0.5rem;
    }
    
    .prompt-selector {
        position: relative;
    }
    
    .prompt-select-row {
        display: flex; 
        gap: 10px; 
        align-items: center;
    }
    
    .prompt-search-dropdown {
        position: absolute; 
        top: 100%; 
        left: 0; 
        width: 100%; 
        background: white; 
        border: 1px solid #ddd; 
        border-radius: 4px; 
        z-index: 100; 
        margin-top: 5px; 
        max-height: 300px; 
        overflow-y: auto; 
        padding: 10px; 
        box-shadow: 0 4px 8px rgba(0,0,0,0.1);
    }
    
    .prompt-search-results {
        display: flex; 
        flex-direction: column; 
        gap: 5px;
    }
    
    .prompt-search-item {
        cursor: pointer; 
        padding: 8px; 
        border-radius: 4px;
    }
    
    .prompt-search-item:hover {
        background-color: #f5f5f5;
    }
    
    .prompt-search-item.active {
        background-color: #e9eef1;
    }
    
    .prompt-item-header {
        display: flex; 
        justify-content: space-between; 
        align-items: center;
    }
    
    .prompt-item-title {
        font-weight: bold;
    }
    
    .prompt-item-model {
        font-size: 0.8rem;
    }
    
    .prompt-item-id {
        font-size: 0.8rem;
        margin-top: 4px;
    }
    
    /* Selected prompt display for edit mode */
    .selected-prompt-display {
        background: #f8f8f8;
        padding: 0.75rem 1rem;
        border-radius: 4px;
        border: 1px solid #e0e0e0;
        width: 100%;
        font-size: 0.95rem;
    }
    
    /* Prompt Preview Styles */
    .prompt-preview {
        margin-top: 1rem;
        border: 1px solid #eee;
        border-radius: 6px;
        overflow: hidden;
    }
    
    .prompt-preview-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 0.5rem 1rem;
        background: #f5f5f5;
        border-bottom: 1px solid #eee;
    }
    
    .preview-title {
        font-weight: bold;
        font-size: 0.9rem;
    }
    
    .preview-model {
        font-size: 0.8rem;
        color: #666;
        background: #e9e9e9;
        padding: 0.2rem 0.5rem;
        border-radius: 4px;
    }
    
    .loading-goal {
        padding: 0.5rem;
        text-align: center;
        font-style: italic;
        color: #666;
        background-color: #f8f9fa;
    }
    
    .messages-container {
        padding: 0.5rem;
        max-height: 300px;
        overflow-y: auto;
    }
    
    .message {
        margin-bottom: 0.75rem;
        border: 1px solid #eee;
        border-radius: 6px;
        overflow: hidden;
    }
    
    .message-header {
        display: flex;
        justify-content: space-between;
        padding: 0.25rem 0.5rem;
        font-size: 0.8rem;
        background-color: #f5f5f5;
    }
    
    .message-role {
        font-weight: bold;
        text-transform: capitalize;
    }
    
    .message-index {
        color: #666;
    }
    
    .message-content {
        padding: 0.5rem;
        font-size: 0.85rem;
        white-space: pre-wrap;
    }
    
    .message.system .message-header {
        background-color: #e9ecef;
        color: #495057;
    }
    
    .message.user .message-header {
        background-color: #e7f5ff;
        color: #0d6efd;
    }
    
    .message.assistant .message-header {
        background-color: #d4edda;
        color: #28a745;
    }
    
    .no-messages {
        padding: 1rem;
        text-align: center;
        color: #666;
        font-style: italic;
    }
    
    .form-check {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        cursor: pointer;
    }
    
    .form-check-label {
        margin-bottom: 0;
    }
    
    .error {
        color: #721c24;
        background-color: #f8d7da;
        border: 1px solid #f5c6cb;
        padding: 0.75rem;
        margin-bottom: 1rem;
        border-radius: 4px;
    }
</style> 