package config

import (
	"os"

	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// Config is a type for general configuration
type Config struct {
	HTTPPort         string            `yaml:"httpPort" envconfig:"HTTP_PORT"`
	GRPCPort         string            `yaml:"grpcPort" envconfig:"GRPC_PORT"`
	PromPort         string            `yaml:"promPort" envconfig:"PROM_PORT"`
	OcAgentHost      string            `yaml:"ocAgentHost" envconfig:"OC_AGENT_HOST"`
	AppName          string            `yaml:"appName" envconfig:"APP_NAME"`
	GinMode          string            `yaml:"ginMode" envconfig:"GIN_MODE"`
	DBConfig         *DBConfig         `yaml:"dbConfig"`
	LocalCacheConfig *LocalCacheConfig `yaml:"localCacheConfig"`
	RedisConfig      *RedisConfig      `yaml:"redisConfig"`
	Resolver         string            `yaml:"resolver" envconfig:"RESOLVER"`
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
	Addrs              string `yaml:"addrs" envconfig:"REDIS_ADDRS"`
	Password           string `yaml:"password" envconfig:"REDIS_PASSWORD"`
	DB                 int    `yaml:"db" envconfig:"REDIS_DB"`
	PoolSize           int    `yaml:"poolSize" envconfig:"REDIS_POOL_SIZE"`
	MaxRetries         int    `yaml:"maxRetries" envconfig:"REDIS_MAX_RETRIES"`
	ExpirationSeconds  int64  `yaml:"expirationSeconds" envconfig:"REDIS_EXPIRATION_SECONDS"`
	IdleTimeoutSeconds int64  `yaml:"idleTimeoutSeconds" envconfig:"REDIS_IDLE_TIMEOUT_SECONDS"`
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
	config.Logger = newLogger(config.AppName, config.GinMode)
	log.SetOutput(config.Logger.Writer)

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
