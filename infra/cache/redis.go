package cache

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"
	"strings"
	"time"

	"github.com/go-redis/redis/extra/redisotel/v8"
	"github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	"github.com/minghsu0107/saga-product/config"
)

var (
	RedisClient redis.UniversalClient
	//ErrRedisUnlockFail is redis unlock fail error
	ErrRedisUnlockFail = errors.New("redis unlock fail")
	// ErrRedisCmdNotFound is redis command not found error
	ErrRedisCmdNotFound = errors.New("redis command not found; supports only SET and DELETE")
)

// RedisCache is the interface of redis cache
type RedisCache interface {
	Get(ctx context.Context, key string, dst interface{}) (bool, error)
	Exist(ctx context.Context, key string) (bool, error)
	Set(ctx context.Context, key string, val interface{}) error
	BFReserve(ctx context.Context, key string, errorRate float64, capacity int64) error
	BFInsert(ctx context.Context, key string, errorRate float64, capacity int64, items ...interface{}) error
	BFAdd(ctx context.Context, key string, item interface{}) error
	BFExist(ctx context.Context, key string, item interface{}) (bool, error)
	CFReserve(ctx context.Context, key string, capacity int64, bucketSize, maxIterations int) error
	CFAdd(ctx context.Context, key string, item interface{}) error
	CFExist(ctx context.Context, key string, item interface{}) (bool, error)
	CFDel(ctx context.Context, key string, item interface{}) error
	IncrBy(ctx context.Context, key string, val int64) error
	Delete(ctx context.Context, key string) error
	GetMutex(mutexname string) *redsync.Mutex
	ExecPipeLine(ctx context.Context, cmds *[]RedisCmd) error
	Publish(ctx context.Context, topic string, payload interface{}) error
}

// RedisCacheImpl is the redis cache client type
type RedisCacheImpl struct {
	client     redis.UniversalClient
	rs         *redsync.Redsync
	expiration int64
}

// RedisOpType is the redis operation type
type RedisOpType int

const (
	// SET represents set operation
	SET RedisOpType = iota
	// DELETE represents delete operation
	DELETE
	// INCRBY represents incrBy operation
	INCRBY
)

// RedisPayload is a abstract interface for payload type
type RedisPayload interface {
	Payload()
}

// RedisSetPayload is the payload type of set method
type RedisSetPayload struct {
	RedisPayload
	Key string
	Val interface{}
}

// RedisDeletePayload is the payload type of delete method
type RedisDeletePayload struct {
	RedisPayload
	Key string
}

// RedisIncrByPayload is the payload type of incrBy method
type RedisIncrByPayload struct {
	RedisPayload
	Key string
	Val int64
}

// Payload implements abstract interface
func (RedisSetPayload) Payload() {}

// Payload implements abstract interface
func (RedisDeletePayload) Payload() {}

// Payload implements abstract interface
func (RedisIncrByPayload) Payload() {}

// RedisCmd represents an operation and its payload
type RedisCmd struct {
	OpType  RedisOpType
	Payload RedisPayload
}

// RedisPipelineCmd is redis pipeline command type
type RedisPipelineCmd struct {
	OpType RedisOpType
	Cmd    interface{}
}

func NewRedisClient(config *config.Config) (redis.UniversalClient, error) {
	RedisClient = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:         getServerAddrs(config.RedisConfig.Addrs),
		Password:      config.RedisConfig.Password,
		PoolSize:      config.RedisConfig.PoolSize,
		MaxRetries:    config.RedisConfig.MaxRetries,
		IdleTimeout:   time.Duration(config.RedisConfig.IdleTimeoutSeconds) * time.Second,
		ReadOnly:      true,
		RouteRandomly: true,
	})
	ctx := context.Background()
	pong, err := RedisClient.Ping(ctx).Result()
	if err == redis.Nil || err != nil {
		return nil, err
	}
	RedisClient.AddHook(redisotel.NewTracingHook())
	config.Logger.ContextLogger.WithField("type", "setup:redis").Info("successful redis connection: " + pong)
	return RedisClient, nil
}

// NewRedisCache is the factory of redis cache
func NewRedisCache(config *config.Config, client redis.UniversalClient) RedisCache {
	pool := goredis.NewPool(client)
	rs := redsync.New(pool)

	return &RedisCacheImpl{
		client:     client,
		rs:         rs,
		expiration: config.RedisConfig.ExpirationSeconds,
	}
}

// Get returns true if the key already exists and set dst to the corresponding value
func (rc *RedisCacheImpl) Get(ctx context.Context, key string, dst interface{}) (bool, error) {
	val, err := rc.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	} else if err != nil {
		return false, err
	} else {
		json.Unmarshal([]byte(val), dst)
	}
	return true, nil
}

// Exist checks whether a key exists
func (rc *RedisCacheImpl) Exist(ctx context.Context, key string) (bool, error) {
	numExistKey, err := rc.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	exist := (numExistKey == 1)
	return exist, nil
}

// Set sets a key-value pair
func (rc *RedisCacheImpl) Set(ctx context.Context, key string, val interface{}) error {
	strVal, err := json.Marshal(val)
	if err != nil {
		return err
	}
	if err := rc.client.Set(ctx, key, strVal, getRandomExpiration(rc.expiration)).Err(); err != nil {
		return err
	}
	return nil
}

func (rc *RedisCacheImpl) BFReserve(ctx context.Context, key string, errorRate float64, capacity int64) error {
	if err := rc.client.Do(ctx, "bf.reserve", key, errorRate, capacity).Err(); err != nil {
		return err
	}
	return nil
}

func (rc *RedisCacheImpl) BFInsert(ctx context.Context, key string, errorRate float64, capacity int64, items ...interface{}) error {
	args := []interface{}{"bf.insert", key, "capacity", capacity, "error", errorRate, "items"}
	args = append(args, items...)
	if err := rc.client.Do(ctx, args...).Err(); err != nil {
		return err
	}
	return nil
}

func (rc *RedisCacheImpl) BFAdd(ctx context.Context, key string, item interface{}) error {
	if err := rc.client.Do(ctx, "bf.add", key, item).Err(); err != nil {
		return err
	}
	return nil
}

func (rc *RedisCacheImpl) BFExist(ctx context.Context, key string, item interface{}) (bool, error) {
	res, err := rc.client.Do(ctx, "bf.exists", key, item).Int()
	if err != nil {
		return false, err
	}
	return (res == 1), nil
}

func (rc *RedisCacheImpl) CFReserve(ctx context.Context, key string, capacity int64, bucketSize, maxIterations int) error {
	if err := rc.client.Do(ctx, "cf.reserve", key, capacity, "BUCKETSIZE", bucketSize, "MAXITERATIONS", maxIterations).Err(); err != nil {
		return err
	}
	return nil
}

func (rc *RedisCacheImpl) CFAdd(ctx context.Context, key string, item interface{}) error {
	if err := rc.client.Do(ctx, "cf.add", key, item).Err(); err != nil {
		return err
	}
	return nil
}

func (rc *RedisCacheImpl) CFExist(ctx context.Context, key string, item interface{}) (bool, error) {
	res, err := rc.client.Do(ctx, "cf.exists", key, item).Int()
	if err != nil {
		return false, err
	}
	return (res == 1), nil
}

func (rc *RedisCacheImpl) CFDel(ctx context.Context, key string, item interface{}) error {
	if err := rc.client.Do(ctx, "cf.del", key, item).Err(); err != nil {
		return err
	}
	return nil
}

func (rc *RedisCacheImpl) IncrBy(ctx context.Context, key string, val int64) error {
	return rc.client.IncrBy(ctx, key, val).Err()
}

// Delete deletes a key
func (rc *RedisCacheImpl) Delete(ctx context.Context, key string) error {
	if err := rc.client.Del(ctx, key).Err(); err != nil {
		return err
	}
	return nil
}

func (rc *RedisCacheImpl) GetMutex(mutexname string) *redsync.Mutex {
	return rc.rs.NewMutex(mutexname, redsync.WithExpiry(5*time.Second))
}

// ExecPipeLine execute the given commands in a pipline
func (rc *RedisCacheImpl) ExecPipeLine(ctx context.Context, cmds *[]RedisCmd) error {
	pipe := rc.client.Pipeline()
	var pipelineCmds []RedisPipelineCmd
	for _, cmd := range *cmds {
		switch cmd.OpType {
		case SET:
			strVal, err := json.Marshal(cmd.Payload.(RedisSetPayload).Val)
			if err != nil {
				return err
			}
			pipelineCmds = append(pipelineCmds, RedisPipelineCmd{
				OpType: SET,
				Cmd:    pipe.Set(ctx, cmd.Payload.(RedisSetPayload).Key, strVal, getRandomExpiration(rc.expiration)),
			})
		case DELETE:
			pipelineCmds = append(pipelineCmds, RedisPipelineCmd{
				OpType: DELETE,
				Cmd:    pipe.Del(ctx, cmd.Payload.(RedisDeletePayload).Key),
			})
		case INCRBY:
			payload := cmd.Payload.(RedisIncrByPayload)
			pipelineCmds = append(pipelineCmds, RedisPipelineCmd{
				OpType: INCRBY,
				Cmd:    pipe.IncrBy(ctx, payload.Key, payload.Val),
			})
		default:
			return ErrRedisCmdNotFound
		}
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}

	for _, executedCmd := range pipelineCmds {
		switch executedCmd.OpType {
		case SET:
			if err := executedCmd.Cmd.(*redis.StatusCmd).Err(); err != nil {
				return err
			}
		case DELETE, INCRBY:
			if err := executedCmd.Cmd.(*redis.IntCmd).Err(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (rc *RedisCacheImpl) Publish(ctx context.Context, topic string, payload interface{}) error {
	strVal, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return rc.client.Publish(ctx, topic, strVal).Err()
}

func getRandomExpiration(expiration int64) time.Duration {
	return time.Duration(expiration+rand.Int63n(10)) * time.Second
}

func getServerAddrs(addrs string) []string {
	return strings.Split(addrs, ",")
}
