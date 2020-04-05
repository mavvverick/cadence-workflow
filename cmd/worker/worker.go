package worker

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/YOVO-LABS/workflow/config"
	ca "github.com/YOVO-LABS/workflow/internal/adapter"

	"go.uber.org/cadence/worker"
	"go.uber.org/zap"
)

//WorkerInterface ...
type WorkerInterface interface {
	Init(taskList string)

	Start(verbose, workerType string)
}

//Worker ...
type Worker struct {
	config         config.AppConfig
	taskList       string
	cadenceAdapter ca.CadenceAdapter
}

//New ...
func New(configPath string) WorkerInterface {
	var appConfig config.AppConfig
	appConfig.LoadConfig(configPath)

	return &Worker{
		config: appConfig,
	}
}

// Init ...
func (w *Worker) Init(tasklist string) {
	//start dependency injection
	w.cadenceAdapter.Setup(&w.config.Cadence)
	w.taskList = tasklist
}

//Start ...
func (w *Worker) Start(verbose, workerType string) {
	// Configure worker options.
	workerOptions := worker.Options{
		MetricsScope:          w.cadenceAdapter.Scope,
		EnableLoggingInReplay: true,
	}
	if verbose == "0" {
		workerOptions.Logger = zap.NewNop()
	} else {
		workerOptions.Logger = w.cadenceAdapter.Logger
	}

	if workerType == "activity" {
		workerOptions.EnableSessionWorker = true
		workerOptions.DisableWorkflowWorker = true
		workerOptions.DisableActivityWorker = false
		workerOptions.MaxConcurrentSessionExecutionSize = 1
		workerOptions.WorkerStopTimeout = time.Second * 10
	} else if workerType == "workflow" {
		workerOptions.EnableSessionWorker = false
		workerOptions.DisableWorkflowWorker = false
		workerOptions.DisableActivityWorker = true
		workerOptions.WorkerStopTimeout = time.Second * 10
	}

	cadenceWorker := worker.New(w.cadenceAdapter.ServiceClient, w.config.Cadence.Domain, w.taskList, workerOptions)
	err := cadenceWorker.Start()
	if err != nil {
		w.cadenceAdapter.Logger.Error("Failed to start workers.", zap.Error(err))
		panic("Failed to start workers")
	}

	done := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint
		time.Sleep(time.Second * 5)
		cadenceWorker.Stop()
		close(done)
	}()
	<-done
}
