package jobprocessor

import (
	"time"

	"jobprocessor/api/model"

	"github.com/pborman/uuid"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// ApplicationName is the name of the tasklist
const (
	jobServerURL    = "http://localhost:4000"
	TaskList        = "JobProcessing"
	ApplicationName = "JobProcessor"
)

// HostID represents the hostname or the ip address of the worker
var HostID = ApplicationName + "_" + uuid.New()

func init() {
	workflow.Register(Workflow)
}

// Workflow workflow
func Workflow(ctx workflow.Context, jobID string, format model.Format) (result string, err error) {
	creationActivityOptions := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute * 10,
		StartToCloseTimeout:    time.Minute * 10,
		HeartbeatTimeout:       time.Second * 20,
	}

	createJobContext := workflow.WithActivityOptions(ctx, creationActivityOptions)
	createJoblogger := workflow.GetLogger(createJobContext)
	createJoblogger.Info("Starting workflow")
	createJoblogger.Info(format.Source)

	sessionOptions := &workflow.SessionOptions{
		CreationTimeout:  time.Minute * 10,
		ExecutionTimeout: time.Minute * 10,
	}
	createJobSessionCtx, err := workflow.CreateSession(createJobContext, sessionOptions)
	if err != nil {
		return "", err
	}
	defer workflow.CompleteSession(createJobSessionCtx)

	err = workflow.ExecuteActivity(createJobSessionCtx, createJobActivity, jobID).Get(createJobSessionCtx, nil)
	if err != nil {
		createJoblogger.Error("Failed to execute createJobActivity function", zap.Error(err))
		return "", err
	}

	processingActivityOptions := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute * 10,
		StartToCloseTimeout:    time.Minute * 10,
	}
	processJobContext := workflow.WithActivityOptions(ctx, processingActivityOptions)
	logger := workflow.GetLogger(processJobContext)

	processJobSessionOptions := &workflow.SessionOptions{
		CreationTimeout:  time.Minute * 10,
		ExecutionTimeout: time.Minute * 10,
	}

	processJobSessionContext, err := workflow.CreateSession(processJobContext, processJobSessionOptions)
	if err != nil {
		logger.Error("Failed to create session context", zap.Error(err))
	}
	defer workflow.CompleteSession(processJobSessionContext)

	var status string
	err = workflow.ExecuteActivity(processJobSessionContext, waitForDecisionActivity,
		jobID).Get(processJobSessionContext, &status)
	if err != nil {
		return "", err
	}

	if status != "APPROVED" {
		logger.Info("Workflow completed.", zap.String("JobStatus", status))
		return "", nil
	}

	var filePath string
	err = workflow.ExecuteActivity(processJobSessionContext, downloadFileActivity,
		jobID, format.Source).Get(processJobSessionContext, &filePath)

	if err != nil {
		logger.Info("Workflow completed with failed downloadFileActivity", zap.Error(err))
		return "", err
	}

	var encodeFlag string
	err = workflow.ExecuteActivity(processJobSessionContext, compressFileActivity,
		jobID, filePath, format).Get(processJobSessionContext, &encodeFlag)
	if err != nil || encodeFlag == "FAILED" {
		logger.Info("Workflow completed with failed processJobSessionContext", zap.Error(err))
		return "", err
	}

	err = workflow.ExecuteActivity(processJobSessionContext, uploadFileActivity,
		jobID, format).Get(processJobSessionContext, nil)
	if err != nil {
		logger.Info("Workflow completed with failed uploadFileActivity", zap.Error(err))
		return "", err
	}

	err = workflow.ExecuteActivity(processJobSessionContext, migrateToColdLineActivity,
		jobID, format).Get(processJobSessionContext, nil)

	if err != nil {
		logger.Info("Workflow completed with failed migrateToColdLineActivity", zap.Error(err))
		return "", err
	}

	return "COMPLETED", nil
}
