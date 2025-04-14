

<script lang="ts">
    import { onMount } from 'svelte';
    
    let { isOpen, title, onClose=null, children } = $props<{
        isOpen: boolean;
        title: string;
        onClose: () => void;
        children: any;
    }>();
    const close=()=>{isOpen=false}
    
    $effect(() => {
        isOpen ? document.body.style.overflow = 'hidden' : document.body.style.overflow = 'unset'
        return (()=>document.body.style.overflow = 'unset')
    });
</script>

{#if isOpen}
    <div class="modal-backdrop" onclick={onClose || close}>
        <div class="modal-content" onclick={(e) => e.stopPropagation()}>
            <div class="modal-header">
                <h3>{title}</h3>
                <button class="close-button" onclick={onClose || close}>&times;</button>
            </div>
            <div class="modal-body">
                {@render children()}
            </div>
        </div>
    </div>
{/if}

<style>
    .modal-backdrop {
        position: fixed;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        background-color: rgba(0, 0, 0, 0.5);
        z-index: 100;
        overflow-y: auto;
        padding-top:3rem;
    }
    
    .modal-content {
        margin: auto;
        background-color: white;
        border-radius: 8px;
        box-shadow: 0 2px 10px rgba(0, 0, 0, 0.2);
        max-width: 90%;
        width: fit-content;
    }
    
    .modal-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 1rem 1.5rem;
        border-bottom: 1px solid #eee;
    }
    
    .modal-header h3 {
        margin: 0;
    }
    .close-button {
        background: none;
        border: none;
        font-size: 1.5rem;
        cursor: pointer;
        padding: 0.2rem;
        line-height: 1;
        aspect-ratio: 1/1;
        height: 1em;
        display: flex;
        align-items: center;
        justify-content: center;
    }
    
    .modal-body {
        padding: 1.5rem;
    }
</style>
