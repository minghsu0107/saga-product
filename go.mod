module github.com/minghsu0107/saga-product

go 1.15

require (
	github.com/allegro/bigcache/v3 v3.0.0
	github.com/gin-gonic/gin v1.7.1
	github.com/go-redis/redis/v8 v8.8.2
	github.com/go-redsync/redsync/v4 v4.3.0
	github.com/golang/protobuf v1.5.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/minghsu0107/saga-account v0.0.0-20210514033625-fcd69d4f7369
	github.com/minghsu0107/saga-pb v0.4.0
	github.com/sirupsen/logrus v1.8.1
	github.com/slok/go-http-metrics v0.9.0
	github.com/sony/sonyflake v1.0.0
	go.opencensus.io v0.23.0
	google.golang.org/grpc v1.37.1
	gopkg.in/yaml.v2 v2.4.0
	gorm.io/driver/mysql v1.0.5
	gorm.io/gorm v1.21.8
)
