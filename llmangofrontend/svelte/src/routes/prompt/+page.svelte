<script lang="ts">
    import { onMount } from 'svelte';
    import { llmangoAPI, type Prompt } from '$lib/classes/llmangoAPI.svelte';
    import PromptCard from '$lib/PromptCard.svelte';
    
    let searchTerm: string = $state('');
    let loading: boolean = $state(true);
    
    const filteredPrompts = $derived.by(() => {
        if (!searchTerm.trim()) {
            return Object.values(llmangoAPI.prompts);
        }
        
        const term = searchTerm.toLowerCase();
        return Object.values(llmangoAPI.prompts).filter(prompt => 
            prompt.UID?.toLowerCase().includes(term) || 
            prompt.model?.toLowerCase().includes(term)
        );
    });
    
    onMount(async () => {
        try {
            await llmangoAPI.initialize()
            loading = false;
        } catch (error) {
            console.error('Error loading prompts:', error);
            loading = false;
        }
    });
</script>

<svelte:head>
    <title>Prompts | LLMango</title>
</svelte:head>

<div>
    <h1>Prompts</h1>
    
    <input 
        type="text" 
        class="search-input" 
        bind:value={searchTerm} 
        placeholder="Search prompts by UID or model..." 
    />
    <div class="card-container">

    {#if loading}
        <div class="loading">Loading prompts...</div>
    {:else if filteredPrompts.length === 0}
        <div>No prompts found. {searchTerm ? 'Try a different search term.' : 'Create your first prompt!'}</div>
    {:else}
            {#each filteredPrompts as prompt}
                <PromptCard {prompt}/>
            {/each}
    {/if}
</div>
</div>
