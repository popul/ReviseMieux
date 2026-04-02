package dto

// ErrorResponse is the standard error response.
type ErrorResponse struct {
	Error   string `json:"error" example:"resource not found"`
	Details string `json:"details,omitempty" example:"chapter with id abc not found"`
}

// HealthResponse is the response for the health check endpoint.
type HealthResponse struct {
	Status  string `json:"status" example:"ok"`
	Version string `json:"version,omitempty" example:"0.1.0"`
}

// MessageResponse is a generic message response.
type MessageResponse struct {
	Message string `json:"message" example:"operation successful"`
}

// DevTokenResponse is the response for the dev token endpoint.
type DevTokenResponse struct {
	Token  string `json:"token"`
	UserID string `json:"user_id"`
}
