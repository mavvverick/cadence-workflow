package cron

import (
	"context"
	"encoding/json"
	"os"

	ca "github.com/YOVO-LABS/workflow/common/cadence"
	ka "github.com/YOVO-LABS/workflow/common/messaging"
	"github.com/uber/cadence/common"
	"go.uber.org/cadence"
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"

	"time"
)

const writeWorkflowInfoName = "writeWorkflowInfo"

func init() {
	workflow.RegisterWithOptions(Workflow, workflow.RegisterOptions{Name: "Cron Scaler"})

	activity.RegisterWithOptions(
		writeWorkflowInfo,
		activity.RegisterOptions{Name: writeWorkflowInfoName},
	)
}

func Workflow(ctx workflow.Context, jobID string) error {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Second * 20,
		HeartbeatTimeout:       time.Second * 10,
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
	logger := workflow.GetLogger(jobCtx)

	err := workflow.ExecuteActivity(jobCtx, writeWorkflowInfo, jobID).Get(jobCtx, nil)
	if err != nil {
		logger.Error("Failed to execute writeWorkflowInfo function", zap.Error(err))
		return err
	}

	return  nil
}

func writeWorkflowInfo(ctx context.Context, jobID string) error {
	var workflowInfo WorkflowExecution

	cadenceClient  := ctx.Value("cadenceClient").(ca.CadenceAdapter)
	kafkaClient  := ctx.Value("kafkaClient").(ka.KafkaAdapter)

	duration :=60

	requestClosedWorkflow := &shared.ListClosedWorkflowExecutionsRequest{
		MaximumPageSize: common.Int32Ptr(int32(10000)),
		StartTimeFilter: &shared.StartTimeFilter{
			EarliestTime: common.Int64Ptr(time.Now().Add(time.Duration(-duration)*time.Minute).UnixNano()),
			LatestTime:   common.Int64Ptr(time.Now().UnixNano()),
		},
	}

	for true {
		listClosedWorkflow, err := cadenceClient.CadenceClient.ListClosedWorkflow(ctx, requestClosedWorkflow)
		if err != nil {
			return err
		}
		if len(listClosedWorkflow.Executions) != 0  {
			workflowInfo.Total += len(listClosedWorkflow.Executions)
			for _, w := range listClosedWorkflow.Executions {
				switch *w.CloseStatus {
				case shared.WorkflowExecutionCloseStatusCompleted:
					workflowInfo.Completed++
				case shared.WorkflowExecutionCloseStatusFailed:
					workflowInfo.Failed++
				case shared.WorkflowExecutionCloseStatusCanceled:
					workflowInfo.Cancelled++
				case shared.WorkflowExecutionCloseStatusTerminated:
					workflowInfo.Terminated++
				case shared.WorkflowExecutionCloseStatusTimedOut:
					workflowInfo.Timeout++
				}
			}
		} else {
			break
		}

		if len(listClosedWorkflow.NextPageToken) != 0 {
			requestClosedWorkflow.NextPageToken = listClosedWorkflow.NextPageToken
		} else {
			break
		}
	}

	requestOpenWorkflow := &shared.ListOpenWorkflowExecutionsRequest{
		MaximumPageSize: common.Int32Ptr(int32(3)),
		StartTimeFilter: &shared.StartTimeFilter{
			EarliestTime: common.Int64Ptr(time.Now().Add(time.Duration(-duration)*time.Minute).UnixNano()),
			LatestTime:   common.Int64Ptr(time.Now().UnixNano()),
		},
	}

	for true {
		listOpenWorkflow, err := cadenceClient.CadenceClient.ListOpenWorkflow(ctx, requestOpenWorkflow)
		if err != nil {
			return err
		}
		if len(listOpenWorkflow.Executions) != 0  {
			workflowInfo.Open = len(listOpenWorkflow.Executions)
			workflowInfo.Total += workflowInfo.Open
		}
		if len(listOpenWorkflow.NextPageToken) != 0 {
			requestOpenWorkflow.NextPageToken = listOpenWorkflow.NextPageToken
		} else {
			break
		}
	}

	//taskListResponse, err := b.CadenceAdapter.CadenceClient.DescribeTaskList(ctx, jp.TaskList, shared.TaskListTypeActivity)
	//workflowInfo.Pollers = len(taskListResponse.Pollers)

	//push message to kafka
	kafkaMsg, err := json.Marshal(&workflowInfo)
	if err != nil {
		return err
	}

	if kafkaMsg != nil {
		err = kafkaClient.Producer.Publish(os.Getenv("CRON_TOPIC"), "video", string(kafkaMsg))
		if err != nil {
			return err
		}
		//err = kafkaClient.Producer.Close()
		//if err != nil {
		//	return err
		//}
	}
	return nil
}
