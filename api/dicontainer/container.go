package dicontainer

import (
	"jobprocessor/api/controller"
	"jobprocessor/api/service"
	"jobprocessor/config"
	"jobprocessor/internal/adapter"
)

// ServiceContainer resolve all dependencies between controller, service, infrastructure except application level dependencies such us logging, config and etc ...
type ServiceContainer struct {
	config config.AppConfig

	//controllers
	LeaderboardController  *controller.LeaderboardController
	JobProcessorController *controller.JobProcessorController
	HTTPErrorController    *controller.HTTPErrorController
}

// NewServiceContainer ...
func NewServiceContainer(config config.AppConfig) *ServiceContainer {
	return &ServiceContainer{
		config: config,
	}
}

//InitDependenciesInjection ...
func (container *ServiceContainer) InitDependenciesInjection() {
	//Initializing base controller
	baseController := controller.BaseController{Config: container.config}

	//Initializing Clients
	var cadenceClient adapter.CadenceAdapter
	cadenceClient.Setup(&container.config.Cadence)

	var kafkaClient adapter.KafkaAdapter
	kafkaClient.Setup(&container.config.Kafka)

	//Services
	leaderboardService := &service.LeaderboardService{CadenceAdapter: cadenceClient, Logger: container.config.Logger}
	jobprocessorService := &service.JobProcessorService{CadenceAdapter: cadenceClient, Logger: container.config.Logger}

	//Initializing controllers
	container.LeaderboardController = &controller.LeaderboardController{BaseController: baseController,
		LeaderboardService: leaderboardService}
	container.JobProcessorController = &controller.JobProcessorController{BaseController: baseController,
		JobProcessorService: jobprocessorService}
}
