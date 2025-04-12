<script lang="ts">
    import { openrouter } from '$lib/classes/openrouter.svelte';
    import { onMount } from 'svelte';

    let searchQuery = $state('');

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
                <details class="model-card">
                    <summary>
                        <div class="model-header">
                            <h3>{model.name}</h3>
                            <div class="model-date">{new Date(model.created * 1000).toLocaleDateString()}</div>
                        </div>
                        <div class="details-indicator">
                            <span class="details-text">details</span>
                            <span class="expand-icon">▼</span>
                        </div>
                    </summary>
                    
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
                                        <div>Prompt: ${typeof model.pricing.prompt === 'number' ? model.pricing.prompt.toFixed(7) : 'N/A'} / token</div>
                                        <div>Completion: ${typeof model.pricing.completion === 'number' ? model.pricing.completion.toFixed(7) : 'N/A'} / token</div>
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
                </details>
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
        transition: all 0.2s ease;
        position: relative;
    }

    .model-card:hover {
        box-shadow: 0 2px 8px rgba(0,0,0,0.1);
    }

    summary {
        list-style: none;
        cursor: pointer;
        width: 100%;
    }

    summary::-webkit-details-marker {
        display: none;
    }

    .model-header {
        display: flex;
        justify-content: space-between;
        align-items: start;
        margin-bottom: .5rem;
        gap:.5rem;
    }

    .model-header h3 {
        margin: 0;
        font-size: 1.1rem;
    }
    .model-date{
        font-size: .8rem;
        font-weight: 300;
        color:grey;
    }

    .details-indicator {
        display: flex;
        align-items: center;
        justify-content: center;
        gap:.5rem;
        color: #aaa;
    }
    
    .details-text {
        font-size: 0.7rem;
    }

    .expand-icon {
        font-size: 0.8rem;
        transition: transform 0.2s ease;
    }

    details[open] .expand-icon {
        transform: rotate(180deg);
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

    hr {
        border: none;
        border-top: 1px solid #eee;
        margin: 10px 0;
    }
</style>