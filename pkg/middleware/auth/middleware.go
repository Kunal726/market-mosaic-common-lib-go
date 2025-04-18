package auth

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/Kunal726/market-mosaic-common-lib-go/pkg/auth"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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
func GetUserFromContext(c *gin.Context) (*auth.TokenValidationResponse, bool) {
	user, exists := c.Get(auth.UserContextKey)
	if !exists {
		return nil, false
	}

	validationResp, ok := user.(*auth.TokenValidationResponse)
	return validationResp, ok
}

// MustGetUserFromContext retrieves the user from the context or panics if not found
func MustGetUserFromContext(c *gin.Context) *auth.TokenValidationResponse {
	user, exists := GetUserFromContext(c)
	if !exists {
		panic(auth.ErrUserNotFoundInContext)
	}
	return user
}
