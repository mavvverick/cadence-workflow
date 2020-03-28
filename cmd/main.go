package main

import (
	"flag"

	server "github.com/YOVO-LABS/workflow/cmd/server"
	worker "github.com/YOVO-LABS/workflow/cmd/workflow"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	var service string
	var configFilePath string
	var serverPort string
	var tasklist string
	flag.StringVar(&service, "service", "", "Name of the service to start (app server or workflow workerworker)")
	flag.StringVar(&configFilePath, "config", "./config", "absolute path to the configuration file")
	flag.StringVar(&serverPort, "port", "4000", "port on which server runs")
	flag.StringVar(&tasklist, "tasklist", "", "Name of the tasklist")
	flag.Parse()

	if service == "app" {
		application := server.New(configFilePath)
		// init necessary module before start
		application.Init()
		// start http server
		application.Start(serverPort)

	} else if service == "worker" {
		worker := worker.New(configFilePath)
		worker.Init(tasklist)
		worker.Start()

		// The workers are supposed to be long running process that should not exit.
		select {}
	}

}
