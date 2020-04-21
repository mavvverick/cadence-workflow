package messaging

import (
	"context"

	kafka "github.com/segmentio/kafka-go"
)

//KafkaConsumer ...
type KafkaConsumer struct {
	Reader *kafka.Reader
}

type KafkaMessage struct {
	Key   []byte
	Value []byte
}

//NewConsumer creates consumer object to read message
func NewConsumer(kc *KafkaConfig) *KafkaConsumer {
	kc.Validate()
	brokers := kc.getBrokers(kc.Brokers)

	reader := kafka.NewReader(
		kafka.ReaderConfig{
			Brokers:   brokers,
			Topic:     kc.Topic,
			Partition: 0,
			MinBytes:  10e3, // 10KB
			MaxBytes:  10e6, //10 MB
		})

	return &KafkaConsumer{
		Reader: reader,
	}
}

//Consume reads messages from a kafka topic
func (kp *KafkaConsumer) Consume(ctx context.Context) (*KafkaMessage, error) {
	msg, err := kp.Reader.ReadMessage(ctx)

	kafkaMessage := &KafkaMessage{
		Key:   msg.Key,
		Value: msg.Value,
	}

	if err != nil {
		return nil, err
	}
	return kafkaMessage, err
}
