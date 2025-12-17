package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	// Create miniredis server
	mr, err := miniredis.Run()
	require.NoError(t, err)

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return client, mr
}

// Test 1: RateLimitMiddleware - Under Limit
func TestRateLimitMiddleware_UnderLimit(t *testing.T) {
	client, mr := setupRedis(t)
	defer mr.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RateLimitMiddleware(client, 5)) // 5 requests per minute
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// Make 3 requests (under limit of 5)
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:1234"
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Request %d should succeed", i+1)
	}
}

// Test 2: RateLimitMiddleware - Exceeded Limit
func TestRateLimitMiddleware_ExceededLimit(t *testing.T) {
	client, mr := setupRedis(t)
	defer mr.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RateLimitMiddleware(client, 3)) // 3 requests per minute
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// Make 3 requests (at limit)
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:1234"
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "Request %d should succeed", i+1)
	}

	// 4th request should be rate limited
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Body.String(), "Too many requests")
}

// Test 3: RateLimitMiddleware - Different IPs Independent
func TestRateLimitMiddleware_DifferentIPs(t *testing.T) {
	client, mr := setupRedis(t)
	defer mr.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RateLimitMiddleware(client, 2)) // 2 requests per minute
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// IP 1: Make 2 requests (at limit)
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// IP 2: Should still be able to make requests
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.2:5678"
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "Different IP should have independent rate limit")

	// IP 1: 3rd request should be blocked
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

// Test 4: RateLimitMiddleware - Redis Unavailable (Should Allow)
func TestRateLimitMiddleware_RedisUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RateLimitMiddleware(nil, 1)) // nil Redis client
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// Should allow requests when Redis is unavailable
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Should allow requests when Redis is unavailable")
}

// Test 5: RateLimitMiddleware - Key Expiration
func TestRateLimitMiddleware_KeyExpiration(t *testing.T) {
	client, mr := setupRedis(t)
	defer mr.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RateLimitMiddleware(client, 2)) // 2 requests per minute
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// Make 2 requests (at limit)
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:1234"
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// Fast-forward time in miniredis (simulate 1 minute passing)
	mr.FastForward(61 * time.Second) // 61 seconds

	// Should be able to make requests again after expiration
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Should allow requests after key expiration")
}

// Test 6: AuthRateLimitMiddleware - Under Limit
func TestAuthRateLimitMiddleware_UnderLimit(t *testing.T) {
	client, mr := setupRedis(t)
	defer mr.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuthRateLimitMiddleware(client, 3)) // 3 requests per minute
	router.POST("/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// Make 2 requests (under limit)
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/login", nil)
		req.RemoteAddr = "127.0.0.1:1234"
		req.Header.Set("User-Agent", "TestAgent/1.0")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}
}

// Test 7: AuthRateLimitMiddleware - Exceeded Limit with Retry-After
func TestAuthRateLimitMiddleware_ExceededWithRetryAfter(t *testing.T) {
	client, mr := setupRedis(t)
	defer mr.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuthRateLimitMiddleware(client, 2)) // 2 requests per minute
	router.POST("/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// Make 2 requests (at limit)
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/login", nil)
		req.RemoteAddr = "127.0.0.1:1234"
		req.Header.Set("User-Agent", "TestAgent/1.0")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// 3rd request should be rate limited with Retry-After header
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	req.Header.Set("User-Agent", "TestAgent/1.0")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Body.String(), "Too many requests")
	assert.Equal(t, "60", w.Header().Get("Retry-After"), "Should have Retry-After header")
}

// Test 8: AuthRateLimitMiddleware - Different User Agents Independent
func TestAuthRateLimitMiddleware_DifferentUserAgents(t *testing.T) {
	client, mr := setupRedis(t)
	defer mr.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuthRateLimitMiddleware(client, 2)) // 2 requests per minute
	router.POST("/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// User Agent 1: Make 2 requests (at limit)
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/login", nil)
		req.RemoteAddr = "127.0.0.1:1234"
		req.Header.Set("User-Agent", "Browser/1.0")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// User Agent 2: Should still be able to make requests (different key)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	req.Header.Set("User-Agent", "Mobile/2.0")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "Different User-Agent should have independent rate limit")
}

// Test 9: AuthRateLimitMiddleware - Redis Unavailable
func TestAuthRateLimitMiddleware_RedisUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuthRateLimitMiddleware(nil, 1)) // nil Redis client
	router.POST("/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Should allow requests when Redis is unavailable")
}

// Test 10: RateLimitMiddleware - Concurrent Requests
func TestRateLimitMiddleware_ConcurrentRequests(t *testing.T) {
	client, mr := setupRedis(t)
	defer mr.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RateLimitMiddleware(client, 5)) // 5 requests per minute
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// Make 5 concurrent requests
	results := make(chan int, 5)
	for i := 0; i < 5; i++ {
		go func() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "127.0.0.1:1234"
			router.ServeHTTP(w, req)
			results <- w.Code
		}()
	}

	// Collect results
	successCount := 0
	for i := 0; i < 5; i++ {
		code := <-results
		if code == http.StatusOK {
			successCount++
		}
	}

	// All 5 should succeed (at limit)
	assert.Equal(t, 5, successCount, "All requests at limit should succeed")

	// 6th request should be blocked
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

// Test 11: AuthRateLimitMiddleware - Brute Force Scenario
func TestAuthRateLimitMiddleware_BruteForceProtection(t *testing.T) {
	client, mr := setupRedis(t)
	defer mr.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuthRateLimitMiddleware(client, 5)) // 5 login attempts per minute
	router.POST("/login", func(c *gin.Context) {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false})
	})

	// Simulate brute force attack (10 failed login attempts)
	successCount := 0
	blockedCount := 0

	for i := 0; i < 10; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/login", nil)
		req.RemoteAddr = "192.168.1.100:1234"
		req.Header.Set("User-Agent", "AttackBot/1.0")
		router.ServeHTTP(w, req)

		if w.Code == http.StatusUnauthorized {
			successCount++ // Request went through
		} else if w.Code == http.StatusTooManyRequests {
			blockedCount++ // Blocked by rate limit
		}
	}

	// First 5 should go through, next 5 should be blocked
	assert.Equal(t, 5, successCount, "First 5 requests should go through")
	assert.Equal(t, 5, blockedCount, "Next 5 requests should be blocked by rate limit")
}

// Test 12: RateLimitMiddleware - Zero Limit (Block All)
func TestRateLimitMiddleware_ZeroLimit(t *testing.T) {
	client, mr := setupRedis(t)
	defer mr.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RateLimitMiddleware(client, 0)) // 0 requests per minute (block all)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// First request should be blocked (count starts at 0, but 0 >= 0)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code, "Zero limit should block all requests")
}
