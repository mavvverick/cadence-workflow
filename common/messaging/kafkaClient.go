package messaging

//KafkaAdapter ...
type KafkaAdapter struct {
	Producer *KafkaProducer
	Consumer *KafkaConsumer
}

// Setup ...
func (c *KafkaAdapter) Setup(config *KafkaConfig) {
	producer := NewProducer(config)
	c.Producer = producer

	consumer := NewConsumer(config)
	c.Consumer = consumer
	return
}
