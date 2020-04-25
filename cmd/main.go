package main

import (
	"flag"
	"os"

	server "github.com/YOVO-LABS/workflow/api"
	worker "github.com/YOVO-LABS/workflow/workflows"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	var service string
	var configFilePath string
	var serverPort string
	var tasklist string
	var logger string
	flag.StringVar(&service, "service", "", "Name of the service to start (app server or workflow workerworker)")
	flag.StringVar(&configFilePath, "config", "./config", "absolute path to the configuration file")
	flag.StringVar(&serverPort, "port", "4000", "port on which server runs")
	flag.StringVar(&tasklist, "tasklist", "", "Name of the tasklist")
	flag.StringVar(&logger, "v", "0", "Logger enable/disable")
	flag.Parse()

	if os.Getenv("SERVICE") != "" {
		service = os.Getenv("SERVICE")
	}

	if os.Getenv("TASKLIST") != "" {
		tasklist = os.Getenv("TASKLIST")
	}

	if os.Getenv("VERBOSE") != "" {
		logger = os.Getenv("VERBOSE")
	}

	if service == "app" || os.Getenv("SERVICE") == "app" {
		application := server.New(configFilePath)
		// init necessary module before start
		application.Init()
		// start http server
		application.Start(serverPort)
	} else if service == "workflow" || os.Getenv("SERVICE") == "workflow" {
		worker := worker.New(configFilePath)
		worker.Init(tasklist)
		worker.Start(logger, service)
	} else if service == "activity" || os.Getenv("SERVICE") == "activity" {
		worker := worker.New(configFilePath)
		worker.Init(tasklist)
		worker.Start(logger, service)
	}
}
