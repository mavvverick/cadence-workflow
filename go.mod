module github.com/YOVO-LABS/workflow

go 1.13

require (
	cloud.google.com/go/storage v1.6.0
	github.com/disintegration/imaging v1.6.2
	github.com/fogleman/gg v1.3.0
	github.com/go-chi/chi v4.0.3+incompatible
	github.com/go-chi/cors v1.0.1
	github.com/go-redis/redis/v7 v7.2.0
	github.com/go-sql-driver/mysql v1.4.1
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0
	github.com/golang/protobuf v1.3.3
	github.com/google/uuid v1.1.1
	github.com/jinzhu/gorm v1.9.12
	github.com/joho/godotenv v1.3.0
	github.com/pborman/uuid v1.2.0
	github.com/segmentio/kafka-go v0.3.5
	github.com/shopspring/decimal v0.0.0-20200227202807-02e2044944cc
	github.com/uber-go/tally v3.3.15+incompatible
	github.com/uber/cadence v0.11.0
	go.uber.org/cadence v0.11.2
	go.uber.org/yarpc v1.44.0
	go.uber.org/zap v1.14.1
	golang.org/x/net v0.0.0-20200222125558-5a598a2470a0
	google.golang.org/api v0.18.0
	google.golang.org/grpc v1.27.1
	gopkg.in/validator.v2 v2.0.0-20191107172027-c3144fdedc21
)

replace github.com/apache/thrift => github.com/apache/thrift v0.0.0-20190309152529-a9b748bb0e02
