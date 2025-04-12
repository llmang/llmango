<script lang="ts">
    import { onMount } from 'svelte';
    
    export let jsonText: string;
    let formattedJson = '';

    const syntaxHighlight = (json: string): string => {
        try {
            const jsonData = JSON.parse(json);
            const formatted = JSON.stringify(jsonData, null, 2);
            
            return formatted.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, function (match) {
                let cls = 'json-number'; // Number
                if (/^"/.test(match)) {
                    if (/:$/.test(match)) {
                        cls = 'json-key'; // Key
                    } else {
                        cls = 'json-string'; // String
                    }
                } else if (/true|false/.test(match)) {
                    cls = 'json-boolean'; // Boolean
                } else if (/null/.test(match)) {
                    cls = 'json-null'; // null
                }
                return `<span class="${cls}">${match}</span>`;
            });
        } catch (e) {
            console.error('JSON parse error:', e);
            return json;
        }
    };

    onMount(() => {
        formattedJson = syntaxHighlight(jsonText);
    });
</script>

<style>
    .json-preview {
        background-color: #f8f8f8;
        border-radius: 4px;
        padding: 1rem;
        font-family: monospace;
        font-size: 0.8rem;
        max-height: 200px;
        overflow: auto;
        white-space: pre-wrap;
        margin-bottom: 1rem;
        line-height: 1.4;
    }
    
    :global(.json-key) { 
        color: #444;
    }
    :global(.json-string) { 
        color: #286;
    }
    :global(.json-number) { 
        color: #07a;
    }
    :global(.json-boolean) { 
        color: #905;
    }
    :global(.json-null) { 
        color: #666;
    }
</style>

<pre class="json-preview">{@html formattedJson}</pre>
