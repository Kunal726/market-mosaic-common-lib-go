package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Kunal726/market-mosaic-common-lib-go/pkg/auth"
	"github.com/Kunal726/market-mosaic-common-lib-go/pkg/zookeeper"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	pb "github.com/Kunal726/market-mosaic-common-lib-go/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// Middleware handles token validation and user context
type Middleware struct {
	client *auth.Client
	logger *zap.Logger
}

// NewMiddleware creates a new instance of AuthMiddleware
func NewMiddleware(client *auth.Client, logger *zap.Logger) *Middleware {
	return &Middleware{
		client: client,
		logger: logger,
	}
}

// extractToken extracts JWT token from header or cookie
func (m *Middleware) extractToken(c *gin.Context) (string, error) {
	// First try to get token from Authorization header
	authHeader := c.GetHeader(auth.AuthorizationHeader)
	if authHeader != "" {
		if !strings.HasPrefix(authHeader, auth.BearerPrefix) {
			return "", errors.New(auth.ErrInvalidHeaderFormat)
		}
		return strings.TrimPrefix(authHeader, auth.BearerPrefix), nil
	}

	// If not in header, try to get from cookie
	cookie, err := c.Cookie(auth.JWTCookieName)
	if err != nil {
		if err == http.ErrNoCookie {
			return "", errors.New(auth.ErrTokenNotFound)
		}
		return "", fmt.Errorf("error reading cookie: %w", err)
	}

	return cookie, nil
}

// isDevelopmentMode checks if the application is running in development mode
func (m *Middleware) isDevelopmentMode() bool {
	env := os.Getenv("ENV")
	if env == "" {
		env = os.Getenv("APP_ENV")
	}
	return env == auth.EnvDevelopment
}

// createMockUser creates a mock user for development mode
func (m *Middleware) createMockUser() *auth.TokenValidationResponse {
	return &auth.TokenValidationResponse{
		Valid:       true,
		Username:    "dev-user",
		UserID:      1,
		Email:       "dev@example.com",
		Name:        "Development User",
		Authorities: []string{"ROLE_USER", "ROLE_ADMIN"},
	}
}

// ValidateToken middleware validates the token and sets user context
func (m *Middleware) ValidateToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if we're in development mode
		if m.isDevelopmentMode() {
			m.logger.Info("Running in development mode - using mock authentication")
			c.Set(auth.UserContextKey, m.createMockUser())
			c.Next()
			return
		}

		token, err := m.extractToken(c)
		if err != nil {
			m.logger.Error("Failed to extract token", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": auth.ErrNoTokenProvided,
			})
			return
		}

		validationResp, err := m.client.ValidateToken(token)
		if err != nil {
			m.logger.Error("Token validation failed", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": auth.ErrInvalidToken,
			})
			return
		}

		if !validationResp.Valid {
			m.logger.Error("Token is invalid")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": auth.ErrInvalidToken,
			})
			return
		}

		// Set user context for downstream handlers
		c.Set(auth.UserContextKey, validationResp)
		c.Next()
	}
}

// GetUserFromContext retrieves the user from the context
func GetUserFromContext(c *gin.Context) (*pb.TokenResponse, bool) {
	user, exists := c.Get(auth.UserContextKey)
	if !exists {
		return nil, false
	}

	validationResp, ok := user.(*pb.TokenResponse)
	return validationResp, ok
}

// MustGetUserFromContext retrieves the user from the context or panics if not found
func MustGetUserFromContext(c *gin.Context) *pb.TokenResponse {
	user, exists := GetUserFromContext(c)
	if !exists {
		panic(auth.ErrUserNotFoundInContext)
	}
	return user
}


// ValidateToken middleware validates the token and sets user context
func (m *Middleware) ValidateTokenGrpc(zkClient *zookeeper.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if we're in development mode
		if m.isDevelopmentMode() {
			m.logger.Info("Running in development mode - using mock authentication")
			c.Set(auth.UserContextKey, m.createMockUser())
			c.Next()
			return
		}

		token, err := m.extractToken(c)

		if err != nil {
			m.logger.Error("Failed to extract token", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": auth.ErrNoTokenProvided,
			})
			return
		}

		authServiceUrl, err := zkClient.GetStringValueByKey("AUTH_SERV_URL", true)
		if err != nil {
			authServiceUrl = "localhost:9090"
			m.logger.Error("Unable to get Auth Srevice Url From Congig", zap.Error(err))
		}

		conn, err := grpc.NewClient(authServiceUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Failed to Connect Auth Servers"})
			return
		}
		defer conn.Close()

		client := pb.NewAuthServiceClient(conn)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		md := metadata.Pairs("Cookie", auth.JWTCookieName + "=" + token)
		ctx = metadata.NewOutgoingContext(ctx, md)

		resp, err := client.ValidateToken(ctx, &pb.TokenRequest{})

		if err != nil || !resp.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		if !resp.Valid {
			m.logger.Error("Token is invalid")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": auth.ErrInvalidToken,
			})
			return
		}

		c.Set(auth.UserContextKey, resp)
		c.Next()
	}
}