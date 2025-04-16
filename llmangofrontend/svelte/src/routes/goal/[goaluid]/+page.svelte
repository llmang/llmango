<svelte:head>
    <title>{goal?.title || 'Goal Details'} - LLMango</title>
</svelte:head>

<script lang="ts">
    import FormatJson from '$lib/FormatJson.svelte';
    import type { Goal, Prompt } from '$lib/classes/llmangoAPI.svelte';
    import { llmangoAPI } from '$lib/classes/llmangoAPI.svelte';
    import { onMount } from 'svelte';
    import { page } from '$app/state';
    import { llmangoLogging, type Log, type SpendResponse } from '$lib/classes/llmangoLogging.svelte';
    import LogTable from '$lib/LogTable.svelte';
    import PromptModal from '$lib/PromptModal.svelte';
    import { base } from '$app/paths';
    import PromptCard from '$lib/PromptCard.svelte';

    let goaluid = $derived(page.params.goaluid);

    // Initial states with safe defaults
    let goal = $derived<Goal | undefined>(llmangoAPI?.goals?.[goaluid]);
    let prompts = $derived<Prompt[] | null>(llmangoAPI?.promptsByGoalUID?.[goaluid] || null);

    let loading = $state<boolean>(true);
    let error = $state<string | null>(null);
    let logs = $state<Log[]>([]);  // Initialize as empty array
    let spendResponse = $state<SpendResponse | null >(null)
    let logsLoading = $state<boolean>(false);
    // Modal state
    let promptModalOpen = $state<boolean>(false);
    let modalPrompt = $state<Prompt | null>(null);
    let isViewMode = $state<boolean>(false);
    
    // --- Simple Inline Edit State ---
    let isEditing = $state(false);
    let editableTitle = $state('');
    let editableDescription = $state('');
    let isSaving = $state(false);
    let editError = $state<string | null>(null);

    // Initialize editable fields when goal loads or changes (if not editing)
    $effect(() => {
        if (goal && !isEditing) {
            editableTitle = goal.title;
            editableDescription = goal.description;
        }
    });
    
    // Load data on component mount
    onMount(async () => {
        loading = true;
        try {
            await llmangoAPI.initialize()
            logsLoading = true;
            const logsResponse = await llmangoLogging.getGoalLogs(goaluid, {
                includeRaw: true,
                limit: 5,
                offset: 0
            });
            logs = logsResponse?.logs 
            spendResponse = await llmangoLogging.getSpend({goalUID:goaluid})
        } catch (e) {
            error = e instanceof Error ? e.message : 'Failed to load data';
        }finally{
            loading = false;
            logsLoading=false
        }
    });

    // Open modal to create a new prompt
    function openCreatePromptModal() {
        modalPrompt = null;
        isViewMode = false;
        promptModalOpen = true;
    }

    // --- Simple Save Function ---
    async function saveGoalChanges() {
        if (!goal || !goaluid || !editableTitle.trim()) {
            editError = "Title cannot be empty.";
            return;
        }
        isSaving = true;
        editError = null;
        try {
            await llmangoAPI.updateGoal(goaluid, editableTitle.trim(), editableDescription.trim());
            isEditing = false; // Exit edit mode on success
        } catch (e) {
            editError = e instanceof Error ? e.message : 'Failed to save changes';
            console.error("Save failed:", e);
        } finally {
            isSaving = false;
        }
    }
    
    function cancelEdit() {
        isEditing = false;
        editError = null;
        // Reset values from the potentially updated goal state
        if (goal) {
            editableTitle = goal.title;
            editableDescription = goal.description;
        }
    }
</script>
<style>
    .prompt-card-wrapper{
        position:relative;
        height: fit-content;
        width: fit-content;
    }
    
    .goal-page {
        max-width: 1200px;
        margin: 0 auto;
        padding: 1rem;
    }
    
    .loading, .error {
        text-align: center;
        padding: 2rem;
        background-color: #f8f9fa;
        border-radius: 8px;
        margin: 2rem 0;
    }
    
    .error {
        color: #721c24;
        background-color: #f8d7da;
    }

    .description {
        color: #6c757d;
        margin: 0;
    }

    
    .meta-info {
        display: flex;
        gap: 2rem;
        margin-bottom: 2rem;
        color: #6c757d;
    }
    
    .examples {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 1rem;
        margin-bottom: 2rem;
    }
    
    
    pre {
        background-color: #f1f3f5;
        padding: 1rem;
        border-radius: 4px;
        overflow: auto;
        white-space: pre-wrap;
        word-break: break-word;
    }
    
    .empty-state {
        background-color: #f8f9fa;
        border-radius: 8px;
        padding: 2rem;
        text-align: center;
        margin: 2rem 0;
    }
    
    .goal-debug {
        border-top: 1px solid #eee;
        padding-top: 0.5rem;
    }
    
    .goal-debug summary {
        cursor: pointer;
        color: #666;
        font-size: 0.8rem;
    }

    .no-items {
        padding: 1rem;
        text-align: center;
        color: #777;
        font-style: italic;
    }

    /* Add minimal styles for editing */
    .edit-container {
        margin-bottom: 1rem;
        position: relative;
    }
    .edit-actions {
        margin-top: 0.5rem;
        display: flex;
        gap: 0.5rem;
    }
    .edit-actions button {
        padding: 0.25rem 0.75rem;
    }
    .edit-error {
        color: red;
        font-size: 0.9em;
        margin-top: 0.5rem;
    }
    .title-input, .desc-textarea {
        width: 100%;
        padding: 0.5rem;
        border: 1px solid #ccc;
        border-radius: 4px;
        margin-bottom: 0.5rem;
    }
    .title-input {
        font-size: 1.75rem;
        font-weight: 500;
    }
    .desc-textarea {
        min-height: 60px; /* Basic height */
        resize: vertical; /* Allow vertical resize */
        font-family: inherit; /* Match surrounding text */
        line-height: 1.5;
    }
    /* Slightly adjust header alignment */
    .page-header {
        position: relative;
    }
    .edit-button {
        position: absolute;
        top: 0;
        right: 0;
        padding: 0.25rem 0.5rem;
    }
</style> 
<div class="goal-page">
    {#if loading}
        <div class="loading">Loading goal data...</div>
    {:else if error}
        <div class="error">
            <h2>Error</h2>
            <p>{error}</p>
            <a href="{base}/goal">Back to Goals</a>
        </div>
    {:else}
        <div class="edit-container">
            {#if isEditing}
                <div>
                    <input 
                        type="text" 
                        bind:value={editableTitle} 
                        class="title-input"
                        placeholder="Goal Title"
                        disabled={isSaving}
                    />
                    <textarea 
                        bind:value={editableDescription} 
                        class="desc-textarea"
                        placeholder="Goal Description"
                        disabled={isSaving}
                    ></textarea>
                    <div class="edit-actions">
                        <button onclick={saveGoalChanges} disabled={isSaving || !editableTitle.trim()} class="btn btn-success btn-sm">
                            {isSaving ? 'Saving...' : 'Save'}
                        </button>
                        <button onclick={cancelEdit} disabled={isSaving} class="btn btn-secondary btn-sm">Cancel</button>
                    </div>
                    {#if editError}
                        <p class="edit-error">Error: {editError}</p>
                    {/if}
                </div>
            {:else}
                <div class="page-header">
                    <h1>{goal?.title}</h1>
                    <p class="description">{goal?.description}</p>
                    <button onclick={() => { isEditing = true; editError = null; }} class="btn btn-outline-secondary btn-sm edit-button">Edit</button>
                </div>
            {/if}
        </div>

        <div class="spend-data">
            Total Spend: ${spendResponse?.spend?.toFixed(3)} ({spendResponse?.count} runs)
        </div>

        <div class="meta-info">
            <div class="meta-item">
                <strong>Goal ID:</strong> {goal?.UID}
            </div>
            <div class="meta-item">
                <strong>Prompts:</strong> {prompts?.length}
            </div>
        </div>

        <h2>Example</h2>
        <div class="examples">
            <div class="example-panel">
                <div class="item-title">Input</div>
                <pre>{JSON.stringify(goal?.exampleInput, null, 2)}</pre>
            </div>
            <div class="example-panel">
                <div class="item-title">Output</div>
                <pre>{JSON.stringify(goal?.exampleOutput, null, 2)}</pre>
            </div>
        </div>

        <h2>Prompts</h2>
        {#if !prompts || prompts.length === 0}
            <div class="empty-state">
                <p>No prompts found for this goal.</p>
                <button class="btn btn-primary" onclick={openCreatePromptModal}>Create First Prompt</button>
            </div>
        {:else}
            <div class="card-container">
                <button onclick={openCreatePromptModal} class="card new-item-card">
                    <div>+</div>
                    <div>Create New Prompt</div>
                </button>
                {#each prompts as prompt}
                    <div class="prompt-card-wrapper">
                        <PromptCard {prompt} editable={true} />
                    </div>
                {/each}
            </div>
        {/if}
        
        <!-- Logs Section -->
        <div class="item-title">Recent Logs</div>
        {#if logsLoading}
            <div class="loading">Loading logs...</div>
        {:else if logs && logs.length > 0}
            <LogTable logs={logs || []} />
        {:else}
            <div class="no-items">No logs found for this goal</div>
        {/if}
        
        <!-- Debug Info -->
        <details class="goal-debug">
            <summary>Debug Info</summary>
            <FormatJson jsonText={JSON.stringify(goal)} />
        </details>
    {/if}
</div>

<!-- Prompt Modal -->
{#if goal}
    <PromptModal 
        isOpen={promptModalOpen}
        goalUID={goaluid}
        prompt={modalPrompt}
        onClose={() => promptModalOpen = false}
    />
{/if}

