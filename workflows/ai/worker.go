package ai

import (
	"context"
	"fmt"
	ca "github.com/YOVO-LABS/workflow/common/cadence"
	ka "github.com/YOVO-LABS/workflow/common/messaging"
	"github.com/YOVO-LABS/workflow/config"
	"github.com/YOVO-LABS/workflow/internal/grpc"
	"go.uber.org/cadence/worker"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

//Worker ...
type Worker struct {
	config         config.AppConfig
	taskList       string
	cadenceAdapter ca.CadenceAdapter
	kafkaAdapter   ka.KafkaAdapter
	options        worker.Options
}

//New ...
func New(configPath string) *Worker {
	var appConfig config.AppConfig
	appConfig.LoadConfig(configPath)

	return &Worker{
		config: appConfig,
	}
}

// Init ...
func (w *Worker) Init(tasklist, verbose, workerType string) {
	//start dependency injection
	w.cadenceAdapter.Setup(&w.config.Cadence)
	w.kafkaAdapter.Setup(&w.config.Kafka)
	w.taskList = tasklist
	workerOptions := worker.Options {
		MetricsScope:          w.cadenceAdapter.Scope,
		EnableLoggingInReplay: true,
	}
	if verbose == "0" {
		workerOptions.Logger = zap.NewNop()
	} else {
		workerOptions.Logger = w.cadenceAdapter.Logger
	}

	if workerType == "activity" {
		mlClientConn, err := grpc.PredictgRPCConnection()
		if err != nil {
			fmt.Println("Error ml client ", err)
		}
		ctx := context.WithValue(context.Background(), "mlClient", mlClientConn)
		ctx = context.WithValue(ctx, "kafkaClient", w.kafkaAdapter)
		workerOptions.BackgroundActivityContext = ctx

		workerOptions.EnableSessionWorker = true
		workerOptions.DisableWorkflowWorker = true
		workerOptions.DisableActivityWorker = false
		workerOptions.MaxConcurrentSessionExecutionSize = 10
		workerOptions.WorkerStopTimeout = time.Second * 10

	} else if workerType == "workflow" {
		workerOptions.EnableSessionWorker = false
		workerOptions.DisableWorkflowWorker = false
		workerOptions.DisableActivityWorker = true
		workerOptions.WorkerStopTimeout = time.Second * 10
	}
	w.options = workerOptions
}

//Start ...
func (w *Worker) Start() {
	cadenceWorker := worker.New(w.cadenceAdapter.ServiceClient, w.config.Cadence.Domain, w.taskList, w.options)
	err := cadenceWorker.Start()
	if err != nil {
		w.cadenceAdapter.Logger.Error("Failed to start workers.", zap.Error(err))
		panic("Failed to start workers")
	}

	done := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		killSignal := <-sigint
		switch killSignal {
		case os.Interrupt:
			log.Print("Got SIGINT...")
		case syscall.SIGTERM:
			log.Print("Got SIGTERM...")
		}
		time.Sleep(time.Second * 5)
		cadenceWorker.Stop()
		close(done)
	}()
	<-done
}
