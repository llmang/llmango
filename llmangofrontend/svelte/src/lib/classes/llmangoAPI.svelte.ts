import { API_URL } from '$env/static/public';

// Types matching the backend structure
export type Prompt = {
    UID: string;
    goalUID: string;
    model: string;
    parameters: PromptParameters;
    messages: any[];
    weight: number;
    isCanary: boolean;
    maxRuns: number;
    totalRuns: number;
}

// Updated to match our types.ts Goal interface
export type Goal = {
    UID: string;
    prompts: Record<string, string>; // Maps promptUID to promptUID for easy lookup
    title: string;
    description: string;
    inputOutput: { // Changed from exampleInput/exampleOutput directly
        inputExample: any;
        outputExample: any;
    };
}

export class PromptParameters {
    temperature?: number;
    max_tokens?: number;
    top_p?: number;
    frequency_penalty?: number;
    presence_penalty?: number;
    
    constructor(params: Partial<PromptParameters> = {}) {
        this.temperature = params.temperature;
        this.max_tokens = params.max_tokens;
        this.top_p = params.top_p;
        this.frequency_penalty = params.frequency_penalty;
        this.presence_penalty = params.presence_penalty;
    }
    
    // Convert to a plain object, removing undefined values
    toJSON(): Record<string, any> {
        const result: Record<string, any> = {};
        
        if (this.temperature !== undefined) result.temperature = this.temperature;
        if (this.max_tokens !== undefined) result.max_tokens = this.max_tokens;
        if (this.top_p !== undefined) result.top_p = this.top_p;
        if (this.frequency_penalty !== undefined) result.frequency_penalty = this.frequency_penalty;
        if (this.presence_penalty !== undefined) result.presence_penalty = this.presence_penalty;
        
        return result;
    }
    
    // Create from a plain object or existing parameters
    static fromObject(obj: any): PromptParameters {
        if (!obj) return new PromptParameters();
        return new PromptParameters({
            temperature: obj.temperature,
            max_tokens: obj.max_tokens,
            top_p: obj.top_p,
            frequency_penalty: obj.frequency_penalty,
            presence_penalty: obj.presence_penalty
        });
    }
}



class LLMangoAPI{
    prompts = $state<Record<string, Prompt>>({});
    goals = $state<Record<string, Goal>>({});
    promptsByGoalUID = $derived.by<Record<string, Prompt[]>>(() => {
        const result: Record<string, Prompt[]> = {};
        Object.values(this.prompts).forEach((prompt) => {
            if (!result[prompt.goalUID]) {
                result[prompt.goalUID] = [];
            }
            result[prompt.goalUID].push(prompt);
        });
        return result;
    })
    isLoaded = false;
    private baseUrl: string;
    private initializationPromise: Promise<void> | null = null;

    constructor() {
        this.baseUrl = API_URL || "/mango/api";
    }

    async initialize(): Promise<void> {
        if (this.initializationPromise) {
            return this.initializationPromise;
        }
        
        this.initializationPromise = Promise.resolve().then(() => {
            return this.loadAllData(false);
        });
        
        return this.initializationPromise;
    }
    
    async reload(): Promise<void> {
        await this.loadAllData()
    }

    private async loadAllData(force: boolean = false): Promise<void> {
        if (this.isLoaded && !force) return;
        
        try {
            const [goalsResponse, promptsResponse] = await Promise.all([
                fetch(`${this.baseUrl}/goals`),
                fetch(`${this.baseUrl}/prompts`)
            ]);
            
            if (!goalsResponse.ok || !promptsResponse.ok) {
                throw new Error("Failed to fetch data");
            }
            
            const goalsData = await goalsResponse.json(); // Fetch goals data first
            this.goals = goalsData.reduce((acc: Record<string, Goal>, goal: Goal) => { // Reduce into a Record
                acc[goal.UID] = goal;
                return acc;
            }, {});

            const promptsData = await promptsResponse.json();
            this.prompts = promptsData.reduce((acc: Record<string, Prompt>, prompt: Prompt) => {
                acc[prompt.UID] = prompt;
                return acc;
            }, {});

            this.isLoaded = true;
        } catch (error) {
            console.error('Failed to load data:', error);
            throw error;
        }
    }


    updateGoal = async (goalUID: string, title: string, description: string): Promise<void> => {
        const url = `${this.baseUrl}/goal/${goalUID}/update`;
        const updateData = { title, description }; // Send only title and description
        
        const response = await fetch(url, {
            method: 'POST',
            headers: { // Ensure correct content type
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(updateData)
        });
        
        if (!response.ok) {
            throw new Error(`Failed to update goal: ${response.statusText}`);
        }
        
        // Update local state optimistically or after confirmation
        if (this.goals[goalUID]) {
            // Create a new object to trigger reactivity
            this.goals[goalUID] = { 
                ...this.goals[goalUID], 
                title: title, 
                description: description 
            };
        }
    }

    createPrompt = async (prompt: Prompt): Promise<void> => {
        const url = `${this.baseUrl}/prompt/create`;
        const response = await fetch(url, {
            method: 'POST',
            body: JSON.stringify(prompt)
        });
        
        if (!response.ok) {
            throw new Error(`Failed to create prompt: ${response.statusText}`);
        }
        
        this.prompts[prompt.UID] = prompt;
    }

    updatePrompt = async (promptUID: string, prompt: Prompt): Promise<void> => {
        const url = `${this.baseUrl}/prompts/${promptUID}/update`;
        const response = await fetch(url, {
            method: 'POST',
            body: JSON.stringify(prompt)
        });
        
        if (!response.ok) {
            throw new Error(`Failed to update prompt: ${response.statusText}`);
        }
        
        this.prompts[promptUID] = prompt;
    }

    deletePrompt = async (promptUID: string): Promise<void> => {
        const url = `${this.baseUrl}/prompt/delete`;
        const response = await fetch(url, {
            method: 'POST',
            body: JSON.stringify({ promptUID })
        });
        
        if (!response.ok) {
            throw new Error(`Failed to delete prompt: ${response.statusText}`);
        }
        
        delete this.prompts[promptUID];
    }

    updateAPIKey = async (apiKey: string): Promise<void> => {
        const url = `${this.baseUrl}/update-key`;
        const response = await fetch(url, {
            method: 'POST',
            body: JSON.stringify({ apiKey })
        });
        
        if (!response.ok) {
            const error = await response.text();
            throw new Error(`Failed to update API key: ${error}`);
        }
    }
}

const llmangoAPI = new LLMangoAPI();
llmangoAPI.initialize()
export {llmangoAPI}
export default llmangoAPI