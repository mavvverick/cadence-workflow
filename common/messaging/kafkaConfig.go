package messaging

import (
	"encoding/json"
	"errors"
	"fmt"
)

type (
	// KafkaConfig describes the configuration needed to connect to all kafka clusters
	KafkaConfig struct {
		Topic 	string
		Brokers string
	}

	kafkaBrokers struct {
		ip []string
	}
)

// Validate will validate config for kafka
func (k *KafkaConfig) Validate() {
	if len(k.Brokers) == 0 {
		fmt.Println(errors.New("Empty Broker"))
	}
	if len(k.Topic) == 0 {
		fmt.Println(errors.New("Empty Topic"))
	}
}

func (k *KafkaConfig) getBrokersForKafkaCluster(brokers []byte) []string {
	var kb kafkaBrokers
	json.Unmarshal(brokers, &kb.ip)
	return kb.ip
}
