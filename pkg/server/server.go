package server

import (
	"reflect"

	"github.com/Kunal726/market-mosaic-common-lib-go/pkg/middleware"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// Server represents the HTTP server configuration
type Server struct {
	engine *gin.Engine
	port   string
}

// NewServer creates and configures a new server instance
func NewServer(logger *zap.Logger, port string) *Server {
	if port == "" {
		port = ":8080"
	}

	engine := setupGinEngine(logger)

	return &Server{
		engine: engine,
		port:   port,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	return s.engine.Run(s.port)
}

// Engine returns the underlying gin engine
func (s *Server) Engine() *gin.Engine {
	return s.engine
}

// setupGinEngine configures the Gin engine with middleware
func setupGinEngine(logger *zap.Logger) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	// Setup validator
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := fld.Tag.Get("json")
			if name == "" {
				name = fld.Name
			}
			return name
		})
	}

	engine.Use(
		gin.Recovery(),
		middleware.LoggingMiddleware(logger),
		middleware.ErrorHandler(logger),
	)

	return engine
}
