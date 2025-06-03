package llmangoagents

import "encoding/json"

// what is an agent
// 1. an llm model and a system message
// 2. a group of tools that it can call
//
// Base Tools: Return To User, continue,
//
//
//
//
// some agents continue running repeatedly, sometimes when an agent responds it returns its value to the last agent in the system.
// after a tool is called it returns its value to the agent and then continues with another message.

type Tool interface {
	run(json.RawMessage) json.RawMessage
	getSchema() (json.RawMessage, error)
	getDescription() string
	getName() string
}

type Agent struct {
	Name          string
	SystemMessage string
	Model         string
	Parameters    string
	Tools         []Tool
}

type WorkflowStep struct {
	maxRuns          int
	maxCost          int
	entryAgent       Agent
	validationAgent  Agent                       //optional or your entry agent can judge itself //this can be either a tool or a schema
	validationSchema *func(json.RawMessage) bool //
}

type Workflow struct {
	entryAgent Agent
	exitAgent  Agent
	totalSpend int //milicents like openrouter spec
	totalSteps int //steps
}

//Flow works like this. We have an entrypoint agent at each step this organizes the process or can be a single step agent for simple tasks basically decides whether to call tools or to return
//if we have a exitpoint this acts as a validator agent which either accepts or denys the results
//we use opneai spec for the tool calling if possible. or a generalizable format
//for certain tools we can "initialize them " which puts a key or other thing into a map for them to access thus this can limit the destructive power of llms by forcing create or update to be limited to that key thus managing their control in sql environment
