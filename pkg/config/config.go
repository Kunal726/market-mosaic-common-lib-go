package config

import (
	"sync"

	"github.com/Kunal726/market-mosaic-common-lib-go/pkg/zookeeper"
	"go.uber.org/zap"
)

var (
	instance *ConfigManager
	once     sync.Once
)

// ConfigManager manages the ZooKeeper configuration
type ConfigManager struct {
	zkClient *zookeeper.Client
	logger   *zap.Logger
}

// NewConfigManager creates a new ConfigManager instance
func NewConfigManager(logger *zap.Logger) (*ConfigManager, error) {
	var err error
	once.Do(func() {
		zkClient, zkErr := zookeeper.NewClient(logger)
		if zkErr != nil {
			err = zkErr
			return
		}
		instance = &ConfigManager{
			zkClient: zkClient,
			logger:   logger,
		}
	})
	return instance, err
}

// GetInstance returns the singleton instance of ConfigManager
func GetInstance() *ConfigManager {
	return instance
}

// GetStringValueByKey retrieves a string value from service configuration
func (cm *ConfigManager) GetStringValueByKey(key string, isCommon bool) (string, error) {
	return cm.zkClient.GetStringValueByKey(key, isCommon)
}

// GetConfigValueByKey retrieves a configuration value from common configuration
func (cm *ConfigManager) GetConfigValueByKey(key string, isCommon bool) (interface{}, error) {
	return cm.zkClient.GetConfigValueByKey(key, isCommon)
}

// RefreshData manually triggers a refresh of the configurations
func (cm *ConfigManager) RefreshData() {
	cm.logger.Info("Manual configuration refresh triggered")
	cm.zkClient.RefreshData()
}

// Close closes the ZooKeeper connection
func (cm *ConfigManager) Close() {
	if cm.zkClient != nil {
		cm.zkClient.Close()
	}
}
