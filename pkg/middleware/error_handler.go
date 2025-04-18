package middleware

import (
	"net/http"

	"github.com/Kunal726/market-mosaic-common-lib-go/pkg/dtos"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ErrorHandler is a middleware that handles errors and returns appropriate responses
func ErrorHandler(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			logger.Error("Request error", zap.Error(err))

			// Create error response
			response := dtos.BaseResponseDTO{
				Status:  false,
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			}

			// Send error response
			c.JSON(http.StatusInternalServerError, response)
		}
	}
}
