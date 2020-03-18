package main

import (
	"flag"

	app "github.com/YOVO-LABS/workflow/cmd/server"
)

func main() {
	var configFilePath string
	var serverPort string
	flag.StringVar(&configFilePath, "config", "./config", "absolute path to the configuration file")
	flag.StringVar(&serverPort, "server_port", "4000", "port on which server runs")
	flag.Parse()

	application := app.New(configFilePath)

	// init necessary module before start
	application.Init()

	// start http server
	application.Start(serverPort)
}
