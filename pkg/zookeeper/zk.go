package zookeeper

import (
	"fmt"
	"os"
	"time"

	"github.com/samuel/go-zookeeper/zk"
	"go.uber.org/zap"
)

// Client represents a ZooKeeper client
type Client struct {
	conn     *zk.Conn
	config   *Config
	cache    *ConfigCache
	stopChan chan struct{}
	logger   *zap.Logger
}

// NewClient creates a new ZooKeeper client
func NewClient(logger *zap.Logger) (*Client, error) {
	config := DefaultConfig()

	// Set configuration from environment variables
	config.Hosts = []string{fmt.Sprintf("%s:%s",
		os.Getenv("ZK_HOST"),
		os.Getenv("ZK_PORT"))}
	config.ServiceName = os.Getenv("SERVICE_NAME")
	config.CommonLibName = os.Getenv("COMMON_LIB_NAME")

	if config.ServiceName == "" {
		return nil, fmt.Errorf("SERVICE_NAME environment variable is required")
	}

	// Connect to ZooKeeper
	conn, _, err := zk.Connect(config.Hosts, config.SessionTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ZooKeeper: %w", err)
	}

	client := &Client{
		conn:     conn,
		config:   config,
		cache:    NewConfigCache(logger),
		stopChan: make(chan struct{}),
		logger:   logger,
	}

	// Initial load of configurations
	if err := client.loadConfigurations(); err != nil {
		return nil, fmt.Errorf("failed to load initial configurations: %w", err)
	}

	// Start background refresh
	go client.startConfigRefresh()

	return client, nil
}

// loadConfigurations loads both service and common configurations
func (c *Client) loadConfigurations() error {
	// Load service config
	servicePath := fmt.Sprintf("/config/%s/config-properties", c.config.ServiceName)
	data, _, err := c.conn.Get(servicePath)
	if err != nil && err != zk.ErrNoNode {
		return fmt.Errorf("failed to get service config: %w", err)
	}
	if err == nil {
		if err := c.cache.UpdateServiceConfig(data); err != nil {
			return fmt.Errorf("failed to update service config: %w", err)
		}
	}

	// Load common config
	if c.config.CommonLibName != "" {
		commonPath := fmt.Sprintf("/config/application/%s", c.config.CommonLibName)
		data, _, err := c.conn.Get(commonPath)
		if err != nil && err != zk.ErrNoNode {
			return fmt.Errorf("failed to get common config: %w", err)
		}
		if err == nil {
			if err := c.cache.UpdateCommonConfig(data); err != nil {
				return fmt.Errorf("failed to update common config: %w", err)
			}
		}
	}

	return nil
}

// startConfigRefresh starts a background goroutine to refresh configurations
func (c *Client) startConfigRefresh() {
	ticker := time.NewTicker(c.config.RefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := c.loadConfigurations(); err != nil {
				c.logger.Error("Failed to refresh configurations", zap.Error(err))
			}
		case <-c.stopChan:
			return
		}
	}
}

// GetStringValueByKey retrieves a string value from service configuration
func (c *Client) GetStringValueByKey(key string, isCommon bool) (string, error) {
	value, exists := c.cache.GetConfig(isCommon, key)
	if !exists {
		return "", fmt.Errorf("key %s not found in configuration", key)
	}

	if strValue, ok := value.(string); ok {
		return strValue, nil
	}
	return fmt.Sprintf("%v", value), nil
}

// GetConfigValueByKey retrieves a configuration value from common configuration
func (c *Client) GetConfigValueByKey(key string, isCommon bool) (any, error) {
	value, exists := c.cache.GetConfig(isCommon, key)
	if !exists {
		return nil, fmt.Errorf("key %s not found in common configuration", key)
	}
	return value, nil
}

// RefreshData manually triggers a refresh of the configurations
func (c *Client) RefreshData() {
	c.logger.Info("Manual ZooKeeper config refresh triggered")
	if err := c.loadConfigurations(); err != nil {
		c.logger.Error("Failed to refresh configurations", zap.Error(err))
	}
}

// Get retrieves the data and stat of a node
func (c *Client) Get(path string) ([]byte, *zk.Stat, error) {
	data, stat, err := c.conn.Get(path)
	if err == zk.ErrNoNode {
		return nil, nil, fmt.Errorf("node %s does not exist", path)
	} else if err != nil {
		return nil, nil, fmt.Errorf("failed to get node %s: %w", path, err)
	}
	return data, stat, nil
}

// GetChildren retrieves the children of a node
func (c *Client) GetChildren(path string) ([]string, error) {
	children, _, err := c.conn.Children(path)
	if err == zk.ErrNoNode {
		return nil, fmt.Errorf("node %s does not exist", path)
	} else if err != nil {
		return nil, fmt.Errorf("failed to get children of node %s: %w", path, err)
	}
	return children, nil
}

// Exists checks if a node exists
func (c *Client) Exists(path string) (bool, error) {
	exists, _, err := c.conn.Exists(path)
	if err != nil {
		return false, fmt.Errorf("failed to check existence of node %s: %w", path, err)
	}
	return exists, nil
}

// Close closes the ZooKeeper connection and stops the refresh goroutine
func (c *Client) Close() {
	close(c.stopChan)
	c.conn.Close()
}
