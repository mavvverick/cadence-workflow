package dicontainer

import (
	"github.com/YOVO-LABS/workflow/api/controller"
	ca "github.com/YOVO-LABS/workflow/common/cadence"
	"github.com/YOVO-LABS/workflow/config"
	//ka "github.com/YOVO-LABS/workflow/common/messaging"
	"github.com/YOVO-LABS/workflow/internal/service"
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
	var cadenceClient ca.CadenceAdapter
	cadenceClient.Setup(&container.config.Cadence)

	//var kafkaClient ka.KafkaAdapter
	//kafkaClient.Setup(&container.config.Kafka)

	//Services
	leaderboardService := &service.LeaderboardService{CadenceAdapter: cadenceClient, Logger: container.config.Logger}
	jobprocessorService := &service.JobProcessorService{
		CadenceAdapter: cadenceClient,
		//KafkaAdapter: kafkaClient,
		Logger: container.config.Logger,
	}

	//Initializing controllers
	container.HTTPErrorController = &controller.HTTPErrorController{BaseController: baseController}
	container.LeaderboardController = &controller.LeaderboardController{BaseController: baseController,
		LeaderboardService: leaderboardService}
	container.JobProcessorController = &controller.JobProcessorController{BaseController: baseController,
		JobProcessorService: jobprocessorService}

}
