<script lang="ts">
    import { base } from '$app/paths';
    import Card from './Card.svelte';
    import type { Goal } from './classes/llmangoAPI.svelte';
    import FormatJson from './FormatJson.svelte';

    let { goal } = $props<{
        goal: Goal;
    }>();
</script>

<Card 
    title={goal?.title || 'Untitled Goal'} 
    description={goal?.description || 'No description'} 
    href={`${base}/goal/${goal?.UID || ''}`}
>
    <div class="goal-info uid">UID: {goal?.UID || 'unknown'}</div>
    <div class="goal-info solutions-count">
        {goal?.solutions ? Object.keys(goal.solutions).length : 0} solutions
    </div>
    <details class="debug-details" onclick={(e)=>{e.stopPropagation()}}>
        <summary>Debug Info</summary>
        <FormatJson jsonText={JSON.stringify(goal, null, 2)}></FormatJson>
    </details>
</Card>

<style>
    .goal-info {
        font-size: 0.8rem;
        margin-bottom: .5em;
        color: #666;
    }

    .uid {
        background: #f1f3f5;
        padding: 0.2rem 0.5rem;
        border-radius: 4px;
        font-family: monospace;
    }

    .solutions-count {
        background: #e9ecef;
        padding: 0.2rem 0.5rem;
        border-radius: 4px;
    }

    .debug-details {
        margin-top: 1rem;
        border-top: 1px solid #eee;
        padding-top: 0.5rem;
    }
    
    .debug-details summary {
        cursor: pointer;
        color: #666;
        font-size: 0.8rem;
    }
    

</style>
