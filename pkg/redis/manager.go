package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Manager handles Redis operations
type Manager struct {
	client *redis.Client
	logger *zap.Logger
}

// NewManager creates a new Redis manager
func NewManager(client *redis.Client, logger *zap.Logger) *Manager {
	return &Manager{
		client: client,
		logger: logger,
	}
}

// Set sets a key-value pair in Redis with optional expiration
func (m *Manager) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	err = m.client.Set(ctx, key, jsonValue, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}

	return nil
}

// Get gets a value from Redis by key and unmarshals it into the provided value
func (m *Manager) Get(ctx context.Context, key string, value any) error {
	jsonValue, err := m.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key %s does not exist", key)
		}
		return fmt.Errorf("failed to get key %s: %w", key, err)
	}

	err = json.Unmarshal(jsonValue, value)
	if err != nil {
		return fmt.Errorf("failed to unmarshal value for key %s: %w", key, err)
	}

	return nil
}

// Delete deletes a key from Redis
func (m *Manager) Delete(ctx context.Context, key string) (bool, error) {
	result, err := m.client.Del(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to delete key %s: %w", key, err)
	}
	return result > 0, nil
}

// Increment increments a counter in Redis
func (m *Manager) Increment(ctx context.Context, key string) (int64, error) {
	result, err := m.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment key %s: %w", key, err)
	}
	return result, nil
}

// Exists checks if a key exists in Redis
func (m *Manager) Exists(ctx context.Context, key string) (bool, error) {
	result, err := m.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check existence of key %s: %w", key, err)
	}
	return result > 0, nil
}

// Expire sets expiration time for a key
func (m *Manager) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	result, err := m.client.Expire(ctx, key, expiration).Result()
	if err != nil {
		return false, fmt.Errorf("failed to set expiration for key %s: %w", key, err)
	}
	return result, nil
}

// GetTimeToLive gets the remaining time to live for a key
func (m *Manager) GetTimeToLive(ctx context.Context, key string) (time.Duration, error) {
	result, err := m.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL for key %s: %w", key, err)
	}
	return result, nil
}

// Close closes the Redis client connection
func (m *Manager) Close() error {
	return m.client.Close()
}
