package messaging

import (
	"context"
	"fmt"

	kafka "github.com/segmentio/kafka-go"
)

//KafkaProducer ...
type KafkaProducer struct {
	Writer *kafka.Writer
}

//NewProducer creates a producer object
func NewProducer(kc *KafkaConfig, cluster, topic string) *KafkaProducer {
	brokers := kc.getBrokersForKafkaCluster(cluster)
	fmt.Println(brokers)
	writer := kafka.NewWriter(
		kafka.WriterConfig{
			Brokers: brokers,
			Topic:   topic,
		})

	return &KafkaProducer{
		Writer: writer,
	}
}

// Publish pushes message to a kafka topic
func (kp *KafkaProducer) Publish(ctx context.Context, key, msg string) error {
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

// Close flushes all buffered messages and closes the writer.
func (kp *KafkaProducer) Close(ctx context.Context) error {
	err := kp.Writer.Close()
	if err != nil {
		panic(err)
	}
	return nil
}
