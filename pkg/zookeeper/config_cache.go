package zookeeper

import (
	"encoding/base64"
	"encoding/json"
	"sync"

	"go.uber.org/zap"
)

// ConfigCache manages the caching of service and common configurations
type ConfigCache struct {
	serviceConfig map[string]any
	commonConfig  map[string]any
	mu            sync.RWMutex
	logger        *zap.Logger
}

// NewConfigCache creates a new ConfigCache instance
func NewConfigCache(logger *zap.Logger) *ConfigCache {
	return &ConfigCache{
		serviceConfig: make(map[string]any),
		commonConfig:  make(map[string]any),
		logger:        logger,
	}
}

// UpdateServiceConfig updates the service configuration cache
func (cc *ConfigCache) UpdateServiceConfig(data []byte) error {
	decodedData, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		cc.logger.Warn("Failed to decode service config", zap.Error(err))
		return err
	}

	var config map[string]any
	if err := json.Unmarshal(decodedData, &config); err != nil {
		cc.logger.Warn("Failed to parse service config", zap.Error(err))
		return err
	}

	cc.mu.Lock()
	cc.serviceConfig = config
	cc.mu.Unlock()
	return nil
}

// UpdateCommonConfig updates the common configuration cache
func (cc *ConfigCache) UpdateCommonConfig(data []byte) error {
	decodedData, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		cc.logger.Warn("Failed to decode common config", zap.Error(err))
		return err
	}

	var config map[string]any
	if err := json.Unmarshal(decodedData, &config); err != nil {
		cc.logger.Warn("Failed to parse common config", zap.Error(err))
		return err
	}

	cc.mu.Lock()
	cc.commonConfig = config
	cc.mu.Unlock()
	return nil
}

// GetServiceConfig retrieves a value from service configuration
func (cc *ConfigCache) getServiceConfig(key string) (any, bool) {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	value, exists := cc.serviceConfig[key]
	return value, exists
}

// GetCommonConfig retrieves a value from common configuration
func (cc *ConfigCache) getCommonConfig(key string) (any, bool) {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	value, exists := cc.commonConfig[key]
	return value, exists
}

// GetConfig retrieves a config based on type
func (cc *ConfigCache) GetConfig(isCommon bool, key string) (any, bool) {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	if isCommon {
		return cc.getCommonConfig(key)
	}
	
	return cc.getServiceConfig(key)

}