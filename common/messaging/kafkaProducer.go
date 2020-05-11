package messaging

import (
	"fmt"
	"os"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	// "github.com/segmentio/kafka-go"
)

//KafkaProducer ...
type KafkaProducer struct {
	// Writer *kafka.Writer
	Writer *kafka.Producer
}

//NewProducer creates a producer object
func NewProducer(kc *KafkaConfig) *KafkaProducer {
	//kc.Validate()

	// brokers := kc.getBrokers(kc.Brokers)

	// mechanism, _ := scram.Mechanism(scram.SHA256, os.Getenv("KAFKA_USERNAME"), os.Getenv("KAFKA_PASS"))
	// dialer := &kafka.Dialer{
	// 	Timeout:       10 * time.Second,
	// 	SASLMechanism: mechanism,
	// 	DualStack:     true,
	// 	TLS: &tls.Config{
	// 		InsecureSkipVerify: true,
	// 	},
	// }

	// writer := kafka.NewWriter(
	// 	kafka.WriterConfig{
	// 		Brokers:  brokers,
	// 		Dialer:   dialer,
	// 		Balancer: &kafka.LeastBytes{},
	// 		Async:    true,
	// 		Topic:    kc.Topic,
	// 	})
	config := &kafka.ConfigMap{
		"metadata.broker.list": os.Getenv("KAFKA_BROKERS"),
		"broker.address.family" : "v4",
		"security.protocol":    "SASL_SSL",
		"sasl.mechanisms":      "SCRAM-SHA-256",
		"sasl.username":        os.Getenv("KAFKA_USERNAME"),
		"sasl.password":        os.Getenv("KAFKA_PASS"),

		//"group.id":             os.Getenv("GROUPID"),
		//"default.topic.config": kafka.ConfigMap{"auto.offset.reset": "earliest"},
		//"debug":            		"generic,broker,security",
	}
	producer, err := kafka.NewProducer(config)
	if err != nil {
		fmt.Printf("Failed to create producer: %s\n", err)
	}

	return &KafkaProducer{
		Writer: producer,
	}
}

// Publish pushes message to a kafka topic
func (kp *KafkaProducer) Publish(topic, key, msg string) error {
	// err := kp.Writer.WriteMessages(
	// 	ctx,
	// 	kafka.Message{
	// 		Key:   []byte(key),
	// 		Value: []byte(msg),
	// 	},
	// )
	err := kp.Writer.Produce(
		&kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &topic,
				Partition: kafka.PartitionAny},
			Value: []byte(msg),
			Key:   []byte(key)}, nil)
	if err != nil {
		return err
	}

	e := <-kp.Writer.Events()
	m := e.(*kafka.Message)

	if m.TopicPartition.Error != nil {
		return m.TopicPartition.Error
	}
	kp.Writer.Flush(1000)
	return nil
}

// Close flushes all buffered messages and closes the writer.
func (kp *KafkaProducer) Close() error {
	kp.Writer.Close()
	//close(kp.Writer.Events())
	// err := kp.Writer.Close()
	// if err != nil {
	// 	panic(err)
	// }
	return nil
}
