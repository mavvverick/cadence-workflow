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
	// kafkaAdapter   ka.KafkaAdapter
	options worker.Options
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
	workerOptions := worker.Options{
		MetricsScope:          w.cadenceAdapter.Scope,
		EnableLoggingInReplay: true,
	}

	if workerType == "activity" {
		gcsClient, err := storage.NewClient(context.Background(),
			option.WithCredentialsJSON([]byte(os.Getenv("GOOGLE_JSON"))))
		if err != nil {
			fmt.Println("Cannot initiate GCS Client")
			return
		}
		var kafkaCallbackClient ka.KafkaAdapter
		kafkaCallbackClient.Setup(&w.config.Kafka, os.Getenv("CB_TOPIC"))
		var kafkaDevCallbackClient ka.KafkaAdapter
		kafkaDevCallbackClient.Setup(&w.config.Kafka, os.Getenv("CB_DEV_TOPIC"))
		var kafkaCronClient ka.KafkaAdapter
		kafkaCronClient.Setup(&w.config.Kafka, os.Getenv("CRON_TOPIC"))

		ctx := context.Background()
		ctx = context.WithValue(ctx, "cadenceClient", w.cadenceAdapter)
		ctx = context.WithValue(ctx, "gcsClient", gcsClient)
		ctx = context.WithValue(ctx, "kafkaCallbackClient", kafkaCallbackClient)
		ctx = context.WithValue(ctx, "kafkaCronClient", kafkaCronClient)
		ctx = context.WithValue(ctx, "kafkaDevCallbackClient", kafkaDevCallbackClient)

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
	if verbose == "0" {
		workerOptions.Logger = zap.NewNop()
	} else {
		workerOptions.Logger = w.cadenceAdapter.Logger
	}
	w.options = workerOptions
	w.taskList = tasklist
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
