<script lang="ts">
    let { text } = $props<{ text: string }>();
    let showTooltip = $state(false);
</script>

<div 
    class="info-icon-wrapper" 
    onmouseenter={() => showTooltip = true}
    onmouseleave={() => showTooltip = false}
    onfocus={() => showTooltip = true} 
    onblur={() => showTooltip = false}
    tabindex="0" 
    role="tooltip"
    aria-describedby="tooltip-content"
>
    <span class="info-icon">i</span>
    {#if showTooltip}
        <div class="tooltip" id="tooltip-content">
            {text}
        </div>
    {/if}
</div>

<style>
    .info-icon-wrapper {
        position: relative;
        display: inline-block;
        margin-left: 0.3rem;
        cursor: help;
    }

    .info-icon {
        border-radius: 50%;
        background-color: grey;
        color:white;
        aspect-ratio: 1/1;
        height: 1.25em;
        width: 1.25em;
        font-size: 0.8em;
        display: flex;
        align-items: center;
        justify-content: center;
        vertical-align: middle;
        box-shadow: 0 0 10px var(--grey-50);
    }

    .tooltip {
        position: absolute;
        bottom: 100%; 
        left: 50%;
        transform: translateX(-50%);
        background-color: #333;
        color: white;
        padding: 8px 12px;
        border-radius: 4px;
        font-size: 0.8rem;
        white-space: pre-wrap; /* Allows text wrapping */
        z-index: 10;
        width: 250px; /* Adjust width as needed */
        margin-bottom: 5px; /* Space between icon and tooltip */
        box-shadow: 0 2px 5px rgba(0,0,0,0.2);
        text-align: left;
        opacity: 0.95;
    }

    .info-icon-wrapper:focus {
        outline: none;
    }
</style> 