package broker

import (
	"context"
	"fmt"
	"strings"

	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/components/metrics"
	"github.com/ThreeDotsLabs/watermill/message"
	conf "github.com/minghsu0107/saga-product/config"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

type RedisPublisher message.Publisher

var (
	ResultPublisher RedisPublisher
	RedisClient     redis.UniversalClient
)

// NewRedisPublisher returns a redis publisher for event streaming
func NewRedisPublisher(config *conf.Config) (RedisPublisher, error) {
	var err error
	ctx := context.Background()
	RedisClient = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:         getServerAddrs(config.RedisConfig.Addrs),
		Password:      config.RedisConfig.Password,
		PoolSize:      config.RedisConfig.PoolSize,
		MaxRetries:    config.RedisConfig.MaxRetries,
		ReadOnly:      true,
		RouteRandomly: true,
	})
	pong, err := RedisClient.Ping(ctx).Result()
	if err == redis.Nil || err != nil {
		return nil, err
	}
	redisotel.InstrumentTracing(RedisClient)
	config.Logger.ContextLogger.WithField("type", "setup:redis").Info("successful redis connection: " + pong)

	publisherConfig := redisstream.PublisherConfig{
		Client:     RedisClient,
		Marshaller: &redisstream.DefaultMarshallerUnmarshaller{},
		Maxlens: map[string]int64{
			conf.PurchaseResultTopic: config.RedisConfig.Publisher.PurchaseResultTopicMaxlen,
		},
	}
	ResultPublisher, err = redisstream.NewPublisher(publisherConfig, logger)
	if err != nil {
		return nil, err
	}

	registry, ok := prom.DefaultRegisterer.(*prom.Registry)
	if !ok {
		return nil, fmt.Errorf("prometheus type casting error")
	}
	metricsBuilder := metrics.NewPrometheusMetricsBuilder(registry, config.App, "pubsub")
	ResultPublisher, err = metricsBuilder.DecoratePublisher(ResultPublisher)
	if err != nil {
		return nil, err
	}
	return ResultPublisher, nil
}

func getServerAddrs(addrs string) []string {
	return strings.Split(addrs, ",")
}
