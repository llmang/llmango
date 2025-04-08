package templates

const LogsTemplates = `
{{define "log-viewer"}}
<div 
    x-data="{ 
        logs: [],
        currentPage: 1,
        totalPages: 1,
        isLoading: false,
        filterOptions: {{toJSON .FilterOptions}},
        async loadLogs() {
            this.isLoading = true;
            try {
                let url = '{{.BaseRoute}}/api/logs';
                
                // Add filter parameters
                const params = new URLSearchParams();
                params.append('page', this.currentPage.toString());
                params.append('perPage', '5');
                
                // Add all filter options to params
                Object.entries(this.filterOptions).forEach(([key, value]) => {
                    if (value !== null && value !== undefined) {
                        params.append(key, value.toString());
                    }
                });
                
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
        }
    }"
    x-init="loadLogs()"
>
    <div class="log-viewer">
        <div class="logs-container">
            <template x-if="logs.length > 0">
                <div class="log-table">
                    <div class="log-header-row">
                        <div class="log-cell">Time</div>
                        <div class="log-cell">IDs</div>
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
                <p>No logs found.</p>
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
{{end}}
`
