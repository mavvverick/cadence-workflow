package messaging

type (
	// KafkaConfig describes the configuration needed to connect to all kafka clusters
	KafkaConfig struct {
		Clusters map[string]ClusterConfig `yaml:"clusters"`
		Topics   map[string]TopicConfig   `yaml:"topics"`
	}

	// ClusterConfig describes the configuration for a single Kafka cluster
	ClusterConfig struct {
		Brokers []string `yaml:"brokers"`
	}

	// TopicConfig describes the mapping from topic to Kafka cluster
	TopicConfig struct {
		Cluster string `yaml:"cluster"`
	}
)

// Validate will validate config for kafka
func (k *KafkaConfig) Validate(checkCluster bool, checkTopic bool) {
	if len(k.Clusters) == 0 {
		panic("Empty Kafka Cluster Config")
	}
	if len(k.Topics) == 0 {
		panic("Empty Topics Config")
	}
}

func (k *KafkaConfig) getKafkaClusterForTopic(topic string) string {
	return k.Topics[topic].Cluster
}

func (k *KafkaConfig) getBrokersForKafkaCluster(kafkaCluster string) []string {
	return k.Clusters[kafkaCluster].Brokers
}
