package adapter

import (
	"github.com/YOVO-LABS/workflow/common/messaging"
)

//KafkaAdapter ...
type KafkaAdapter struct {
	Producer *messaging.KafkaProducer
	Consumer *messaging.KafkaConsumer
	// tlsConfig     *tls.Config
	// client        uberKafkaClient.Client
	// metricsClient metrics.Client
	// logger log.Logger
}

// Setup ...
func (c *KafkaAdapter) Setup(config *messaging.KafkaConfig) {
	producer := messaging.NewProducer(config, "dev", "wallet-dev")
	c.Producer = producer

	consumer := messaging.NewConsumer(config, "dev", "wallet-dev")
	c.Consumer = consumer
}
