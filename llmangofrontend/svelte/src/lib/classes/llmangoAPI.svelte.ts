import { API_URL } from '$env/static/public';

// Types matching the backend structure
export type Prompt = {
    UID: string;
    goalUID:string;
    model: string;
    parameters: PromptParameters;
    messages: any[];
}

export type Solution = {
    promptUID: string;
    weight: number;
    isCanary: boolean;
    maxRuns: number;
    totalRuns: number;
}

// Updated to match our types.ts Goal interface
export type Goal = {
    UID: string;
    solutions: Record<string, Solution>;
    title: string;
    description: string;
    exampleInput: any;
    exampleOutput: any;
}

export class PromptParameters {
    temperature?: number;
    max_tokens?: number;
    top_p?: number;
    frequency_penalty?: number;
    presence_penalty?: number;
    
    constructor(params: Partial<PromptParameters> = {}) {
        this.temperature = params.temperature !== undefined ? params.temperature : 0.7;
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

class LLMangoAPI {
    private static instance: LLMangoAPI;
    private baseUrl: string;
    private goals: Record<string, Goal> = {};
    private prompts: Record<string, Prompt> = {};

    private constructor() {
        this.baseUrl = API_URL || "/mango/api";
    }

    public static getInstance = (): LLMangoAPI => {
        if (!LLMangoAPI.instance) {
            LLMangoAPI.instance = new LLMangoAPI();
        }
        return LLMangoAPI.instance;
    }

    private fetch = async <T>(endpoint: string, options?: RequestInit): Promise<T> => {
        let url: string;
        if (endpoint.startsWith('/')) {
            // If endpoint starts with /, append it to the baseUrl
            url = `${this.baseUrl}${endpoint}`;
        } else {
            // Otherwise, join with a /
            url = `${this.baseUrl}/${endpoint}`;
        }

        try {
            const response = await fetch(url, {
                ...options,
                credentials: 'include',
                headers: {
                    'Content-Type': 'application/json',
                    ...options?.headers
                }
            });

            if (!response.ok) {
                throw new Error(response.statusText || `Request failed with status ${response.status}`);
            }

            const data = await response.json();
            return data as T;
        } catch (error) {
            console.error(`API request failed: ${endpoint}`, error);
            throw error;
        }
    }

    // Goals API
    getAllGoals = async (refresh: boolean = false): Promise<Record<string, Goal>> => {
        if (refresh || !this.goals || Object.keys(this.goals).length === 0) {
            try {
                this.goals = await this.fetch<Record<string, Goal>>('/goals');
            } catch (error) {
                console.error('Failed to fetch goals:', error);
                throw error;
            }
        }
        return this.goals;
    }

    getGoal = async (goalUID: string): Promise<Goal | null> => {
        try {
            if (this.goals[goalUID]) {
                return this.goals[goalUID];
            }
            const goal = await this.fetch<Goal>(`/goal/${goalUID}`);
            if (!goal) {
                return null;
            }
            // Ensure solutions is initialized
            if (!goal.solutions) {
                goal.solutions = {};
            }
            if (this.goals && Object.keys(this.goals).length > 0) {
                this.goals[goalUID] = goal;
            }
            return goal;
        } catch (error) {
            console.error(`Failed to fetch goal ${goalUID}:`, error);
            return null;
        }
    }

    updateGoal = async (goalUID: string, Goal: Partial<Goal>): Promise<void> => {
        await this.fetch(`/goal/${goalUID}/update`, {
            method: 'POST',
            body: JSON.stringify(Goal)
        });
        if (this.goals[goalUID]) {
            this.goals[goalUID] = Object.assign({}, this.goals[goalUID], Goal);
        }
    }

    // Prompts API
    getAllPrompts = async (refresh: boolean = false): Promise<Record<string, Prompt>> => {
        if (refresh || !this.prompts || Object.keys(this.prompts).length === 0) {
            this.prompts = await this.fetch<Record<string, Prompt>>('/prompts');
        }
        return this.prompts;
    }
    getPrompt = async (promptUID: string): Promise<Prompt> => {
        if (this.prompts[promptUID]) {
            return this.prompts[promptUID];
        }
        const prompt = await this.fetch<Prompt>(`/prompts/${promptUID}`);
        if (this.prompts && Object.keys(this.prompts).length > 0) {
            this.prompts[promptUID] = prompt;
        }
        return prompt;
    }

    createPrompt = async (prompt: Prompt): Promise<void> => {
        await this.fetch('/prompt/create', {
            method: 'POST',
            body: JSON.stringify(prompt)
        });
        this.prompts[prompt.UID] = prompt;
    }

    updatePrompt = async (promptUID: string, prompt: Prompt): Promise<void> => {
        await this.fetch(`/prompts/${promptUID}/update`, {
            method: 'POST',
            body: JSON.stringify(prompt)
        });
        this.prompts[promptUID] = prompt;
    }

    deletePrompt = async (promptUID: string): Promise<void> => {
        await this.fetch('/prompt/delete', {
            method: 'POST',
            body: JSON.stringify({ promptUID })
        });
        delete this.prompts[promptUID];
    }

    // Logs API methods moved to llmangoLogging.svelte.ts

    // Solutions API
    createSolution = async (goalUID: string, solution: Solution): Promise<void> => {
        await this.fetch('/solution/create', {
            method: 'POST',
            body: JSON.stringify({ goalUID, solution })
        });
        if (this.goals[goalUID]) {
            this.goals[goalUID].solutions[solution.promptUID] = solution;
        }
    }

    updateSolution = async (solutionUID: string, solution: Solution): Promise<void> => {
        await this.fetch(`/solutions/${solutionUID}/update`, {
            method: 'POST',
            body: JSON.stringify(solution)
        });
        // Find and update the solution in the relevant goal
        for (const goal of Object.values(this.goals)) {
            if (goal.solutions[solutionUID]) {
                goal.solutions[solutionUID] = solution;
                break;
            }
        }
    }

    deleteSolution = async (solutionUID: string): Promise<void> => {
        await this.fetch(`/solutions/${solutionUID}/delete`, {
            method: 'POST'
        });
        // Find and delete the solution from the relevant goal
        for (const goal of Object.values(this.goals)) {
            if (goal.solutions[solutionUID]) {
                delete goal.solutions[solutionUID];
                break;
            }
        }
    }
    // API Key Management
    updateAPIKey = async (apiKey: string): Promise<void> => {
        const response = await this.fetch<Response>('/update-key', {
            method: 'POST',
            body: JSON.stringify({ apiKey })
        });
        
        if (!response.ok) {
            const error = await response.text();
            throw new Error(`Failed to update API key: ${error}`);
        }
    }
}

// Export a singleton instance of the LLMangoAPI class
export const llmangoAPI = LLMangoAPI.getInstance();