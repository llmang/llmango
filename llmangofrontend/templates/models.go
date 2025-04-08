package templates

// ModelsTemplate contains the template for the models page
const ModelsTemplate = `
{{define "models"}}
{{template "header" .}}

<h2>OpenRouter Models</h2>

<div x-data="{
    searchQuery: '',
    expandedModel: null
}">
    <div class="model-controls" style="margin-bottom: 20px;">
        <div style="display: flex; align-items: center; gap: 10px; margin-bottom: 15px;">
            <input 
                type="text" 
                x-model="searchQuery" 
                placeholder="Search models..." 
                class="form-control" 
                style="max-width: 300px;"
            >
            <button 
                class="btn btn-primary" 
                @click="$store.modelStore.fetchModels(true)" 
                x-bind:disabled="$store.modelStore.loading"
            >
                <span x-show="$store.modelStore.loading">Loading...</span>
                <span x-show="!$store.modelStore.loading">Refresh Models</span>
            </button>
        </div>
        
        <p x-show="$store.modelStore.lastFetched">
            Last updated: <span x-text="new Date($store.modelStore.lastFetched).toLocaleString()"></span>
        </p>
        <p x-show="$store.modelStore.error" style="color: red;" x-text="$store.modelStore.error"></p>
    </div>

    <div x-show="$store.modelStore.loading">Loading models...</div>
    
    <div x-show="!$store.modelStore.loading && $store.modelStore.models.length === 0">
        No models available. Click "Refresh Models" to fetch the latest models.
    </div>

    <div x-show="!$store.modelStore.loading && $store.modelStore.models.length > 0">
        <div class="card-container" style="display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 15px;">
            <template x-for="model in $store.modelStore.filteredModels(searchQuery)" :key="model.id">
                <div 
                    class="card" 
                    style="display: flex; flex-direction: column; cursor: pointer; transition: transform 0.2s, box-shadow 0.2s;"
                    :class="{ 'clickable': true }"
                    @click="expandedModel = expandedModel === model.id ? null : model.id"
                >
                    <!-- Always visible header -->
                    <div style="display: flex; justify-content: space-between; align-items: center;">
                        <h3 class="model-name" style="margin: 0;" x-text="model.name"></h3>
                        <div>
                            <small x-text="new Date(model.created * 1000).toLocaleDateString()"></small>
                        </div>
                    </div>
                    
                    <!-- Expandable details -->
                    <div x-show="expandedModel === model.id" x-transition>
                        <hr style="margin: 10px 0;">
                        
                        <div class="model-meta">
                            <div>
                                <strong>ID:</strong> <code x-text="model.id"></code>
                            </div>
                            <div>
                                <strong>Context Length:</strong> <span x-text="model.context_length.toLocaleString()"></span>
                            </div>
                            <div>
                                <strong>Pricing:</strong> 
                                <div style="margin-left: 10px;">
                                    <div>Prompt: $<span x-text="parseFloat(model.pricing.prompt).toFixed(7)"></span> / token</div>
                                    <div>Completion: $<span x-text="parseFloat(model.pricing.completion).toFixed(7)"></span> / token</div>
                                </div>
                            </div>
                        </div>
                        
                        <div class="model-desc" style="margin: 10px 0;" x-text="model.description || 'No description available'"></div>
                        
                        <div class="model-arch">
                            <div x-show="model.architecture">
                                <strong>Modalities:</strong>
                                <div>
                                    <strong>Input:</strong> 
                                    <span x-text="model.architecture.input_modalities ? model.architecture.input_modalities.join(', ') : 'text'"></span>
                                </div>
                                <div>
                                    <strong>Output:</strong> 
                                    <span x-text="model.architecture.output_modalities ? model.architecture.output_modalities.join(', ') : 'text'"></span>
                                </div>
                            </div>
                        </div>
                    </div>
                    
                    <!-- Expansion indicator -->
                    <div style="text-align: center; margin-top: 5px;">
                        <span x-text="expandedModel === model.id ? '▲ Collapse' : '▼ Details'"></span>
                    </div>
                </div>
            </template>
        </div>
    </div>
</div>

{{template "footer"}}
{{end}}
`
