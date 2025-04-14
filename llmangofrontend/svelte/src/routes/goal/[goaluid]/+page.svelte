<svelte:head>
    <title>{goal?.title || 'Goal Details'} - LLMango</title>
</svelte:head>

<script lang="ts">
    import FormatJson from '$lib/FormatJson.svelte';
    import type { Goal, Prompt } from '$lib/classes/llmangoAPI.svelte';
    import { llmangoAPI } from '$lib/classes/llmangoAPI.svelte';
    import { onMount } from 'svelte';
    import { page } from '$app/state';
    import { llmangoLogging, type Log } from '$lib/classes/llmangoLogging.svelte';
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
    let logsLoading = $state<boolean>(false);
    // Modal state
    let promptModalOpen = $state<boolean>(false);
    let modalPrompt = $state<Prompt | null>(null);
    let isViewMode = $state<boolean>(false);
    
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
</script>

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
        <div class="page-header">
            <h1>{goal?.title}</h1>
            <p class="description">{goal?.description}</p>
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
</style> 