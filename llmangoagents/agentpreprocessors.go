package llmangoagents

import "github.com/llmang/llmango/openrouter"

//agent preprocessors run before the agent invokes its first action


func preThink(ag *Agent, input string) string {
	//run the agent with a modified systme prompt
	res, err := ag.Manager.SystemManager.Openrouter.GenerateNonStreamingChatResponse(
		&openrouter.OpenRouterRequest{
			Messages: []openrouter.Message{
				{
					Role:    "system",
					Content: ag.BuildSystemMessage() + "\n You are now in thinking mode. You will be given time to think about how you should best approach this problem and you will work thorugh your thoughts layout out your best strategy in a concise and effective manner. Quickly go over the problem and think of the best solution plan.",
				},
				{
					Role:    "user",
					Content: "For the task:" + input + " \n Think about this and come up with a concise plan.",
				},
			},
		},
	)
	//agent logging etc

	if err != nil {
		//error logging
	}
	//return message
	ag.LogPreProcess(res, err)

	return res.Choices[0].Message
}


func preRagl(ag *Agent, input string) string {
	req :=&openrouter.OpenRouterRequest{
		Messages: []openrouter.Message{
			{
				Role:    "system",
				Content: "You are in retrieval mode. To carry out your task you will first be allowed to gather information by querying a vector database. Turn the input into between 0 and 5 queryable strings ",
			},
			{
				Role:    "user",
				Content: "For the task:" + input + " \n Think about this and come up with a concise plan of action.",
			},
		},
	},
	//openrouter.AddStructuringToMessageFromJson("jsonSchema", request) //either adds dfeault schema in correct spot or uses our custom structured message system
	res, err := ag.Manager.SystemManager.Openrouter.GenerateNonStreamingChatResponse(req)

	if err!=nil{

	}
	var vecretrievse []string
	golang waitgroup{
		wg.add
		res vec.retrieveFromStrings()
		for res
		vectretrieves=append(vecretrives, res)
	}
}


//========================================================
//================Prevalidation tool======================

// PreValidateTool definitions
const (
	PreValidateToolRedirect   = "prevalidate_redirect"
	PreValidateToolError      = "prevalidate_error"
	PreValidateToolReturnBack = "prevalidate_returnback"
)

// PreValidateToolList is the list of special toolcalls for prevalidation
var PreValidateToolList = []string{
	PreValidateToolRedirect,
	PreValidateToolError,
	PreValidateToolReturnBack,
}

func preValidate(ag *Agent, input string) string {
	// Instruct the agent to use a toolcall if the input is not valid, otherwise do nothing.
	req := &openrouter.OpenRouterRequest{
		Messages: []openrouter.Message{
			{
				Role:    "system",
				Content: ag.BuildSystemMessage() + `
You are now in prevalidation mode. Your job is to make sure the following input is worthy and valid given your task and role.
Quickly and concisely respond back with your assertion about if the task is complete.

If it is NOT valid, you MUST respond by calling ONE of the following tools:
- prevalidate_redirect: to redirect to another agent (provide agent_id)
- prevalidate_error: to signal an error (provide error message)
- prevalidate_returnback: to return the task to the last agent (no args)
If the input is valid, do not call any tool and reply with only "VALIDATED" nothing else.`,
			},
			{
				Role:    "user",
				Content: "Input to validate: " + input,
			},
		},
		Tools: []openrouter.Tool{
			{
				Type: "function",
				Function: openrouter.Function{
					Name:        PreValidateToolRedirect,
					Description: "Redirects the task to another agent. Use if the input is not valid for this agent. Provide agent_id.",
					Parameters: map[string]interface{}{
						"type":       "object",
						"properties": map[string]interface{}{"agent_id": map[string]interface{}{"type": "string"}},
						"required":   []string{"agent_id"},
					},
				},
			},
			{
				Type: "function",
				Function: openrouter.Function{
					Name:        PreValidateToolError,
					Description: "Signals an error with the input. Use if the input is invalid or problematic. Provide error message.",
					Parameters: map[string]interface{}{
						"type":       "object",
						"properties": map[string]interface{}{"message": map[string]interface{}{"type": "string"}},
						"required":   []string{"message"},
					},
				},
			},
			{
				Type: "function",
				Function: openrouter.Function{
					Name:        PreValidateToolReturnBack,
					Description: "Returns the task to the last agent. Use if the input should be sent back. No arguments.",
					Parameters: map[string]interface{}{
						"type":       "object",
						"properties": map[string]interface{}{},
						"required":   []string{},
					},
				},
			},
		},
	}

	res, err := ag.Manager.SystemManager.Openrouter.GenerateNonStreamingChatResponse(req)
	ag.LogPreProcess(res, err)
	if err != nil || len(res.Choices) == 0 {
		return ""
	}

	// Parse tool calls from the response
	choice := res.Choices[0]
	if len(choice.ToolCalls) > 0 {
		toolCall := choice.ToolCalls[0]
		switch toolCall.Name {
		case PreValidateToolRedirect:
			agentID, _ := toolCall.Args["agent_id"].(string)
			return "@@REDIRECT:" + agentID
		case PreValidateToolError:
			msg, _ := toolCall.Args["message"].(string)
			return "@@ERROR:" + msg
		case PreValidateToolReturnBack:
			return "@@RETURNBACK"
		}
	}
	return ""
}

