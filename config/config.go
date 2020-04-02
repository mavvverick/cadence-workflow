package config

import (
	"encoding/json"
	"fmt"
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

	kafkaConfig := []byte(os.Getenv("KAFKA_CONFIG"))
	err := json.Unmarshal(kafkaConfig, &h)
	if err != nil {
		fmt.Printf("Unable to decode into struct, %v", err)
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	h.Logger = logger
	logger.Debug("Finished loading Configuration!")
}
