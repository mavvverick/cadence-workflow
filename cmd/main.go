package main

import (
	"flag"
	"os"

	server "github.com/YOVO-LABS/workflow/cmd/server"
	worker "github.com/YOVO-LABS/workflow/cmd/worker"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	var service string
	var configFilePath string
	var serverPort string
	var tasklist string
	var logger string
	var workerType string
	flag.StringVar(&service, "service", "", "Name of the service to start (app server or workflow workerworker)")
	flag.StringVar(&configFilePath, "config", "./config", "absolute path to the configuration file")
	flag.StringVar(&serverPort, "port", "4000", "port on which server runs")
	flag.StringVar(&tasklist, "tasklist", "", "Name of the tasklist")
	flag.StringVar(&logger, "v", "0", "Logger enable/disable")
	flag.StringVar(&workerType, "type", "", "Type of worker {activity/workflow}")
	flag.Parse()

	if os.Getenv("TASKLIST") != "" {
		tasklist = os.Getenv("TASKLIST")
	}

	if os.Getenv("VERBOSE") != "" {
		logger = os.Getenv("VERBOSE")
	}

	if os.Getenv("WORKER_TYPE") != "" {
		workerType = os.Getenv("WORKER_TYPE")
	}

	if service == "app" || os.Getenv("SERVICE") == "app" {
		application := server.New(configFilePath)
		// init necessary module before start
		application.Init()

		// bootstrap workflow worker
		worker := worker.New(configFilePath)
		worker.Init(tasklist)
		worker.Start(logger, workerType)

		// start http server
		application.Start(serverPort)
	} else if service == "worker" || os.Getenv("SERVICE") == "worker" {

		worker := worker.New(configFilePath)
		worker.Init(tasklist)
		worker.Start(logger, workerType)
		// The workers are supposed to be long running process that should not exit.
	}
}
