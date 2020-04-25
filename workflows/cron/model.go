package cron

type WorkflowExecution struct {
	Completed	int
	Failed		int
	Open		int
	Cancelled 	int
	Timeout		int
	Terminated 	int
	Total 		int
	//Pollers		int
}