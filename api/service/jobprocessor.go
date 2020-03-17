package service

import (
	"context"
	"jobprocessor/api/model"
	"jobprocessor/internal/adapter"
	jp "jobprocessor/workflow/jobprocessor"
	"time"

	"github.com/google/uuid"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

//JobProcessorInterface ...
type JobProcessorInterface interface {
	CreateJob(ctx context.Context, query *model.Query) (*workflow.Execution, error)
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
		ExecutionStartToCloseTimeout:    time.Minute,
		DecisionTaskStartToCloseTimeout: time.Minute,
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
