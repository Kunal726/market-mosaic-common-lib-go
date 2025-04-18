package auth

import (
	"time"
)

const (
	// HTTP related constants
	AuthorizationHeader = "Authorization"
	BearerPrefix        = "Bearer "
	JWTCookieName       = "JWT_SESSION"
	UserContextKey      = "user"

	// Environment variables
	EnvDevelopment = "development"
	EnvProduction  = "production"

	// HTTP status messages
	ErrInvalidToken          = "invalid token"
	ErrNoTokenProvided       = "no valid authorization token provided"
	ErrInvalidHeaderFormat   = "invalid authorization header format"
	ErrTokenNotFound         = "no token found in header or cookie"
	ErrUserNotFoundInContext = "user not found in context"

	// Default values
	DefaultHTTPTimeout = 5 * time.Second
)
