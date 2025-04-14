<script lang="ts">
    import { base } from '$app/paths';
    import Card from './Card.svelte';
    import { llmangoAPI, type Goal } from './classes/llmangoAPI.svelte';
    import FormatJson from './FormatJson.svelte';

    let { goal } = $props<{
        goal: Goal;
    }>();

    function getGoalStatus() {
        const prompts = llmangoAPI.promptsByGoalUID[goal?.UID] || [];
        
        if (prompts.length === 0) {
            return { status: 'Inactive', color: 'var(--color-secondary)' };
        }
        
        const hasRunning = prompts.some(p => p.weight > 0 && (!p.isCanary || (p.isCanary && p.maxRuns > (p.totalRuns || 0))));
        const hasCanaries = prompts.some(p => p.isCanary);
        const allStopped = prompts.every(p => p.weight === 0 || (p.isCanary && p.maxRuns <= (p.totalRuns || 0)));
        
        if (hasRunning) {
            return { status: hasCanaries ? 'Active (Canary)' : 'Active', color: 'var(--color-success)' };
        } else if (allStopped) {
            return { status: 'Inactive', color: 'var(--color-danger)' };
        }
        
        return { status: 'Unknown', color: 'var(--color-secondary)' };
    }

    const status = $derived(getGoalStatus());
</script>

<Card 
    title={goal?.title || 'Untitled Goal'} 
    description={goal?.description || 'No description'} 
    href={`${base}/goal/${goal?.UID || ''}`}
>
    <span class="status-badge" style="background-color: color-mix(in srgb, {status.color} 70%, white 30%)">
        <span class="status-dot" style="background-color: {status.color}"></span>
        {status.status}
    </span>
    
    <div class="goal-info uid">UID: {goal?.UID || 'unknown'}</div>
    <div class="bottom-container">
        <div>prompts:{llmangoAPI.promptsByGoalUID[goal?.UID]?.length || 0}</div>
    </div>
</Card>

<style>
    .goal-info {
        font-size: 0.8rem;
        margin-bottom: .5em;
        color: var(--color-text-secondary);
    }

    .uid {
        background: var(--color-bg-secondary);
        padding: 0.2rem 0.5rem;
        border-radius: 4px;
        font-family: monospace;
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
        background: rgba(128, 128, 128, 0.165);
        border-radius: 50%;
        display: inline-block;
    }
    
    .bottom-container {
        display: flex;
        justify-content: space-between;
    }
</style>
