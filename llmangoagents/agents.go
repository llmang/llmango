package llmangoagents

import "encoding/json"

// takes in the agent its going to and the task the agent is asked
type Preprocess func(*Agent, string) (string, error)

//thinking //uses agents llm with a custom system prompt to **think**
//data collectors //uses the string to query a database/datastore

type Agent struct {
	Manager       *WorkflowManager
	Name          string
	SystemMessage string
	Model         string
	Parameters    string
	Tools         map[string]*Tool //abilities for the agent.
	PreProcessors []Preprocess     //in order of their usage
}

type UserAgentInstance struct {
	UserId          string
	WebsocketConnId string
}
//cli user agent instance
//webhook user agent instance
//interface(transmit() recieve())


type WorkflowManager struct {
	UserAgentInstance *UserAgentInstance
	WorkflowUUID      string //highest level wrapper around everthing nothing higher than this.
	TotalCost         int
	TotalSteps        int

	context
	StepUUID          string
	CurrentDepth      int
	Input             json.RawMessage //user data or document data etc.
}


type WorkflowBuilder struct {
	Steps []WorkflowStep
}

func CreateWorkflow() *WorkflowBuilder {
	return &WorkflowBuilder{}
}


//handoff tools are injected by the workflow step
type WorkflowStep struct {
	Agent Agent
	AgentTools []Tool
	HelperAgents []Agent
	AllowHandoffs boolean
	WorkflowTools []Workflow
}

func (ws *WorkflowStep) ValidateWorkflowstep() {
	//make sure there is at least 1 agent
}

type AgentSystemManager struct{
	CompatabillityCutoff int //unix timestamp for last point of compatability (point where users can/cannot pick back up a conversation)
	ActiveWorkflows map[string]struct{timestamp int, *WorkflowManager}
}

func(asm *AgentSystemManager)RebuildWorkflowManager(string)(
	//rebuild the state of the workflow based on the previous calls from database logs. 

)

func(asm *AgentSystemManager)startNewWorkflow(workflow, input){
	//build an instance of a workflow for a user/start the workflow 
}

//user connects if not logged in we issue a cookie to link them to their chat. 
//else we use userid to provide the link

//make sure the user owns the workflow 
//workflows are linked to user/cookieid
func(asm *AgentSystemManager)runWorkflow(workflowUUID, step, depth input)(ouput, error)
	wf := asm.GetWorkflow(workflowUUID)
	//if not found rebuild from database.
	//if not found return error.
	//validate that the workflow is present and the step/depth is present.
	//we could also theoretically just build "part" of the workflow and then search for the rest when needed. 
	err := wf.validate(step, depth, input)
	//validate input at the step of the workflow by walking through
	return wf.start() //runs until it completes or has a return to user action. 
}

func (wf *Workflow)start(step, depth, input)(output, error){
	step = wf.getStep(step, depth)
	var stepResponse
	for steps<totalsteps{
	stepResponse = step.run(input)
	step++
	depth=0
	}
	if stepResponse.toUser = true{
		wf.userAgentInstance.transmit(stepResponse) //send the whole object to the user that contains the step the depth the user interaction the message
	}
	wf.exit()
}

func (st *Step) Run(input){
	if st.allowHandoff
	//check if there is entrynode
	entryAgent=st.node[0]
	if entryAgent{
		entryAgent=entryAgent
	}

	exited:=false
	//step run loop
	for exited==false{
		agentResult, error = entryAgent.run()
		if agentResult.handoff{
			//check node list and verify its proper
			//else inject error tool msg in global context
			//then return to the same node unless max turns is reached in which case add "FINAL CALL sumarize your results you can no longer call any tools your next message must satisfy your goal"
		}
		if no handoff and no errors  exited=true
	}
	//now we run the exit process
	exitStep := 0 
	for exitStep<totalExitSteps{
		if err try again or go to next //run exit step //exit steps can be optional or manditory
		exitStep++
	}
	return(responseData)
}


// func (a *Agent) Preprocess() {
// }

// func (a *Agent) Run() {
// 	//run the actual query
// }
func(ag *Agent) Run{
	res = ag.MakeRequest
	//retry protocol
	//logging protocol
	msgs=[res.msg]
	if res.tools.length>0{
		//run the tool calls async
		waitgroup.add(len(res.tools.length)){
			go callEachTool
			push to msgs
		}
	}

	self contain toolcall errors as the result as they are always just text in the end anyway
	like toolcal resuolt: FAILED TO SUCCESFULLY GET DATA. 

	return msgs
}


func(ag *Agent) MakeRequest{
	

}


//use hashing to allow saving agents in db and not repeating by standardizing save format then we can use this to map what agents are for what flows. 

//toollogging should be selfcontaned but with the traceid 
//agentlogging (llm logging)

//Automatic helper/run fucn that creates separated instances at each level where needed. 