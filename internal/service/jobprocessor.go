package service

import (
	"context"
	"fmt"
	"github.com/uber/cadence/common"
	"go.uber.org/cadence/.gen/go/shared"
	"net/http"
	"strconv"
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
	ListJob(ctx context.Context, m string) (*model.WorkflowExecution, error)
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

func (b *JobProcessorService) ListJob(ctx context.Context, d string) (*model.WorkflowExecution, error) {
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

		if listClosedWorkflow.NextPageToken != nil {
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
		if listOpenWorkflow.NextPageToken != nil {
			requestOpenWorkflow.NextPageToken = listOpenWorkflow.NextPageToken
		} else {
			break
		}
	}

	//push message to kafka
	//kafkaMsg, err := json.Marshal(&workflowInfo)
	//if err != nil {
	//	return nil, err
	//}
	//
	//if kafkaMsg != nil {
	//	err = b.KafkaAdapter.Producer.Publish(ctx, string(time.Now().UnixNano()), string(kafkaMsg))
	//	if err != nil {
	//		return nil, err
	//	}
	//}

	return &workflowInfo, nil
}