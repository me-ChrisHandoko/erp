package jwt

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"backend/internal/config"
)

// TokenService handles JWT token generation and validation
// Supports both HS256 (symmetric) and RS256 (asymmetric) algorithms
// Reference: BACKEND-IMPLEMENTATION.md lines 73-84, 795-920
type TokenService struct {
	config     config.JWTConfig
	privateKey *rsa.PrivateKey // For RS256
	publicKey  *rsa.PublicKey  // For RS256
}

// Claims represents JWT claims for authentication
type Claims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	TenantID string `json:"tenant_id"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// NewTokenService creates a new JWT token service
func NewTokenService(cfg config.JWTConfig) (*TokenService, error) {
	service := &TokenService{
		config: cfg,
	}

	// Load RSA keys if using RS256
	if cfg.Algorithm == "RS256" {
		if err := service.loadRSAKeys(); err != nil {
			return nil, fmt.Errorf("failed to load RSA keys: %w", err)
		}
	}

	return service, nil
}

// GenerateAccessToken generates a new access token
// Expiry is configured via JWT.Expiry (default: 30 minutes)
func (s *TokenService) GenerateAccessToken(userID, email, tenantID, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Email:    email,
		TenantID: tenantID,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.config.Expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	return s.generateToken(claims)
}

// GenerateRefreshToken generates a new refresh token
// Expiry is configured via JWT.RefreshExpiry (default: 30 days)
// Refresh tokens have minimal claims for security
func (s *TokenService) GenerateRefreshToken(userID string) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(now.Add(s.config.RefreshExpiry)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
	}

	var token *jwt.Token
	if s.config.Algorithm == "RS256" {
		token = jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		return token.SignedString(s.privateKey)
	}

	// HS256
	token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.Secret))
}

// ValidateAccessToken validates an access token and returns claims
func (s *TokenService) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, s.keyFunc)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token and returns user ID
func (s *TokenService) ValidateRefreshToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, s.keyFunc)
	if err != nil {
		return "", fmt.Errorf("failed to parse refresh token: %w", err)
	}

	if !token.Valid {
		return "", fmt.Errorf("invalid refresh token")
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return "", fmt.Errorf("invalid refresh token claims")
	}

	return claims.Subject, nil
}

// generateToken generates a token with given claims
func (s *TokenService) generateToken(claims Claims) (string, error) {
	var token *jwt.Token

	if s.config.Algorithm == "RS256" {
		token = jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		return token.SignedString(s.privateKey)
	}

	// HS256
	token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.Secret))
}

// keyFunc returns the key for token validation
func (s *TokenService) keyFunc(token *jwt.Token) (interface{}, error) {
	// Validate algorithm
	if s.config.Algorithm == "RS256" {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.publicKey, nil
	}

	// HS256
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
	return []byte(s.config.Secret), nil
}

// loadRSAKeys loads RSA private and public keys from files
func (s *TokenService) loadRSAKeys() error {
	// Load private key
	privateKeyData, err := os.ReadFile(s.config.PrivateKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read private key file: %w", err)
	}

	privateBlock, _ := pem.Decode(privateKeyData)
	if privateBlock == nil {
		return fmt.Errorf("failed to decode private key PEM")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(privateBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}
	s.privateKey = privateKey

	// Load public key
	publicKeyData, err := os.ReadFile(s.config.PublicKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read public key file: %w", err)
	}

	publicBlock, _ := pem.Decode(publicKeyData)
	if publicBlock == nil {
		return fmt.Errorf("failed to decode public key PEM")
	}

	publicKeyInterface, err := x509.ParsePKIXPublicKey(publicBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}

	publicKey, ok := publicKeyInterface.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("public key is not RSA")
	}
	s.publicKey = publicKey

	return nil
}
