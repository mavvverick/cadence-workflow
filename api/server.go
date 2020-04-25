package api

import (
	"github.com/YOVO-LABS/workflow/api/dicontainer"
	"github.com/YOVO-LABS/workflow/api/router"
	"github.com/YOVO-LABS/workflow/config"
	"net/http"
)

//Application ...
type Application struct {
	config config.AppConfig
	serviceContainer *dicontainer.ServiceContainer
	router           router.RoutingInterface
}

//New ...
func New(configPath string) *Application {
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
