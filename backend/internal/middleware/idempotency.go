package middleware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"backend/pkg/errors"
)

// IdempotencyResponse represents the cached response structure
type IdempotencyResponse struct {
	StatusCode int               `json:"status_code"`
	Body       string            `json:"body"`
	Headers    map[string]string `json:"headers"`
}

// IdempotencyMiddleware implements idempotency for POST operations using Idempotency-Key header
// Reference: ANALYSIS-01-TENANT-COMPANY-SETUP.md Issue #9: Idempotency Support Missing
// Prevents duplicate record creation from duplicate requests
func IdempotencyMiddleware(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only apply to POST, PUT, PATCH methods (state-changing operations)
		if c.Request.Method != http.MethodPost &&
		   c.Request.Method != http.MethodPut &&
		   c.Request.Method != http.MethodPatch {
			c.Next()
			return
		}

		// If Redis is not available, skip idempotency check
		if redisClient == nil {
			c.Next()
			return
		}

		// Get Idempotency-Key header
		idempotencyKey := c.GetHeader("Idempotency-Key")
		if idempotencyKey == "" {
			// No idempotency key provided - process normally
			c.Next()
			return
		}

		// Validate idempotency key format (UUID or similar)
		if len(idempotencyKey) < 16 || len(idempotencyKey) > 128 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   errors.NewBadRequestError("Idempotency-Key must be between 16 and 128 characters"),
			})
			c.Abort()
			return
		}

		ctx := context.Background()

		// Generate unique Redis key combining idempotency key + request hash
		requestHash, err := generateRequestHash(c)
		if err != nil {
			// Failed to hash request - allow through but log error
			c.Next()
			return
		}

		redisKey := fmt.Sprintf("idempotency:%s:%s", idempotencyKey, requestHash)

		// Check if we have a cached response for this idempotency key
		cachedResponse, err := redisClient.Get(ctx, redisKey).Result()
		if err == nil {
			// Found cached response - return it
			var response IdempotencyResponse
			if err := json.Unmarshal([]byte(cachedResponse), &response); err == nil {
				// Set cached headers
				for key, value := range response.Headers {
					c.Header(key, value)
				}
				// Return cached response
				c.Data(response.StatusCode, "application/json", []byte(response.Body))
				c.Abort()
				return
			}
		} else if err != redis.Nil {
			// Redis error (not just cache miss) - allow through but log error
			c.Next()
			return
		}

		// No cached response - process request and cache the response
		// Use custom writer to capture response
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
			headers:        make(map[string]string),
		}
		c.Writer = writer

		// Process the request
		c.Next()

		// Cache the response in Redis with 24h TTL
		response := IdempotencyResponse{
			StatusCode: writer.Status(),
			Body:       writer.body.String(),
			Headers:    writer.headers,
		}

		responseJSON, err := json.Marshal(response)
		if err == nil {
			// Store in Redis with 24 hour TTL
			redisClient.Set(ctx, redisKey, responseJSON, 24*time.Hour)
		}
	}
}

// responseWriter wraps gin.ResponseWriter to capture response body and headers
type responseWriter struct {
	gin.ResponseWriter
	body    *bytes.Buffer
	headers map[string]string
}

// Write captures the response body
func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// WriteHeader captures response headers
func (w *responseWriter) WriteHeader(statusCode int) {
	// Capture important headers
	for key, values := range w.ResponseWriter.Header() {
		if len(values) > 0 {
			w.headers[key] = values[0]
		}
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

// generateRequestHash creates a hash of the request (method + path + body)
func generateRequestHash(c *gin.Context) (string, error) {
	// Read request body
	var bodyBytes []byte
	if c.Request.Body != nil {
		bodyBytes, _ = io.ReadAll(c.Request.Body)
		// Restore request body for further processing
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	// Create hash from method + path + body
	hasher := sha256.New()
	hasher.Write([]byte(c.Request.Method))
	hasher.Write([]byte(c.Request.URL.Path))
	hasher.Write(bodyBytes)

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
