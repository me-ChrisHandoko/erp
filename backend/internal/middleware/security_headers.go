package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	"backend/internal/config"
)

// SecurityHeadersConfig holds configuration for security headers
type SecurityHeadersConfig struct {
	// Enable/disable specific headers
	EnableHSTS              bool
	EnableCSP               bool
	EnableXFrameOptions     bool
	EnableXContentType      bool
	EnableXXSSProtection    bool
	EnableReferrerPolicy    bool
	EnablePermissionsPolicy bool

	// HSTS Configuration
	HSTSMaxAge            int
	HSTSIncludeSubDomains bool
	HSTSPreload           bool

	// CSP Configuration
	CSPDirectives      map[string][]string
	CSPReportOnly      bool
	CSPReportURI       string
	CSPUseNonce        bool
	CSPReportViolation bool

	// Frame Options
	XFrameOptions string // DENY, SAMEORIGIN

	// Referrer Policy
	ReferrerPolicy string

	// Permissions Policy
	PermissionsPolicy string
}

// DefaultSecurityHeadersConfig returns default security headers configuration
// Phase 1: Safe defaults with minimal breaking changes
func DefaultSecurityHeadersConfig() *SecurityHeadersConfig {
	return &SecurityHeadersConfig{
		// Phase 1: Enable safe headers immediately
		EnableXFrameOptions:     true,
		EnableXContentType:      true,
		EnableXXSSProtection:    true,
		EnableReferrerPolicy:    true,
		EnablePermissionsPolicy: true,

		// Phase 2: HSTS (requires SSL setup)
		EnableHSTS:            false, // Enable manually after SSL setup
		HSTSMaxAge:            31536000, // 1 year
		HSTSIncludeSubDomains: false,   // Enable after testing
		HSTSPreload:           false,   // Enable after HSTS stable

		// Phase 3: CSP (requires extensive testing)
		EnableCSP:          false, // Start with Report-Only mode
		CSPReportOnly:      true,  // Monitor violations first
		CSPUseNonce:        true,
		CSPReportViolation: true,
		CSPReportURI:       "/api/v1/csp-report", // CSP violation reporting endpoint
		CSPDirectives:      getDefaultCSPDirectives(),

		// Configuration values
		XFrameOptions:     "DENY",
		ReferrerPolicy:    "strict-origin-when-cross-origin",
		PermissionsPolicy: "camera=(), microphone=(), geolocation=(), payment=()",
	}
}

// ProductionSecurityHeadersConfig returns production-ready configuration
// Use this after all phases tested and validated
func ProductionSecurityHeadersConfig() *SecurityHeadersConfig {
	config := DefaultSecurityHeadersConfig()

	// Production: Enable all security features
	config.EnableHSTS = true
	config.HSTSIncludeSubDomains = true
	config.HSTSPreload = false // Set true only after submitting to preload list

	config.EnableCSP = true
	config.CSPReportOnly = false // Enforce in production after testing

	return config
}

// getDefaultCSPDirectives returns default CSP directives
func getDefaultCSPDirectives() map[string][]string {
	return map[string][]string{
		// Default policy: only allow resources from same origin
		"default-src": {"'self'"},

		// Scripts: self + nonce (no unsafe-inline, no unsafe-eval)
		"script-src": {"'self'", "'strict-dynamic'"},

		// Styles: self + nonce (allow unsafe-inline for compatibility)
		"style-src": {"'self'", "'unsafe-inline'"}, // TODO: Remove unsafe-inline after audit

		// Images: self, data URIs, HTTPS
		"img-src": {"'self'", "data:", "https:"},

		// Fonts: self, data URIs
		"font-src": {"'self'", "data:"},

		// API connections: self + localhost ports for development
		// NOTE: In production, add your API domain (e.g., https://api.yourdomain.com)
		"connect-src": {
			"'self'",
			"http://localhost:8080",  // Backend API (HTTP - development)
			"https://localhost:8080", // Backend API (HTTPS - Phase 2)
			"ws://localhost:8080",    // WebSocket (HTTP - development)
			"wss://localhost:8080",   // WebSocket (HTTPS - Phase 2)
		},

		// Frames: none (prevent clickjacking)
		"frame-ancestors": {"'none'"},

		// Form submissions: self only
		"form-action": {"'self'"},

		// Base URI: self only
		"base-uri": {"'self'"},

		// Object/embed: none
		"object-src": {"'none'"},

		// Upgrade HTTP to HTTPS
		"upgrade-insecure-requests": {},
	}
}

// SecurityHeadersMiddleware adds security headers to HTTP responses
// Reference: OWASP Secure Headers Project
// https://owasp.org/www-project-secure-headers/
func SecurityHeadersMiddleware(cfg *config.Config) gin.HandlerFunc {
	securityConfig := DefaultSecurityHeadersConfig()

	// Use production config in production environment
	if cfg.IsProduction() {
		securityConfig = ProductionSecurityHeadersConfig()
	}

	// Override with user config if available
	if cfg.Security.EnableHSTS {
		securityConfig.EnableHSTS = true
		securityConfig.HSTSMaxAge = cfg.Security.HSTSMaxAge
		securityConfig.HSTSIncludeSubDomains = cfg.Security.HSTSIncludeSubDomains
	}

	if cfg.Security.EnableCSP {
		securityConfig.EnableCSP = true
		securityConfig.CSPReportOnly = cfg.Security.CSPReportOnly
	}

	return func(c *gin.Context) {
		// ========================================================================
		// PHASE 1: SAFE HEADERS (No Breaking Changes)
		// Deploy immediately for instant security improvement
		// ========================================================================

		// X-Frame-Options: Prevent clickjacking
		if securityConfig.EnableXFrameOptions {
			c.Header("X-Frame-Options", securityConfig.XFrameOptions)
		}

		// X-Content-Type-Options: Prevent MIME-type sniffing
		if securityConfig.EnableXContentType {
			c.Header("X-Content-Type-Options", "nosniff")
		}

		// X-XSS-Protection: Enable browser XSS filter (legacy browsers)
		if securityConfig.EnableXXSSProtection {
			c.Header("X-XSS-Protection", "1; mode=block")
		}

		// Referrer-Policy: Control referer information leakage
		if securityConfig.EnableReferrerPolicy {
			c.Header("Referrer-Policy", securityConfig.ReferrerPolicy)
		}

		// Permissions-Policy: Disable unnecessary browser features
		if securityConfig.EnablePermissionsPolicy {
			c.Header("Permissions-Policy", securityConfig.PermissionsPolicy)
		}

		// ========================================================================
		// PHASE 2: HSTS (Requires SSL Certificate)
		// Enable after SSL setup and testing
		// ========================================================================

		if securityConfig.EnableHSTS {
			hstsValue := fmt.Sprintf("max-age=%d", securityConfig.HSTSMaxAge)

			if securityConfig.HSTSIncludeSubDomains {
				hstsValue += "; includeSubDomains"
			}

			if securityConfig.HSTSPreload {
				hstsValue += "; preload"
			}

			c.Header("Strict-Transport-Security", hstsValue)
		}

		// ========================================================================
		// PHASE 3: CONTENT SECURITY POLICY (Requires Extensive Testing)
		// Start with Report-Only mode, switch to enforcement after validation
		// ========================================================================

		if securityConfig.EnableCSP {
			// Generate nonce for inline scripts/styles if enabled
			var nonce string
			if securityConfig.CSPUseNonce {
				nonce = generateCSPNonce()
				c.Set("csp-nonce", nonce) // Store in context for use in templates
			}

			// Build CSP header value
			cspValue := buildCSPHeader(securityConfig, nonce)

			// Add CSP reporting if enabled
			if securityConfig.CSPReportViolation && securityConfig.CSPReportURI != "" {
				cspValue += fmt.Sprintf("; report-uri %s", securityConfig.CSPReportURI)
			}

			// Use Report-Only mode for testing, enforcement mode for production
			if securityConfig.CSPReportOnly {
				c.Header("Content-Security-Policy-Report-Only", cspValue)
			} else {
				c.Header("Content-Security-Policy", cspValue)
			}

			// Expose nonce to client (if needed for dynamic script injection)
			if nonce != "" {
				c.Header("X-CSP-Nonce", nonce)
			}
		}

		// ========================================================================
		// ADDITIONAL SECURITY HEADERS
		// ========================================================================

		// X-Permitted-Cross-Domain-Policies: Restrict cross-domain access
		c.Header("X-Permitted-Cross-Domain-Policies", "none")

		// X-Download-Options: Prevent file download execution in IE
		c.Header("X-Download-Options", "noopen")

		// Cross-Origin-Embedder-Policy: Isolate cross-origin resources
		// Note: May break third-party embeds, test before enabling
		// c.Header("Cross-Origin-Embedder-Policy", "require-corp")

		// Cross-Origin-Opener-Policy: Isolate browsing context
		c.Header("Cross-Origin-Opener-Policy", "same-origin")

		// Cross-Origin-Resource-Policy: Control resource loading
		c.Header("Cross-Origin-Resource-Policy", "same-origin")

		c.Next()
	}
}

// buildCSPHeader builds CSP header value from directives
func buildCSPHeader(config *SecurityHeadersConfig, nonce string) string {
	var policies []string

	for directive, values := range config.CSPDirectives {
		// Skip empty directives
		if len(values) == 0 && directive != "upgrade-insecure-requests" {
			continue
		}

		// Special handling for upgrade-insecure-requests (no value)
		if directive == "upgrade-insecure-requests" {
			policies = append(policies, directive)
			continue
		}

		// Add nonce to script-src and style-src if enabled
		if nonce != "" && (directive == "script-src" || directive == "style-src") {
			values = append(values, fmt.Sprintf("'nonce-%s'", nonce))
		}

		policy := fmt.Sprintf("%s %s", directive, strings.Join(values, " "))
		policies = append(policies, policy)
	}

	return strings.Join(policies, "; ")
}

// generateCSPNonce generates a cryptographically secure nonce for CSP
func generateCSPNonce() string {
	// Generate 16 random bytes
	nonceBytes := make([]byte, 16)
	_, err := rand.Read(nonceBytes)
	if err != nil {
		// Fallback to timestamp-based nonce (not cryptographically secure)
		// This should never happen in practice
		return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", gin.Mode())))
	}

	// Encode to base64
	return base64.StdEncoding.EncodeToString(nonceBytes)
}

// CSPReportHandler handles CSP violation reports
// POST /api/v1/csp-report
func CSPReportHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var report map[string]interface{}

		if err := c.ShouldBindJSON(&report); err != nil {
			c.JSON(400, gin.H{"error": "Invalid CSP report"})
			return
		}

		// Log CSP violation for monitoring
		// TODO: Send to monitoring system (Sentry, DataDog, etc.)
		fmt.Printf("🚨 CSP Violation Report:\n%+v\n", report)

		// Extract key information
		if cspReport, ok := report["csp-report"].(map[string]interface{}); ok {
			fmt.Printf("  Blocked URI: %v\n", cspReport["blocked-uri"])
			fmt.Printf("  Violated Directive: %v\n", cspReport["violated-directive"])
			fmt.Printf("  Document URI: %v\n", cspReport["document-uri"])
			fmt.Printf("  Source File: %v\n", cspReport["source-file"])
			fmt.Printf("  Line Number: %v\n", cspReport["line-number"])
		}

		// Return 204 No Content (standard for CSP reporting)
		c.Status(204)
	}
}

// GetCSPNonce retrieves CSP nonce from context
// Use this in templates to apply nonce to inline scripts/styles
func GetCSPNonce(c *gin.Context) string {
	if nonce, exists := c.Get("csp-nonce"); exists {
		if nonceStr, ok := nonce.(string); ok {
			return nonceStr
		}
	}
	return ""
}
