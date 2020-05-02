package jobprocessor

import (
	"github.com/YOVO-LABS/workflow/workflows/ai"
	"time"

	"go.uber.org/cadence"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

const (
	TaskList                    = "JobProcessor"
	ChildWorkflowExecErrMsg		= "Child Workflow execution failed"
	SessionCreationErrorMsg     = "Session Creation Failed"
	DownloadActivityErrorMsg    = "Failed Download Activity"
	CompressionActivityErrorMsg = "Failed Compression Activity"
	UploadActivityErrorMsg      = "Failed Upload Activity"
	Download                    = "DOWNLOAD"
	Compression                 = "COMPRESSION"
	Upload                      = "UPLOAD"
	Task                        = "task"
	CallbackErrorEvent          = "error"
	Completed                   = "COMPLETED"
)

func init() {
	workflow.RegisterWithOptions(Workflow, workflow.RegisterOptions{Name:TaskList})
	//workflow.RegisterWithOptions(ai.Workflow, workflow.RegisterOptions{Name:ai.Tasklist})
}

// Workflow Session Based to perform download, compression and upload
func Workflow(ctx workflow.Context, jobID string, format Format) error {
	logger := workflow.GetLogger(ctx)
	cb := NewCallbackInfo(&format)
	exec := workflow.GetInfo(ctx).WorkflowExecution

	jobID = exec.ID
	runID := exec.RunID

	cwo := workflow.ChildWorkflowOptions{
		WorkflowID:                     runID,
		TaskList:                       ai.Tasklist,
		ExecutionStartToCloseTimeout:   time.Hour * 24,
		TaskStartToCloseTimeout:        time.Minute * 24,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)
	err := workflow.ExecuteChildWorkflow(ctx, ai.Workflow,
		runID, format.Source).Get(ctx, nil)
	if err != nil {
		logger.Error(ChildWorkflowExecErrMsg, zap.Error(err))
		return cadence.NewCustomError(err.Error(), ChildWorkflowExecErrMsg)
	}

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

	ctx, err = workflow.CreateSession(jobCtx, so)
	if err != nil {
		logger.Error(SessionCreationErrorMsg, zap.Error(err))
		return cadence.NewCustomError(err.Error(), SessionCreationErrorMsg)
	}
	defer workflow.CompleteSession(ctx)

	var dO DownloadObject

	err = workflow.ExecuteActivity(ctx, downloadFileActivity,
		jobID, format.Source, format.Payload, format.WatermarkURL).Get(ctx, &dO)
	if err != nil {
		logger.Error(DownloadActivityErrorMsg, zap.Error(err))
		cb.PushMessage(Download, Task, jobID, CallbackErrorEvent)
		return cadence.NewCustomError(err.Error(), DownloadActivityErrorMsg)
	}

	err = workflow.ExecuteActivity(ctx, compressMediaActivity,
		jobID, dO, format).Get(ctx, nil)
	if err != nil {
		logger.Error(CompressionActivityErrorMsg, zap.Error(err))
		cb.PushMessage(Compression, Task, jobID, CallbackErrorEvent)
		return cadence.NewCustomError(err.Error(), CompressionActivityErrorMsg)
	}

	err = workflow.ExecuteActivity(ctx, uploadFileActivity,
		jobID, dO.VideoPath, format).Get(ctx, nil)
	if err != nil {
		logger.Error(UploadActivityErrorMsg, zap.Error(err))
		cb.PushMessage(Upload, Task, jobID, CallbackErrorEvent)
		return cadence.NewCustomError(err.Error(), UploadActivityErrorMsg)
	}

	cb.PushMessage(Completed, Task, jobID, "saved")
	return nil
}
