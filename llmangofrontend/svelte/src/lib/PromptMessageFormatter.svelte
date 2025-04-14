<script lang="ts">
    import type { Goal } from './classes/llmangoAPI.svelte';

    let { message, goal = null } = $props<{
        message: string;
        goal?: Goal | null;
    }>();

    // RegExp to match {{varname}} pattern (excluding if statements)
    const variablePattern = /\{\{(?!#if|\/if)([^{}]+)\}\}/g;
    
    // RegExp to match {{#if varname}}...{{/if}} pattern
    const ifStatementPattern = /\{\{#if\s+([^{}]+)\}\}([\s\S]*?)\{\{\/if\}\}/g;
    
    // Track collapsed state of if blocks
    const collapsedBlocks = $state<Record<number, boolean>>({});
    
    // Define types for our blocks
    type TextBlock = {
        type: 'text';
        content: string;
    };
    
    type IfBlock = {
        type: 'ifBlock';
        id: number;
        condition: string;
        content: string;
        colorClass: string;
        collapsed: boolean;
    };
    
    type Block = TextBlock | IfBlock;
    
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
    
    // Function to format the message content with highlighted variables and if statements
    function formatMessageWithBlocks(): Block[] {
        if (!message) return [];
        
        const blocks: Block[] = [];
        let lastIndex = 0;
        let match;
        let blockCounter = 0;
        
        // Reset the regex
        ifStatementPattern.lastIndex = 0;
        
        // Process if statements first
        while ((match = ifStatementPattern.exec(message)) !== null) {
            // Add text before the match
            if (match.index > lastIndex) {
                const textBefore = message.substring(lastIndex, match.index);
                blocks.push({
                    type: 'text',
                    content: formatVariables(textBefore)
                });
            }
            
            const [fullMatch, conditionVar, content] = match;
            const isValid = isValidVariable(conditionVar);
            const colorClass = isValid ? 'valid-condition' : 'invalid-condition';
            const id = blockCounter++;
            
            blocks.push({
                type: 'ifBlock',
                id,
                condition: conditionVar,
                content: formatVariables(content),
                colorClass,
                collapsed: collapsedBlocks[id] || false
            });
            
            lastIndex = match.index + fullMatch.length;
        }
        
        // Add remaining text
        if (lastIndex < message.length) {
            const textAfter = message.substring(lastIndex);
            blocks.push({
                type: 'text',
                content: formatVariables(textAfter)
            });
        }
        
        return blocks;
    }
    
    // Format variables in text
    function formatVariables(text: string) {
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
    
    // Toggle collapse state
    function toggleBlock(id: number) {
        collapsedBlocks[id] = !collapsedBlocks[id];
    }
    
    let blocks = $derived(formatMessageWithBlocks());
</script>

<div class="prompt-formatter">
    {#each blocks as block}
        {#if block.type === 'text'}
            <span>{@html block.content}</span>
        {:else if block.type === 'ifBlock'}
            <div class="if-block">
                <div class="if-header">
                    <button class="toggle-arrow" onclick={() => toggleBlock(block.id)}>
                        {block.collapsed ? '▶' : '▼'}
                    </button>
                    <span class="{block.colorClass}">&#123;&#123;#if {block.condition}&#125;&#125;</span>
                </div
                >{#if !block.collapsed}
                    <div class="if-content">
                        {@html block.content}
                    </div>
                {/if}<div class="{block.colorClass}">&#123;&#123;/if&#125;&#125;</div>
            </div>
        {/if}
    {/each}
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
    
    .valid-condition {
        color: #6f42c1;
        font-weight: bold;
        background-color: rgba(111, 66, 193, 0.1);
        padding: 2px 4px;
        border-radius: 3px;
        display: inline-block;
        width: 100%;
    }
    
    .invalid-condition {
        color: #dc3545;
        font-weight: bold;
        background-color: rgba(220, 53, 69, 0.1);
        padding: 2px 4px;
        border-radius: 3px;
        display: inline-block;
    }
    
    .if-block {
        margin: 2px 0;
        border-left: 2px solid #6f42c1;
    }
    
    .if-header {
        display: flex;
        align-items: center;
    }
    
    .toggle-arrow {
        cursor: pointer;
        margin-right: 4px;
        background:
rgba(128, 128, 128, 0.436);
  border-radius:.5em;
  aspect-ratio: 1/1;
  height: 1.5em;
  display: flex;
  justify-content: center;
  align-items: center;

    }
    
    .if-content {
        padding-left: 16px;
    }
</style> 