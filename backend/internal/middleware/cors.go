package middleware

import (
	"github.com/gin-gonic/gin"

	"backend/internal/config"
)

// CORSMiddleware handles Cross-Origin Resource Sharing
// Reference: BACKEND-IMPLEMENTATION.md lines 1257-1277 (CORS Configuration)
func CORSMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range cfg.CORS.AllowedOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				allowed = true
				c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
				break
			}
		}

		if !allowed && len(cfg.CORS.AllowedOrigins) > 0 {
			// Origin not allowed
			c.AbortWithStatus(403)
			return
		}

		// Set CORS headers
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		// Use configured allowed methods
		allowedMethods := "GET, POST, PUT, DELETE, PATCH, OPTIONS"
		if len(cfg.CORS.AllowedMethods) > 0 {
			allowedMethods = ""
			for i, method := range cfg.CORS.AllowedMethods {
				if i > 0 {
					allowedMethods += ","
				}
				allowedMethods += method
			}
		}
		c.Writer.Header().Set("Access-Control-Allow-Methods", allowedMethods)

		// Use configured allowed headers
		allowedHeaders := "Content-Type,Authorization,X-Requested-With,X-CSRF-Token,X-Company-ID"
		if len(cfg.CORS.AllowedHeaders) > 0 {
			allowedHeaders = ""
			for i, header := range cfg.CORS.AllowedHeaders {
				if i > 0 {
					allowedHeaders += ","
				}
				allowedHeaders += header
			}
		}
		c.Writer.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
		c.Writer.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
