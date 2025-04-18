    <script lang="ts">
    import FormatJson from '$lib/FormatJson.svelte';
    import type { Prompt, Goal } from '$lib/classes/llmangoAPI.svelte';
    import { llmangoAPI } from '$lib/classes/llmangoAPI.svelte';
    import { onMount } from 'svelte';
    import { page } from '$app/state';
    import Modal from '$lib/Modal.svelte';
    import PromptMessageFormatter from '$lib/PromptMessageFormatter.svelte';
    import { base } from '$app/paths';
    import Card from '$lib/Card.svelte';
    import { llmangoLogging, type Log } from '$lib/classes/llmangoLogging.svelte';
    import LogTable from '$lib/LogTable.svelte';
    import { goto } from '$app/navigation';
    import PromptModal from '$lib/PromptModal.svelte';


    let promptuid: string = $derived(page.params.promptuid);
    let prompt: Prompt | null = $derived(llmangoAPI.prompts[promptuid] || null);
    let goal: Goal | null = $derived(llmangoAPI.goals[prompt?.goalUID] || null);

    // Initial states - initialize with safe defaults
    let error = $state<string | null>(null);
    let loading = $state(true);
    let logs = $state<Log[]>([]);  // Initialize as empty array
    let logsLoading = $state(false);
    
    // Modal states
    let warningModalOpen = $state(false);
    let editModalOpen = $state(false);
    
    // Load data on component mount
    onMount(async () => {
        loading = true;
        try {
            await llmangoAPI.initialize()
            prompt?.goalUID
            logsLoading = true;
            const logsResponse = await llmangoLogging.getPromptLogs(promptuid, {
                includeRaw: true,
                limit: 5,
                offset: 0
            });
            logs = logsResponse?.logs || [];  // Use nullish coalescing to ensure we always have an array
        } catch (e) {
            error = e instanceof Error ? e.message : 'Failed to load data';
        } finally {
            loading = false;
            logsLoading = false;
            
        }
    });
    
    // Show warning before proceeding to edit
    function showWarning() {
        warningModalOpen = true;
    }
    
    // Proceed from warning to edit modal
    function proceedToEdit() {
        warningModalOpen = false;
        editModalOpen = true;
    }
    
    
    // Delete the prompt
    async function deletePrompt() {
        if (!confirm('⚠️ Are you sure you want to delete this prompt? This action cannot be undone and may affect existing solutions.')) {
            return;
        }
        
        try {
            await llmangoAPI.deletePrompt(promptuid);
            // Redirect to prompts list
            
            goto(base+"/prompt")
        } catch (e) {
            error = e instanceof Error ? e.message : 'Failed to delete prompt';
        }
    }

</script>
<div class="prompt-page">
    <div>Prompt page</div>
    {#if loading}
        <div class="loading">Loading prompt data...</div>
    {:else if error}
        <div class="error">{error}</div>
    {:else if prompt}
        <div class="prompt-header">
            <h2>{prompt.UID || 'Untitled Prompt'}</h2>
        </div>

        <div class="card prompt-card">
            <h3>Details</h3>
            
            <!-- Metadata Section -->
            <div class="metadata-container">
                <div class="metadata-section">
                    <h4>Prompt ID</h4>
                    <div class="item-attribute">{prompt.UID}</div>
                </div>
                <div class="metadata-section">
                    <h4>Model</h4>
                    <div class="item-attribute">{prompt.model || 'Default'}</div>
                </div>
            </div>
            
            <!-- Messages Section -->
            <h3>Messages</h3>
            <div class="messages-container">
                {#if prompt.messages && prompt.messages.length > 0}
                    {#each prompt.messages as message, index}
                        <div class="message {message.role}">
                            <div class="message-header">
                                <span class="message-role">{message.role}</span>
                                <span class="message-index">#{index + 1}</span>
                            </div>
                            <div class="message-content">
                                <PromptMessageFormatter message={message.content} goal={goal}/>
                            </div>
                        </div>
                    {/each}
                {:else}
                    <div class="no-messages">No messages defined for this prompt</div>
                {/if}
            </div>
            
            <!-- Parameters Section -->
            {#if prompt.parameters}
                <h3>Parameters</h3>
                <div class="parameters-container">
                    <FormatJson jsonText={JSON.stringify(prompt.parameters)} />
                </div>
            {/if}
            
            <!-- Goal Section -->
            {#if prompt.goalUID}
                <h3>Associated Goal</h3>
                {#if goal}
                    <Card 
                        title={goal.UID} 
                        description={goal.description || 'No description'} 
                        href={`${base}/goal/${goal.UID}`}
                    >
                    <FormatJson jsonText={JSON.stringify(goal.inputOutput.inputExample || "")} />
                    <FormatJson jsonText={JSON.stringify(goal.inputOutput.outputExample || "")} />
                    </Card>
                {:else}
                    <div class="loading-goal">Loading goal information...</div>
                {/if}
            {:else}
                <div class="no-goal">
                    <span class="muted-text">No goal attached to this prompt</span>
                </div>
            {/if}
            
            <!-- Logs Section -->
            <h3>Recent Logs</h3>
            {#if logsLoading}
                <div class="loading-small">Loading logs...</div>
            {:else if logs && logs.length > 0}
                <LogTable logs={logs || []} />
            {:else}
                <div class="no-logs">No logs found for this prompt</div>
            {/if}
            
            <!-- Unsafe Actions Section -->
            <details class="unsafe-actions">
                <summary>⚠️ Unsafe Actions</summary>
                <div class="unsafe-actions-content">
                    <p class="warning-text">These actions may affect data consistency and cannot be undone. Use with caution.</p>
                    <div class="unsafe-buttons">
                        <button 
                            class="btn btn-warning" 
                            onclick={showWarning}
                        >
                            ⚠️ Edit Prompt
                        </button>
                        <button 
                            class="btn btn-danger" 
                            onclick={deletePrompt}
                        >
                            ⚠️ Delete Prompt
                        </button>
                    </div>
                </div>
            </details>
            
            <!-- Debug Info -->
            <details class="prompt-debug">
                <summary>Debug Info</summary>
                <FormatJson jsonText={JSON.stringify(prompt)} />
            </details>
        </div>
        
        <!-- Warning Modal -->
        <Modal 
            isOpen={warningModalOpen} 
            title="Warning" 
            onClose={() => warningModalOpen = false}
        >
            <div class="warning-section">
                <h4>⚠️ Data Consistency Risk</h4>
                <p>Editing this prompt may affect existing solutions and goals that use it. This action cannot be undone.</p>
                <p>Are you sure you want to proceed?</p>
            </div>
            <div class="modal-actions">
                <button
                    type="button"
                    class="btn btn-secondary"
                    onclick={() => warningModalOpen = false}
                >
                    Cancel
                </button>
                <button
                type="button"
                class="btn btn-primary"
                onclick={proceedToEdit}
            >
                ⚠️ Proceed
            </button>
            </div>
        </Modal>
        
        <!-- Edit Prompt Modal -->
        <PromptModal
            goalUID={prompt.goalUID} 
            {prompt}
            isOpen={editModalOpen}
            onClose={() => editModalOpen = false}
        />
    {:else}
    <div>No prompt  but loaded</div>
    {/if}
</div>

<style>
    .prompt-page {
        margin: 1rem 0;
    }
    
    .prompt-header {
        margin-bottom: 1rem;
    }
    
    .prompt-card {
        padding: 1.5rem;
        margin-bottom: 2rem;
        background: white;
        border-radius: 8px;
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
    }
    
    .metadata-container {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 1.5rem;
        margin: 1.5rem 0 2rem 0;
    }
    
    @media (max-width: 768px) {
        .metadata-container {
            grid-template-columns: 1fr;
        }
    }
    
    .metadata-section h4 {
        margin-top: 0;
        margin-bottom: 0.5rem;
        color: #444;
    }
    
    .messages-container {
        display: flex;
        flex-direction: column;
        gap: 1rem;
        margin-bottom: 1rem;
    }

    .message {
        padding: 1rem;
        border-radius: 0.5rem;
        border: 1px solid #e0e0e0;
    }

    .message.user {
        background-color: #f0f7ff;
    }

    .message.assistant {
        background-color: #f7f7f7;
    }

    .message.system {
        background-color: #fff8e1;
    }

    .message-header {
        display: flex;
        justify-content: space-between;
        margin-bottom: 0.5rem;
        padding-bottom: 0.5rem;
        border-bottom: 1px solid #eee;
    }

    .message-role {
        font-weight: bold;
        text-transform: capitalize;
    }

    .message-content {
        white-space: pre-wrap;
        font-family: monospace;
    }
    
    .no-messages {
        color: #666;
        font-style: italic;
        padding: 1rem;
        background-color: #f7f7f7;
        border-radius: 0.5rem;
    }
    
    .parameters-container {
        margin-bottom: 1rem;
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
    
    .item-attribute {
        background-color: #f5f5f5;
        border-radius: 0.25rem;
        padding: 0.75rem;
        font-family: monospace;
        font-size: 0.8rem;
        max-height: 200px;
        overflow: auto;
        white-space: pre-wrap;
    }

    .loading,
    .error {
        padding: 2rem;
        text-align: center;
        background-color: #f9f9f9;
        border-radius: 8px;
        margin: 2rem 0;
    }
    
    .loading-small {
        padding: 1rem;
        text-align: center;
        background-color: #f9f9f9;
        border-radius: 8px;
        margin: 1rem 0;
        color: #666;
        font-style: italic;
    }
    
    .error {
        color: #dc3545;
        background-color: #f8d7da;
    }
    
    .modal-actions {
        display: flex;
        justify-content: flex-end;
        gap: 0.5rem;
        margin-top: 1rem;
    }
    
    .warning-section {
        background-color: #fff3cd;
        border: 1px solid #ffeeba;
        border-radius: 0.5rem;
        padding: 1.5rem;
        margin-bottom: 1rem;
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
    
    .goal-card {
        display: block;
        padding: 1rem;
        margin-bottom: 1rem;
        background-color: #f8f9fa;
        border: 1px solid #e9ecef;
        border-radius: 8px;
        transition: all 0.2s ease;
        text-decoration: none;
        color: inherit;
    }
    
    .goal-card:hover {
        background-color: #e9ecef;
        transform: translateY(-2px);
        box-shadow: 0 4px 8px rgba(0, 0, 0, 0.05);
    }
    
    .goal-description {
        margin: 0.5rem 0;
        color: #666;
    }
    
    .goal-metadata {
        display: flex;
        gap: 0.5rem;
        margin-top: 0.5rem;
    }
    
    .goal-type {
        font-family: monospace;
        font-size: 0.8rem;
        padding: 0.2rem 0.5rem;
        background-color: #e2e6ea;
        border-radius: 4px;
    }
    
    .goal-error {
        padding: 1rem;
        background-color: #ffeaea;
        border: 1px solid #ffcccc;
        border-radius: 6px;
        color: #d63031;
        margin-bottom: 1rem;
    }
    
    .warning-icon {
        margin-right: 0.5rem;
    }
    
    .no-goal {
        padding: 0.5rem 0;
        margin-bottom: 1rem;
    }
    
    .muted-text {
        color: #6c757d;
        font-size: 0.9rem;
        font-style: italic;
    }
    
    .loading-goal {
        padding: 0.75rem;
        background-color: #f5f5f5;
        border-radius: 4px;
        font-style: italic;
        color: #666;
    }

    .badge {
        background: #e9ecef;
        padding: 0.2rem 0.5rem;
        border-radius: 4px;
        font-size: 0.8rem;
        color: #6c757d;
    }

    .goal-info {
        margin-top: 0.5rem;
    }
    
    .no-logs {
        padding: 1rem;
        text-align: center;
        color: #777;
        font-style: italic;
        background-color: #f9f9f9;
        border-radius: 4px;
    }
</style> 