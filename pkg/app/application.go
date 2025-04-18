package app

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Kunal726/market-mosaic-common-lib-go/pkg/db"
	"github.com/Kunal726/market-mosaic-common-lib-go/pkg/logger"
	"github.com/Kunal726/market-mosaic-common-lib-go/pkg/redis"
	"github.com/Kunal726/market-mosaic-common-lib-go/pkg/tracing"
	"github.com/Kunal726/market-mosaic-common-lib-go/pkg/zookeeper"
	"github.com/joho/godotenv"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Application holds all dependencies
type Application struct {
	Logger        *zap.Logger
	DB            *gorm.DB
	TraceProvider *trace.TracerProvider
	Environment   string
	RedisManager  *redis.Manager
	ZKClient      *zookeeper.Client
}

// NewApplication initializes and returns a new Application instance
func NewApplication() (*Application, error) {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	// Get environment
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}

	// Initialize logger based on environment
	var zapLogger *zap.Logger
	if env == "production" {
		zapLogger = logger.GetProdLogger()
	} else {
		zapLogger = logger.GetSitLogger()
	}

	// Initialize tracing
	traceProvider := tracing.InitTracer()

	// Initialize database
	database, err := db.InitDB()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize ZooKeeper client
	zkClient, err := zookeeper.NewClient(zapLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize ZooKeeper client: %w", err)
	}

	// Initialize Redis configuration
	redisConfig, err := redis.NewConfig(zkClient, zapLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis config: %w", err)
	}

	// Create Redis client
	redisClient := redis.NewClient(redisConfig, zapLogger)

	// Create Redis manager
	redisManager := redis.NewManager(redisClient, zapLogger)

	return &Application{
		Logger:        zapLogger,
		DB:            database,
		TraceProvider: traceProvider,
		Environment:   env,
		RedisManager:  redisManager,
		ZKClient:      zkClient,
	}, nil
}

// Cleanup performs cleanup of application resources
func (app *Application) Cleanup() {
	if err := app.Logger.Sync(); err != nil {
		log.Printf("failed to sync logger: %v", err)
	}

	if app.TraceProvider != nil {
		if err := app.TraceProvider.Shutdown(context.Background()); err != nil {
			log.Printf("failed to shutdown trace provider: %v", err)
		}
	}

	if app.DB != nil {
		sqlDB, err := app.DB.DB()
		if err != nil {
			log.Printf("failed to get database instance: %v", err)
			return
		}
		if err := sqlDB.Close(); err != nil {
			log.Printf("failed to close database connection: %v", err)
		}
	}

	if app.RedisManager != nil {
		if err := app.RedisManager.Close(); err != nil {
			log.Printf("failed to close Redis connection: %v", err)
		}
	}

	if app.ZKClient != nil {
		app.ZKClient.Close()
	}
}
