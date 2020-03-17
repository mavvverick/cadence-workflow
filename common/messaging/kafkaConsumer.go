package messaging

import (
	"context"

	kafka "github.com/segmentio/kafka-go"
)

//KafkaConsumer ...
type KafkaConsumer struct {
	Reader *kafka.Reader
}

//NewConsumer creates consumer object to read message
func NewConsumer(kc *KafkaConfig, broker, topic string) *KafkaConsumer {
	brokers := kc.getBrokersForKafkaCluster(broker)

	reader := kafka.NewReader(
		kafka.ReaderConfig{
			Brokers:   brokers,
			Topic:     topic,
			Partition: 0,
			MinBytes:  10e3, // 10KB
			MaxBytes:  10e6, //10 MB
		})

	return &KafkaConsumer{
		Reader: reader,
	}
}

//Consume reads messages from a kafka topic
func (kp *KafkaProducer) Consume(ctx context.Context, key, msg string) error {
	err := kp.Writer.WriteMessages(
		ctx,
		kafka.Message{
			Key:   []byte(key),
			Value: []byte(msg),
		},
	)
	if err != nil {
		panic(err)
	}
	return nil
}
