package server

import (
	"jobprocessor/api/dicontainer"
	"jobprocessor/api/router"
	config "jobprocessor/config"

	"net/http"
)

//AppInterface ...
type AppInterface interface {
	Init()

	Start(serverPort string) error
}

//Application ...
type Application struct {
	config config.AppConfig

	serviceContainer *dicontainer.ServiceContainer
	router           router.RoutingInterface
}

//New ...
func New(configPath string) AppInterface {
	var appConfig config.AppConfig
	appConfig.LoadConfig(configPath)

	return &Application{
		config: appConfig,
	}
}

// Init ...
func (app *Application) Init() {
	//start dependency injection
	app.serviceContainer = dicontainer.NewServiceContainer(app.config)

	app.serviceContainer.InitDependenciesInjection()

	//initialize new handlers
	app.router = router.NewRouter(app.config)
	app.router.Routes(app.serviceContainer)
}

// Start serve http server
func (app *Application) Start(serverPort string) error {
	return http.ListenAndServe(":"+serverPort, app.router.RouteMultiplexer())
}
