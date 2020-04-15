package model

import "context"

//Workflow ...
type Workflow struct {
	WfID  string `json:"wfId"  validate:"nonzero"`
	RunID string `json:"rid,omitempty"`
}

type DownloadObject struct {
	VideoPath	string
	Watermark 	string
	UserImage 	string
	Background	string
	Font		string
}
// Validate ...
func (c *Workflow) Validate(ctx context.Context) error {
	return ValidateFields(c)
}
