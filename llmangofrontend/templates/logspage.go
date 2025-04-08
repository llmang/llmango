package templates

const LogsPageTemplate = `
{{define "logs"}}
{{template "header" .}}

<div class="page-content">
    <div class="page-header">
        <h1>Logs</h1>
    </div>

    <div 
        x-data="{ 
            logs: [],
            currentPage: 1,
            totalPages: 1,
            isLoading: false,
            filterOptions: {
                goalId: null,
                promptId: null,
                perPage: 10
            },
            goals: [],
            prompts: [],
            async loadLogs() {
                this.isLoading = true;
                try {
                    let url = '{{.BaseRoute}}/api/logs';
                    
                    // Add filter parameters
                    const params = new URLSearchParams();
                    params.append('page', this.currentPage.toString());
                    params.append('perPage', this.filterOptions.perPage.toString());
                    
                    // Add all filter options to params
                    if (this.filterOptions.goalId) {
                        params.append('goalId', this.filterOptions.goalId);
                    }
                    if (this.filterOptions.promptId) {
                        params.append('promptId', this.filterOptions.promptId);
                    }
                    
                    const response = await fetch(url + '?' + params.toString());
                    const data = await response.json();
                    if (data.success) {
                        this.logs = data.data.logs;
                        this.totalPages = data.data.pagination.totalPages;
                    }
                } catch (error) {
                    console.error('Failed to load logs:', error);
                }
                this.isLoading = false;
            },
            async loadGoalsAndPrompts() {
                // Load goals for filtering
                try {
                    const goalsResponse = await fetch('{{.BaseRoute}}/api/goals');
                    const goalsData = await goalsResponse.json();
                    if (goalsData.success) {
                        this.goals = goalsData.data;
                    }
                } catch (error) {
                    console.error('Failed to load goals:', error);
                }
                
                // Load prompts for filtering
                try {
                    const promptsResponse = await fetch('{{.BaseRoute}}/api/prompts');
                    const promptsData = await promptsResponse.json();
                    if (promptsData.success) {
                        this.prompts = promptsData.data;
                    }
                } catch (error) {
                    console.error('Failed to load prompts:', error);
                }
            },
            async prevPage() {
                if (this.currentPage > 1) {
                    this.currentPage--;
                    await this.loadLogs();
                }
            },
            async nextPage() {
                if (this.currentPage < this.totalPages) {
                    this.currentPage++;
                    await this.loadLogs();
                }
            },
            resetFilters() {
                this.filterOptions = {
                    goalId: null,
                    promptId: null,
                    perPage: 10
                };
                this.currentPage = 1;
                this.loadLogs();
            }
        }"
        x-init="loadLogs(); loadGoalsAndPrompts()"
    >
        <!-- Filter Section -->
        <div class="filters-section">
            <h3>Filter Logs</h3>
            <div class="filter-controls">
                <div class="filter-group">
                    <label for="goalFilter">Goal:</label>
                    <select id="goalFilter" x-model="filterOptions.goalId" @change="currentPage = 1; loadLogs()">
                        <option value="">All Goals</option>
                        <template x-for="(goalInfo, goalId) in {{toJSON .Goals}}" :key="goalId">
                            <option :value="goalId" x-text="goalInfo.Title || goalId"></option>
                        </template>
                    </select>
                </div>
                
                <div class="filter-group">
                    <label for="promptFilter">Prompt:</label>
                    <select id="promptFilter" x-model="filterOptions.promptId" @change="currentPage = 1; loadLogs()">
                        <option value="">All Prompts</option>
                        <template x-for="(prompt, promptId) in {{toJSON .Prompts}}" :key="promptId">
                            <option :value="promptId" x-text="prompt.Name || promptId"></option>
                        </template>
                    </select>
                </div>
                
                <div class="filter-group">
                    <label for="perPageFilter">Per Page:</label>
                    <select id="perPageFilter" x-model="filterOptions.perPage" @change="currentPage = 1; loadLogs()">
                        <option value="5">5</option>
                        <option value="10">10</option>
                        <option value="20">20</option>
                        <option value="50">50</option>
                    </select>
                </div>
                
                <button class="btn btn-secondary" @click="resetFilters">Reset Filters</button>
            </div>
        </div>

        <!-- Log Display -->
        <div class="log-viewer">
            <div class="logs-container">
                <template x-if="logs.length > 0">
                    <div class="log-table">
                        <div class="log-header-row">
                            <div class="log-cell">Time</div>
                            <div class="log-cell">Goal ID</div>
                            <div class="log-cell">Prompt ID</div>
                            <div class="log-cell">Tokens</div>
                            <div class="log-cell">Cost</div>
                            <div class="log-cell">Actions</div>
                        </div>
                        <template x-for="log in logs" :key="log.timestamp">
                            <div class="log-row-container">
                                <div class="log-row">
                                    <div class="log-cell" :title="new Date(log.timestamp * 1000).toLocaleString()">
                                        <span x-text="new Date(log.timestamp * 1000).toLocaleString('en-GB', {day: '2-digit', month: '2-digit', year: '2-digit', hour: '2-digit', minute: '2-digit', second: '2-digit'})"></span>
                                    </div>
                                    <div class="log-cell" :title="log.goalUID">
                                        <span x-text="log.goalUID.substring(0, 10) + (log.goalUID.length > 10 ? '...' : '')"></span>
                                    </div>
                                    <div class="log-cell" :title="log.promptUID">
                                        <span x-text="log.promptUID.substring(0, 10) + (log.promptUID.length > 10 ? '...' : '')"></span>
                                    </div>
                                    <div class="log-cell" :title="'Input: ' + log.inputTokens + ' tokens, Output: ' + log.outputTokens + ' tokens'">
                                        <span class="token-count" :title="'Input Tokens: ' + log.inputTokens" x-text="log.inputTokens"></span> / 
                                        <span class="token-count" :title="'Output Tokens: ' + log.outputTokens" x-text="log.outputTokens"></span>
                                    </div>
                                    <div class="log-cell" :title="'Total cost: $' + log.cost">
                                        <span x-text="'$' + log.cost"></span>
                                    </div>
                                    <div class="log-cell">
                                        <button 
                                            class="btn btn-sm btn-secondary"
                                            @click="log.showDetails = !log.showDetails"
                                        >
                                            <span x-text="log.showDetails ? 'Hide' : 'Show'"></span> Details
                                        </button>
                                    </div>
                                </div>
                                <div 
                                    class="log-details"
                                    x-show="log.showDetails"
                                    x-transition
                                >
                                    <div class="details-row">
                                        <div class="details-cell">
                                            <h5>Goal ID</h5>
                                            <pre x-text="log.goalUID"></pre>
                                        </div>
                                        <div class="details-cell">
                                            <h5>Prompt ID</h5>
                                            <pre x-text="log.promptUID"></pre>
                                        </div>
                                    </div>
                                    <div class="log-section">
                                        <h5>Input</h5>
                                        <pre x-text="log.inputObject"></pre>
                                    </div>
                                    <div class="log-section">
                                        <h5>Output</h5>
                                        <pre x-text="log.outputObject"></pre>
                                    </div>
                                    <template x-if="log.error">
                                        <div class="log-section error">
                                            <h5>Error</h5>
                                            <pre x-text="log.error"></pre>
                                        </div>
                                    </template>
                                </div>
                            </div>
                        </template>
                    </div>
                </template>
                <template x-if="logs.length === 0">
                    <p class="no-logs">No logs found. Try adjusting your filters.</p>
                </template>
            </div>

            <div class="pagination" x-show="totalPages > 1">
                <button @click="prevPage()" :disabled="currentPage === 1">Previous</button>
                <span>Page <span x-text="currentPage"></span> of <span x-text="totalPages"></span></span>
                <button @click="nextPage()" :disabled="currentPage === totalPages">Next</button>
            </div>

            <div class="loading" x-show="isLoading">
                Loading logs...
            </div>
        </div>
    </div>
</div>

{{template "footer"}}
{{end}}
`
