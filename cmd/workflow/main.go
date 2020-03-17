package main

import (
	"fmt"
	config "jobprocessor/config"
	ca "jobprocessor/internal/adapter"
	lb "jobprocessor/workflow/leaderboard"

	"go.uber.org/cadence/worker"
	"go.uber.org/zap"
)

func startWorkers(h *ca.CadenceAdapter, taskList string) {
	// Configure worker options.
	workerOptions := worker.Options{
		MetricsScope:          h.Scope,
		Logger:                h.Logger,
		EnableLoggingInReplay: true,
		EnableSessionWorker:   true,
	}

	cadenceWorker := worker.New(h.ServiceClient, h.Config.Domain, taskList, workerOptions)
	err := cadenceWorker.Start()
	if err != nil {
		h.Logger.Error("Failed to start workers.", zap.Error(err))
		panic("Failed to start workers")
	}
}

func main() {
	fmt.Println("Starting Worker..")
	var appConfig config.AppConfig
	appConfig.LoadConfig("./config")
	var cadenceClient ca.CadenceAdapter
	cadenceClient.Setup(&appConfig.Cadence)

	startWorkers(&cadenceClient, lb.TaskList)
	// The workers are supposed to be long running process that should not exit.
	select {}
}
