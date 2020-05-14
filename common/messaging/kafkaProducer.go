package messaging

import (
	"crypto/tls"
	"os"
	"time"

	// "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
	"golang.org/x/net/context"
)

//KafkaProducer ...
type KafkaProducer struct {
	// Writer *kafka.Writer
	Writer *kafka.Writer
}

//NewProducer creates a producer object
func NewProducer(kc *KafkaConfig, topic string) *KafkaProducer {
	kc.Validate()

	brokers := kc.getBrokers(kc.Brokers)

	mechanism, _ := scram.Mechanism(scram.SHA256, os.Getenv("KAFKA_USERNAME"), os.Getenv("KAFKA_PASS"))
	dialer := &kafka.Dialer{
		Timeout:       10 * time.Second,
		SASLMechanism: mechanism,
		DualStack:     true,
		TLS: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	writer := kafka.NewWriter(
		kafka.WriterConfig{
			Brokers:  brokers,
			Dialer:   dialer,
			Balancer: &kafka.LeastBytes{},
			Async:    true,
			Topic:    topic,
		})
	// config := &kafka.ConfigMap{
	// 	"metadata.broker.list":  os.Getenv("KAFKA_BROKERS"),
	// 	"broker.address.family": "v4",
	// 	"security.protocol":     "SASL_SSL",
	// 	"sasl.mechanisms":       "SCRAM-SHA-256",
	// 	"sasl.username":         os.Getenv("KAFKA_USERNAME"),
	// 	"sasl.password":         os.Getenv("KAFKA_PASS"),
	// 	"group.id":              os.Getenv("GROUPID"),
	// 	//"default.topic.config": kafka.ConfigMap{"auto.offset.reset": "earliest"},
	// 	//"debug":            		"generic,broker,security",
	// }
	// producer, err := kafka.NewProducer(config)
	// if err != nil {
	// 	fmt.Printf("Failed to create producer: %s\n", err)
	// }

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
	// err := kp.Writer.Produce(
	// 	&kafka.Message{
	// 		TopicPartition: kafka.TopicPartition{
	// 			Topic:     &topic,
	// 			Partition: kafka.PartitionAny},
	// 		Value: []byte(msg),
	// 		Key:   []byte(key)}, nil)
	// if err != nil {
	// 	return err
	// }

	// e := <-kp.Writer.Events()
	// switch ev := e.(type) {
	// case *kafka.Message:
	// 	m := e.(*kafka.Message)
	// 	if m.TopicPartition.Error != nil {
	// 		return m.TopicPartition.Error
	// 	}
	// case kafka.Error:
	// 	fmt.Printf("kafka error: %v", ev)
	// }

	// kp.Writer.Flush(1000)
	return nil
}

// Close flushes all buffered messages and closes the writer.
func (kp *KafkaProducer) Close() error {
	// kp.Writer.Close()
	//close(kp.Writer.Events())
	err := kp.Writer.Close()
	if err != nil {
		panic(err)
	}
	return nil
}
