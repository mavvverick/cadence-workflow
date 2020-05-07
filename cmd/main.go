package main

import (
	"flag"
	jp "github.com/YOVO-LABS/workflow/workflows/jobprocessor"
	"os"
	"strings"

	server "github.com/YOVO-LABS/workflow/api"
	"github.com/YOVO-LABS/workflow/workflows/ai"
	_ "github.com/joho/godotenv/autoload"
)

const (
	workflow = "workflow"
	activity = "activity"
)

func main() {
	var service string
	var configFilePath string
	var serverPort string
	var tasklist string
	var logger string
	flag.StringVar(&service, "service", "", "Name of the service to start")
	flag.StringVar(&configFilePath, "config", "./config", "path to the configuration file")
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
		application.Init()
		application.Start(serverPort)
	} else {
		if strings.ToLower(tasklist) == strings.ToLower(jp.TaskList) {
			if service == workflow || os.Getenv("SERVICE") == workflow {
				worker := jp.New(configFilePath)
				worker.Init(tasklist, logger, service)
				worker.Start()
			} else if service == activity || os.Getenv("SERVICE") == activity {
				worker := jp.New(configFilePath)
				worker.Init(tasklist, logger, service)
				worker.Start()
			}
		}  else if strings.ToLower(tasklist) == strings.ToLower(ai.Tasklist) {
			if service == workflow || os.Getenv("SERVICE") == workflow {
				worker := ai.New(configFilePath)
				worker.Init(tasklist, logger, service)
				worker.Start()
			} else if service == activity || os.Getenv("SERVICE") == activity {
				worker := ai.New(configFilePath)
				worker.Init(tasklist, logger, service)
				worker.Start()
			}
		}
	}
}
