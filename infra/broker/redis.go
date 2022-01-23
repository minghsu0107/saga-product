package broker

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ThreeDotsLabs/watermill/components/metrics"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-redis/redis/v8"
	conf "github.com/minghsu0107/saga-product/config"
	redistream "github.com/minghsu0107/watermill-redistream/pkg/redis"
	prom "github.com/prometheus/client_golang/prometheus"
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
		IdleTimeout:   time.Duration(config.RedisConfig.IdleTimeoutSeconds) * time.Second,
		ReadOnly:      true,
		RouteRandomly: true,
	})
	pong, err := RedisClient.Ping(ctx).Result()
	if err == redis.Nil || err != nil {
		return nil, err
	}
	config.Logger.ContextLogger.WithField("type", "setup:redis").Info("successful redis connection: " + pong)

	ResultPublisher, err = redistream.NewPublisher(ctx, RedisClient, &redistream.DefaultMarshaler{}, logger)
	if err != nil {
		return nil, err
	}

	registry, ok := prom.DefaultGatherer.(*prom.Registry)
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
