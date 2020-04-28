package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/YOVO-LABS/workflow/workflows/cron"
	"github.com/uber/cadence/common"
	"go.uber.org/cadence/.gen/go/shared"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/YOVO-LABS/workflow/api/model"
	ca "github.com/YOVO-LABS/workflow/common/cadence"
	ka "github.com/YOVO-LABS/workflow/common/messaging"
	"github.com/YOVO-LABS/workflow/common/mysql"
	jp "github.com/YOVO-LABS/workflow/workflows/jobprocessor"

	"github.com/google/uuid"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

//JobProcessorService ...
type JobProcessorService struct {
	CadenceAdapter ca.CadenceAdapter
	KafkaAdapter ka.KafkaAdapter
	DB 	mysql.DB
	Logger *zap.Logger
}

// CreateJob ...
func (b *JobProcessorService) CreateJob(ctx context.Context, queryParams *model.QueryParams) (*workflow.Execution, error) {
	var encodes []jp.Encode
	var videoFormat jp.Format

	query := queryParams.Query
	for _, format := range query.Format {
		var encode jp.Encode
		encode.Destination=format.Destination.URL
		encode.Size=format.Size
		encode.VideoCodec=format.VideoCodec
		encode.FrameRate=format.Framerate
		encode.BitRate=format.Bitrate
		encode.BufferSize=format.Bitrate
		encode.MaxRate=format.Bitrate
		encode.VideoFormat=format.FileExtension

		if (jp.Logo{}) != format.Logo {
			encode.Logo.Source=format.Logo.Source
			videoFormat.WatermarkURL=format.Logo.Source
		}
		encodes = append(encodes, encode)
	}

	videoFormat.Source=query.Source
	videoFormat.CallbackURL=query.CallbackURL
	videoFormat.Payload=query.Payload
	videoFormat.Encode=encodes

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
		videoFormat,
	)
	return execution, err
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
		MaximumPageSize: common.Int32Ptr(int32(10000)),
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
		MaximumPageSize: common.Int32Ptr(int32(10000)),
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

func (l *JobProcessorService) GetData(ctx context.Context, dateRange *model.DataRange) (*[]interface{}, error) {
	var result []interface{}
	rows, err := l.DB.Client.
		Table("job_infos").
		Select("type as Type, size as Size, sum(cost) as Cost, cast(created_at as char) as Date").
		Group("created_at, type, size").
		Where("created_at BETWEEN ? AND ?", &dateRange.Starttime, &dateRange.Endtime).
		Rows()
	if err != nil {
		return nil, err
	}
	//var res *model.JobData

	for rows.Next() {
		var d model.JobData
		rows.Scan(&d.Type, &d.Size, &d.Cost, &d.Date)
		result = append(result, d)
	}
	return &result, nil
}

