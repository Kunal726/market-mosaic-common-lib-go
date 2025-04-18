package dtos

// BaseResponseDTO represents the base response structure
type BaseResponseDTO struct {
	Status  bool   `json:"status"`
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}
