package leaderboard

import (
	"time"

	"github.com/pborman/uuid"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// ApplicationName is the name of the tasklist
const (
	jobServerURL    = "http://localhost:4000"
	TaskList        = "LeaderBoard"
	ApplicationName = "LeaderBoardCalculator"
)

// HostID represents the hostname or the ip address of the worker
var HostID = ApplicationName + "_" + uuid.New()

func init() {
	workflow.RegisterWithOptions(Workflow, workflow.RegisterOptions{Name: ApplicationName})
}

// Workflow workflow
func Workflow(ctx workflow.Context, jobID string) (result string, err error) {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute * 10,
		StartToCloseTimeout:    time.Minute * 10,
		HeartbeatTimeout:       time.Minute * 3,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Starting Workflow")

	sessionOptions := &workflow.SessionOptions{
		CreationTimeout:  time.Minute * 10,
		ExecutionTimeout: time.Minute * 10,
	}

	sessionCtx, err := workflow.CreateSession(ctx, sessionOptions)
	if err != nil {
		return "", err
	}
	defer workflow.CompleteSession(sessionCtx)

	err = workflow.ExecuteActivity(ctx, calculateLeaderBoard, jobID).Get(sessionCtx, nil)
	if err != nil {
		logger.Error("Failed to execute createJobActivity function", zap.Error(err))
		return "", err
	}

	return "COMPLETED", nil
}

//CancelTimer ...
// func CancelTimer(ctx workflow.Context, wfID string) (string, error) {
// 	logger := workflow.GetLogger(ctx)
// 	newCtx := workflow.WithWorkflowID(ctx, wfID)

// 	execInfo := workflow.GetInfo(newCtx).WorkflowExecution
// 	sigName := fmt.Sprintf("sig.%v", execInfo.RunID)
// 	sigCh := workflow.GetSignalChannel(ctx, sigName)

// 	selector := workflow.NewSelector(newCtx)

// 	_, cancelTimerHandler := workflow.WithCancel(newCtx)

// 	selector.AddReceive(sigCh, func(c workflow.Channel, more bool) {
// 		logger.Info("Cancel outstanding timer.")
// 		cancelTimerHandler()
// 	})
// 	return "", nil

// }

// err = workflow.ExecuteActivity(ctx, pushLeaderboardScore, chInfo).Get(ctx, nil)
// if err != nil {
// 	logger.Error("Failed to execute pushLeaderboardScore function", zap.Error(err))
// 	return "", err
// }

// currentDate := time.Now() //2020-03-06 HH:MM:SS Z
// timer := strings.Split(timerDuration, ":")
// timerHour, _ := strconv.Atoi(timer[0])
// timerMinute, _ := strconv.Atoi(timer[1])
// timerSecond, _ := strconv.Atoi(timer[2])

// year, month, day := currentDate.Year(), currentDate.Month(), currentDate.Day() //2020, 03, 06
// logger.Info(time.Date(year, month, day, timerHour, timerMinute, timerSecond, int(0), time.Local).String())
// waitDuration := time.Until(time.Date(year, month, day, timerHour, timerMinute, timerSecond, int(0), time.Local)) // until timer hour
// logger.Info(waitDuration.String())

// selector := workflow.NewSelector(ctx)
// timerFuture := workflow.NewTimer(ctx, time.Second*waitDuration)
// logger.Info((time.Second * waitDuration).String())

// selector.AddFuture(timerFuture, func(f workflow.Future) {
// workflow.ExecuteActivity(ctx, pushLeaderboardScore, chInfo).Get(ctx, nil)
// })

// selector.Select(ctx)
