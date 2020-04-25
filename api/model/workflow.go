package model

import (
	"context"
)

//Workflow ...
type Workflow struct {
	WfID  string `json:"wfId"  validate:"nonzero"`
	RunID string `json:"rid,omitempty"`
}

type Meta struct {
	Duration float64
	Size     float64
	Bitrate  int
}

type DataRange struct {
	Starttime 	int64	`json:"starttime,omitempty"`
	Endtime		int64	`json:"endtime,omitempty"`
	Duration	int64	`json:"duration,omitempty"`
}

type JobCount struct {
	Success 	int
	Failed 		int
}

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

// Validate ...
func (c *Workflow) Validate(ctx context.Context) error {
	return ValidateFields(c)
}

func (c *DataRange) Validate(ctx context.Context) error {
	return ValidateFields(c)
}
