package redis

import (
	"fmt"
	"strconv"

	"github.com/Kunal726/market-mosaic-common-lib-go/pkg/zookeeper"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Config represents Redis configuration
type Config struct {
	Host string
	Port int
}

// NewConfig creates a new Redis configuration from ZooKeeper
func NewConfig(zkClient *zookeeper.Client, logger *zap.Logger) (*Config, error) {
	redisConfig, err := zkClient.GetConfigValueByKey("REDIS_CONFIG", true)
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis config from ZooKeeper: %w", err)
	}

	configMap, ok := redisConfig.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid Redis config format")
	}

	host, ok := configMap["host"].(string)
	if !ok {
		host = "localhost"
		logger.Warn("Using default Redis host: localhost")
	}

	portStr, ok := configMap["port"].(string)
	if !ok {
		portStr = "6379"
		logger.Warn("Using default Redis port: 6379")
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid Redis port: %w", err)
	}

	return &Config{
		Host: host,
		Port: port,
	}, nil
}

// NewClient creates a new Redis client
func NewClient(config *Config, logger *zap.Logger) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	logger.Info("Redis client initialized",
		zap.String("host", config.Host),
		zap.Int("port", config.Port))

	return client
}
