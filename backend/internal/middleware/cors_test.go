package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"backend/internal/config"
)

// Test 1: CORSMiddleware - Allowed Origin
func TestCORSMiddleware_AllowedOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		CORS: config.CORSConfig{
			AllowedOrigins: []string{"http://localhost:3000", "https://example.com"},
		},
	}

	router := gin.New()
	router.Use(CORSMiddleware(cfg))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, "GET, POST, PUT, DELETE, PATCH, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Content-Type, Authorization, X-Requested-With", w.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "86400", w.Header().Get("Access-Control-Max-Age"))
}

// Test 2: CORSMiddleware - Wildcard Origin
func TestCORSMiddleware_WildcardOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		CORS: config.CORSConfig{
			AllowedOrigins: []string{"*"},
		},
	}

	router := gin.New()
	router.Use(CORSMiddleware(cfg))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	testOrigins := []string{
		"http://localhost:3000",
		"https://example.com",
		"https://subdomain.example.com",
		"http://any-origin.com",
	}

	for _, origin := range testOrigins {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", origin)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Origin %s should be allowed with wildcard", origin)
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

// Test 3: CORSMiddleware - Blocked Origin
func TestCORSMiddleware_BlockedOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		CORS: config.CORSConfig{
			AllowedOrigins: []string{"http://localhost:3000"},
		},
	}

	router := gin.New()
	router.Use(CORSMiddleware(cfg))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://evil-site.com") // Not in allowed list
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

// Test 4: CORSMiddleware - No Origin Header
func TestCORSMiddleware_NoOriginHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		CORS: config.CORSConfig{
			AllowedOrigins: []string{"http://localhost:3000"},
		},
	}

	router := gin.New()
	router.Use(CORSMiddleware(cfg))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	// No Origin header set
	router.ServeHTTP(w, req)

	// Should be blocked if origins list is not empty
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// Test 5: CORSMiddleware - Empty Allowed Origins (Allow All)
func TestCORSMiddleware_EmptyAllowedOrigins(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		CORS: config.CORSConfig{
			AllowedOrigins: []string{}, // Empty list
		},
	}

	router := gin.New()
	router.Use(CORSMiddleware(cfg))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://any-origin.com")
	router.ServeHTTP(w, req)

	// Empty list allows all origins
	assert.Equal(t, http.StatusOK, w.Code)
}

// Test 6: CORSMiddleware - Preflight Request (OPTIONS)
func TestCORSMiddleware_PreflightRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		CORS: config.CORSConfig{
			AllowedOrigins: []string{"http://localhost:3000"},
		},
	}

	router := gin.New()
	router.Use(CORSMiddleware(cfg))
	router.POST("/api/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/api/data", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, "GET, POST, PUT, DELETE, PATCH, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Content-Type, Authorization, X-Requested-With", w.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "86400", w.Header().Get("Access-Control-Max-Age"))
}

// Test 7: CORSMiddleware - Multiple Allowed Origins
func TestCORSMiddleware_MultipleAllowedOrigins(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		CORS: config.CORSConfig{
			AllowedOrigins: []string{
				"http://localhost:3000",
				"http://localhost:8080",
				"https://app.example.com",
			},
		},
	}

	router := gin.New()
	router.Use(CORSMiddleware(cfg))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// Test each allowed origin
	allowedOrigins := []string{
		"http://localhost:3000",
		"http://localhost:8080",
		"https://app.example.com",
	}

	for _, origin := range allowedOrigins {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", origin)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Origin %s should be allowed", origin)
		assert.Equal(t, origin, w.Header().Get("Access-Control-Allow-Origin"))
	}

	// Test blocked origin
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://blocked-origin.com")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// Test 8: CORSMiddleware - Preflight Blocked Origin
func TestCORSMiddleware_PreflightBlockedOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		CORS: config.CORSConfig{
			AllowedOrigins: []string{"http://localhost:3000"},
		},
	}

	router := gin.New()
	router.Use(CORSMiddleware(cfg))
	router.POST("/api/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/api/data", nil)
	req.Header.Set("Origin", "http://evil-site.com")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

// Test 9: CORSMiddleware - POST Request with CORS
func TestCORSMiddleware_POSTRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		CORS: config.CORSConfig{
			AllowedOrigins: []string{"https://frontend.example.com"},
		},
	}

	router := gin.New()
	router.Use(CORSMiddleware(cfg))
	router.POST("/api/create", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"id": "123"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/create", nil)
	req.Header.Set("Origin", "https://frontend.example.com")
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "https://frontend.example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
}

// Test 10: CORSMiddleware - Case Sensitive Origin Matching
func TestCORSMiddleware_CaseSensitiveOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		CORS: config.CORSConfig{
			AllowedOrigins: []string{"http://localhost:3000"},
		},
	}

	router := gin.New()
	router.Use(CORSMiddleware(cfg))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// Origins are case-sensitive - this should be blocked
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://LOCALHOST:3000") // Different case
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// Test 11: CORSMiddleware - All HTTP Methods
func TestCORSMiddleware_AllHTTPMethods(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		CORS: config.CORSConfig{
			AllowedOrigins: []string{"http://localhost:3000"},
		},
	}

	router := gin.New()
	router.Use(CORSMiddleware(cfg))

	handler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"method": c.Request.Method})
	}

	router.GET("/test", handler)
	router.POST("/test", handler)
	router.PUT("/test", handler)
	router.DELETE("/test", handler)
	router.PATCH("/test", handler)

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(method, "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Method %s should be allowed", method)
		assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

// Test 12: CORSMiddleware - Max Age Header
func TestCORSMiddleware_MaxAgeHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		CORS: config.CORSConfig{
			AllowedOrigins: []string{"http://localhost:3000"},
		},
	}

	router := gin.New()
	router.Use(CORSMiddleware(cfg))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	router.ServeHTTP(w, req)

	assert.Equal(t, "86400", w.Header().Get("Access-Control-Max-Age"), "Max age should be 24 hours (86400 seconds)")
}

// Test 13: CORSMiddleware - Credentials Header Always Set
func TestCORSMiddleware_CredentialsHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		CORS: config.CORSConfig{
			AllowedOrigins: []string{"http://localhost:3000"},
		},
	}

	router := gin.New()
	router.Use(CORSMiddleware(cfg))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	router.ServeHTTP(w, req)

	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"), "Credentials should always be true")
}

// Test 14: CORSMiddleware - Allowed Headers
func TestCORSMiddleware_AllowedHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		CORS: config.CORSConfig{
			AllowedOrigins: []string{"*"},
		},
	}

	router := gin.New()
	router.Use(CORSMiddleware(cfg))
	router.POST("/api", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/api", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	router.ServeHTTP(w, req)

	allowedHeaders := w.Header().Get("Access-Control-Allow-Headers")
	assert.Contains(t, allowedHeaders, "Content-Type")
	assert.Contains(t, allowedHeaders, "Authorization")
	assert.Contains(t, allowedHeaders, "X-Requested-With")
}

// Test 15: CORSMiddleware - Wildcard with Preflight
func TestCORSMiddleware_WildcardPreflight(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		CORS: config.CORSConfig{
			AllowedOrigins: []string{"*"},
		},
	}

	router := gin.New()
	router.Use(CORSMiddleware(cfg))
	router.POST("/api/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": "test"})
	})

	// Preflight request with any origin
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/api/data", nil)
	req.Header.Set("Origin", "http://any-random-origin.com")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}
