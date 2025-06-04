package llmangoagents

import "encoding/json"


type ToolConstructor func(*WorkflowManager) *Tool

//Agent specific tools
//Thinking - ability to pause and think
//
//When subAgents
//    -- Transmit Back partial data
//    -- Optional Handoff to other agent

// Cloud function tool
//allow suers to have arbritrary endpoints for tools
//input json output json.
//basically just ability to use llambda or cloudflare workers to host funcs if they are heavy/to allow for 3rd party functinos easier. create func from url instead of something else pre provides self contained workspace.

//searchtool
//bing or google

type searchTool struct{}

func (s *searchTool) Run(input json.RawMessage) (json.RawMessage, error) {
	// TODO: Implement search functionality
	return json.RawMessage(`{"result": "search not implemented"}`), nil
}
func (s *searchTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"query": {
				"type": "string",
				"description": "The search query to execute"
			}
		},
		"required": ["query"]
	}`)
}

func (s *searchTool) Description() string {
	return "Search for information using a query string"
}
func (s *searchTool) Name() string {
	return "search"
}

func NewBingTool(manager *WorkflowManager) *Tool {
	return &Tool{
		ToolUid: "bing_search",
		Function: func(input json.RawMessage) (json.RawMessage, error) {
			//add the user id
			//filter the messagae
			//log the usage in the manager
			//do the actino and return
			return json.RawMessage(""), nil
		},
		InputSchema:  "",
		OutputSchema: "",
	}
}

// var BingSearchTool Tool = &searchTool{}
// TODO: Implement proper Tool interface methods (lowercase) when tooling system is ready

//scrape tool
// craw4ai scrape of pages

//perplexity serach type resaerch tools
//grok research agent tool

//realtime data pulling??? twitter?? weather etc???

////language foucsed tools for just me
//get words
//get sentences for word
//search grammar books
