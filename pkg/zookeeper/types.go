package zookeeper

import "time"

const (
	// DefaultRefreshInterval is the default time interval for configuration refresh
	DefaultRefreshInterval = 30 * time.Second

	// DefaultConnectionTimeout is the default timeout for ZooKeeper connection
	DefaultConnectionTimeout = 10 * time.Second

	// DefaultSessionTimeout is the default session timeout for ZooKeeper
	DefaultSessionTimeout = 30 * time.Second
)

// Config represents the ZooKeeper client configuration
type Config struct {
	Hosts             []string
	ServiceName       string
	CommonLibName     string
	RefreshInterval   time.Duration
	ConnectionTimeout time.Duration
	SessionTimeout    time.Duration
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		RefreshInterval:   DefaultRefreshInterval,
		ConnectionTimeout: DefaultConnectionTimeout,
		SessionTimeout:    DefaultSessionTimeout,
	}
}
