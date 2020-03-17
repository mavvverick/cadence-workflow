module jobprocessor

go 1.13

require (
	cloud.google.com/go/storage v1.6.0
	github.com/Shopify/sarama v1.23.0
	github.com/confluentinc/confluent-kafka-go v1.3.0
	github.com/go-chi/chi v4.0.3+incompatible
	github.com/go-chi/cors v1.0.1
	github.com/go-redis/redis v6.15.7+incompatible
	github.com/go-redis/redis/v7 v7.2.0
	github.com/gogo/status v1.1.0
	github.com/google/uuid v1.1.1
	github.com/jinzhu/gorm v1.9.12
	github.com/pborman/uuid v1.2.0
	github.com/segmentio/kafka-go v0.3.5
	github.com/shopspring/decimal v0.0.0-20200227202807-02e2044944cc
	github.com/spf13/viper v1.6.2
	github.com/uber-go/kafka-client v0.2.3-0.20191018205945-8b3555b395f9
	github.com/uber-go/tally v3.3.15+incompatible
	github.com/uber/cadence v0.11.0
	github.com/uber/ringpop-go v0.8.5
	go.uber.org/cadence v0.11.0
	go.uber.org/yarpc v1.44.0
	go.uber.org/zap v1.14.0
	google.golang.org/grpc v1.27.1
	gopkg.in/validator.v2 v2.0.0-20180514200540-135c24b11c19
)

replace github.com/apache/thrift => github.com/apache/thrift v0.0.0-20190309152529-a9b748bb0e02
