<script lang="ts">
    import { openrouter } from '$lib/classes/openrouter.svelte';
    import { onMount } from 'svelte';

    let searchQuery = $state('');
    let expandedModel = $state<string | null>(null);

    onMount(async () => {
        await openrouter.initialize();
    });

    $effect(() => {
        if (openrouter.error) {
            console.error('Error loading models:', openrouter.error);
        }
    });

    // Filter models based on search query
    const filteredModels = () => {
        if (!searchQuery) return openrouter.models;
        
        const query = searchQuery.toLowerCase();
        return openrouter.models.filter(model => 
            model.name.toLowerCase().includes(query) || 
            model.id.toLowerCase().includes(query)
        );
    };
</script>

<h1>OpenRouter Models</h1>

<div class="models-container">
    <div class="search-container">
        <div class="search-controls">
            <input 
                type="text" 
                bind:value={searchQuery}
                placeholder="Search models..." 
                class="form-control" 
            />
            <button 
                class="btn btn-primary" 
                onclick={() => openrouter.reload()}
                disabled={openrouter.loading}
            >
                {#if openrouter.loading}
                    Loading...
                {:else}
                    Refresh Models
                {/if}
            </button>
        </div>
        
        {#if openrouter.lastFetched}
            <p>Last updated: {new Date(openrouter.lastFetched).toLocaleString()}</p>
        {/if}
        {#if openrouter.error}
            <p class="error-text">{openrouter.error}</p>
        {/if}
    </div>

    {#if openrouter.loading}
        <div class="loading">Loading models...</div>
    {:else if openrouter.models.length === 0}
        <div class="no-models">
            No models available. Click "Refresh Models" to fetch the latest models.
        </div>
    {:else}
        <div class="models-grid">
            {#each filteredModels() as model (model.id)}
                <div 
                    class="model-card" 
                    class:expanded={expandedModel === model.id}
                    onclick={() => expandedModel = expandedModel === model.id ? null : model.id}
                >
                    <div class="model-header">
                        <h3>{model.name}</h3>
                        <div>
                            <small>{new Date(model.created * 1000).toLocaleDateString()}</small>
                        </div>
                    </div>
                    
                    {#if expandedModel === model.id}
                        <div class="model-details">
                            <hr />
                            
                            <div class="model-info">
                                <div>
                                    <strong>ID:</strong> <code>{model.id}</code>
                                </div>
                                <div>
                                    <strong>Context Length:</strong> {model.context_length?.toLocaleString()}
                                </div>
                                {#if model.pricing}
                                    <div>
                                        <strong>Pricing:</strong> 
                                        <div class="pricing-details">
                                            <div>Prompt: ${model.pricing.prompt?.toFixed(7)} / token</div>
                                            <div>Completion: ${model.pricing.completion?.toFixed(7)} / token</div>
                                        </div>
                                    </div>
                                {/if}
                            </div>
                            
                            <div class="model-description">
                                {model.description || 'No description available'}
                            </div>
                            
                            {#if model.architecture}
                                <div class="model-architecture">
                                    <strong>Modalities:</strong>
                                    <div>
                                        <strong>Input:</strong> 
                                        {model.architecture.input_modalities?.join(', ') || 'text'}
                                    </div>
                                    <div>
                                        <strong>Output:</strong> 
                                        {model.architecture.output_modalities?.join(', ') || 'text'}
                                    </div>
                                </div>
                            {/if}
                        </div>
                    {/if}
                    
                    <div class="expand-indicator">
                        {expandedModel === model.id ? '▲ Collapse' : '▼ Details'}
                    </div>
                </div>
            {/each}
        </div>
    {/if}
</div>

<style>
    .search-container {
        margin-bottom: 20px;
    }

    .search-controls {
        display: flex;
        align-items: center;
        gap: 10px;
        margin-bottom: 15px;
    }

    .form-control {
        padding: 8px 12px;
        border: 1px solid #ddd;
        border-radius: 4px;
        font-size: 1rem;
        max-width: 300px;
    }

    .btn {
        padding: 8px 16px;
        border-radius: 4px;
        border: none;
        cursor: pointer;
        font-size: 1rem;
    }

    .btn-primary {
        background-color: #007bff;
        color: white;
    }

    .btn-primary:disabled {
        opacity: 0.7;
        cursor: not-allowed;
    }

    .error-text {
        color: #dc3545;
        margin-top: 10px;
    }

    .loading {
        text-align: center;
        padding: 20px;
        color: #666;
    }

    .no-models {
        text-align: center;
        padding: 20px;
        color: #666;
    }

    /* Model Grid */
    .models-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
        gap: 15px;
    }

    /* Model Card */
    .model-card {
        background: white;
        border: 1px solid #eee;
        border-radius: 8px;
        padding: 15px;
        cursor: pointer;
        transition: all 0.2s ease;
    }

    .model-card:hover {
        box-shadow: 0 2px 8px rgba(0,0,0,0.1);
    }

    .model-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
    }

    .model-header h3 {
        margin: 0;
        font-size: 1.1rem;
    }

    .model-details {
        margin-top: 10px;
    }

    .model-info {
        margin: 10px 0;
    }

    .model-info code {
        background: #f5f5f5;
        padding: 2px 4px;
        border-radius: 3px;
    }

    .pricing-details {
        margin-left: 10px;
    }

    .model-description {
        margin: 10px 0;
        color: #666;
    }

    .model-architecture {
        margin-top: 10px;
    }

    .expand-indicator {
        text-align: center;
        margin-top: 10px;
        color: #666;
        font-size: 0.9rem;
    }

    hr {
        border: none;
        border-top: 1px solid #eee;
        margin: 10px 0;
    }
</style>