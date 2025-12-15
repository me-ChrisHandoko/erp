package response

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError represents an application-level error
type AppError struct {
	StatusCode int
	Code       string
	Message    string
	Err        error
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates a new application error
func NewAppError(statusCode int, code, message string, err error) *AppError {
	return &AppError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
		Err:        err,
	}
}

// Common error constructors

func ErrBadRequest(message string) *AppError {
	return NewAppError(http.StatusBadRequest, "BAD_REQUEST", message, nil)
}

func ErrUnauthorized(message string) *AppError {
	if message == "" {
		message = "Unauthorized access"
	}
	return NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", message, nil)
}

func ErrForbidden(message string) *AppError {
	if message == "" {
		message = "Access forbidden"
	}
	return NewAppError(http.StatusForbidden, "FORBIDDEN", message, nil)
}

func ErrNotFound(resource string) *AppError {
	message := fmt.Sprintf("%s not found", resource)
	return NewAppError(http.StatusNotFound, "NOT_FOUND", message, nil)
}

func ErrConflict(message string) *AppError {
	return NewAppError(http.StatusConflict, "CONFLICT", message, nil)
}

func ErrValidation(message string) *AppError {
	return NewAppError(http.StatusUnprocessableEntity, "VALIDATION_ERROR", message, nil)
}

func ErrInternalServer(err error) *AppError {
	return NewAppError(http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error", err)
}

func ErrTenantNotActive(tenantID string) *AppError {
	message := fmt.Sprintf("Tenant %s is not active or subscription has expired", tenantID)
	return NewAppError(http.StatusForbidden, "TENANT_NOT_ACTIVE", message, nil)
}

func ErrInsufficientPermission(permission string) *AppError {
	message := fmt.Sprintf("Insufficient permission: %s required", permission)
	return NewAppError(http.StatusForbidden, "INSUFFICIENT_PERMISSION", message, nil)
}

func ErrInvalidCredentials() *AppError {
	return NewAppError(http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password", nil)
}

func ErrTokenExpired() *AppError {
	return NewAppError(http.StatusUnauthorized, "TOKEN_EXPIRED", "Token has expired", nil)
}

func ErrInvalidToken() *AppError {
	return NewAppError(http.StatusUnauthorized, "INVALID_TOKEN", "Invalid token", nil)
}

func ErrDuplicateEntry(field string) *AppError {
	message := fmt.Sprintf("%s already exists", field)
	return NewAppError(http.StatusConflict, "DUPLICATE_ENTRY", message, nil)
}

func ErrInsufficientStock(productName string, available, required int) *AppError {
	message := fmt.Sprintf("Insufficient stock for %s: available %d, required %d", productName, available, required)
	return NewAppError(http.StatusConflict, "INSUFFICIENT_STOCK", message, nil)
}

func ErrBatchExpired(batchNumber string) *AppError {
	message := fmt.Sprintf("Batch %s has expired", batchNumber)
	return NewAppError(http.StatusConflict, "BATCH_EXPIRED", message, nil)
}

func ErrInvalidStatusTransition(from, to string) *AppError {
	message := fmt.Sprintf("Invalid status transition from %s to %s", from, to)
	return NewAppError(http.StatusConflict, "INVALID_STATUS_TRANSITION", message, nil)
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// GetAppError extracts AppError from error chain
func GetAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return nil
}
