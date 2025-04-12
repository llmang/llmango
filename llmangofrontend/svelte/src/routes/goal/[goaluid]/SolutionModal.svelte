<script lang="ts">
    import { llmangoAPI } from '$lib/classes/llmangoAPI.svelte';
    import type { Prompt, Solution } from '$lib/classes/llmangoAPI.svelte';
    import Modal from '$lib/Modal.svelte';
    import { onMount } from 'svelte';
    
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
    
    // Initialize form when editing
    onMount(() => {
        if (mode === 'edit' && currentSolution) {
            selectedPromptUID = currentSolution.promptUID;
            weight = currentSolution.weight;
            isCanary = currentSolution.isCanary;
            maxRuns = currentSolution.maxRuns;
        } else {
            // Reset form for create mode
            selectedPromptUID = '';
            weight = 1;
            isCanary = false;
            maxRuns = 10;
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
        if (!selectedPromptUID) {
            error = 'Please select a prompt';
            return;
        }
        
        loading = true;
        error = null;
        
        try {
            const solution: Solution = {
                promptUID: selectedPromptUID,
                weight: weight,
                isCanary: isCanary,
                maxRuns: maxRuns,
                totalRuns: 0 // Default for new solutions
            };
            
            if (mode === 'create') {
                await llmangoAPI.createSolution(goalUID, solution);
            } else if (mode === 'edit' && currentSolutionId) {
                await llmangoAPI.updateSolution(currentSolutionId, solution, goalUID);
            }
            
            onClose();
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
                </div>
                
                {#if showSearch}
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