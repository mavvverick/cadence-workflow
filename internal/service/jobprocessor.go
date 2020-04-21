package service

import (
	"context"
	"errors"
	"fmt"
	cron "github.com/YOVO-LABS/workflow/workflows/cron"
	"github.com/uber/cadence/common"
	"go.uber.org/cadence/.gen/go/shared"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/YOVO-LABS/workflow/api/model"
	"github.com/YOVO-LABS/workflow/internal/adapter"
	"github.com/YOVO-LABS/workflow/internal/handler"
	jp "github.com/YOVO-LABS/workflow/workflows/jobprocessor"

	"github.com/google/uuid"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

//JobProcessorInterface ...
type JobProcessorInterface interface {
	CreateJob(ctx context.Context, queryParams *model.QueryParams) (*workflow.Execution, error)
	NotifyJobStateChange(w http.ResponseWriter, r *http.Request) error
	GetJobInfo(ctx context.Context, workflowOption *model.Workflow) (interface{}, error)
	JobStatusCount(ctx context.Context, m string) (*model.WorkflowExecution, error)
	GetLogs(ctx context.Context, st, du string) error
	CreateCron(ctx context.Context, cronTime string) (*workflow.Execution, error)
}

//JobProcessorService ...
type JobProcessorService struct {
	CadenceAdapter adapter.CadenceAdapter
	KafkaAdapter adapter.KafkaAdapter
	Logger         *zap.Logger
}

// CreateJob ...
func (b *JobProcessorService) CreateJob(ctx context.Context, queryParams *model.QueryParams) (*workflow.Execution, error) {
	var encodes []model.Encode
	videoFormat := model.NewVideoFormat()

	query := queryParams.Query
	for _, format := range query.Format {
		mp4EncodeParams := model.NewMP4Encode()
		mp4EncodeParams.
			SetDestination(format.Destination.URL).
			SetSize(format.Size).
			SetVideoCodec(format.VideoCodec).
			SetFrameRate(format.Framerate).
			SetBitRate(format.Bitrate).
			SetBufferSize(format.Bitrate).
			SetMaxRate(format.Bitrate).
			SetVideoFormat(format.FileExtension)
		if (model.Logo{}) != format.Logo {
			mp4EncodeParams.SetWatermarkURL(format.Logo.Source)
			videoFormat.SetFormatWatermarkURL(format.Logo.Source)
		}
		encodes = append(encodes, mp4EncodeParams.GetEncode())
	}

	videoFormat.
		SetFormatSource(query.Source).
		SetFormatCallbackURL(query.CallbackURL).
		SetFormatPayload(query.Payload).
		SetFormatEncode(encodes)

	workflowOptions := client.StartWorkflowOptions{
		ID:                              "jobProcessing_" + uuid.New().String(),
		TaskList:                        jp.TaskList,
		ExecutionStartToCloseTimeout:    time.Hour * 24,
		DecisionTaskStartToCloseTimeout: time.Minute * 24,
	}

	execution, err := b.CadenceAdapter.CadenceClient.StartWorkflow(
		context.Background(),
		workflowOptions,
		jp.Workflow,
		uuid.New().String(),
		videoFormat.GetFormat(),
	)
	return execution, err
}

// NotifyJobStateChange ...
func (b *JobProcessorService) NotifyJobStateChange(w http.ResponseWriter, r *http.Request) error {
	isAPICall := r.URL.Query().Get("is_api_call") == "true"
	id := r.URL.Query().Get("id")
	actionType := r.URL.Query().Get("type")

	allExpense := handler.AllExpense
	oldState, ok := allExpense[id]
	if !ok {
		fmt.Println("ERROR:INVALID_ID")
		return nil
	}

	const (
		created   = "CREATED"
		approved  = "APPROVED"
		rejected  = "REJECTED"
		completed = "COMPLETED"
	)

	switch actionType {
	case "approve":
		allExpense[id] = approved
	case "reject":
		allExpense[id] = rejected
	case "processed":
		allExpense[id] = completed
	}
	if isAPICall {
		fmt.Fprint(w, "SUCCEED")
	} else {
		handler.ListHandler(w, r)
	}

	if oldState == created && (allExpense[id] == approved || allExpense[id] == rejected) {
		token, ok := handler.TokenMap[id]
		if !ok {
			fmt.Printf("Invalid id:%s\n", id)
			return nil
		}
		err := b.CadenceAdapter.CadenceClient.CompleteActivity(context.Background(), token, string(allExpense[id]), nil)
		if err != nil {
			fmt.Printf("Failed to complete activity with error: %+v\n", err)
		} else {
			fmt.Printf("Successfully complete activity: %s\n", token)
		}
	}
	fmt.Printf("Set state for %s from %s to %s.\n", id, oldState, allExpense[id])
	return nil
	// report state change
}

// GetJobInfo ...
func (b *JobProcessorService) GetJobInfo(ctx context.Context, workflowOption *model.Workflow) (interface{}, error) {
	describeWorkflowExecution, err := b.CadenceAdapter.CadenceClient.DescribeWorkflowExecution(ctx, workflowOption.WfID, workflowOption.RunID)
	if err != nil {
		return "nil", err
	}
	execTime := describeWorkflowExecution.WorkflowExecutionInfo.GetCloseTime() - describeWorkflowExecution.WorkflowExecutionInfo.GetStartTime()
	return execTime, nil
}

func (b *JobProcessorService) JobStatusCount(ctx context.Context, d string) (*model.WorkflowExecution, error) {
	var workflowInfo model.WorkflowExecution

	duration, err := strconv.Atoi(d)
	if err != nil {
		return nil, err
	}

	requestClosedWorkflow := &shared.ListClosedWorkflowExecutionsRequest{
		MaximumPageSize: common.Int32Ptr(int32(10000)),
		StartTimeFilter: &shared.StartTimeFilter{
			EarliestTime: common.Int64Ptr(time.Now().Add(time.Duration(-duration)*time.Minute).UnixNano()),
			LatestTime:   common.Int64Ptr(time.Now().UnixNano()),
		},
	}

	for true {
		listClosedWorkflow, err := b.CadenceAdapter.CadenceClient.ListClosedWorkflow(ctx, requestClosedWorkflow)
		if err != nil {
			return nil, err
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
		listOpenWorkflow, err := b.CadenceAdapter.CadenceClient.ListOpenWorkflow(ctx, requestOpenWorkflow)
		if err != nil {
			return nil, err
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
	//kafkaMsg, err := json.Marshal(&workflowInfo)
	//if err != nil {
	//	return nil, err
	//}
	//
	//if kafkaMsg != nil {
	//	err = b.KafkaAdapter.Producer.Publish(ctx, "video", string(kafkaMsg))
	//	if err != nil {
	//		return nil, err
	//	}
	//}
	return &workflowInfo, nil
}

func (b *JobProcessorService) GetLogs(ctx context.Context, st, du string) error {
	var workflowInfo model.WorkflowExecution

	duration, err := strconv.Atoi(du)
	if err != nil {
		return err
	}

	starttime, err := time.Parse("2006-01-02T15:04:05", st)
	if err != nil {
		return err
	}

	f, err := os.Create("/tmp/cadence-logs.csv")
	if err != nil {
		return err
	}

	requestClosedWorkflow := &shared.ListClosedWorkflowExecutionsRequest{
		MaximumPageSize: common.Int32Ptr(int32(10000)),
		StartTimeFilter: &shared.StartTimeFilter{
			EarliestTime: common.Int64Ptr(starttime.Add(time.Duration(-duration)*time.Minute).UnixNano()),
			LatestTime:   common.Int64Ptr(starttime.UnixNano()),
		},
	}

	for true {
		listClosedWorkflow, err := b.CadenceAdapter.CadenceClient.ListClosedWorkflow(ctx, requestClosedWorkflow)
		if err != nil {
			return err
		}
		for _, w := range listClosedWorkflow.Executions {
			fmt.Fprintln(f, w)
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
			EarliestTime: common.Int64Ptr(starttime.Add(time.Duration(-duration)*time.Minute).UnixNano()),
			LatestTime:   common.Int64Ptr(starttime.UnixNano()),
		},
	}

	for true {
		listOpenWorkflow, err := b.CadenceAdapter.CadenceClient.ListOpenWorkflow(ctx, requestOpenWorkflow)
		if err != nil {
			return err
		}
		for _, w := range listOpenWorkflow.Executions {
			fmt.Fprintln(f, w)
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
	return nil
}

//CreateCron ...
func (l *JobProcessorService) CreateCron(ctx context.Context, cronTime string) (*workflow.Execution, error) {
	cronSchedule := strings.Split(cronTime, " ")
	if len(cronSchedule) == 0 || len(cronSchedule) < 2 {
		return nil, errors.New("invalid cron expression")
	}
	cronTime = fmt.Sprintf("%s %s %s %s %s", cronSchedule[0], cronSchedule[1], "*", "*", "*")

	workflowOptions := client.StartWorkflowOptions{
		ID:                              "CRON_" + uuid.New().String(),
		TaskList:                        jp.TaskList,
		ExecutionStartToCloseTimeout:    time.Hour * 24,
		DecisionTaskStartToCloseTimeout: time.Hour * 24,
		CronSchedule:                    cronTime,
	}

	execution, err := l.CadenceAdapter.CadenceClient.StartWorkflow(
		context.Background(),
		workflowOptions,
		cron.Workflow,
		uuid.New().String(),
	)
	if err != nil {
		return nil, err
	}
	return execution, nil
}

