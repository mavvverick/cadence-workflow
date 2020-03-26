package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/YOVO-LABS/workflow/config"
	"github.com/YOVO-LABS/workflow/internal/adapter"

	lb "github.com/YOVO-LABS/workflow/workflows/jobprocessor"

	"github.com/google/uuid"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

//LeaderboardInterface ...
type LeaderboardInterface interface {
	CreateCron(ctx context.Context, cronTime string) (*workflow.Execution, error)
	TerminateCron(ctx context.Context, wfID string) error
}

//LeaderboardService ...
type LeaderboardService struct {
	CadenceAdapter adapter.CadenceAdapter
	Logger         *zap.Logger
}

//NewLeaderboardService ...
func NewLeaderboardService(config config.AppConfig) LeaderboardService {
	return LeaderboardService{
		Logger: config.Logger,
	}
}

//CreateCron ...
func (l *LeaderboardService) CreateCron(ctx context.Context, cronTime string) (*workflow.Execution, error) {
	cronSchedule := strings.Split(cronTime, ":")
	cronTime = fmt.Sprintf("%s %s %s %s %s", cronSchedule[1], cronSchedule[0], "*", "*", "*")

	workflowOptions := client.StartWorkflowOptions{
		ID:                              "leaderBoardProcess_" + uuid.New().String(),
		TaskList:                        lb.TaskList,
		ExecutionStartToCloseTimeout:    time.Hour * 2,
		DecisionTaskStartToCloseTimeout: time.Hour * 2,
		CronSchedule:                    cronTime,
	}

	execution, err := l.CadenceAdapter.CadenceClient.StartWorkflow(
		context.Background(),
		workflowOptions,
		lb.Workflow,
		uuid.New().String(),
	)
	if err != nil {
		return nil, err
	}
	return execution, nil
}

//TerminateCron ...
func (l *LeaderboardService) TerminateCron(ctx context.Context, wfID string) error {
	err := l.CadenceAdapter.CadenceClient.TerminateWorkflow(
		context.Background(),
		wfID,
		"",
		"",
		nil,
	)

	if err != nil {
		return err
	}
	return nil
}
