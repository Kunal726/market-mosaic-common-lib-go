package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// TokenValidationResponse represents the response from the auth service
type TokenValidationResponse struct {
	Valid       bool     `json:"valid"`
	Username    string   `json:"username"`
	UserID      int      `json:"userId"`
	Email       string   `json:"email"`
	Name        string   `json:"name"`
	Authorities []string `json:"authorities"`
}

// Client handles communication with the auth service
type Client struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
}

// ClientOption represents a function that configures the Client
type ClientOption func(*Client)

// WithTimeout sets the HTTP client timeout
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// NewClient creates a new instance of AuthClient
func NewClient(baseURL string, logger *zap.Logger, opts ...ClientOption) *Client {
	client := &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: DefaultHTTPTimeout,
		},
		logger: logger,
	}

	// Apply options
	for _, opt := range opts {
		opt(client)
	}

	return client
}

// ValidateToken validates the given token with the auth service
func (c *Client) ValidateToken(token string) (*TokenValidationResponse, error) {
	if token == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}

	c.logger.Info("Validating token with auth service")

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/validate", c.baseURL), nil)
	if err != nil {
		c.logger.Error("Failed to create request", zap.Error(err))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add the token to the Authorization header
	req.Header.Set(AuthorizationHeader, fmt.Sprintf("%s%s", BearerPrefix, token))

	// Also set the token as a cookie
	cookie := &http.Cookie{
		Name:     JWTCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	req.AddCookie(cookie)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to send request", zap.Error(err))
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("Auth service returned non-200 status code",
			zap.Int("status_code", resp.StatusCode))
		return nil, fmt.Errorf("auth service returned status code: %d", resp.StatusCode)
	}

	var validationResp TokenValidationResponse
	if err := json.NewDecoder(resp.Body).Decode(&validationResp); err != nil {
		c.logger.Error("Failed to decode response", zap.Error(err))
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	c.logger.Info("Token validation completed",
		zap.Bool("valid", validationResp.Valid),
		zap.String("username", validationResp.Username))

	return &validationResp, nil
}
