package logger

import (
	"fmt"
	"os"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LoggerConfig represents the configuration for the logger
type LoggerConfig struct {
	LogDir      string
	LogFileName string
	MaxSize     int
	MaxBackups  int
	MaxAge      int
	Compress    bool
}

// GetSitLogger creates a logger for SIT environment
func GetSitLogger() *zap.Logger {
	config := LoggerConfig{
		LogDir:      "target",
		LogFileName: fmt.Sprintf("target/%s.log", os.Getenv("SERVICE_NAME")),
		MaxSize:     500,
		MaxBackups:  100,
		MaxAge:      30,
		Compress:    true,
	}
	return createLogger(config, zapcore.DebugLevel)
}

// GetProdLogger creates a logger for Production environment
func GetProdLogger() *zap.Logger {
	config := LoggerConfig{
		LogDir:      "applogs/jiopay_logs",
		LogFileName: "applogs/jiopay_logs/sample-gp-proj-serv.log",
		MaxSize:     500,
		MaxBackups:  100,
		MaxAge:      30,
		Compress:    true,
	}
	return createLogger(config, zapcore.InfoLevel)
}

// createLogger creates a new logger instance with the given configuration
func createLogger(config LoggerConfig, logLevel zapcore.Level) *zap.Logger {
	// Create log directory if it doesn't exist
	err := os.MkdirAll(config.LogDir, os.ModePerm)
	if err != nil {
		fmt.Printf("could not create directory %s: %v", config.LogDir, err)
	}

	// Configure log rotation
	logWriter := &lumberjack.Logger{
		Filename:   config.LogFileName,
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
		Compress:   config.Compress,
	}

	// Set up the logger configuration
	zapConfig := zap.NewProductionConfig()
	zapConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006/01/02 15:04:05.0000")
	zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	zapConfig.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	zapConfig.EncoderConfig.FunctionKey = "func"

	// Create a core with the lumberjack writer and encoder
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(zapConfig.EncoderConfig),
		zapcore.AddSync(logWriter),
		logLevel,
	)

	// Create and return the logger instance
	return zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(0),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
}
