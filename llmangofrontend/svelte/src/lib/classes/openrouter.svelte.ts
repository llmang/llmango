export interface OpenRouterModel {
    id: string;
    name: string;
    created: number;
    description?: string;
    context_length?: number;
    pricing?: {
        prompt?: number;
        completion?: number;
    };
    architecture?: {
        input_modalities?: string[];
        output_modalities?: string[];
    };
    [key: string]: any; // For additional properties
}

class OpenRouter {
    models = $state<OpenRouterModel[]>([]);
    modelsMap = $state<Record<string, OpenRouterModel>>({});
    loading = $state(false);
    error = $state<string | null>(null);
    lastFetched = $state<string | null>(null);
    hasModels = $state(false);
    initialized = $state<Promise<boolean> | null>(null);
    
    // Initialize the store with cached data and optionally fetch fresh data
    initialize = async (): Promise<boolean> => {
        // Return existing initialization promise if already in progress
        if (this.initialized) {
            return this.initialized;
        }
        
        // Create a new initialization promise
        this.initialized = this.#initialize();
        return this.initialized;
    }
    
    #initialize = async (): Promise<boolean> => {
        // Try to load from cache first
        const cached = localStorage.getItem('openrouter_models');
        
        if (cached) {
            try {
                const data = JSON.parse(cached);
                this.models = data.models || [];
                this.lastFetched = data.lastFetched;
                this.hasModels = (data.models || []).length > 0;
                // Build the models map
                this.#buildModelsMap();
            } catch (err) {
                console.error('Error parsing cached models:', err);
            }
        }

        // If we have models and the cache isn't stale, return early
        if (this.models.length > 0 && !this.#isCacheStale()) {
            return true;
        }
        
        // Otherwise, fetch fresh data
        return this.reload();
    }
    
    // Check if the cache is stale (older than 24 hours)
    #isCacheStale = (): boolean => {
        if (!this.lastFetched) return true;
        
        const lastFetchedDate = new Date(this.lastFetched);
        const now = new Date();
        // Check if last fetch was more than 24 hours ago
        return (now.getTime() - lastFetchedDate.getTime()) > (24 * 60 * 60 * 1000);
    }
    
    // Build the models map for efficient lookup
    #buildModelsMap = (): void => {
        const map: Record<string, OpenRouterModel> = {};
        for (const model of this.models) {
            map[model.id] = model;
        }
        this.modelsMap = map;
    }
    
    // Load models (initialize if needed)
    load = async (): Promise<OpenRouterModel[]> => {
        await this.initialize();
        return this.models;
    }
    
    // Forcefully reload/refetch models from OpenRouter API
    reload = async (): Promise<boolean> => {
        if (this.loading) {
            return false; // Already loading
        }

        this.loading = true;
        this.error = null;

        try {
            const response = await fetch('https://openrouter.ai/api/v1/models');
            if (!response.ok) {
                throw new Error('Failed to fetch models: ' + response.status);
            }
            
            const data = await response.json();
            this.models = data.data || [];
            this.hasModels = this.models.length > 0;
            this.lastFetched = new Date().toISOString();
            
            // Build the models map
            this.#buildModelsMap();
            
            // Save to localStorage for caching
            localStorage.setItem('openrouter_models', JSON.stringify({
                models: this.models,
                lastFetched: this.lastFetched
            }));
            
            return true;
        } catch (err) {
            this.error = err instanceof Error ? err.message : String(err);
            console.error('Error fetching models:', err);
            return false;
        } finally {
            this.loading = false;
        }
    }
    
    // Filter models based on a search query
    filterModels = (query = ''): OpenRouterModel[] => {
        if (!query) return [...this.models].sort((a, b) => b.created - a.created);
        
        const lowerQuery = query.toLowerCase();
        return this.models
            .filter(model => 
                model.id.toLowerCase().includes(lowerQuery) || 
                model.name.toLowerCase().includes(lowerQuery)
            )
            .sort((a, b) => b.created - a.created);
    }

    // Check if a model ID exists
    hasModel = (modelId: string): boolean => {
        return modelId in this.modelsMap;
    }

    // Get a model by ID
    getModel = (modelId: string): OpenRouterModel | undefined => {
        return this.modelsMap[modelId];
    }
}

// Create and export a singleton instance
export const openrouter = new OpenRouter();
