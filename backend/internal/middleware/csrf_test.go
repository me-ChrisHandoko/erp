package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestCSRFMiddleware_ValidToken(t *testing.T) {
	// Setup
	router := setupTestRouter()

	router.Use(CSRFMiddleware())
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Generate CSRF token
	token, err := GenerateCSRFToken()
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Create request with matching cookie and header
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  "csrf_token",
		Value: token,
	})
	req.Header.Set("X-CSRF-Token", token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCSRFMiddleware_MissingCookie(t *testing.T) {
	// Setup
	router := setupTestRouter()

	router.Use(CSRFMiddleware())
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create request with header but no cookie
	token, _ := GenerateCSRFToken()
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("X-CSRF-Token", token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert - should fail with 403
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCSRFMiddleware_MissingHeader(t *testing.T) {
	// Setup
	router := setupTestRouter()

	router.Use(CSRFMiddleware())
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create request with cookie but no header
	token, _ := GenerateCSRFToken()
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  "csrf_token",
		Value: token,
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert - should fail with 403
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCSRFMiddleware_MismatchedTokens(t *testing.T) {
	// Setup
	router := setupTestRouter()

	router.Use(CSRFMiddleware())
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create request with different tokens
	cookieToken, _ := GenerateCSRFToken()
	headerToken, _ := GenerateCSRFToken()

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  "csrf_token",
		Value: cookieToken,
	})
	req.Header.Set("X-CSRF-Token", headerToken)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert - should fail with 403
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCSRFMiddleware_SafeMethods(t *testing.T) {
	// Setup
	router := setupTestRouter()

	router.Use(CSRFMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	router.HEAD("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.OPTIONS("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Test GET (should pass without CSRF token)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test HEAD (should pass without CSRF token)
	req = httptest.NewRequest(http.MethodHead, "/test", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test OPTIONS (should pass without CSRF token)
	req = httptest.NewRequest(http.MethodOptions, "/test", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGenerateCSRFToken(t *testing.T) {
	// Generate multiple tokens and ensure they're different
	token1, err1 := GenerateCSRFToken()
	token2, err2 := GenerateCSRFToken()

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotEmpty(t, token1)
	assert.NotEmpty(t, token2)
	assert.NotEqual(t, token1, token2, "Tokens should be unique")

	// Token should be base64 URL encoded (32 bytes = 43 chars base64)
	assert.Greater(t, len(token1), 40)
}

func TestSetCSRFCookie(t *testing.T) {
	// Setup
	router := setupTestRouter()

	router.GET("/set-csrf", func(c *gin.Context) {
		token, _ := GenerateCSRFToken()
		SetCSRFCookie(c, token, false) // secure = false for testing
		c.Status(http.StatusOK)
	})

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/set-csrf", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	cookies := w.Result().Cookies()
	assert.NotEmpty(t, cookies)

	var csrfCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "csrf_token" {
			csrfCookie = cookie
			break
		}
	}

	assert.NotNil(t, csrfCookie, "CSRF cookie should be set")
	assert.NotEmpty(t, csrfCookie.Value)
	assert.False(t, csrfCookie.HttpOnly, "CSRF cookie should NOT be httpOnly")
	assert.Equal(t, "/", csrfCookie.Path)
}

func TestClearCSRFCookie(t *testing.T) {
	// Setup
	router := setupTestRouter()

	router.GET("/clear-csrf", func(c *gin.Context) {
		ClearCSRFCookie(c, false) // secure = false for testing
		c.Status(http.StatusOK)
	})

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/clear-csrf", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	cookies := w.Result().Cookies()

	var csrfCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "csrf_token" {
			csrfCookie = cookie
			break
		}
	}

	assert.NotNil(t, csrfCookie, "CSRF cookie should be present")
	assert.Equal(t, -1, csrfCookie.MaxAge, "Cookie should be expired")
}

func TestSecureCompare(t *testing.T) {
	// Test equal strings
	assert.True(t, secureCompare("test123", "test123"))

	// Test different strings (same length)
	assert.False(t, secureCompare("test123", "test456"))

	// Test different lengths
	assert.False(t, secureCompare("test", "testing"))
	assert.False(t, secureCompare("testing", "test"))

	// Test empty strings
	assert.True(t, secureCompare("", ""))
	assert.False(t, secureCompare("test", ""))
	assert.False(t, secureCompare("", "test"))
}
