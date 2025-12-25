package middleware

import (
	"bytes"
	"encoding/json"
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

func setupIdempotencyTestRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	// Create miniredis server
	mr, err := miniredis.Run()
	require.NoError(t, err)

	// Create Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return redisClient, mr
}

// Test Issue #9 Fix: Idempotency middleware returns cached response for duplicate requests
func TestIdempotencyMiddleware_DuplicateRequestReturnsCached(t *testing.T) {
	redisClient, mr := setupIdempotencyTestRedis(t)
	defer mr.Close()

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Counter to track how many times handler is called
	handlerCallCount := 0

	router.Use(IdempotencyMiddleware(redisClient))
	router.POST("/api/users", func(c *gin.Context) {
		handlerCallCount++
		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"data": gin.H{
				"id":   "user-123",
				"name": "John Doe",
			},
		})
	})

	idempotencyKey := "test-idempotency-key-12345"
	requestBody := []byte(`{"name": "John Doe", "email": "john@example.com"}`)

	// First request - should execute handler
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(requestBody))
	req1.Header.Set("Idempotency-Key", idempotencyKey)
	req1.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusCreated, w1.Code)
	assert.Equal(t, 1, handlerCallCount, "Handler should be called once")

	var response1 map[string]interface{}
	json.Unmarshal(w1.Body.Bytes(), &response1)
	assert.True(t, response1["success"].(bool))

	// Second request with same idempotency key - should return cached response
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(requestBody))
	req2.Header.Set("Idempotency-Key", idempotencyKey)
	req2.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusCreated, w2.Code)
	assert.Equal(t, 1, handlerCallCount, "Handler should still only be called once (cached response)")

	var response2 map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &response2)
	assert.True(t, response2["success"].(bool))
	assert.Equal(t, response1, response2, "Second response should be identical to first")
}

// Test Issue #9 Fix: Different idempotency keys process independently
func TestIdempotencyMiddleware_DifferentKeysProcessIndependently(t *testing.T) {
	redisClient, mr := setupIdempotencyTestRedis(t)
	defer mr.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()

	handlerCallCount := 0

	router.Use(IdempotencyMiddleware(redisClient))
	router.POST("/api/users", func(c *gin.Context) {
		handlerCallCount++
		c.JSON(http.StatusCreated, gin.H{"success": true, "call": handlerCallCount})
	})

	requestBody := []byte(`{"name": "Test"}`)

	// Request 1 with key A (min 16 chars)
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(requestBody))
	req1.Header.Set("Idempotency-Key", "test-key-A-1234567890")
	router.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusCreated, w1.Code)
	assert.Equal(t, 1, handlerCallCount)

	// Request 2 with key B - should execute handler
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(requestBody))
	req2.Header.Set("Idempotency-Key", "test-key-B-1234567890")
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusCreated, w2.Code)
	assert.Equal(t, 2, handlerCallCount, "Different idempotency key should execute handler")
}

// Test Issue #9 Fix: Different request bodies with same key are treated as different requests
func TestIdempotencyMiddleware_DifferentBodiesDifferentRequests(t *testing.T) {
	redisClient, mr := setupIdempotencyTestRedis(t)
	defer mr.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()

	handlerCallCount := 0

	router.Use(IdempotencyMiddleware(redisClient))
	router.POST("/api/users", func(c *gin.Context) {
		handlerCallCount++
		c.JSON(http.StatusCreated, gin.H{"success": true, "call": handlerCallCount})
	})

	idempotencyKey := "same-key-1234567890"

	// Request 1 with body A
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/api/users", bytes.NewBufferString(`{"name": "Alice"}`))
	req1.Header.Set("Idempotency-Key", idempotencyKey)
	router.ServeHTTP(w1, req1)

	assert.Equal(t, 1, handlerCallCount)

	// Request 2 with body B - same key but different body
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/api/users", bytes.NewBufferString(`{"name": "Bob"}`))
	req2.Header.Set("Idempotency-Key", idempotencyKey)
	router.ServeHTTP(w2, req2)

	assert.Equal(t, 2, handlerCallCount, "Different body should be treated as different request")
}

// Test Issue #9 Fix: No idempotency key processes normally
func TestIdempotencyMiddleware_NoIdempotencyKeyProcessesNormally(t *testing.T) {
	redisClient, mr := setupIdempotencyTestRedis(t)
	defer mr.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()

	handlerCallCount := 0

	router.Use(IdempotencyMiddleware(redisClient))
	router.POST("/api/users", func(c *gin.Context) {
		handlerCallCount++
		c.JSON(http.StatusCreated, gin.H{"success": true, "call": handlerCallCount})
	})

	requestBody := []byte(`{"name": "Test"}`)

	// Request 1 without idempotency key
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(requestBody))
	router.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusCreated, w1.Code)
	assert.Equal(t, 1, handlerCallCount)

	// Request 2 without idempotency key - should execute handler again
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(requestBody))
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusCreated, w2.Code)
	assert.Equal(t, 2, handlerCallCount, "Without idempotency key, should execute every time")
}

// Test Issue #9 Fix: GET requests bypass idempotency (no caching)
func TestIdempotencyMiddleware_GETRequestsBypass(t *testing.T) {
	redisClient, mr := setupIdempotencyTestRedis(t)
	defer mr.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()

	handlerCallCount := 0

	router.Use(IdempotencyMiddleware(redisClient))
	router.GET("/api/users", func(c *gin.Context) {
		handlerCallCount++
		c.JSON(http.StatusOK, gin.H{"success": true, "call": handlerCallCount})
	})

	// Request 1 with idempotency key (should be ignored for GET)
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/api/users", nil)
	req1.Header.Set("Idempotency-Key", "test-key")
	router.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusOK, w1.Code)
	assert.Equal(t, 1, handlerCallCount)

	// Request 2 with same key - should execute again (GET bypasses idempotency)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/api/users", nil)
	req2.Header.Set("Idempotency-Key", "test-key")
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Equal(t, 2, handlerCallCount, "GET requests should bypass idempotency")
}

// Test Issue #9 Fix: Invalid idempotency key returns 400 Bad Request
func TestIdempotencyMiddleware_InvalidKeyReturns400(t *testing.T) {
	redisClient, mr := setupIdempotencyTestRedis(t)
	defer mr.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(IdempotencyMiddleware(redisClient))
	router.POST("/api/users", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"success": true})
	})

	// Test: Idempotency key too short
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/api/users", bytes.NewBufferString(`{}`))
	req1.Header.Set("Idempotency-Key", "short")
	router.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusBadRequest, w1.Code)

	// Test: Idempotency key too long
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/api/users", bytes.NewBufferString(`{}`))
	longKey := string(make([]byte, 129)) // 129 characters
	req2.Header.Set("Idempotency-Key", longKey)
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusBadRequest, w2.Code)
}

// Test Issue #9 Fix: Redis unavailability gracefully allows requests
func TestIdempotencyMiddleware_RedisUnavailableAllowsRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	handlerCallCount := 0

	// Use nil Redis client
	router.Use(IdempotencyMiddleware(nil))
	router.POST("/api/users", func(c *gin.Context) {
		handlerCallCount++
		c.JSON(http.StatusCreated, gin.H{"success": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/users", bytes.NewBufferString(`{}`))
	req.Header.Set("Idempotency-Key", "test-key-12345")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, 1, handlerCallCount, "Should allow request when Redis unavailable")
}

// Test Issue #9 Fix: Cached responses expire after 24 hours
func TestIdempotencyMiddleware_CacheExpiresAfter24Hours(t *testing.T) {
	redisClient, mr := setupIdempotencyTestRedis(t)
	defer mr.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()

	handlerCallCount := 0

	router.Use(IdempotencyMiddleware(redisClient))
	router.POST("/api/users", func(c *gin.Context) {
		handlerCallCount++
		c.JSON(http.StatusCreated, gin.H{"success": true, "call": handlerCallCount})
	})

	idempotencyKey := "expiry-test-key-1234567"
	requestBody := []byte(`{"test": "data"}`)

	// First request
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(requestBody))
	req1.Header.Set("Idempotency-Key", idempotencyKey)
	router.ServeHTTP(w1, req1)

	assert.Equal(t, 1, handlerCallCount)

	// Fast-forward time by 25 hours
	mr.FastForward(25 * time.Hour)

	// Request after expiry should execute handler again
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(requestBody))
	req2.Header.Set("Idempotency-Key", idempotencyKey)
	router.ServeHTTP(w2, req2)

	assert.Equal(t, 2, handlerCallCount, "Should execute handler after cache expiry")
}

// Test Issue #9 Fix: PUT and PATCH methods also support idempotency
func TestIdempotencyMiddleware_PUTAndPATCHSupported(t *testing.T) {
	redisClient, mr := setupIdempotencyTestRedis(t)
	defer mr.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()

	putCallCount := 0
	patchCallCount := 0

	router.Use(IdempotencyMiddleware(redisClient))
	router.PUT("/api/users/:id", func(c *gin.Context) {
		putCallCount++
		c.JSON(http.StatusOK, gin.H{"success": true})
	})
	router.PATCH("/api/users/:id", func(c *gin.Context) {
		patchCallCount++
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// Test PUT with idempotency
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("PUT", "/api/users/123", bytes.NewBufferString(`{"name": "Updated"}`))
	req1.Header.Set("Idempotency-Key", "put-key-1234567890")
	router.ServeHTTP(w1, req1)

	assert.Equal(t, 1, putCallCount)

	// Duplicate PUT request
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("PUT", "/api/users/123", bytes.NewBufferString(`{"name": "Updated"}`))
	req2.Header.Set("Idempotency-Key", "put-key-1234567890")
	router.ServeHTTP(w2, req2)

	assert.Equal(t, 1, putCallCount, "PUT should use cached response")

	// Test PATCH with idempotency
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("PATCH", "/api/users/123", bytes.NewBufferString(`{"status": "active"}`))
	req3.Header.Set("Idempotency-Key", "patch-key-1234567890")
	router.ServeHTTP(w3, req3)

	assert.Equal(t, 1, patchCallCount)

	// Duplicate PATCH request
	w4 := httptest.NewRecorder()
	req4, _ := http.NewRequest("PATCH", "/api/users/123", bytes.NewBufferString(`{"status": "active"}`))
	req4.Header.Set("Idempotency-Key", "patch-key-1234567890")
	router.ServeHTTP(w4, req4)

	assert.Equal(t, 1, patchCallCount, "PATCH should use cached response")
}
