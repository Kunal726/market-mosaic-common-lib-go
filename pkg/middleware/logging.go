package middleware

import (
	"os"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

// LoggingMiddleware creates a middleware that adds tracing and logging context to the request
func LoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the current trace and span from the Gin context
		tracerName := os.Getenv("SERVICE_NAME") + "-tracer"
		tracer := otel.Tracer(tracerName)

		// Start a new span
		_, span := tracer.Start(c.Request.Context(), "traceName")
		defer span.End()

		// Get trace and span IDs
		traceID := span.SpanContext().TraceID().String()
		spanID := span.SpanContext().SpanID().String()

		// Add trace_id and span_id to the logger
		loggerWithFields := logger.With(
			zap.String("trace_id", traceID),
			zap.String("span_id", spanID),
		)

		// Pass the logger with these fields to the context
		c.Set("logger", loggerWithFields)

		// Proceed to the next middleware/handler
		c.Next()
	}
}
