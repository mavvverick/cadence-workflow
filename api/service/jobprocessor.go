package service

import (
	"context"
	"fmt"
	"net/http"
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
	CreateJob(ctx context.Context, query *model.Query) (*workflow.Execution, error)
	NotifyJobStateChange(w http.ResponseWriter, r *http.Request) error
}

//JobProcessorService ...
type JobProcessorService struct {
	CadenceAdapter adapter.CadenceAdapter
	Logger         *zap.Logger
}

// CreateJob ...
func (b *JobProcessorService) CreateJob(ctx context.Context, query *model.Query) (*workflow.Execution, error) {
	var encodes []model.Encode
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
			SetVideoFormat(format.FileExtension).
			GetEncode()
		encodes = append(encodes, mp4EncodeParams.GetEncode())
	}

	videoFormat := model.NewVideoFormat()
	videoFormat.
		SetFormatSource(query.Source).
		SetFormatEncode(encodes).
		GetFormat()

	workflowOptions := client.StartWorkflowOptions{
		ID:                              "jobProcessing_" + uuid.New().String(),
		TaskList:                        jp.TaskList,
		ExecutionStartToCloseTimeout:    time.Minute * 5,
		DecisionTaskStartToCloseTimeout: time.Minute * 5,
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

	fmt.Println(id, actionType, isAPICall)

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
