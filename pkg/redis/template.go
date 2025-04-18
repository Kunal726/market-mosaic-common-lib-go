package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Template represents a Redis template for operations
type Template struct {
	client *redis.Client
	logger *zap.Logger
}

// NewTemplate creates a new Redis template
func NewTemplate(client *redis.Client, logger *zap.Logger) *Template {
	return &Template{
		client: client,
		logger: logger,
	}
}

// Set sets a key-value pair in Redis
func (t *Template) Set(ctx context.Context, key string, value any) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	err = t.client.Set(ctx, key, jsonValue, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}

	return nil
}

// Get gets a value from Redis by key
func (t *Template) Get(ctx context.Context, key string, value any) error {
	jsonValue, err := t.client.Get(ctx, key).Bytes()
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
func (t *Template) Delete(ctx context.Context, key string) error {
	err := t.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete key %s: %w", key, err)
	}

	return nil
}

// Close closes the Redis client connection
func (t *Template) Close() error {
	return t.client.Close()
}
