package config

import (
	"os"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// Config is a type for general configuration
type Config struct {
	App              string            `yaml:"app" envconfig:"APP"`
	GinMode          string            `yaml:"ginMode" envconfig:"GIN_MODE"`
	HTTPPort         string            `yaml:"httpPort" envconfig:"HTTP_PORT"`
	GRPCPort         string            `yaml:"grpcPort" envconfig:"GRPC_PORT"`
	PromPort         string            `yaml:"promPort" envconfig:"PROM_PORT"`
	OcAgentHost      string            `yaml:"ocAgentHost" envconfig:"OC_AGENT_HOST"`
	DBConfig         *DBConfig         `yaml:"dbConfig"`
	LocalCacheConfig *LocalCacheConfig `yaml:"localCacheConfig"`
	RedisConfig      *RedisConfig      `yaml:"redisConfig"`
	NATSConfig       *NATSConfig       `yaml:"natsConfig"`
	RPCEndpoints     *RPCEndpoints     `yaml:"rpcEndpoints"`
	ServiceOptions   *ServiceOptions   `yaml:"serviceOptions"`
	Logger           *Logger
}

// DBConfig is database config type
type DBConfig struct {
	Dsn          string `yaml:"dsn" envconfig:"DB_DSN"`
	MaxIdleConns int    `yaml:"maxIdleConns" envconfig:"DB_MAX_IDLE_CONNS"`
	MaxOpenConns int    `yaml:"maxOpenConns" envconfig:"DB_MAX_OPEN_CONNS"`
}

// LocalCacheConfig defines cache related settings
type LocalCacheConfig struct {
	ExpirationSeconds int64 `yaml:"expirationSeconds" envconfig:"LOCAL_CACHE_EXPIRATION_SECONDS"`
}

// RedisConfig is redis config type
type RedisConfig struct {
	Addrs              string      `yaml:"addrs" envconfig:"REDIS_ADDRS"`
	Password           string      `yaml:"password" envconfig:"REDIS_PASSWORD"`
	DB                 int         `yaml:"db" envconfig:"REDIS_DB"`
	PoolSize           int         `yaml:"poolSize" envconfig:"REDIS_POOL_SIZE"`
	MaxRetries         int         `yaml:"maxRetries" envconfig:"REDIS_MAX_RETRIES"`
	ExpirationSeconds  int64       `yaml:"expirationSeconds" envconfig:"REDIS_EXPIRATION_SECONDS"`
	IdleTimeoutSeconds int64       `yaml:"idleTimeoutSeconds" envconfig:"REDIS_IDLE_TIMEOUT_SECONDS"`
	Bloom              *RedisBloom `yaml:"bloom"`
}

// RedisBloom filter config
type RedisBloom struct {
	Activate  bool    `yaml:"activate" envconfig:"REDIS_BLOOM_ACTIVATE"`
	ErrorRate float64 `yaml:"errorRate" envconfig:"REDIS_BLOOM_ERROR_RATE"`
	Capacity  int64   `yaml:"capacity" envconfig:"REDIS_BLOOM_CAPACITY"`
}

// NATSConfig wraps NATS client configurations
type NATSConfig struct {
	ClusterID  string          `yaml:"clusterID" envconfig:"NATS_CLUSTER_ID"`
	URL        string          `yaml:"url" envconfig:"NATS_URL"`
	ClientID   string          `yaml:"clientID" envconfig:"NATS_CLIENT_ID"`
	Subscriber *NATSSubscriber `yaml:"subscriber"`
}

type NATSSubscriber struct {
	QueueGroup  string `yaml:"queueGroup" envconfig:"NATS_SUBSCRIBER_QUEUE_GROUP"`
	DurableName string `yaml:"durableName" envconfig:"NATS_SUBSCRIBER_DURABLE_NAME"`
	Count       int    `yaml:"count" envconfig:"NATS_SUBSCRIBER_COUNT"`
}

// RPCEndpoints wraps all rpc server urls
type RPCEndpoints struct {
	AuthSvcHost    string `yaml:"authSvcHost" envconfig:"RPC_AUTH_SVC_HOST"`
	ProductSvcHost string `yaml:"productSvcHost" envconfig:"RPC_PRODUCT_SVC_HOST"`
}

// ServiceOptions defines options for rpc calls between services
type ServiceOptions struct {
	Rps           int `yaml:"rps" envconfig:"SERVICE_OPTIONS_RPS"`
	TimeoutSecond int `yaml:"timeoutSecond" envconfig:"SERVICE_OPTIONS_TIMEOUT_SECOND"`
	Timeout       time.Duration
}

// NewConfig is the factory of Config instance
func NewConfig() (*Config, error) {
	var config Config
	if err := readFile(&config); err != nil {
		return nil, err
	}
	if err := readEnv(&config); err != nil {
		return nil, err
	}
	config.Logger = newLogger(config.App, config.GinMode)
	log.SetOutput(config.Logger.Writer)
	if config.NATSConfig.ClientID == "" {
		config.NATSConfig.ClientID = watermill.NewShortUUID()
	}

	return &config, nil
}

func readFile(config *Config) error {
	f, err := os.Open("config.yml")
	if err != nil {
		return err
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(config)
	if err != nil {
		return err
	}
	return nil
}

func readEnv(config *Config) error {
	err := envconfig.Process("", config)
	if err != nil {
		return err
	}
	return nil
}
