<script lang="ts">
    import type { Log } from './classes/llmangoLogging.svelte';
    import { slide } from 'svelte/transition';
    
    let { log } = $props<{
        log: Log;
    }>();
    
    let expanded = $state(false);
    
    function toggleExpanded() {
        expanded = !expanded;
    }
</script>

<div class="log-container">
    <div class="log-row">
        <div class="log-cell" title={new Date(log.timestamp * 1000).toLocaleString()}>
            {new Date(log.timestamp * 1000).toLocaleString('en-GB', {
                day: '2-digit', 
                month: '2-digit', 
                year: '2-digit', 
                hour: '2-digit', 
                minute: '2-digit', 
                second: '2-digit'
            })}
        </div>
        <div class="log-cell" title={log.goalUID}>
            {log.goalUID.substring(0, 10)}{log.goalUID.length > 10 ? '...' : ''}
        </div>
        <div class="log-cell" title={log.promptUID}>
            {log.promptUID.substring(0, 10)}{log.promptUID.length > 10 ? '...' : ''}
        </div>
        <div class="log-cell" title={`Input: ${log.inputTokens} tokens, Output: ${log.outputTokens} tokens`}>
            <span class="token-count" title={`Input Tokens: ${log.inputTokens}`}>{log.inputTokens}</span> / 
            <span class="token-count" title={`Output Tokens: ${log.outputTokens}`}>{log.outputTokens}</span>
        </div>
        <div class="log-cell" title={`Total cost: $${log.cost}`}>
            ${log.cost}
        </div>
        <div class="log-cell">
            <button 
                class="btn btn-sm btn-secondary"
                onclick={toggleExpanded}
            >
                {expanded ? 'Hide' : 'Show'} Details
            </button>
        </div>
    </div>
    
    {#if expanded}
        <div class="log-details" transition:slide>
            <div class="details-row">
                <div class="details-cell">
                    <h5>Goal ID</h5>
                    <pre>{log.goalUID}</pre>
                </div>
                <div class="details-cell">
                    <h5>Prompt ID</h5>
                    <pre>{log.promptUID}</pre>
                </div>
            </div>
            <div class="log-section">
                <h5>Input</h5>
                <pre>{log.inputObject}</pre>
            </div>
            <div class="log-section">
                <h5>Output</h5>
                <pre>{log.outputObject}</pre>
            </div>
            {#if log.error}
                <div class="log-section error">
                    <h5>Error</h5>
                    <pre>{log.error}</pre>
                </div>
            {/if}
        </div>
    {/if}
</div>

<style>
    .log-container {
        border-bottom: 1px solid #eee;
    }
    
    .log-row {
        display: flex;
        padding: 0.5rem;
        align-items: center;
        font-size: 0.85rem;
    }
    
    .log-cell {
        flex: 1;
        padding: 0 0.5rem;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }
    
    .token-count {
        display: inline-block;
        min-width: 2.5rem;
        text-align: right;
    }
    
    .log-details {
        width: 100%;
        padding: 1rem;
        background-color: #f8f8f8;
        border-top: 1px solid #ddd;
        font-size: .8em;
    }
    
    .details-row {
        display: flex;
        gap: 1rem;
        margin-bottom: 1rem;
    }
    
    .details-cell {
        flex: 1;
    }
    
    .details-cell h5 {
        margin: 0 0 0.5rem 0;
        color: #666;
    }
    
    .details-cell pre {
        background-color: #fff;
        padding: 0.5rem;
        border-radius: 0.25rem;
        overflow-x: auto;
        margin: 0;
        white-space: pre-wrap;
        word-wrap: break-word;
    }
    
    .log-section {
        margin-bottom: 1rem;
    }
    
    .log-section:last-child {
        margin-bottom: 0;
    }
    
    .log-section h5 {
        margin: 0 0 0.5rem 0;
        color: #666;
    }
    
    .log-section pre {
        background-color: #fff;
        padding: 0.5rem;
        border-radius: 0.25rem;
        overflow-x: auto;
        margin: 0;
        white-space: pre-wrap;
        word-wrap: break-word;
    }
    
    .log-section.error {
        color: #dc3545;
    }
    
    button {
        padding: 0.25rem 0.5rem;
        font-size: 0.75rem;
        border-radius: 0.25rem;
        background-color: #6c757d;
        color: white;
        border: none;
        cursor: pointer;
    }
    
    button:hover {
        background-color: #5a6268;
    }
</style>