package cache

import (
	"context"
	"encoding/json"
	"runtime"
	"time"

	retry "github.com/avast/retry-go"
	"github.com/go-redis/redis/v8"
	"github.com/minghsu0107/saga-product/config"
	"github.com/minghsu0107/saga-product/pkg/workerpool"
)

// LocalCacheCleaner subscribes to InvalidationTopic and invalidates entries
type LocalCacheCleaner interface {
	SubscribeInvalidationEvent() error
	Close()
}

// LocalCacheCleanerImpl implements CacheCleaner interface
type LocalCacheCleanerImpl struct {
	client *redis.ClusterClient
	lc     LocalCache
	pool   *workerpool.Pool
}

func NewLocalCacheCleaner(client *redis.ClusterClient, lc LocalCache) LocalCacheCleaner {
	return &LocalCacheCleanerImpl{
		client: client,
		lc:     lc,
	}
}

// SubscribeInvalidationEvent starts a consumer that subscribe to cache invalidation event from redis
func (l *LocalCacheCleanerImpl) SubscribeInvalidationEvent() error {
	ctx := context.Background()
	pubsub := l.client.Subscribe(ctx, config.InvalidationTopic)
	_, err := pubsub.Receive(ctx)
	if err != nil {
		return err
	}
	channel := pubsub.Channel()
	l.pool = workerpool.NewPool(ctx, workerpool.Option{NumberWorker: runtime.NumCPU()})
	l.pool.Start()

	for msg := range channel {
		var keys []string
		err := json.Unmarshal([]byte(msg.Payload), &keys)
		if err != nil {
			continue
		}
		for _, key := range keys {
			task := l.cleanTask(context.Background(), key)
			l.pool.Do(task)
		}
	}
	return nil
}

func (l *LocalCacheCleanerImpl) Close() {
	l.pool.Stop()
}

// cleanTask closure
func (l *LocalCacheCleanerImpl) cleanTask(ctx context.Context, key string) *workerpool.Task {
	return workerpool.NewTask(ctx, func(ctx context.Context) (interface{}, error) {
		return nil, retry.Do(
			func() error {
				return l.lc.Delete(key)
			},
			retry.Attempts(3),
			retry.DelayType(retry.RandomDelay),
			retry.MaxJitter(10*time.Millisecond),
		)
	})
}
