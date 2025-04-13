<script lang="ts">
    import { onMount } from 'svelte';
    import {base} from "$app/paths"
    import { llmangoAPI, type Prompt, type Goal } from '$lib/classes/llmangoAPI.svelte';
    import PromptCard from '$lib/PromptCard.svelte';
    import GoalCard from '$lib/GoalCard.svelte';
    import { llmangoLogging, type Log } from '$lib/classes/llmangoLogging.svelte';
    import LogTable from '$lib/LogTable.svelte';

    let loading = $state(true);
    let error = $state<string | null>(null);
    let recentLogs = $state<Log[]>([]);
    let logsLoading = $state(false);


    onMount(async () => {
        try {
            await llmangoAPI.initialize()
        } catch (e) {
            error = e instanceof Error ? e.message : 'Failed to load data';
        }
        try {
            logsLoading = true;
            const logsResponse = await llmangoLogging.getAllLogs({ includeRaw: true, limit: 5, offset: 0 });
            recentLogs = logsResponse?.logs ?? [];
        } catch (logError) {
            console.error('Failed to load recent logs:', logError);
            recentLogs = [];
        } finally {
            logsLoading = false;
            loading = false;
        }
    });
</script>

<div class="home-page">
    <div class="page-header"><h1>LLMango Dashboard</h1></div>

    {#if loading}
        <p>Loading...</p>
    {:else if error}
        <p class="error">{error}</p>
    {:else}
        <div class="section">
            <div class="section-header">
                <h2>Recent Logs</h2>
                <a href={`${base}/logs`}>View All</a>
            </div>
            {#if logsLoading}
                <p>Loading logs...</p>
            {:else if recentLogs.length > 0}
                <LogTable logs={recentLogs} cells={5} />
            {:else}
                <p>No logs available</p>
            {/if}
        </div>
        <div class="section">
            <div class="section-header">
                <h2>Recent Goals</h2>
                <a href={`${base}/goal`}>View All</a>
            </div>
            <div class="card-container">
                {#if Object.keys(llmangoAPI).length > 0}
                    {#each Object.entries(llmangoAPI.goals).slice(0, 2) as [id, goal]}
                        {#if goal}
                            <GoalCard {goal} />
                        {/if}
                    {/each}
                {:else}
                    <p>No goals available</p>
                {/if}
            </div>
        </div>
        <div class="section">
            <div class="section-header">
                <h2>Recent Prompts</h2>
                <a href={`${base}/prompt`}>View All</a>
            </div>
            <div class="card-container">
                {#if Object.keys(llmangoAPI.prompts).length > 0}
                    {#each Object.entries(llmangoAPI.prompts).slice(0, 2) as [id, prompt]}
                        <PromptCard {prompt} goal={llmangoAPI.goals[prompt.goalUID]} />
                    {/each}
                {:else}
                    <p>No prompts available</p>
                {/if}
            </div>
        </div>
        <hr />

        <div class="section">
            <h2>How LLMango Works</h2>
            <p>LLMango streamlines the process of integrating and managing Large Language Models (LLMs) in your applications:</p>
            <ol>
                <li><strong>Define Goals:</strong> Define goals in your code, using input and output structs to specify the data for your LLM tasks.</li>
                <li><strong>Add Prompts:</strong> Add prompts that utilize JSON string tags as variable names for variable replacement, using the format <code>&#123;&#123;variableName&#125;&#125;</code> to ensure proper escaping in Svelte.</li>
                <li><strong>Save Configuration:</strong> Select a method to save your configuration. You can choose JSON, SQLite, or build your own adapter to ensure persistence in case of a restart or crash.</li>
                <li><strong>Store and View Logs:</strong> Choose a way to store and view logs. The default is SQLite, but it's an interface, so you can build a logger for any system. The default LLMango frontend provides a way to view your log results.</li>
                <li><strong>Enhance and Analyze:</strong> Easily add new prompts, perform data analysis on your prompts, run tests, and compare results of different prompts or prompt versions.</li>
            </ol>
            <p>This approach ensures easy addition of new prompts and models, complete observability, and robust management of your LLM integrations.</p>
        </div>
    {/if}
</div>

<style>

    h2{
        margin:0;
    }

    .section {
        margin-bottom: 3rem;
    }

    .section-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 1rem;
    }

    .section-header a {
        color: #007bff;
        text-decoration: none;
        font-size: 0.9rem;
    }

    .section-header a:hover {
        text-decoration: underline;
    }
</style>