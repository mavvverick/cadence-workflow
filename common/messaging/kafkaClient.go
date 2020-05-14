package messaging

//KafkaAdapter ...
type KafkaAdapter struct {
	Producer *KafkaProducer
	Consumer *KafkaConsumer
}

// Setup ...
func (c *KafkaAdapter) Setup(config *KafkaConfig, topic string) {
	producer := NewProducer(config, topic)
	c.Producer = producer

	//consumer := NewConsumer(config)
	//c.Consumer = consumer
	return
}
