package model

import "context"

//Cron ...
type Cron struct {
	Time string `json:"time"  validate:"nonzero"`
}

// Validate ...
func (c *Cron) Validate(ctx context.Context) error {
	return ValidateFields(c)
}

//GetTime ...
func (c *Cron) GetTime(ctx context.Context) string {
	return c.Time
}
