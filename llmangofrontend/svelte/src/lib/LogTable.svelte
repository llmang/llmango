<script lang="ts">
    import type { Log } from './classes/llmangoLogging.svelte';
    import LogComponent from './Log.svelte';
    
    let { logs = [], cells = 0 } = $props<{ logs: Log[], cells?: number }>();

    // Create empty cells array if needed
    let emptyCells = $derived.by(() => {
        const emptyCount = Math.max(0, cells - logs.length);
        return Array(emptyCount).fill(null);
    });
</script>
<div class="logs-container">
    <div class="log-table">
        <div class="log-header-row">
            <div class="log-cell">Time</div>
            <div class="log-cell">Goal ID</div>
            <div class="log-cell">Prompt ID</div>
            <div class="log-cell">Tokens</div>
            <div class="log-cell">Cost</div>
            <div class="log-cell">Actions</div>
        </div>
        
        {#each logs as log}
            <LogComponent {log} />
        {/each}
        
        {#each emptyCells as _, i (i)}
            <div class="log-row empty-row">
                <div class="log-cell"></div>
                <div class="log-cell"></div>
                <div class="log-cell"></div>
                <div class="log-cell"></div>
                <div class="log-cell"></div>
                <div class="log-cell"></div>
            </div>
        {/each}
    </div>
    
    {#if logs.length === 0}
        <p class="no-logs" >No logs found. Try adjusting your filters.</p>
    {/if}
</div>


<style>
    .logs-container {
        margin-bottom: 20px;
        position: relative;
    }
    
    .log-table {
        width: 100%;
        border-collapse: collapse;
    }
    
    .log-header-row {
        display: flex;
        padding: 0.5rem;
        background-color: #f5f5f5;
        font-weight: bold;
        border-bottom: 2px solid #ddd;
    }
    
    .log-cell {
        flex: 1;
        padding: 0 0.5rem;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }
    
    .log-row {
        display: flex;
        padding: 0.5rem;
        border-bottom: 1px solid #ddd;
    }
    
    .empty-row {
        background-color: #fafafa;
        color: transparent;
        height: 2.5rem;
    }
    
    .no-logs {
        position: absolute;
        top: 50%;
        left: 50%;
        transform: translate(-50%, -50%);
        padding: 20px;
        color: #666;
        font-style: italic;
        background-color: rgba(255, 255, 255, 0.8);
        border-radius: 4px;
        z-index: 1;
    }
</style> 