<svelte:head>
    <title>Goals | LLMango</title>
</svelte:head>

<script lang="ts">
    import { onMount } from 'svelte';
    import { llmangoAPI, type Goal } from '$lib/classes/llmangoAPI.svelte';
    import GoalCard from '$lib/GoalCard.svelte';
    
    let searchTerm: string = $state('');
    let loading: boolean = $state(true);
    const filteredGoals = $derived( Object.values(llmangoAPI.goals).filter((goal: Goal) => 
            !searchTerm.trim() || 
            goal.title?.toLowerCase().includes(searchTerm.toLowerCase()) || 
            goal.description?.toLowerCase().includes(searchTerm.toLowerCase()) ||
            goal.UID?.toLowerCase().includes(searchTerm.toLowerCase())
        )
    );
    
    onMount(async () => {
        try {
            await llmangoAPI.initialize()
            console.log("intiialized")
        } catch (error) {
            console.error('Error loading goals:', error);
        }finally{
            loading = false;
        }
    });
</script>


<div>
    <h1>Goals</h1>
    
    <input 
        type="text" 
        class="search-input" 
        bind:value={searchTerm} 
        placeholder="Search goals by name, description or UID..." 
    />
    
    {#if loading}
        <div class="loading">Loading goals...</div>
    {:else if filteredGoals.length === 0}
        <div>No goals found. {searchTerm ? 'Try a different search term.' : 'Create your first goal!'}</div>
    {:else}
        <div class="card-container">
            {#each filteredGoals as goal}
                <GoalCard {goal} />
            {/each}
        </div>
    {/if}
</div>
