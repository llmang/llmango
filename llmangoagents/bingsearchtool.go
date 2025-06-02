package llmangoagents

import "encoding/json"

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

var BingSearchTool Tool = &searchTool{}
