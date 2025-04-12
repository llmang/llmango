<script lang="ts">
    let { title, description, href=null, onclick=null, children } = $props();
</script>

{#if href}
    <a href={href} class="clickable card">
        <h3>{title}</h3>
        <p>{description}</p>
        {@render children()}
    </a>
{:else if onclick}
    <button class="clickable button-wrapper card" onclick={onclick}>
        <h3>{title}</h3>
        <p>{description}</p>
        {@render children()}
    </button>
{:else}
    <div class="card">
        <h3>{title}</h3>
        <p>{description}</p>
        {@render children()}
    </div>
{/if}

<style>
    .card {
        min-height: 200px;
        text-decoration: none;
        background: white;
        border-radius: 8px;
        box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
        padding: 1rem;
        transition: transform 0.2s ease-in-out;
    }

    .card:hover {
        transform: translateY(-2px);
        box-shadow: 0 4px 8px rgba(123, 123, 123, 0.2);
    }

    h3 {
        margin: 0;
        font-size: 1.1rem;
        color: #333;
        margin-bottom:.5rem;
    }

    p {
        margin: 0;
        font-size: 0.9rem;
        color: #666;
        margin-bottom:1rem;
    }


</style>