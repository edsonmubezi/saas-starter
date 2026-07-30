package handler

// Common response models for Swagger documentation
// These models ensure consistency across all API endpoints

// SuccessResponse represents a standard successful API response
type SuccessResponse struct {
	Status  int         `json:"status" example:"200"`
	Message string      `json:"message" example:"Operation completed successfully"`
	Data    interface{} `json:"data,omitempty" swaggertype:"object"`
}

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Status  int          `json:"status" example:"400"`
	Message string       `json:"message" example:"Invalid request"`
	Errors  []FieldError `json:"errors,omitempty"`
}

// FieldError represents a field-level validation error
type FieldError struct {
	Field   string `json:"field" example:"email"`
	Message string `json:"message" example:"Email is required"`
}

// PaginatedResponse represents a paginated list response
type PaginatedResponse struct {
	Status     int            `json:"status" example:"200"`
	Message    string         `json:"message" example:"Data retrieved successfully"`
	Data       interface{}    `json:"data" swaggertype:"array,object"`
	Pagination PaginationInfo `json:"pagination"`
}

// PaginationInfo contains pagination metadata
type PaginationInfo struct {
	Page       int  `json:"page" example:"1"`
	PageSize   int  `json:"page_size" example:"10"`
	TotalCount int  `json:"total_count" example:"100"`
	TotalPages int  `json:"total_pages" example:"10"`
	HasNext    bool `json:"has_next" example:"true"`
	HasPrev    bool `json:"has_prev" example:"false"`
}

// MessageResponse represents a simple message response
type MessageResponse struct {
	Status  int    `json:"status" example:"200"`
	Message string `json:"message" example:"Operation completed successfully"`
}

// PasswordResetSuccessResponse for password reset operations
type PasswordResetSuccessResponse struct {
	Status  int    `json:"status" example:"200"`
	Message string `json:"message" example:"If the email exists, a password reset link has been sent"`
}

// PasswordResetErrorResponse for password reset errors
type PasswordResetErrorResponse struct {
	Status  int    `json:"status" example:"400"`
	Message string `json:"message" example:"Invalid or expired token"`
}

// TokenVerificationResponse for reset token verification
type TokenVerificationResponse struct {
	Valid     bool   `json:"valid" example:"true"`
	ExpiresAt string `json:"expires_at,omitempty" example:"2025-10-20T12:30:00Z"`
	Message   string `json:"message,omitempty" example:"Token is valid"`
}

// AuthSuccessResponse for authentication success
type AuthSuccessResponse struct {
	Status       int    `json:"status" example:"200"`
	Message      string `json:"message" example:"Login successful"`
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// ValidationErrorResponse for validation failures
type ValidationErrorResponse struct {
	Status  int                    `json:"status" example:"400"`
	Message string                 `json:"message" example:"Validation failed"`
	Errors  map[string]interface{} `json:"errors"`
}
