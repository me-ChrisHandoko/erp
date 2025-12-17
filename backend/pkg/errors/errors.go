package errors

import (
	"fmt"
	"net/http"
)

// AppError represents a custom application error
// Reference: BACKEND-IMPLEMENTATION.md lines 1347-1524
type AppError struct {
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	StatusCode int               `json:"-"`
	Details    []ValidationError `json:"details,omitempty"`
	Internal   error             `json:"-"`
}

// ValidationError represents a field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if len(e.Details) > 0 {
		// Include validation details in error message
		msg := e.Message + ":"
		for _, detail := range e.Details {
			msg += fmt.Sprintf(" %s - %s;", detail.Field, detail.Message)
		}
		return msg
	}
	return e.Message
}

// Predefined error constructors

// NewValidationError creates a validation error
func NewValidationError(details []ValidationError) *AppError {
	return &AppError{
		Code:       "VALIDATION_ERROR",
		Message:    "Validation failed",
		StatusCode: http.StatusBadRequest,
		Details:    details,
	}
}

// NewAuthenticationError creates an authentication error
func NewAuthenticationError(message string) *AppError {
	return &AppError{
		Code:       "AUTHENTICATION_ERROR",
		Message:    message,
		StatusCode: http.StatusUnauthorized,
	}
}

// NewAuthorizationError creates an authorization error
func NewAuthorizationError(message string) *AppError {
	return &AppError{
		Code:       "AUTHORIZATION_ERROR",
		Message:    message,
		StatusCode: http.StatusForbidden,
	}
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Code:       "NOT_FOUND",
		Message:    resource + " not found",
		StatusCode: http.StatusNotFound,
	}
}

// NewInternalError creates an internal server error
func NewInternalError(err error) *AppError {
	return &AppError{
		Code:       "INTERNAL_ERROR",
		Message:    "Internal server error",
		StatusCode: http.StatusInternalServerError,
		Internal:   err,
	}
}

// NewRateLimitError creates a rate limit error
func NewRateLimitError() *AppError {
	return &AppError{
		Code:       "RATE_LIMIT_EXCEEDED",
		Message:    "Too many requests",
		StatusCode: http.StatusTooManyRequests,
	}
}

// NewAccountLockedError creates an account locked error with tier information
func NewAccountLockedError(tier int, retryAfterSeconds int, attemptsCount int64) *AppError {
	var message string
	if tier > 0 {
		message = fmt.Sprintf("Account locked (Tier %d). Too many failed login attempts (%d). Please try again in %d seconds.",
			tier, attemptsCount, retryAfterSeconds)
	} else {
		message = fmt.Sprintf("Account temporarily locked due to multiple failed login attempts. Please try again in %d seconds.",
			retryAfterSeconds)
	}

	return &AppError{
		Code:       "ACCOUNT_LOCKED",
		Message:    message,
		StatusCode: http.StatusForbidden,
	}
}

// NewConflictError creates a conflict error (e.g., duplicate email)
func NewConflictError(message string) *AppError {
	return &AppError{
		Code:       "CONFLICT",
		Message:    message,
		StatusCode: http.StatusConflict,
	}
}

// NewBadRequestError creates a bad request error
func NewBadRequestError(message string) *AppError {
	return &AppError{
		Code:       "BAD_REQUEST",
		Message:    message,
		StatusCode: http.StatusBadRequest,
	}
}

// NewSubscriptionError creates a subscription error
func NewSubscriptionError(message string) *AppError {
	return &AppError{
		Code:       "SUBSCRIPTION_ERROR",
		Message:    message,
		StatusCode: http.StatusPaymentRequired,
	}
}

// NewCSRFError creates a CSRF validation error
func NewCSRFError(message string) *AppError {
	return &AppError{
		Code:       "CSRF_ERROR",
		Message:    message,
		StatusCode: http.StatusForbidden,
	}
}
