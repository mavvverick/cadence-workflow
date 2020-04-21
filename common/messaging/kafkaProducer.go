package messaging

import (
	"context"
	"crypto/tls"
	"github.com/segmentio/kafka-go/sasl/scram"
	"time"

	"os"

	"github.com/segmentio/kafka-go"
)

//KafkaProducer ...
type KafkaProducer struct {
	Writer *kafka.Writer
}

//NewProducer creates a producer object
func NewProducer(kc *KafkaConfig) *KafkaProducer {
	kc.Validate()

	brokers := kc.getBrokers(kc.Brokers)

	mechanism, _ := scram.Mechanism(scram.SHA256, os.Getenv("KAFKA_USERNAME"), os.Getenv("KAFKA_PASS"))
	dialer := &kafka.Dialer {
		Timeout:   10 * time.Second,
		SASLMechanism: mechanism,
		DualStack: true,
		TLS: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	writer := kafka.NewWriter(
		kafka.WriterConfig{
			Brokers: brokers,
			Dialer:   dialer,
			Balancer: &kafka.LeastBytes{},
			Async:    true,
			Topic:   kc.Topic,

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
		return err
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
