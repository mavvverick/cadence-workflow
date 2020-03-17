package model

import "context"

//Workflow ...
type Workflow struct {
	WfID  string `json:"wfId"  validate:"nonzero"`
	RunID string `json:"rid,omitempty"`
}

// Validate ...
func (c *Workflow) Validate(ctx context.Context) error {
	return ValidateFields(c)
}
