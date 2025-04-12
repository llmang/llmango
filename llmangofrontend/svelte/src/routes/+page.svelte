<script lang="ts">
    import { onMount } from 'svelte';
    import {base} from "$app/paths"
    import { llmangoAPI, type Prompt, type Goal, type Solution } from '$lib/classes/llmangoAPI.svelte';
    import PromptCard from '$lib/PromptCard.svelte';
    import GoalCard from '$lib/GoalCard.svelte';
    import { llmangoLogging, type Log } from '$lib/classes/llmangoLogging.svelte';
    import LogTable from '$lib/LogTable.svelte';

    let recentPrompts = $state<Record<string, Prompt>>({});
    let recentGoals = $state<Record<string, Goal>>({});
    let loading = $state(true);
    let error = $state<string | null>(null);
    let recentLogs = $state<Log[]>([]);
    let logsLoading = $state(false);

    // Helper function to check if a goal uses a specific prompt
    function goalUsesPrompt(goal: Goal, promptUID: string): boolean {
        if (!goal || !goal.solutions) return false;
        return Object.values(goal.solutions).some((solution: Solution) => 
            solution.promptUID === promptUID
        );
    }

    // Helper function to find a goal related to a prompt
    function findRelatedGoal(goals: Record<string, Goal>, promptUID: string): Goal | undefined {
        return Object.values(goals).find(goal => goalUsesPrompt(goal, promptUID));
    }

    onMount(async () => {
        try {
            recentPrompts = await llmangoAPI.getAllPrompts();
            recentGoals = await llmangoAPI.getAllGoals();
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
                {#if Object.keys(recentGoals).length > 0}
                    {#each Object.entries(recentGoals).slice(0, 2) as [id, goal]}
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
                {#if Object.keys(recentPrompts).length > 0}
                    {#each Object.entries(recentPrompts).slice(0, 2) as [id, prompt]}
                        {@const relatedGoal = findRelatedGoal(recentGoals, prompt.UID)}
                        <PromptCard {prompt} goal={relatedGoal} />
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
                <li><strong>Define Goals:</strong> Start by defining goal structs directly in your Go code. These structs specify the desired inputs, outputs, and validation logic for your LLM tasks. The spec of each llmango route is inferred through its struct type and JSON tags.</li>
                <li><strong>Generate Config:</strong> Run the <code>llmango</code> CLI tool. This tool analyzes your goal structs and generates a central <code>llmango.json</code> configuration file.</li>
                <li><strong>Add Prompts:</strong> Populate the <code>llmango.json</code> file with specific prompts for your defined goals. You can do this manually or use the LLMango frontend for a more interactive experience.</li>
                <li><strong>Run & Observe:</strong> LLMango takes over from here. It automatically:
                    <ul>
                        <li>Logs all LLM requests and responses.</li>
                        <li>Provides observability into model performance and costs.</li>
                        <li>Allows easy creation and testing of new prompts against your goals.</li>
                        <li>Facilitates switching between different LLM providers (like those supported by OpenRouter) without changing your core application code.</li>
                    </ul>
                </li>
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