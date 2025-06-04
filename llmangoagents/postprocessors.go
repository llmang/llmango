package llmangoagents

//these are ran after the step completes but before the next step.
//they can inject extra context messages along with the message.

func SummarizationPostProcesssor() string {
	return "@@CONTEXTMESSAGE: _________"
}
