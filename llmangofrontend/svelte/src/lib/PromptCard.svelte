<script lang="ts">
    import { base } from '$app/paths';
    import Card from './Card.svelte';
    import PromptMessageFormatter from './PromptMessageFormatter.svelte';
    import type { Prompt, Goal } from './classes/llmangoAPI.svelte';
    import { llmangoAPI } from './classes/llmangoAPI.svelte';
    import { onMount } from 'svelte';
    import PromptModal from './PromptModal.svelte';
    import StopPropigation from './StopPropigation.svelte';

    let { prompt, goal, editable = false } = $props<{
        prompt: Prompt;
        goal?: Goal | null;
        editable?: boolean;
    }>();

    let isModalOpen = $state(false);

    onMount(()=>{
        if (!goal && prompt.goalUID && llmangoAPI?.goals[prompt.goalUID]){
        goal=llmangoAPI.goals[prompt.goalUID]
        }
    });

    function getPromptStatus(): { status: string; color: string } {
        if (prompt.weight === 0) {
            return { status: 'Stopped', color: 'var(--color-danger)' };
        }
        
        if (prompt.isCanary) {
            if (prompt.maxRuns === 0) {
                return { status: 'Stopped', color: 'var(--color-danger)' };
            }
            if (prompt.maxRuns <= (prompt.totalRuns || 0)) {
                return { status: 'Finished', color: 'var(--color-success)' };
            }
            return { status: 'Running', color: 'var(--color-success)' };
        }
        
        if (prompt.weight > 0) {
            return { status: 'Running', color: 'var(--color-success)' };
        }
        return { status: 'Unknown', color: 'var(--color-secondary)' };
    }

    const status = $derived(getPromptStatus());
</script>

<style>
    .weight-badge {
        padding: 0.2rem 0.5rem;
        font-size: 0.8rem;
        color: var(--color-text-secondary);
        font-weight: 800;
    }

    .status-badge {        
        position: absolute;
        top: .75rem;
        right: .75rem;
        z-index: 1;
        color: white;
        padding: 0.2rem 0.5rem;
        border-radius: 4px;
        font-weight: 500;
        display: flex;
        align-items: center;
        gap: 0.3rem;
        font-size: 0.8rem;
    }

    .status-dot {
        width: .5rem;
        height: .5rem;
        background:rgba(128, 128, 128, 0.165);
        border-radius: 50%;
        display: inline-block;
    }

    .prompt-info {
        display: flex;
        justify-content: space-between;
        font-size: 0.8rem;
        color: var(--color-text-secondary);
        margin-bottom: 0.5rem;
    }

    .info-value {
        background: var(--color-bg-secondary);
        padding: 0.2rem 0.5rem;
        border-radius: 4px;
    }

    .message-count {
        color: var(--color-secondary);
    }
    
    .message-preview {
        margin-bottom: 0.5rem;
        margin-bottom: 0.5rem;
    }
    
    .preview-content {
        font-size: 0.85rem;
        color: var(--color-text-primary);
        display: -webkit-box;
        -webkit-line-clamp: 4;
        line-clamp:4;
        -webkit-box-orient: vertical;
        overflow: hidden;
        text-overflow: ellipsis;
        max-height: 6em;
        font-family: monospace;
        background-color: var(--color-bg-light);
        padding: 0.5rem;
        border-radius: 4px;
        margin: 0;
        line-height: 1.5;
        word-break: break-word;
    }

    .runs-count {
        font-size: 0.8rem;
        color: var(--color-secondary);
    }
    .bottom-container{
        display: flex;
        justify-content: space-between;
    }

    .edit-button {
        background-color: var(--color-primary);
        color: white;
        border: none;
        padding: 0.5em;
        min-width: 5em;
        font-size:1rem;
        border-radius: 4px;
        cursor: pointer;
        transition: background-color 0.2s;
    }

    .edit-button:hover {
        background-color: var(--color-primary-dark);
    }
</style>
<Card 
    title={prompt.UID} 
    description={prompt.model} 
    href={`${base}/prompt/${prompt.UID}`}
    >
        <span class="status-badge" style="background-color: color-mix(in srgb, {status.color} 70%, white 30%)">
            <span class="status-dot" style="background-color: {status.color}"></span>
            {status.status}
        </span>
        
    <div class="message-preview">
        {#if prompt.messages && prompt.messages.length > 0}
            <div class="preview-content">
                <PromptMessageFormatter 
                    message={prompt.messages[0].content} 
                    goal={goal} 
                />
            </div>
        {:else}
            <p class="preview-content">No messages</p>
        {/if}
    </div>
    
    <div class="prompt-info">
        <span class="info-value">{prompt.model}</span>
        <span class="message-count">{prompt.messages.length} messages</span>
    </div>
    
    {#if prompt.goalUID}
        <div class="prompt-info">
            <span class="info-value">Goal: {prompt.goalUID}</span>
        </div>
    {/if}
    <div class="bottom-container">
        <div>
            <span class="runs-count">
                Runs: {prompt.totalRuns || 0}
                {#if prompt.isCanary}
                    / {prompt.maxRuns}
                {/if}
            </span>
            <span class="weight-badge">Weight: {prompt.weight}</span>
        </div>
    {#if editable}
        <div class="edit-button-container">
            <StopPropigation>
                <button class="edit-button" onclick={() => isModalOpen = true}>
                    Edit
                </button>
            </StopPropigation>
        </div>
    {/if}
</div>
</Card>

{#if isModalOpen}
    <PromptModal 
        isOpen={isModalOpen}
        goalUID={prompt.goalUID}
        prompt={prompt}
        onClose={() => isModalOpen = false}
        onSave={(updatedPrompt) => {
            prompt = updatedPrompt;
            isModalOpen = false;
        }}
    />
{/if}

