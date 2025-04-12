import { API_URL } from '$env/static/public';

// Types for log filtering and logs
export type LogFilter = {
    minTimestamp?: number;
    maxTimestamp?: number;
    goalUID?: string;
    promptUID?: string;
    includeRaw: boolean;
    limit: number;
    offset: number;
}

export type Log = {
    timestamp: number;
    goalUID: string;
    promptUID: string;
    rawInput: string;
    inputObject: string;
    rawOutput: string;
    outputObject: string;
    inputTokens: number;
    outputTokens: number;
    cost: number;
    requestTime: number;
    generationTime: number;
    error: string;
}

// Pagination information for API responses
export type PaginationResponse = {
    total: number;
    page: number;
    perPage: number;
    totalPages: number;
}

// Combined response structure for log queries
export type LogResponse = {
    logs: Log[];
    pagination: PaginationResponse;
}

class LLMangoLogging {
    private static instance: LLMangoLogging;
    private baseUrl: string;

    private constructor() {
        this.baseUrl = API_URL || `/mango/api`;
    }

    public static getInstance = (): LLMangoLogging => {
        if (!LLMangoLogging.instance) {
            LLMangoLogging.instance = new LLMangoLogging();
        }
        return LLMangoLogging.instance;
    }

    private fetch = async <T>(endpoint: string, options?: RequestInit): Promise<T> => {
        let url: string;
        if (endpoint.startsWith('/')) {
            url = `${this.baseUrl}${endpoint}`;
        } else {
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

    // Get all logs with filter
    getAllLogs = async (filter: LogFilter): Promise<LogResponse> => {
        return this.fetch<LogResponse>('/logs', {
            method: 'POST',
            body: JSON.stringify(filter)
        });
    }

    // Get logs for a specific goal
    getGoalLogs = async (goalUID: string, filter: Omit<LogFilter, 'goalUID'>): Promise<LogResponse> => {
        return this.fetch<LogResponse>(`/logs/goal/${goalUID}`, {
            method: 'POST',
            body: JSON.stringify(filter)
        });
    }

    // Get logs for a specific prompt
    getPromptLogs = async (promptUID: string, filter: Omit<LogFilter, 'promptUID'>): Promise<LogResponse> => {
        return this.fetch<LogResponse>(`/logs/prompt/${promptUID}`, {
            method: 'POST',
            body: JSON.stringify(filter)
        });
    }
}

// Export a singleton instance of the LLMangoLogging class
export const llmangoLogging = LLMangoLogging.getInstance(); 