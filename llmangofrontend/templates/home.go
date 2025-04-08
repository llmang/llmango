package templates

// HomeTemplate contains the home page template
const HomeTemplate = `
{{define "home"}}
{{template "header" .}}
    <div class="section-header" style="display: flex; justify-content: space-between; align-items: flex-end; margin-bottom: 1rem;">
        <h2>Recent Goals</h2>
        <a href="{{.BaseRoute}}/goals" style="font-size: 0.9rem;">View All</a>
    </div>
    <div class="card-container">
        {{$count := 0}}
        {{range $id, $goalAny := .Goals}}
            {{if lt $count 3}}
                {{$goal := getGoalInfo $goalAny}}
                {{template "goal-card" dict "ID" $id "Goal" $goal "BaseRoute" $.BaseRoute}}
                {{$count = inc $count}}
            {{end}}
        {{else}}
            <p>No goals available</p>
        {{end}}
    </div>

    <div class="section-header" style="display: flex; justify-content: space-between; align-items: flex-end; margin-bottom: 1rem; margin-top: 2rem;">
        <h2>Recent Prompts</h2>
        <a href="{{.BaseRoute}}/prompts" style="font-size: 0.9rem;">View All</a>
    </div>
    <div class="card-container">
        {{$count := 0}}
        {{range $id, $prompt := .Prompts}}
            {{if lt $count 3}}
                {{template "prompt-card" dict "ID" $id "Prompt" $prompt "BaseRoute" $.BaseRoute}}
                {{$count = inc $count}}
            {{end}}
        {{else}}
            <p>No prompts available</p>
        {{end}}
    </div>

    <hr>

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

    {{template "card-styles"}}
    {{template "json-formatter"}}
{{end}}
`
