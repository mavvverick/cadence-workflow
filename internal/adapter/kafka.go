package adapter

import (
	"github.com/YOVO-LABS/workflow/common/messaging"
)

//KafkaAdapter ...
type KafkaAdapter struct {
	Producer *messaging.KafkaProducer
	Consumer *messaging.KafkaConsumer
}

// Setup ...
func (c *KafkaAdapter) Setup(config *messaging.KafkaConfig) {
	producer := messaging.NewProducer(config)
	c.Producer = producer

	consumer := messaging.NewConsumer(config)
	c.Consumer = consumer
	return
}
