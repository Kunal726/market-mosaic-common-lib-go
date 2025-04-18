package db

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Kunal726/market-mosaic-common-lib-go/pkg/zookeeper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// DBConfig holds the database configuration
type DBConfig struct {
	URL         string `json:"url"`
	Username    string `json:"userName"`
	Password    string `json:"password"`
	MaxPoolSize int    `json:"maxPoolSize"`
}

// NewDBConfig creates a new DBConfig from ZooKeeper configuration
func NewDBConfig(zkClient *zookeeper.Client) (*DBConfig, error) {
	dbConfig, err := zkClient.GetConfigValueByKey("DB_CONFIG")
	if err != nil {
		return nil, fmt.Errorf("failed to get DB_CONFIG from ZooKeeper: %w", err)
	}

	configMap, ok := dbConfig.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid DB config format")
	}

	// Type assertions for each field
	url, ok := configMap["goUrl"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid url in DB config")
	}

	username, ok := configMap["userName"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid userName in DB config")
	}

	password, ok := configMap["password"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid password in DB config")
	}

	maxPoolSizeStr, ok := configMap["maxPoolSize"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid maxPoolSize in DB config")
	}

	maxPoolSize, err := strconv.Atoi(maxPoolSizeStr)
	if err != nil {
		return nil, fmt.Errorf("invalid maxPoolSize value: %w", err)
	}

	return &DBConfig{
		URL:         url,
		Username:    username,
		Password:    password,
		MaxPoolSize: maxPoolSize,
	}, nil
}

// InitDB initializes and returns a new database connection with connection pooling
func InitDB(zkClient *zookeeper.Client) (*gorm.DB, error) {
	config, err := NewDBConfig(zkClient)
	if err != nil {
		return nil, fmt.Errorf("failed to get DB config: %w", err)
	}

	// Configure GORM with connection pooling
	gormConfig := &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	}

	// Create MySQL DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Username, config.Password, config.URL, "your_db_name")

	// Open database connection with connection pooling
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       dsn,
		DefaultStringSize:         256,
		DisableDatetimePrecision:  true,
		DontSupportRenameIndex:    true,
		DontSupportRenameColumn:   true,
		SkipInitializeWithVersion: false,
	}), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get the underlying *sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying *sql.DB: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(config.MaxPoolSize / 2)
	sqlDB.SetMaxOpenConns(config.MaxPoolSize)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}
