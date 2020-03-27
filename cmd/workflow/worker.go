package workflow

import (
	"github.com/YOVO-LABS/workflow/config"
	ca "github.com/YOVO-LABS/workflow/internal/adapter"

	"go.uber.org/cadence/worker"
	"go.uber.org/zap"
)

//WorkerInterface ...
type WorkerInterface interface {
	Init(taskList string)

	Start()
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
func (w *Worker) Start() {
	// Configure worker options.
	workerOptions := worker.Options{
		MetricsScope:          w.cadenceAdapter.Scope,
		Logger:                w.cadenceAdapter.Logger,
		EnableLoggingInReplay: true,
		EnableSessionWorker:   true,
	}

	cadenceWorker := worker.New(w.cadenceAdapter.ServiceClient, w.config.Cadence.Domain, w.taskList, workerOptions)
	err := cadenceWorker.Start()
	if err != nil {
		w.cadenceAdapter.Logger.Error("Failed to start workers.", zap.Error(err))
		panic("Failed to start workers")
	}
}
