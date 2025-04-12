<script lang="ts">
    import { llmangoLogging, type Log, type LogFilter, type PaginationResponse } from '$lib/classes/llmangoLogging.svelte';
    import { llmangoAPI, type Goal, type Prompt } from '$lib/classes/llmangoAPI.svelte';
    import LogTable from '$lib/LogTable.svelte';
    import { fade } from 'svelte/transition';
    
    // Page state
    let logs = $state<Log[]>([]);
    let isLoading = $state(false);
    let isMounted = $state(false);
    let currentPage = $state(1);
    let totalPages = $state(1);
    let totalLogs = $state(0);
    let perPage = $state(10);
    
    // Goals and prompts for filters
    let goals = $state<Record<string, Goal>>({});
    let prompts = $state<Record<string, Prompt>>({});
    
    // Filter state
    let filter = $state<LogFilter>({
        includeRaw: true,
        limit: 10,
        offset: 0
    });
    
    // Load initial data
    async function initialize() {
        isLoading = true;
        try {
            // Load goals and prompts for filters
            goals = await llmangoAPI.getAllGoals();
            prompts = await llmangoAPI.getAllPrompts();
            
            // Load initial logs
            await loadLogs();
        } catch (error) {
            console.error('Failed to initialize:', error);
        } finally {
            isLoading = false;
        }
    }
    
    // Load logs with current filter
    async function loadLogs() {
        isLoading = true;
        try {
            // Calculate offset based on current page
            filter.offset = (currentPage - 1) * filter.limit;
            
            // Fetch logs
            const logResponse = await llmangoLogging.getAllLogs(filter);
            logs = logResponse.logs;
            
            // Update pagination state from response
            currentPage = logResponse.pagination.page;
            totalPages = logResponse.pagination.totalPages;
            totalLogs = logResponse.pagination.total;
            perPage = logResponse.pagination.perPage;
        } catch (error) {
            console.error('Failed to load logs:', error);
            logs = [];
        } finally {
            isLoading = false;
            isMounted=true;
        }
    }
    
    // Handle filter changes
    function handleFilterChange() {
        currentPage = 1;
        loadLogs();
    }
    
    // Navigate pages
    function prevPage() {
        if (currentPage > 1) {
            currentPage--;
            loadLogs();
        }
    }
    
    function nextPage() {
        if (currentPage < totalPages) {
            currentPage++;
            loadLogs();
        }
    }
    
    // Reset filters
    function resetFilters() {
        filter = {
            includeRaw: true,
            limit: 10,
            offset: 0
        };
        currentPage = 1;
        loadLogs();
    }
    
    // Initialize on mount
    initialize();
</script>

<div class="logs-page">
    <h1>Logs</h1>
    
    <!-- Filters -->
    <div class="filters-section">
        <div class="item-title">Filter Logs</div>
        <div class="filter-controls">
            <div class="filter-group">
                <label for="goalFilter">Goal:</label>
                <select 
                    id="goalFilter" 
                    bind:value={filter.goalUID} 
                    onchange={handleFilterChange}
                >
                    <option value={undefined}>All Goals</option>
                    {#each Object.entries(goals) as [id, goal]}
                        <option value={id}>{goal.title || id}</option>
                    {/each}
                </select>
            </div>
            
            <div class="filter-group">
                <label for="promptFilter">Prompt:</label>
                <select 
                    id="promptFilter" 
                    bind:value={filter.promptUID} 
                    onchange={handleFilterChange}
                >
                    <option value={undefined}>All Prompts</option>
                    {#each Object.entries(prompts) as [id, prompt]}
                        <option value={id}>{prompt.UID || id}</option>
                    {/each}
                </select>
            </div>
            
            <div class="filter-group">
                <label for="perPageFilter">Per Page:</label>
                <select 
                    id="perPageFilter" 
                    bind:value={filter.limit}
                    onchange={handleFilterChange}
                >
                    <option value={5}>5</option>
                    <option value={10}>10</option>
                    <option value={20}>20</option>
                    <option value={50}>50</option>
                </select>
            </div>
            
            <button class="btn btn-secondary" onclick={resetFilters}>Reset Filters</button>
        </div>
    </div>
    
    <!-- Loading indicator -->
    <LogTable logs={logs || []} cells={perPage} />
    <!-- Pagination -->
    {#if totalPages > 1 || totalLogs > 0}
        <div class="pagination">
            <button 
                onclick={prevPage}
                disabled={currentPage <= 1}
            >
                Previous
            </button>
            
            <span>Page {currentPage} of {totalPages} (Total: {totalLogs} logs)</span>
            
            <button 
                onclick={nextPage}
                disabled={currentPage >= totalPages}
            >
                Next
            </button>
        </div>
    {/if}
    {#if isLoading}
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
    
    .filters-section h3 {
        margin-top: 0;
        margin-bottom: 15px;
    }
    
    .filter-controls {
        display: flex;
        flex-wrap: wrap;
        gap: 15px;
        align-items: flex-end;
    }
    
    .filter-group {
        display: flex;
        flex-direction: column;
        width: 10rem;
    }
    
    .filter-group label {
        margin-bottom: 5px;
        font-size: 0.9rem;
    }
    
    select {
        padding: 0.375rem 0.75rem;
        font-size: 0.9rem;
        border: 1px solid #ced4da;
        border-radius: 0.25rem;
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
