package ai

import (
	"strings"
	"time"

	"go.uber.org/cadence"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

const (
	Tasklist                         = "AI"
	SessionCreationErrorMsg          = "Session Creation Failed"
	CheckNSFWActivityErrorMsg        = "Failed CheckNSFW Activity"
	CorrectWatermarkActivityErrorMsg = "Failed Correct Watermark Activity"
)

func init() {
	workflow.RegisterWithOptions(Workflow, workflow.RegisterOptions{Name: Tasklist})
}

// Workflow Session Based to perform nsfw check and watermark correction
func Workflow(ctx workflow.Context, jobID string, payload string) error {
	logger := workflow.GetLogger(ctx)
	//cb := jobprocessor.NewCallbackInfo(&format)
	exec := workflow.GetInfo(ctx).WorkflowExecution

	jobID = exec.ID

	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute * 5,
		ScheduleToCloseTimeout: time.Minute * 5,
		HeartbeatTimeout:       time.Minute * 3,
		RetryPolicy: &cadence.RetryPolicy{
			InitialInterval:          time.Second,
			BackoffCoefficient:       2.0,
			MaximumInterval:          time.Minute * 5,
			ExpirationInterval:       time.Hour * 10,
			MaximumAttempts:          2,
			NonRetriableErrorReasons: []string{"bad-error"},
		},
	}
	jobCtx := workflow.WithActivityOptions(ctx, ao)

	so := &workflow.SessionOptions{
		CreationTimeout:  time.Hour * 24,
		ExecutionTimeout: time.Minute * 5,
		HeartbeatTimeout: time.Minute * 3,
	}

	ctx, err := workflow.CreateSession(jobCtx, so)
	if err != nil {
		logger.Error(SessionCreationErrorMsg, zap.Error(err))
		return cadence.NewCustomError(err.Error(), SessionCreationErrorMsg)
	}
	defer workflow.CompleteSession(ctx)

	postID := strings.Split(payload, "|")[0]
	err = workflow.ExecuteActivity(ctx, checkNSFWAndLogoActivity,
		jobID, postID).Get(ctx, nil)
	if err != nil {
		logger.Error(CheckNSFWActivityErrorMsg, zap.Error(err))
		//cb.PushMessage(NSFW, Task, jobID, CallbackErrorEvent)
		return cadence.NewCustomError(err.Error(), CheckNSFWActivityErrorMsg)
	}

	//cb.PushMessage(Completed, Task, jobID, "saved")
	return nil
}
