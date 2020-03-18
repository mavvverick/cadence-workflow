package config

import (
	"fmt"

	messaging "github.com/YOVO-LABS/workflow/common/messaging"

	"github.com/spf13/viper"
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
	viper.SetConfigName("application")
	viper.AddConfigPath(configPath)
	viper.AutomaticEnv()
	// viper.SetConfigType("yml")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file, %s", err)
	}

	err := viper.Unmarshal(&h)
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
