package config

import (
	"os"

	messaging "github.com/YOVO-LABS/workflow/common/messaging"

	"go.uber.org/zap"
)

// CadenceConfig ...
type CadenceConfig struct {
	Domain   string
	Service  string
	HostPort string
}

//AppConfig ...
type AppConfig struct {
	Env            string
	WorkerTaskList string
	Cadence        CadenceConfig
	Kafka          messaging.KafkaConfig
	Logger         *zap.Logger
}

// LoadConfig setup the config for the code run
func (h *AppConfig) LoadConfig(configPath string) {
	h.Cadence.Domain = os.Getenv("CADENCE_DOMAIN")
	h.Cadence.HostPort = os.Getenv("CADENCE_HOST")
	h.Cadence.Service = os.Getenv("CADENCE_SERVICE")

	h.Kafka.Brokers = os.Getenv("KAFKA_BROKERS")
	h.Kafka.Topic = os.Getenv("KAFKA_TOPIC_TEST")

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	h.Logger = logger
	logger.Debug("Finished loading Configuration!")
}
