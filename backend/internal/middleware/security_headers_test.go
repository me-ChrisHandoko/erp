package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"backend/internal/config"
)

func TestSecurityHeadersMiddleware_Phase1Headers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Server: config.ServerConfig{
			Environment: "development",
		},
		Security: config.SecurityConfig{
			EnableXFrameOptions:     true,
			EnableXContentType:      true,
			EnableXXSSProtection:    true,
			EnableReferrerPolicy:    true,
			EnablePermissionsPolicy: true,
		},
	}

	router := gin.New()
	router.Use(SecurityHeadersMiddleware(cfg))
	router.GET("/test", func(c *gin.Context) {
		c.String(200, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// Assert Phase 1 headers are present
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
	assert.Contains(t, w.Header().Get("Permissions-Policy"), "camera=()")
	assert.Contains(t, w.Header().Get("Permissions-Policy"), "microphone=()")

	// Additional security headers
	assert.Equal(t, "none", w.Header().Get("X-Permitted-Cross-Domain-Policies"))
	assert.Equal(t, "noopen", w.Header().Get("X-Download-Options"))
	assert.Equal(t, "same-origin", w.Header().Get("Cross-Origin-Opener-Policy"))
	assert.Equal(t, "same-origin", w.Header().Get("Cross-Origin-Resource-Policy"))
}

func TestSecurityHeadersMiddleware_HSTS_Disabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Server: config.ServerConfig{
			Environment: "development",
		},
		Security: config.SecurityConfig{
			EnableHSTS: false, // Disabled
		},
	}

	router := gin.New()
	router.Use(SecurityHeadersMiddleware(cfg))
	router.GET("/test", func(c *gin.Context) {
		c.String(200, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// HSTS should NOT be present when disabled
	assert.Empty(t, w.Header().Get("Strict-Transport-Security"))
}

func TestSecurityHeadersMiddleware_HSTS_Enabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Server: config.ServerConfig{
			Environment: "production",
		},
		Security: config.SecurityConfig{
			EnableHSTS:            true,
			HSTSMaxAge:            31536000,
			HSTSIncludeSubDomains: true,
			HSTSPreload:           false,
		},
	}

	router := gin.New()
	router.Use(SecurityHeadersMiddleware(cfg))
	router.GET("/test", func(c *gin.Context) {
		c.String(200, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// HSTS should be present with correct values
	hstsHeader := w.Header().Get("Strict-Transport-Security")
	assert.Contains(t, hstsHeader, "max-age=31536000")
	assert.Contains(t, hstsHeader, "includeSubDomains")
	assert.NotContains(t, hstsHeader, "preload")
}

func TestSecurityHeadersMiddleware_HSTS_WithPreload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Server: config.ServerConfig{
			Environment: "production",
		},
		Security: config.SecurityConfig{
			EnableHSTS:            true,
			HSTSMaxAge:            31536000,
			HSTSIncludeSubDomains: true,
			HSTSPreload:           true,
		},
	}

	router := gin.New()
	router.Use(SecurityHeadersMiddleware(cfg))
	router.GET("/test", func(c *gin.Context) {
		c.String(200, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// HSTS should include preload
	hstsHeader := w.Header().Get("Strict-Transport-Security")
	assert.Contains(t, hstsHeader, "max-age=31536000")
	assert.Contains(t, hstsHeader, "includeSubDomains")
	assert.Contains(t, hstsHeader, "preload")
}

func TestSecurityHeadersMiddleware_CSP_Disabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Server: config.ServerConfig{
			Environment: "development",
		},
		Security: config.SecurityConfig{
			EnableCSP: false, // Disabled
		},
	}

	router := gin.New()
	router.Use(SecurityHeadersMiddleware(cfg))
	router.GET("/test", func(c *gin.Context) {
		c.String(200, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// CSP should NOT be present when disabled
	assert.Empty(t, w.Header().Get("Content-Security-Policy"))
	assert.Empty(t, w.Header().Get("Content-Security-Policy-Report-Only"))
}

func TestSecurityHeadersMiddleware_CSP_ReportOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Server: config.ServerConfig{
			Environment: "development",
		},
		Security: config.SecurityConfig{
			EnableCSP:     true,
			CSPReportOnly: true,
		},
	}

	router := gin.New()
	router.Use(SecurityHeadersMiddleware(cfg))
	router.GET("/test", func(c *gin.Context) {
		c.String(200, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// CSP should be in Report-Only mode
	assert.Empty(t, w.Header().Get("Content-Security-Policy"))
	cspReportOnly := w.Header().Get("Content-Security-Policy-Report-Only")
	assert.NotEmpty(t, cspReportOnly)
	assert.Contains(t, cspReportOnly, "default-src 'self'")
	assert.Contains(t, cspReportOnly, "frame-ancestors 'none'")
}

func TestSecurityHeadersMiddleware_CSP_Enforcement(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Server: config.ServerConfig{
			Environment: "production",
		},
		Security: config.SecurityConfig{
			EnableCSP:     true,
			CSPReportOnly: false, // Enforcement mode
		},
	}

	router := gin.New()
	router.Use(SecurityHeadersMiddleware(cfg))
	router.GET("/test", func(c *gin.Context) {
		c.String(200, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// CSP should be in enforcement mode
	assert.Empty(t, w.Header().Get("Content-Security-Policy-Report-Only"))
	csp := w.Header().Get("Content-Security-Policy")
	assert.NotEmpty(t, csp)
	assert.Contains(t, csp, "default-src 'self'")
	assert.Contains(t, csp, "frame-ancestors 'none'")
	assert.Contains(t, csp, "upgrade-insecure-requests")
}

func TestSecurityHeadersMiddleware_CSP_WithNonce(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Server: config.ServerConfig{
			Environment: "production",
		},
		Security: config.SecurityConfig{
			EnableCSP:     true,
			CSPReportOnly: false,
		},
	}

	router := gin.New()
	router.Use(SecurityHeadersMiddleware(cfg))
	router.GET("/test", func(c *gin.Context) {
		// Check if nonce is available in context
		nonce := GetCSPNonce(c)
		c.String(200, "Nonce: %s", nonce)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// CSP should contain nonce
	csp := w.Header().Get("Content-Security-Policy")
	assert.Contains(t, csp, "script-src")
	assert.Contains(t, csp, "'nonce-")

	// X-CSP-Nonce header should be present
	nonce := w.Header().Get("X-CSP-Nonce")
	assert.NotEmpty(t, nonce)

	// Response body should contain the same nonce
	assert.Contains(t, w.Body.String(), "Nonce: "+nonce)
}

func TestCSPReportHandler_ValidReport(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/csp-report", CSPReportHandler())

	// Sample CSP violation report
	reportJSON := `{
		"csp-report": {
			"document-uri": "https://example.com/page",
			"violated-directive": "script-src 'self'",
			"blocked-uri": "https://evil.com/malicious.js",
			"source-file": "https://example.com/page",
			"line-number": 10
		}
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/csp-report", strings.NewReader(reportJSON))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Should return 204 No Content
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestCSPReportHandler_InvalidReport(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/csp-report", CSPReportHandler())

	// Invalid JSON
	reportJSON := `{invalid json`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/csp-report", strings.NewReader(reportJSON))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Should return 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetCSPNonce_Present(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("csp-nonce", "test-nonce-123")

	nonce := GetCSPNonce(c)
	assert.Equal(t, "test-nonce-123", nonce)
}

func TestGetCSPNonce_Missing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	nonce := GetCSPNonce(c)
	assert.Empty(t, nonce)
}

func TestDefaultSecurityHeadersConfig(t *testing.T) {
	config := DefaultSecurityHeadersConfig()

	// Phase 1 headers should be enabled
	assert.True(t, config.EnableXFrameOptions)
	assert.True(t, config.EnableXContentType)
	assert.True(t, config.EnableXXSSProtection)
	assert.True(t, config.EnableReferrerPolicy)
	assert.True(t, config.EnablePermissionsPolicy)

	// Phase 2 HSTS should be disabled by default
	assert.False(t, config.EnableHSTS)

	// Phase 3 CSP should be disabled by default
	assert.False(t, config.EnableCSP)
	assert.True(t, config.CSPReportOnly) // But Report-Only when enabled
}

func TestProductionSecurityHeadersConfig(t *testing.T) {
	config := ProductionSecurityHeadersConfig()

	// All Phase 1 headers should be enabled
	assert.True(t, config.EnableXFrameOptions)
	assert.True(t, config.EnableXContentType)

	// HSTS should be enabled in production
	assert.True(t, config.EnableHSTS)
	assert.True(t, config.HSTSIncludeSubDomains)
	assert.False(t, config.HSTSPreload) // Not preloaded by default

	// CSP should be enabled in production (enforcement mode)
	assert.True(t, config.EnableCSP)
	assert.False(t, config.CSPReportOnly)
}

func TestGenerateCSPNonce(t *testing.T) {
	nonce1 := generateCSPNonce()
	nonce2 := generateCSPNonce()

	// Nonces should be unique
	assert.NotEqual(t, nonce1, nonce2)

	// Nonces should be non-empty
	assert.NotEmpty(t, nonce1)
	assert.NotEmpty(t, nonce2)

	// Nonces should be base64 encoded (no special chars except +/=)
	assert.Regexp(t, "^[A-Za-z0-9+/=]+$", nonce1)
	assert.Regexp(t, "^[A-Za-z0-9+/=]+$", nonce2)
}
