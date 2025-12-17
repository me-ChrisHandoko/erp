package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"

	"backend/pkg/errors"
)

// CSRFMiddleware implements CSRF protection using double-submit cookie pattern
// Reference: BACKEND-IMPLEMENTATION.md lines 437-579
//
// How it works:
// 1. Server generates a random CSRF token on login/session start
// 2. Token is sent to client in two ways:
//    - As a cookie (NOT httpOnly, so JavaScript can read it)
//    - Client must send this token back in X-CSRF-Token header
// 3. Server validates that cookie value matches header value
//
// This prevents CSRF because:
// - Attacker cannot read the cookie (Same-Origin Policy)
// - Attacker cannot set the X-CSRF-Token header in cross-origin requests
func CSRFMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip CSRF for safe methods (GET, HEAD, OPTIONS)
		// These methods should not modify state
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Get CSRF token from cookie
		cookieToken, err := c.Cookie("csrf_token")
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   errors.NewCSRFError("CSRF token missing in cookie"),
			})
			c.Abort()
			return
		}

		// Get CSRF token from header
		headerToken := c.GetHeader("X-CSRF-Token")
		if headerToken == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   errors.NewCSRFError("CSRF token required in X-CSRF-Token header"),
			})
			c.Abort()
			return
		}

		// Validate tokens match (constant-time comparison to prevent timing attacks)
		if !secureCompare(cookieToken, headerToken) {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   errors.NewCSRFError("CSRF token mismatch"),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GenerateCSRFToken creates a new cryptographically secure CSRF token
// Returns a base64-encoded random string (32 bytes = 44 characters in base64)
func GenerateCSRFToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// SetCSRFCookie sets the CSRF token as a cookie
// Important: This cookie is NOT httpOnly because the frontend needs to read it
// to send in the X-CSRF-Token header
func SetCSRFCookie(c *gin.Context, token string, secure bool) {
	c.SetCookie(
		"csrf_token",       // name
		token,              // value
		24*60*60,           // maxAge (24 hours in seconds)
		"/",                // path
		"",                 // domain (empty = current domain)
		secure,             // secure (HTTPS only in production)
		false,              // httpOnly = FALSE (frontend needs to read it!)
	)

	// Set SameSite=Strict for additional CSRF protection
	c.SetSameSite(http.SameSiteStrictMode)
}

// ClearCSRFCookie clears the CSRF token cookie
func ClearCSRFCookie(c *gin.Context, secure bool) {
	c.SetCookie(
		"csrf_token",
		"",
		-1, // maxAge -1 means delete
		"/",
		"",
		secure,
		false,
	)
}

// secureCompare performs constant-time comparison of two strings
// This prevents timing attacks where an attacker could measure
// how long the comparison takes to guess the token character by character
func secureCompare(a, b string) bool {
	// If lengths differ, comparison will fail anyway
	if len(a) != len(b) {
		return false
	}

	// XOR all bytes and accumulate result
	// This ensures comparison time is constant regardless of where strings differ
	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}

	return result == 0
}
