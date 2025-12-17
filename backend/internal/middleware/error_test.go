package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"backend/pkg/errors"
)

// Test 1: ErrorHandlerMiddleware - Normal Request (No Error)
func TestErrorHandlerMiddleware_NoError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(ErrorHandlerMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"success":true`)
}

// Test 2: ErrorHandlerMiddleware - Panic Recovery
func TestErrorHandlerMiddleware_PanicRecovery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(ErrorHandlerMiddleware())
	router.GET("/panic", func(c *gin.Context) {
		panic("something went wrong!")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "INTERNAL_ERROR")
	assert.Contains(t, w.Body.String(), "An unexpected error occurred")
	assert.Contains(t, w.Body.String(), `"success":false`)
}

// Test 3: ErrorHandlerMiddleware - String Panic
func TestErrorHandlerMiddleware_StringPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(ErrorHandlerMiddleware())
	router.GET("/panic", func(c *gin.Context) {
		panic("database connection failed")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "INTERNAL_ERROR")
}

// Test 4: ErrorHandlerMiddleware - Error Object Panic
func TestErrorHandlerMiddleware_ErrorPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(ErrorHandlerMiddleware())
	router.GET("/panic", func(c *gin.Context) {
		panic(fmt.Errorf("critical error occurred"))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// Test 5: ErrorHandlerMiddleware - Custom AppError
func TestErrorHandlerMiddleware_CustomAppError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(ErrorHandlerMiddleware())
	router.GET("/error", func(c *gin.Context) {
		// Add custom error to context
		c.Error(errors.NewAuthenticationError("Invalid credentials"))
		c.Abort()
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/error", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "AUTHENTICATION_ERROR")
	assert.Contains(t, w.Body.String(), "Invalid credentials")
	assert.Contains(t, w.Body.String(), `"success":false`)
}

// Test 6: ErrorHandlerMiddleware - Validation Error
func TestErrorHandlerMiddleware_ValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(ErrorHandlerMiddleware())
	router.POST("/validate", func(c *gin.Context) {
		c.Error(errors.NewValidationError([]errors.ValidationError{
			{Field: "email", Message: "Email is required"},
		}))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/validate", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "VALIDATION_ERROR")
	assert.Contains(t, w.Body.String(), "Email is required")
}

// Test 7: ErrorHandlerMiddleware - Authorization Error
func TestErrorHandlerMiddleware_AuthorizationError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(ErrorHandlerMiddleware())
	router.GET("/forbidden", func(c *gin.Context) {
		c.Error(errors.NewAuthorizationError("Insufficient permissions"))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/forbidden", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "AUTHORIZATION_ERROR")
	assert.Contains(t, w.Body.String(), "Insufficient permissions")
}

// Test 8: ErrorHandlerMiddleware - Not Found Error
func TestErrorHandlerMiddleware_NotFoundError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(ErrorHandlerMiddleware())
	router.GET("/resource", func(c *gin.Context) {
		c.Error(errors.NewNotFoundError("User with ID 123"))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/resource", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "NOT_FOUND")
	assert.Contains(t, w.Body.String(), "User with ID 123")
}

// Test 9: ErrorHandlerMiddleware - Rate Limit Error
func TestErrorHandlerMiddleware_RateLimitError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(ErrorHandlerMiddleware())
	router.GET("/api", func(c *gin.Context) {
		c.Error(errors.NewRateLimitError())
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Body.String(), "RATE_LIMIT_EXCEEDED")
	assert.Contains(t, w.Body.String(), "Too many requests")
}

// Test 10: ErrorHandlerMiddleware - Generic Error
func TestErrorHandlerMiddleware_GenericError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(ErrorHandlerMiddleware())
	router.GET("/error", func(c *gin.Context) {
		// Add generic Go error (not AppError)
		c.Error(fmt.Errorf("database connection timeout"))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/error", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "INTERNAL_ERROR")
	assert.Contains(t, w.Body.String(), "database connection timeout")
}

// Test 11: ErrorHandlerMiddleware - Multiple Errors (Last One Wins)
func TestErrorHandlerMiddleware_MultipleErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(ErrorHandlerMiddleware())
	router.GET("/multi-error", func(c *gin.Context) {
		c.Error(errors.NewValidationError([]errors.ValidationError{
			{Field: "test", Message: "First error"},
		}))
		c.Error(errors.NewAuthenticationError("Second error"))
		c.Error(errors.NewAuthorizationError("Third error"))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/multi-error", nil)
	router.ServeHTTP(w, req)

	// Should return the LAST error
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "AUTHORIZATION_ERROR")
	assert.Contains(t, w.Body.String(), "Third error")
}

// Test 12: ErrorHandlerMiddleware - No Errors Added
func TestErrorHandlerMiddleware_NoErrorsAdded(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(ErrorHandlerMiddleware())
	router.GET("/test", func(c *gin.Context) {
		// Handler completes without adding errors
		// No c.JSON() call - testing empty response
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// Should return 200 with empty body
	assert.Equal(t, http.StatusOK, w.Code)
}

// Test 13: ErrorHandlerMiddleware - Conflict Error
func TestErrorHandlerMiddleware_ConflictError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(ErrorHandlerMiddleware())
	router.POST("/create", func(c *gin.Context) {
		c.Error(errors.NewConflictError("User with email test@example.com already exists"))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/create", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "CONFLICT")
	assert.Contains(t, w.Body.String(), "User with email test@example.com already exists")
}

// Test 14: ErrorHandlerMiddleware - Panic After Response Sent
func TestErrorHandlerMiddleware_PanicAfterResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(ErrorHandlerMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
		// Panic after response sent
		panic("panic after response")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// Panic is recovered but response already sent
	assert.Equal(t, http.StatusOK, w.Code)
}

// Test 15: ErrorHandlerMiddleware - Chain of Handlers
func TestErrorHandlerMiddleware_HandlerChain(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(ErrorHandlerMiddleware())

	// First handler adds error and aborts
	handler1 := func(c *gin.Context) {
		c.Error(errors.NewValidationError([]errors.ValidationError{
			{Field: "input", Message: "Invalid input"},
		}))
		c.Abort()
	}

	// Second handler should not execute due to error
	handler2 := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"executed": true})
	}

	router.GET("/chain", handler1, handler2)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/chain", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "VALIDATION_ERROR")
	assert.NotContains(t, w.Body.String(), "executed")
}
