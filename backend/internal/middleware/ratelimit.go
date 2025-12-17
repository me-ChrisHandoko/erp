package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"backend/pkg/errors"
)

// RateLimitMiddleware implements rate limiting using Redis
// Reference: BACKEND-IMPLEMENTATION.md lines 1235-1256 (Rate Limiting)
func RateLimitMiddleware(redisClient *redis.Client, requestsPerMinute int) gin.HandlerFunc {
	return func(c *gin.Context) {
		// If Redis is not available, skip rate limiting
		if redisClient == nil {
			c.Next()
			return
		}

		// Get client identifier (IP address)
		clientID := c.ClientIP()
		key := fmt.Sprintf("rate_limit:%s", clientID)

		ctx := context.Background()

		// Get current count
		count, err := redisClient.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			// Redis error, allow request but log error
			// logger.Error("Redis error in rate limit", "error", err)
			c.Next()
			return
		}

		// Check if limit exceeded
		if count >= requestsPerMinute {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   errors.NewRateLimitError(),
			})
			c.Abort()
			return
		}

		// Increment counter
		pipe := redisClient.Pipeline()
		pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, time.Minute)
		_, err = pipe.Exec(ctx)
		if err != nil {
			// Redis error, allow request but log error
			// logger.Error("Redis error in rate limit increment", "error", err)
			c.Next()
			return
		}

		c.Next()
	}
}

// AuthRateLimitMiddleware implements stricter rate limiting for auth endpoints
// Reference: BACKEND-IMPLEMENTATION.md lines 1098-1134 (Brute Force Protection)
func AuthRateLimitMiddleware(redisClient *redis.Client, requestsPerMinute int) gin.HandlerFunc {
	return func(c *gin.Context) {
		// If Redis is not available, skip rate limiting
		if redisClient == nil {
			c.Next()
			return
		}

		// Get client identifier (IP address + user agent for better tracking)
		clientID := fmt.Sprintf("%s:%s", c.ClientIP(), c.Request.UserAgent())
		key := fmt.Sprintf("auth_rate_limit:%s", clientID)

		ctx := context.Background()

		// Get current count
		count, err := redisClient.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			// Redis error, allow request but log error
			c.Next()
			return
		}

		// Check if limit exceeded
		if count >= requestsPerMinute {
			retryAfter := 60 // seconds
			c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   errors.NewRateLimitError(),
			})
			c.Abort()
			return
		}

		// Increment counter
		pipe := redisClient.Pipeline()
		pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, time.Minute)
		_, err = pipe.Exec(ctx)
		if err != nil {
			// Redis error, allow request but log error
			c.Next()
			return
		}

		c.Next()
	}
}
