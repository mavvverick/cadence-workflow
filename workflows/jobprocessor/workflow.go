package jobprocessor

import (
	"os"
	"strings"
	"time"

	"github.com/YOVO-LABS/workflow/proto/dense"

	"go.uber.org/cadence"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

const (
	TaskList                    = "JobProcessor"
	ChildWorkflowExecErrMsg     = "Child Workflow execution failed"
	SessionCreationErrorMsg     = "Session Creation Failed"
	DownloadActivityErrorMsg    = "Failed Download Activity"
	CompressionActivityErrorMsg = "Failed Compression Activity"
	UploadActivityErrorMsg      = "Failed Upload Activity"
	Download                    = "DOWNLOAD"
	Compression                 = "COMPRESSION"
	Upload                      = "UPLOAD"
	Task                        = "task"
	CallbackErrorEvent          = "error"
	CallbackRejectEvent         = "AI_REJECTED"
	Completed                   = "COMPLETED"
)

func init() {
	workflow.RegisterWithOptions(Workflow, workflow.RegisterOptions{Name: TaskList})
	//workflow.RegisterWithOptions(ai.Workflow, workflow.RegisterOptions{Name:ai.Tasklist})
}

// Workflow Session Based to perform transcoding with AI child workflow to detect nsfw
func Workflow(ctx workflow.Context, jobID string, format Format) error {
	bucketName := strings.Split(strings.Split(format.Source, "//")[1], ".")[0]
	logger := workflow.GetLogger(ctx)
	cb := NewCallbackInfo(&format, bucketName)
	exec := workflow.GetInfo(ctx).WorkflowExecution

	jobID = exec.ID
	runID := exec.RunID
	so := &workflow.SessionOptions{
		CreationTimeout:  time.Hour * 24,
		ExecutionTimeout: time.Minute * 5,
		HeartbeatTimeout: time.Minute * 3,
	}
	ctx, err := workflow.CreateSession(ctx, so)
	if err != nil {
		logger.Error(SessionCreationErrorMsg, zap.Error(err))
		return cadence.NewCustomError(err.Error(), SessionCreationErrorMsg)
	}
	defer workflow.CompleteSession(ctx)

	cwo := workflow.ChildWorkflowOptions{
		WorkflowID:                   runID,
		TaskList:                     "AI",
		ExecutionStartToCloseTimeout: time.Hour * 24,
		TaskStartToCloseTimeout:      time.Minute * 24,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	payloadSplit := strings.Split(format.Payload, "|")
	if len(payloadSplit) == 4 && payloadSplit[3] == "true" && os.Getenv("Is_AI_ALLOW") == "true" {
		var predictResult dense.Response
		err = workflow.ExecuteChildWorkflow(ctx, "AI",
			runID, format.Payload, bucketName, cb).Get(ctx, &predictResult)
		if err != nil {
			logger.Error(ChildWorkflowExecErrMsg, zap.Error(err))
			if cadence.IsCustomError(err) {
				return cadence.NewCustomError(ChildWorkflowExecErrMsg, err.Error())
			}
			_, cancel := workflow.WithCancel(ctx)
			cancel()
			return cadence.NewCanceledError(ChildWorkflowExecErrMsg, err.Error())
		}
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
	ctx = workflow.WithActivityOptions(ctx, ao)

	var dO DownloadObject

	err = workflow.ExecuteActivity(ctx, downloadFileActivity,
		jobID, format.Source, format.Payload, format.WatermarkURL, cb).Get(ctx, &dO)
	if err != nil {
		logger.Error(DownloadActivityErrorMsg, zap.Error(err))
		return cadence.NewCustomError(err.Error(), DownloadActivityErrorMsg)
	}

	err = workflow.ExecuteActivity(ctx, compressMediaActivity,
		jobID, dO, format, cb).Get(ctx, nil)
	if err != nil {
		logger.Error(CompressionActivityErrorMsg, zap.Error(err))
		return cadence.NewCustomError(err.Error(), CompressionActivityErrorMsg)
	}

	err = workflow.ExecuteActivity(ctx, uploadFileActivity,
		jobID, dO.VideoPath, format, cb).Get(ctx, nil)
	if err != nil {
		logger.Error(UploadActivityErrorMsg, zap.Error(err))
		return cadence.NewCustomError(err.Error(), UploadActivityErrorMsg)
	}
	return nil
}
