package jobprocessor

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloud.google.com/go/storage"
	ca "github.com/YOVO-LABS/workflow/common/cadence"
	ka "github.com/YOVO-LABS/workflow/common/messaging"
	"github.com/YOVO-LABS/workflow/config"
	"google.golang.org/api/option"

	"go.uber.org/cadence/worker"
	"go.uber.org/zap"
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
	gcsClient, err := storage.NewClient(context.Background(),
		option.WithCredentialsJSON([]byte(os.Getenv("GOOGLE_JSON"))))
	if err != nil {
		fmt.Println("Cannot initiate GCS Client")
		return
	}
	w.taskList = tasklist
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
		ctx := context.WithValue(context.Background(), "kafkaClient", w.kafkaAdapter)
		ctx = context.WithValue(ctx, "cadenceClient", w.cadenceAdapter)
		ctx = context.WithValue(ctx, "gcsClient", gcsClient)

		workerOptions.BackgroundActivityContext = ctx

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
	w.options = workerOptions
}

//Start ...
func (w *Worker) Start() {
	// Configure worker options.

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
