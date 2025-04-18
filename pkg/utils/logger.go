package utils

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func GetLoggerFromContext(c *gin.Context) *zap.Logger {
	// Retrieve the logger from the context and assert it to *zap.Logger
	logger := c.MustGet("logger")
	return logger.(*zap.Logger)
}
