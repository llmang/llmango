<script>
    import { base } from '$app/paths';

    let {children}=$props()
</script>

<header>
    <div style="display: flex; align-items: center; gap: 0.5rem;">
        <img src="https://public.llmang.com/logos/llmango.png" alt="LLMango Logo" style="height: 2rem;">
        <a href="{base}/"><h1>LLMan<span>go</span></h1></a>
    </div>
    <nav>
        <a href="{base}/">Home</a>
        <a href="{base}/prompt">Prompts</a>
        <a href="{base}/goal">Goals</a>
        <a href="{base}/models">Models</a>
        <a href="{base}/logs">Logs</a>
        <a>|</a>
        <!-- Will need to implement model store functionality separately -->
        <span>
            <span title="OpenRouter models">models</span>
            <span title="Refresh OpenRouter models" style="cursor: pointer; margin-left: 5px;">🔄</span>
        </span>
    </nav>
</header>
<main>
    {@render children()}
</main>
<footer>
    <hr />
    <p>LLMango - A toolkit for building with LLMs in Go</p>
</footer>

<style>
    header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 1rem;
        padding-bottom: 0.5rem;
        border-bottom: 1px solid #eaeaea;
    }

    header h1 {
        margin: 0;
        font-size: 1.5rem;
    }

    nav {
        display: flex;
        gap: 1rem;
        align-items: center;
    }

    a {
        text-decoration: none;
        color: #0070f3;
    }

    a:hover {
        text-decoration: underline;
    }

    main {
        min-height: 80vh;
    }

    footer {
        margin-top: 2rem;
        color: #666;
        font-size: 0.9rem;
        text-align: center;
    }

    footer hr {
        border: none;
        border-top: 1px solid #eaeaea;
        margin: 1rem 0;
    }
</style>