<script lang="ts">
    import FormatJson from '$lib/FormatJson.svelte';
    import Card from '$lib/Card.svelte';
    import type { Goal, Prompt, Solution } from '$lib/classes/llmangoAPI.svelte';
    import { llmangoAPI } from '$lib/classes/llmangoAPI.svelte';
    import SolutionModal from './SolutionModal.svelte';
    import { onMount } from 'svelte';
    import { page } from '$app/state';
    import PromptCard from '$lib/PromptCard.svelte';
    import { llmangoLogging, type Log } from '$lib/classes/llmangoLogging.svelte';
    import LogTable from '$lib/LogTable.svelte';
    import { base } from '$app/paths';
    import StopPropagation from '$lib/StopPropigation.svelte';

    let goaluid = $derived(page.params.goaluid);

    // Initial states with safe defaults
    let goal = $state<Goal | null>(null);
    let prompts = $state<Record<string, Prompt>>({});
    let loading = $state(true);
    let error = $state<string | null>(null);
    let logs = $state<Log[]>([]);  // Initialize as empty array
    let logsLoading = $state(false);
    
    // Load data on component mount
    onMount(async () => {
        try {
            // Fetch goal and prompts data in parallel
            const [fetchedGoal, fetchedPrompts] = await Promise.all([
                llmangoAPI.getGoal(goaluid),
                llmangoAPI.getAllPrompts()
            ]);
            
            if (!fetchedGoal) {
                error = `Goal with ID ${goaluid} not found`;
            } else {
                goal = fetchedGoal;
                prompts = fetchedPrompts || {};
                
                // Fetch logs for this goal
                logsLoading = true;
                try {
                    const logsResponse = await llmangoLogging.getGoalLogs(goaluid, {
                        includeRaw: true,
                        limit: 5,
                        offset: 0
                    });
                    logs = logsResponse?.logs || [];  // Use nullish coalescing to ensure we always have an array
                } catch (logError) {
                    console.error('Failed to load logs:', logError);
                    logs = []; // Ensure logs is at least an empty array on error
                } finally {
                    logsLoading = false;
                }
            }
        } catch (e) {
            error = e instanceof Error ? e.message : 'Failed to load data';
        } finally {
            loading = false;
        }
    });

    const openEditSolutionModal = (solutionId: string, solution: Solution) => {
        currentSolutionId = solutionId;
        currentSolution = solution;
        editSolutionModalOpen = true;
    };

    const closeNewSolutionModal = () => {
        newSolutionModalOpen = false;
    };

    const closeEditSolutionModal = () => {
        editSolutionModalOpen = false;
        currentSolutionId = '';
        currentSolution = null;
    };

    let newSolutionModalOpen = $state(false);
    let editSolutionModalOpen = $state(false);
    let currentSolutionId = $state('');
    let currentSolution = $state<Solution | null>(null);

    // Added helper function to compute solution status badge
    function getSolutionStatus(solution: Solution): { label: string, dotColor: string, bgColor: string } {
        if (solution.weight === 0) {
            return { label: "Stopped", dotColor: "#6c757d", bgColor: "#e9ecef" };
        }
        if (solution.isCanary) {
            if (solution.totalRuns >= solution.maxRuns) {
                return { label: "Completed", dotColor: "#007bff", bgColor: "#cce5ff" };
            } else {
                return { label: "In Progress", dotColor: "#fd7e14", bgColor: "#ffe5d1" };
            }
        } else {
            return { label: "Running", dotColor: "#28a745", bgColor: "#d4edda" };
        }
    }
</script>

<div class="goal-page">
    {#if loading}
        <div class="loading">Loading goal data...</div>
    {:else if error}
        <div class="error">{error}</div>
    {:else if goal}
        <div class="goal-header">
            <h2>Goal: <span class="goal-uid">{goaluid}</span></h2>
        </div>

        <div class="card goal-card">
            <div class="item-title">Title</div>
            <div class="item">{goal.title || 'Untitled Goal'}</div>
            <div class="item-title">Description</div>
            <div class="item">{goal.description || 'No description'}</div>
            <div class="input-output">
                <div class="ioside">
                    <div class="item-title">Input</div>
                    <FormatJson jsonText={JSON.stringify(goal.exampleInput || "")} />
                </div>
                <div class="ioside">
                    <div class="item-title">Output</div>
                    <FormatJson jsonText={JSON.stringify(goal.exampleOutput || "")} />
                </div>
            </div>

            <div class="item-title">Solutions <span style="font-size:.8em">({goal.solutions ? Object.keys(goal.solutions).length : 0})</span></div>
            <div class="card-container">
                <button class="button-wrapper card new-item-card" onclick={() => newSolutionModalOpen = true}>
                    <div>+</div>
                    <div>Add New Solution</div>
                </button>
                {#if goal.solutions && Object.keys(goal.solutions).length > 0}
                    {#each Object.entries(goal.solutions) as [solutionId, solutionObj]}
                        {@const solution = solutionObj as Solution}
                        {@const status = getSolutionStatus(solution)}
                        <Card 
                            title={solution.promptUID || 'No Prompt'}
                            description={""}
                            href={solution.promptUID ? `${base}/prompt/${solution.promptUID}` : undefined}
                        >
                            <div class="solution-badge" style="background-color: {status.bgColor};">
                                <span class="badge-label">{status.label}</span>
                                <span class="badge-dot" style="background-color: {status.dotColor};"></span>
                            </div>
                            <div class="solution-meta">
                                <span>Prompt UID: {solution.promptUID || 'None'}</span>
                                <span>Weight: {solution.weight}</span>
                                {#if solution.isCanary}
                                    <span>Runs: {solution.totalRuns}/{solution.maxRuns}</span>
                                    <span>Type: Canary</span>
                                {:else}
                                    <span>Type: Standard</span>
                                {/if}
                            </div>
                            
                            <div class="edit-button-container">
                                <StopPropagation>
                                    <button 
                                        class="btn-edit" 
                                        onclick={() => openEditSolutionModal(solutionId, solution)}
                                        title="Edit Solution"
                                    >
                                        Edit
                                    </button>
                                </StopPropagation>
                            </div>
                        </Card>
                    {/each}
                {/if}
            </div>
            
            <!-- Related Prompts Section -->
            <div class="item-title">Related Prompts</div>
            <div class="card-container">
                {#if Object.keys(prompts).length > 0}
                    {@const matchingPrompts = Object.entries(prompts).filter(([_, prompt]) => prompt?.goalUID === goaluid)}
                    {#if matchingPrompts.length > 0}
                        {#each matchingPrompts as [promptUID, prompt]}
                            <PromptCard {prompt}/>
                        {/each}
                    {:else}
                        <div class="no-items">No related prompts found</div>
                    {/if}
                {:else}
                    <div class="no-items">No prompts loaded</div>
                {/if}
            </div>
            
            <!-- Logs Section -->
            <div class="item-title">Recent Logs</div>
            {#if logsLoading}
                <div class="loading">Loading logs...</div>
            {:else if logs && logs.length > 0}
                <LogTable logs={logs || []} />
            {:else}
                <div class="no-items">No logs found for this goal</div>
            {/if}
            
            <!-- Debug Info -->
            <details class="goal-debug">
                <summary>Debug Info</summary>
                <FormatJson jsonText={JSON.stringify(goal)} />
            </details>
        </div>

        <!-- Solution Modals -->
        <SolutionModal 
            isOpen={newSolutionModalOpen}
            mode="create"
            prompts={prompts}
            currentSolution={null}
            currentSolutionId=""
            goalUID={goaluid}
            onClose={closeNewSolutionModal}
        />

        <SolutionModal 
            goalUID={goaluid}
            isOpen={editSolutionModalOpen}
            mode="edit"
            prompts={prompts}
            currentSolution={currentSolution}
            currentSolutionId={currentSolutionId}
            onClose={closeEditSolutionModal}
        />
    {/if}
</div>

<style>

    .goal-page {
        margin: 1rem 0;
    }
    
    .goal-header {
        margin-bottom: 1rem;
    }

    .goal-card {
        padding: 1.5rem;
        margin-bottom: 2rem;
        background: white;
        border-radius: 8px;
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
    }
    

    .goal-debug {
        border-top: 1px solid #eee;
        padding-top: 0.5rem;
    }
    
    .goal-debug summary {
        cursor: pointer;
        color: #666;
        font-size: 0.8rem;
    }


    .loading,
    .error {
        padding: 2rem;
        text-align: center;
        background-color: #f9f9f9;
        border-radius: 8px;
    }
    
    .error {
        color: #dc3545;
        background-color: #f8d7da;
    }

    .goal-uid {
        color: #777;
        font-weight: normal;
    }
    

    .input-output {
        flex-wrap: wrap;
        display: flex;
        gap: 1rem;
        margin-top: 1rem;
        margin-bottom: 1rem;
    }
    
    .ioside {
        flex:1;
    }
 

    .no-items {
        padding: 1rem;
        text-align: center;
        color: #777;
        font-style: italic;
    }
    
    .prompt-indicator {
        background-color: #007bff;
    }

    .solution-badge {
        position: absolute;
        top: 10px;
        right: 10px;
        display: flex;
        align-items: center;
        padding: 2px 6px;
        border-radius: 12px;
        font-size: 0.75rem;
        color: #333;
    }
    .badge-label {
        margin-right: 4px;
    }
    .badge-dot {
        width: 8px;
        height: 8px;
        border-radius: 50%;
    }
    .solution-meta {
        margin-top: 0.5rem;
        font-size: 0.85rem;
        color: #666;
        display: flex;
        flex-wrap: wrap;
        gap: 1rem;
    }
    .solution-meta span {
        background-color: #f9f9f9;
        padding: 0.2rem 0.5rem;
        border-radius: 4px;
    }
    
    /* Edit button styles */
    .edit-button-container {
        position: absolute;
        bottom: 10px;
        right: 10px;
    }
    
    .btn-edit {
        background-color: #f8f9fa;
        border: 1px solid #ced4da;
        color: #495057;
        padding: 4px 10px;
        border-radius: 4px;
        font-size: 0.8rem;
        cursor: pointer;
        transition: all 0.2s ease;
    }
    
    .btn-edit:hover {
        background-color: #e9ecef;
        border-color: #adb5bd;
    }
</style> 