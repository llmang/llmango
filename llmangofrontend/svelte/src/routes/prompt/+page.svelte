<script lang="ts">
    import { onMount } from 'svelte';
    import { llmangoAPI, type Prompt } from '$lib/classes/llmangoAPI.svelte';
    import PromptCard from '$lib/PromptCard.svelte';
    import CreatePromptModal from '$lib/CreatePromptModal.svelte';
    
    let prompts: Record<string, Prompt> = $state({});
    let searchTerm: string = $state('');
    let loading: boolean = $state(true);
    let newPromptOpen = $state(false);
    
    const filteredPrompts = $derived.by(() => {
        if (!searchTerm.trim()) {
            return Object.values(prompts);
        }
        
        const term = searchTerm.toLowerCase();
        return Object.values(prompts).filter(prompt => 
            prompt.UID?.toLowerCase().includes(term) || 
            prompt.model?.toLowerCase().includes(term)
        );
    });
    
    onMount(async () => {
        try {
            prompts = await llmangoAPI.getAllPrompts();
            loading = false;
        } catch (error) {
            console.error('Error loading prompts:', error);
            loading = false;
        }
    });

    const newPrompt = () => {
        newPromptOpen = true;
    }

    const handleClose = () => {
        newPromptOpen = false;
    }

    const handleSave = async (prompt: Prompt) => {
        // Refresh the prompts list after saving
        prompts = await llmangoAPI.getAllPrompts();
    }
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
        <button onclick={newPrompt} class="card new-item-card">
            <div>+</div>
            <div>Create New Prompt</div>
        </button>
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

<CreatePromptModal 
    isOpen={newPromptOpen}
    mode="create"
    onClose={handleClose}
    onSave={handleSave}
/>
</div>
