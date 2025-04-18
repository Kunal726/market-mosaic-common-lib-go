package dtos

// Authority represents a granted authority from Spring Security
type Authority struct {
	Authority string `json:"authority"`
}

// TokenValidationResponse represents the response from auth service's validate endpoint
type TokenValidationResponse struct {
	Valid       bool        `json:"valid"`
	Username    string      `json:"username"`
	UserID      int64       `json:"userId"`
	Email       string      `json:"email"`
	Name        string      `json:"name"`
	Authorities []Authority `json:"authorities"`
}
