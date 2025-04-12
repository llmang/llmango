<script lang="ts">
    import { llmangoLogging, type Log, type LogFilter, type PaginationResponse } from '$lib/classes/llmangoLogging.svelte';
    import { llmangoAPI, type Goal, type Prompt } from '$lib/classes/llmangoAPI.svelte';
    import LogTable from '$lib/LogTable.svelte';
    import { fade } from 'svelte/transition';
    import FilterSelect from '$lib/FilterSelect.svelte';
    import { onMount, untrack } from 'svelte';
    
    // Page state
    let logs = $state<Log[]>([]);
    let logsLoading = $state(false);
    let isMounted = $state(false);
    let paginationResponse = $state<PaginationResponse>({
        page: 1,
        totalPages: 1,
        total: 0,
        perPage: 10
    });
    
    // Goals and prompts for filters
    let goals = $state<Record<string, Goal>>({});
    let prompts = $state<Record<string, Prompt>>({});
    
    // Filter state
    let filter = $state<LogFilter>({
        includeRaw: true,
        limit: 10,
        offset: 0
    });


    $effect(()=>{
        if(isMounted===false) return
        loadLogs()
    })
    
    // Load logs with current filter
    async function loadLogs() {
        logsLoading = true;
        try {
            // Fetch logs
            const logResponse = await llmangoLogging.getAllLogs(filter);
            if (logResponse.pagination.totalPages != 0 && filter.offset / filter.limit > logResponse.pagination.totalPages) {
                filter.offset = (logResponse.pagination.totalPages - 1) * filter.limit;
            }
            logs = logResponse.logs;
            paginationResponse=logResponse.pagination
            
        } catch (error) {
            console.error('Failed to load logs:', error);
            logs = [];
        } finally {
            logsLoading = false;
        }
    }
    
    // Navigate pages
    function prevPage() {
        filter.offset = paginationResponse.page-1 * filter.limit;
    }
    
    function nextPage() {
        filter.offset = paginationResponse.page * filter.limit;
    }
    
    // Reset filters
    function resetFilters() {
        filter = {
            includeRaw: true,
            limit: 5,
            offset: 0
        };
        paginationResponse.page = 5;
    }

    onMount(async()=>{
        logsLoading = true;
        isMounted=false
        try {
            // Load goals and prompts for filters
            goals = await llmangoAPI.getAllGoals();
            prompts = await llmangoAPI.getAllPrompts();
            isMounted=true
        } catch (error) {
            console.error('Failed to initialize:', error);
        } finally {
            logsLoading = false;
        }
    })
    
</script>

<div class="logs-page">
    <div class="page-header"><h1>Logs</h1></div>
    <!-- Filters -->
    <div class="filters-section">
        <div class="item-title">Filter Logs</div>
        <div class="filter-controls">
            <div class="filter-selects">
            <FilterSelect id="goalFilter" label="Goal" bind:value={filter.goalUID}>
                <option value={undefined}>All Goals</option>
                {#each Object.entries(goals) as [id, goal]}
                    <option value={id}>{goal.title || id}</option>
                {/each}
            </FilterSelect>
            <FilterSelect id="promptFilter" label="Prompt" bind:value={filter.promptUID}>
                <option value={undefined}>All Prompts</option>
                {#each Object.entries(prompts) as [id, prompt]}
                    <option value={id}>{prompt.UID || id}</option>
                {/each}
            </FilterSelect>
            <FilterSelect id="paginationResponse.perPageFilter" label="Per Page" bind:value={filter.limit}>
                <option value={5}>5</option>
                <option value={10}>10</option>
                <option value={20}>20</option>
                <option value={50}>50</option>
            </FilterSelect>
        </div>
            <button class="btn btn-secondary" onclick={resetFilters}>Reset Filters</button>
        </div>
    </div>
    
    <!-- Loading indicator -->
    <LogTable logs={logs || []} cells={paginationResponse.perPage} />
    <!-- Pagination -->
    {#if paginationResponse.totalPages > 1 || paginationResponse.totalLogs > 0}
        <div class="pagination">
            <button 
                onclick={prevPage}
                disabled={paginationResponse.page <= 1}
            >
                Previous
            </button>
            
            <span>Page {paginationResponse.page} of {paginationResponse.totalPages} (Total: {paginationResponse.total} logs)</span>
            
            <button 
                onclick={nextPage}
                disabled={paginationResponse.page >= paginationResponse.totalPages}
            >
                Next
            </button>
        </div>
    {/if}
    {#if logsLoading}
        <div class="loading" in:fade={{duration:300}}>Loading logs...</div>
    {/if}
</div>

<style>
    h1 {
        margin-bottom: 1rem
    }
    
    .filters-section {
        background-color: #f8f8f8;
        border-radius: 5px;
        padding: 15px;
        margin-bottom: 20px;
    }

    
    .filter-controls {
        display: flex;
        flex-wrap: wrap;
        gap: 15px;
        align-items: flex-end;
    }
    
 .filter-selects{
    display: flex; gap:1rem;
    flex-wrap: wrap;
 }
    
    button {
        padding: 0.375rem 0.75rem;
        font-size: 0.9rem;
        border-radius: 0.25rem;
        background-color: #007bff;
        color: white;
        border: none;
        cursor: pointer;
    }
    
    button:hover:not(:disabled) {
        background-color: #0069d9;
    }
    
    button:disabled {
        background-color: #6c757d;
        cursor: not-allowed;
        opacity: 0.65;
    }
    
    .loading {
        text-align: center;
        padding: 20px;
        color: #666;
        font-style: italic;
    }
    
    .pagination {
        display: flex;
        justify-content: center;
        align-items: center;
        gap: 1rem;
        margin-top: 1rem;
    }
</style>
