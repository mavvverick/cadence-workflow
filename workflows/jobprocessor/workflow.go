package jobprocessor

import (
	"time"

	"github.com/YOVO-LABS/workflow/api/model"
	"github.com/YOVO-LABS/workflow/internal/handler"
	"github.com/pborman/uuid"
	"go.uber.org/cadence"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// ApplicationName is the name of the tasklist
const (
	jobServerURL    = "http://localhost:4000"
	TaskList        = "JobProcessor"
	ApplicationName = "JobProcessor"
)

// HostID represents the hostname or the ip address of the worker
var HostID = ApplicationName + "_" + uuid.New()

func init() {
	workflow.Register(Workflow)
}

//ExecutionTime ..
type ExecutionTime struct {
	DownloadActivity float64
	EncodeActivity   float64
	UploadActivity   float64
	WFID             string
	FFExec0          string
	FFExec1          string
}

// Workflow workflow
func Workflow(ctx workflow.Context, jobID string, format model.Format) (result string, err error) {
	// creationActivityOptions := workflow.ActivityOptions{
	// 	ScheduleToStartTimeout: time.Minute * 30,
	// 	StartToCloseTimeout:    time.Minute * 30,
	// 	HeartbeatTimeout:       time.Minute * 30,
	// }

	// createJobContext := workflow.WithActivityOptions(ctx, creationActivityOptions)
	// createJoblogger := workflow.GetLogger(createJobContext)
	// createJoblogger.Info("Starting workflow")
	// createJoblogger.Info(format.Source)

	// sessionOptions := &workflow.SessionOptions{
	// 	CreationTimeout:  time.Minute * 30,
	// 	ExecutionTimeout: time.Minute * 30,
	// }
	// createJobSessionCtx, err := workflow.CreateSession(createJobContext, sessionOptions)
	// if err != nil {
	// 	return "", err
	// }
	// defer workflow.CompleteSession(createJobSessionCtx)

	// err = workflow.ExecuteActivity(createJobSessionCtx, createJobActivity, jobID).Get(createJobSessionCtx, nil)
	// if err != nil {
	// 	createJoblogger.Error("Created New Job", zap.Error(err))
	// 	return "", err
	// }
	cb := handler.NewCallbackInfo(&format)

	processingActivityOptions := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute * 30,
		StartToCloseTimeout:    time.Minute * 5,
		ScheduleToCloseTimeout: time.Minute * 30,
		HeartbeatTimeout:       time.Minute * 30,
		RetryPolicy: &cadence.RetryPolicy{
			InitialInterval:          time.Second,
			BackoffCoefficient:       2.0,
			MaximumInterval:          time.Minute * 4,
			ExpirationInterval:       time.Minute * 10,
			NonRetriableErrorReasons: []string{"bad-error"},
		},
	}
	processJobContext := workflow.WithActivityOptions(ctx, processingActivityOptions)
	logger := workflow.GetLogger(processJobContext)

	processJobSessionOptions := &workflow.SessionOptions{
		CreationTimeout:  time.Minute * 30,
		ExecutionTimeout: time.Minute * 30,
		HeartbeatTimeout: time.Minute * 30,
	}

	processJobSessionContext, err := workflow.CreateSession(processJobContext, processJobSessionOptions)
	if err != nil {
		logger.Error("Failed to create session context", zap.Error(err))
	}
	defer workflow.CompleteSession(processJobSessionContext)

	// var status string
	// err = workflow.ExecuteActivity(processJobSessionContext, waitForDecisionActivity,
	// 	jobID).Get(processJobSessionContext, &status)
	// if err != nil {
	// 	cb.PushMessage(err.Error(), "task", jobID, "error", format.Encode)
	// 	return "", err
	// }

	// if status != "APPROVED" {
	// 	logger.Info("Workflow completed.", zap.String("JobStatus", status))
	// 	return "", nil
	// }

	jobID = workflow.GetInfo(ctx).WorkflowExecution.ID
	var filePath string
	err = workflow.ExecuteActivity(processJobSessionContext, downloadFileActivity,
		jobID, format.Source).Get(processJobSessionContext, &filePath)

	if err != nil {
		cb.PushMessage("DOWNLOAD", "task", jobID, "error", format.Encode)
		logger.Info("Workflow completed with failed downloadFileActivity", zap.Error(err))
		return "", cadence.NewCustomError(err.Error(), "Failed DownloadActivity")
	}

	err = workflow.ExecuteActivity(processJobSessionContext, compressFileActivity,
		jobID, filePath, format).Get(processJobSessionContext, nil)
	if err != nil {
		cb.PushMessage("COMPRESSION", "task", jobID, "error", format.Encode)
		logger.Info("Workflow completed with failed compressFileActivity", zap.Error(err))
		return "", cadence.NewCustomError(err.Error(), "Failed compressFileActivity")
	}

	err = workflow.ExecuteActivity(processJobSessionContext, uploadFileActivity,
		jobID, filePath, format).Get(processJobSessionContext, nil)
	if err != nil {
		cb.PushMessage("UPLOADING", "task", jobID, "error", format.Encode)
		logger.Info("Workflow completed with failed uploadFileActivity", zap.Error(err))
		return "", cadence.NewCustomError(err.Error(), "Failed uploadFileActivity")
	}

	cb.PushMessage("COMPLETED", "task", jobID, "saved", format.Encode)

	// err = workflow.ExecuteActivity(processJobSessionContext, migrateToColdLineActivity,
	// 	jobID, format).Get(processJobSessionContext, nil)
	// if err != nil {
	// 	logger.Info("Workflow completed with failed migrateToColdLineActivity", zap.Error(err))
	// 	return "", err
	// }

	return "COMPLETED", nil
}
