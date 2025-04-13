<script lang="ts">
    import { base } from '$app/paths';
    import Card from './Card.svelte';
    import PromptMessageFormatter from './PromptMessageFormatter.svelte';
    import type { Prompt, Goal } from './classes/llmangoAPI.svelte';
    import { llmangoAPI } from './classes/llmangoAPI.svelte';
    import { onMount } from 'svelte';

    let { prompt, goal } = $props<{
        prompt: Prompt;
        goal?: Goal | null;
    }>();

    onMount(()=>{
        if (!goal && prompt.goalUID && llmangoAPI?.goals[prompt.goalUID]){
            goal=llmangoAPI.goals[prompt.goalUID]
        }
    })
</script>

<Card 
    title={prompt.UID} 
    description={prompt.model} 
    href={`${base}/prompt/${prompt.UID}`}
>
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
</Card>

<style>
    .prompt-info {
        display: flex;
        justify-content: space-between;
        font-size: 0.8rem;
        color: #666;
        margin-top: 0.5rem;
    }

    .info-value {
        background: #e9ecef;
        padding: 0.2rem 0.5rem;
        border-radius: 4px;
    }

    .message-count {
        color: #6c757d;
    }
    
    .message-preview {
        margin-top: 0.5rem;
        margin-bottom: 0.5rem;
    }
    
    .preview-content {
        font-size: 0.85rem;
        color: #333;
        display: -webkit-box;
        -webkit-line-clamp: 4;
        line-clamp:4;
        -webkit-box-orient: vertical;
        overflow: hidden;
        text-overflow: ellipsis;
        max-height: 6em;
        font-family: monospace;
        background-color: #f8f8f8;
        padding: 0.5rem;
        border-radius: 4px;
        margin: 0;
        line-height: 1.5;
        word-break: break-word;
    }
</style>
