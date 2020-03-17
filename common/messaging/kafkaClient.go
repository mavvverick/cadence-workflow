package messaging

// kafka "github.com/segmentio/kafka-go"

type kafkaClient struct {
	config *KafkaConfig
}

// var _ Client = (*kafkaClient)(nil)

// NewKafkaClient is used to create an instance of KafkaClient
