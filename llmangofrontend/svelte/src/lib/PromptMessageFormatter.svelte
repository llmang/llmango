<script lang="ts">
    import type { Goal } from './classes/llmangoAPI.svelte';

    let { message, goal = null } = $props<{
        message: string;
        goal?: Goal | null;
    }>();

    // RegExp to match {{varname}} pattern
    const variablePattern = /\{\{([^{}]+)\}\}/g;
    
    // Function to validate if a variable exists in the goal's sampleInput
    function isValidVariable(variable: string): boolean {
        if (!goal || !goal.exampleInput) return false;
        
        // Handle nested properties (e.g., "user.name")
        const parts = variable.split('.');
        let current: any = goal.exampleInput;
        
        for (const part of parts) {
            if (current === null || current === undefined || typeof current !== 'object') {
                return false;
            }
            
            if (!(part in current)) {
                return false;
            }
            
            current = current[part];
        }
        
        return true;
    }
    
    // Function to format the message content with highlighted variables
    function formatMessage(text: string): string {
        if (!text) return '';
        
        return text.replace(variablePattern, (match, variableName) => {
            // If no goal is provided, use a neutral color
            if (!goal) {
                return `<span class="neutral-var" title="${variableName}">${match}</span>`;
            }
            
            const isValid = isValidVariable(variableName);
            const colorClass = isValid ? 'valid-var' : 'invalid-var';
            
            return `<span class="${colorClass}" title="${variableName}">${match}</span>`;
        });
    }
    
        let formattedContent = $derived(formatMessage(message))
</script>

<div class="prompt-formatter">
    {@html formattedContent}
</div>

<style>
    .prompt-formatter {
        font-family: monospace;
        white-space: pre-wrap;
        word-break: break-word;
        line-height: 1.5;
    }
    
    :global(.valid-var) {
        color: #28a745;
        font-weight: bold;
        background-color: rgba(40, 167, 69, 0.1);
        padding: 0 2px;
        border-radius: 3px;
    }
    
    :global(.invalid-var) {
        color: #dc3545;
        font-weight: bold;
        background-color: rgba(220, 53, 69, 0.1);
        padding: 0 2px;
        border-radius: 3px;
    }
    
    :global(.neutral-var) {
        color: #0d6efd;
        font-weight: bold;
        background-color: rgba(13, 110, 253, 0.1);
        padding: 0 2px;
        border-radius: 3px;
    }
</style> 