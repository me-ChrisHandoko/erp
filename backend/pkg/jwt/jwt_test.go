package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"backend/internal/config"
)

// Test 1: NewTokenService - HS256 Initialization
func TestNewTokenService_HS256(t *testing.T) {
	cfg := config.JWTConfig{
		Secret:        "test-secret-key-for-jwt-testing",
		Algorithm:     "HS256",
		Expiry:        15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	}

	service, err := NewTokenService(cfg)
	require.NoError(t, err, "Should create HS256 token service")
	assert.NotNil(t, service)
	assert.Equal(t, "HS256", service.config.Algorithm)
	assert.Nil(t, service.privateKey, "HS256 should not have RSA keys")
	assert.Nil(t, service.publicKey, "HS256 should not have RSA keys")
}

// Test 2: NewTokenService - RS256 Initialization with Valid Keys
func TestNewTokenService_RS256_ValidKeys(t *testing.T) {
	// Create temporary directory for test keys
	tempDir := t.TempDir()
	privateKeyPath := filepath.Join(tempDir, "private.pem")
	publicKeyPath := filepath.Join(tempDir, "public.pem")

	// Generate test RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Write private key
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})
	err = os.WriteFile(privateKeyPath, privateKeyPEM, 0600)
	require.NoError(t, err)

	// Write public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err)
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})
	err = os.WriteFile(publicKeyPath, publicKeyPEM, 0644)
	require.NoError(t, err)

	// Create token service with RS256
	cfg := config.JWTConfig{
		Algorithm:      "RS256",
		PrivateKeyPath: privateKeyPath,
		PublicKeyPath:  publicKeyPath,
		Expiry:         15 * time.Minute,
		RefreshExpiry:  7 * 24 * time.Hour,
	}

	service, err := NewTokenService(cfg)
	require.NoError(t, err, "Should create RS256 token service")
	assert.NotNil(t, service)
	assert.Equal(t, "RS256", service.config.Algorithm)
	assert.NotNil(t, service.privateKey, "RS256 should have private key")
	assert.NotNil(t, service.publicKey, "RS256 should have public key")
}

// Test 3: NewTokenService - RS256 with Missing Keys
func TestNewTokenService_RS256_MissingKeys(t *testing.T) {
	cfg := config.JWTConfig{
		Algorithm:      "RS256",
		PrivateKeyPath: "/nonexistent/private.pem",
		PublicKeyPath:  "/nonexistent/public.pem",
		Expiry:         15 * time.Minute,
		RefreshExpiry:  7 * 24 * time.Hour,
	}

	service, err := NewTokenService(cfg)
	assert.Error(t, err, "Should fail with missing key files")
	assert.Nil(t, service)
	assert.Contains(t, err.Error(), "failed to load RSA keys")
}

// Test 4: GenerateAccessToken - HS256 Token Generation
func TestGenerateAccessToken_HS256(t *testing.T) {
	cfg := config.JWTConfig{
		Secret:        "test-secret-key-for-jwt-testing",
		Algorithm:     "HS256",
		Expiry:        15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	}

	service, err := NewTokenService(cfg)
	require.NoError(t, err)

	// Generate token
	userID := "user-123"
	email := "test@example.com"
	tenantID := "tenant-456"
	role := "ADMIN"

	token, err := service.GenerateAccessToken(userID, email, tenantID, role)
	require.NoError(t, err, "Should generate access token")
	assert.NotEmpty(t, token, "Token should not be empty")

	// Verify token format (JWT has 3 parts separated by dots)
	assert.Contains(t, token, ".")
}

// Test 5: ValidateAccessToken - Valid Token with All Claims
func TestValidateAccessToken_ValidToken(t *testing.T) {
	cfg := config.JWTConfig{
		Secret:        "test-secret-key-for-jwt-testing",
		Algorithm:     "HS256",
		Expiry:        15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	}

	service, err := NewTokenService(cfg)
	require.NoError(t, err)

	// Generate token
	userID := "user-123"
	email := "test@example.com"
	tenantID := "tenant-456"
	role := "ADMIN"

	token, err := service.GenerateAccessToken(userID, email, tenantID, role)
	require.NoError(t, err)

	// Validate and extract claims
	claims, err := service.ValidateAccessToken(token)
	require.NoError(t, err, "Should validate token successfully")
	assert.Equal(t, userID, claims.UserID, "UserID should match")
	assert.Equal(t, email, claims.Email, "Email should match")
	assert.Equal(t, tenantID, claims.TenantID, "TenantID should match")
	assert.Equal(t, role, claims.Role, "Role should match")
	assert.NotNil(t, claims.ExpiresAt, "Should have expiry time")
	assert.NotNil(t, claims.IssuedAt, "Should have issued time")
}

// Test 6: ValidateAccessToken - Expired Token
func TestValidateAccessToken_ExpiredToken(t *testing.T) {
	cfg := config.JWTConfig{
		Secret:        "test-secret-key-for-jwt-testing",
		Algorithm:     "HS256",
		Expiry:        1 * time.Millisecond, // Very short expiry
		RefreshExpiry: 7 * 24 * time.Hour,
	}

	service, err := NewTokenService(cfg)
	require.NoError(t, err)

	// Generate token
	token, err := service.GenerateAccessToken("user-123", "test@example.com", "tenant-456", "ADMIN")
	require.NoError(t, err)

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	// Validate expired token
	claims, err := service.ValidateAccessToken(token)
	assert.Error(t, err, "Should reject expired token")
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "failed to parse token")
}

// Test 7: ValidateAccessToken - Invalid Signature (Tampered Token)
func TestValidateAccessToken_InvalidSignature(t *testing.T) {
	cfg := config.JWTConfig{
		Secret:        "test-secret-key-for-jwt-testing",
		Algorithm:     "HS256",
		Expiry:        15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	}

	service, err := NewTokenService(cfg)
	require.NoError(t, err)

	// Generate valid token
	token, err := service.GenerateAccessToken("user-123", "test@example.com", "tenant-456", "ADMIN")
	require.NoError(t, err)

	// Tamper with token signature
	tamperedToken := token[:len(token)-5] + "AAAAA"

	// Validate tampered token
	claims, err := service.ValidateAccessToken(tamperedToken)
	assert.Error(t, err, "Should reject tampered token")
	assert.Nil(t, claims)
}

// Test 8: ValidateAccessToken - Wrong Secret Key
func TestValidateAccessToken_WrongSecret(t *testing.T) {
	cfg := config.JWTConfig{
		Secret:        "test-secret-key-for-jwt-testing",
		Algorithm:     "HS256",
		Expiry:        15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	}

	service, err := NewTokenService(cfg)
	require.NoError(t, err)

	// Generate token
	token, err := service.GenerateAccessToken("user-123", "test@example.com", "tenant-456", "ADMIN")
	require.NoError(t, err)

	// Create service with different secret
	wrongCfg := config.JWTConfig{
		Secret:        "different-secret-key",
		Algorithm:     "HS256",
		Expiry:        15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	}
	wrongService, err := NewTokenService(wrongCfg)
	require.NoError(t, err)

	// Validate with wrong secret
	claims, err := wrongService.ValidateAccessToken(token)
	assert.Error(t, err, "Should reject token signed with different secret")
	assert.Nil(t, claims)
}

// Test 9: ValidateAccessToken - Malformed Token
func TestValidateAccessToken_MalformedToken(t *testing.T) {
	cfg := config.JWTConfig{
		Secret:        "test-secret-key-for-jwt-testing",
		Algorithm:     "HS256",
		Expiry:        15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	}

	service, err := NewTokenService(cfg)
	require.NoError(t, err)

	testCases := []struct {
		name  string
		token string
	}{
		{"Empty token", ""},
		{"Invalid format", "not.a.valid.jwt.token"},
		{"Missing parts", "header.payload"},
		{"Random string", "random-string-not-jwt"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			claims, err := service.ValidateAccessToken(tc.token)
			assert.Error(t, err, "Should reject malformed token: %s", tc.name)
			assert.Nil(t, claims)
		})
	}
}

// Test 10: GenerateRefreshToken - HS256 Refresh Token Generation
func TestGenerateRefreshToken_HS256(t *testing.T) {
	cfg := config.JWTConfig{
		Secret:        "test-secret-key-for-jwt-testing",
		Algorithm:     "HS256",
		Expiry:        15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	}

	service, err := NewTokenService(cfg)
	require.NoError(t, err)

	// Generate refresh token
	userID := "user-123"
	refreshToken, err := service.GenerateRefreshToken(userID)
	require.NoError(t, err, "Should generate refresh token")
	assert.NotEmpty(t, refreshToken, "Refresh token should not be empty")
}

// Test 11: ValidateRefreshToken - Valid Refresh Token
func TestValidateRefreshToken_ValidToken(t *testing.T) {
	cfg := config.JWTConfig{
		Secret:        "test-secret-key-for-jwt-testing",
		Algorithm:     "HS256",
		Expiry:        15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	}

	service, err := NewTokenService(cfg)
	require.NoError(t, err)

	// Generate refresh token
	userID := "user-123"
	refreshToken, err := service.GenerateRefreshToken(userID)
	require.NoError(t, err)

	// Validate refresh token
	extractedUserID, err := service.ValidateRefreshToken(refreshToken)
	require.NoError(t, err, "Should validate refresh token")
	assert.Equal(t, userID, extractedUserID, "UserID should match")
}

// Test 12: ValidateRefreshToken - Expired Refresh Token
func TestValidateRefreshToken_ExpiredToken(t *testing.T) {
	cfg := config.JWTConfig{
		Secret:        "test-secret-key-for-jwt-testing",
		Algorithm:     "HS256",
		Expiry:        15 * time.Minute,
		RefreshExpiry: 1 * time.Millisecond, // Very short expiry
	}

	service, err := NewTokenService(cfg)
	require.NoError(t, err)

	// Generate refresh token
	refreshToken, err := service.GenerateRefreshToken("user-123")
	require.NoError(t, err)

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	// Validate expired refresh token
	userID, err := service.ValidateRefreshToken(refreshToken)
	assert.Error(t, err, "Should reject expired refresh token")
	assert.Empty(t, userID)
}

// Test 13: Algorithm Confusion Attack Prevention
func TestAlgorithmConfusion_Prevention(t *testing.T) {
	cfg := config.JWTConfig{
		Secret:        "test-secret-key-for-jwt-testing",
		Algorithm:     "HS256",
		Expiry:        15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	}

	service, err := NewTokenService(cfg)
	require.NoError(t, err)

	// Create a token with different algorithm (security attack simulation)
	claims := Claims{
		UserID:   "attacker",
		Email:    "attacker@example.com",
		TenantID: "stolen-tenant",
		Role:     "OWNER",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Try to create token with different algorithm (none algorithm attack)
	noneToken := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	noneTokenString, err := noneToken.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	// Validate - should reject due to algorithm mismatch
	extractedClaims, err := service.ValidateAccessToken(noneTokenString)
	assert.Error(t, err, "Should reject token with different algorithm")
	assert.Nil(t, extractedClaims)
	assert.Contains(t, err.Error(), "unexpected signing method")
}

// Test 14: RS256 - Full Workflow (Generate and Validate)
func TestRS256_FullWorkflow(t *testing.T) {
	// Create temporary directory for test keys
	tempDir := t.TempDir()
	privateKeyPath := filepath.Join(tempDir, "private.pem")
	publicKeyPath := filepath.Join(tempDir, "public.pem")

	// Generate test RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Write private key
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})
	err = os.WriteFile(privateKeyPath, privateKeyPEM, 0600)
	require.NoError(t, err)

	// Write public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err)
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})
	err = os.WriteFile(publicKeyPath, publicKeyPEM, 0644)
	require.NoError(t, err)

	// Create token service with RS256
	cfg := config.JWTConfig{
		Algorithm:      "RS256",
		PrivateKeyPath: privateKeyPath,
		PublicKeyPath:  publicKeyPath,
		Expiry:         15 * time.Minute,
		RefreshExpiry:  7 * 24 * time.Hour,
	}

	service, err := NewTokenService(cfg)
	require.NoError(t, err)

	// Generate access token
	userID := "user-123"
	email := "test@example.com"
	tenantID := "tenant-456"
	role := "ADMIN"

	token, err := service.GenerateAccessToken(userID, email, tenantID, role)
	require.NoError(t, err, "Should generate RS256 access token")
	assert.NotEmpty(t, token)

	// Validate token
	claims, err := service.ValidateAccessToken(token)
	require.NoError(t, err, "Should validate RS256 token")
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, tenantID, claims.TenantID)
	assert.Equal(t, role, claims.Role)

	// Generate and validate refresh token
	refreshToken, err := service.GenerateRefreshToken(userID)
	require.NoError(t, err, "Should generate RS256 refresh token")

	extractedUserID, err := service.ValidateRefreshToken(refreshToken)
	require.NoError(t, err, "Should validate RS256 refresh token")
	assert.Equal(t, userID, extractedUserID)
}

// Test 15: RS256 - Invalid Public Key (Security Test)
func TestRS256_InvalidPublicKey(t *testing.T) {
	// Create temporary directory for test keys
	tempDir := t.TempDir()
	privateKeyPath := filepath.Join(tempDir, "private.pem")
	publicKeyPath := filepath.Join(tempDir, "public.pem")

	// Generate first RSA key pair
	privateKey1, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Generate second RSA key pair (different keys)
	privateKey2, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Write private key 1
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey1)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})
	err = os.WriteFile(privateKeyPath, privateKeyPEM, 0600)
	require.NoError(t, err)

	// Write public key 2 (mismatch!)
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey2.PublicKey)
	require.NoError(t, err)
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})
	err = os.WriteFile(publicKeyPath, publicKeyPEM, 0644)
	require.NoError(t, err)

	// Create token service with mismatched keys
	cfg := config.JWTConfig{
		Algorithm:      "RS256",
		PrivateKeyPath: privateKeyPath,
		PublicKeyPath:  publicKeyPath,
		Expiry:         15 * time.Minute,
		RefreshExpiry:  7 * 24 * time.Hour,
	}

	service, err := NewTokenService(cfg)
	require.NoError(t, err)

	// Generate token (signed with privateKey1)
	token, err := service.GenerateAccessToken("user-123", "test@example.com", "tenant-456", "ADMIN")
	require.NoError(t, err)

	// Validate with publicKey2 - should fail
	claims, err := service.ValidateAccessToken(token)
	assert.Error(t, err, "Should reject token signed with different key")
	assert.Nil(t, claims)
}

// Test 16: Token Expiry Timing Precision
func TestTokenExpiry_TimingPrecision(t *testing.T) {
	cfg := config.JWTConfig{
		Secret:        "test-secret-key-for-jwt-testing",
		Algorithm:     "HS256",
		Expiry:        2 * time.Second, // 2 second expiry for testing
		RefreshExpiry: 7 * 24 * time.Hour,
	}

	service, err := NewTokenService(cfg)
	require.NoError(t, err)

	// Generate token
	token, err := service.GenerateAccessToken("user-123", "test@example.com", "tenant-456", "ADMIN")
	require.NoError(t, err)

	// Validate immediately - should pass
	claims, err := service.ValidateAccessToken(token)
	require.NoError(t, err, "Should validate immediately after generation")
	assert.NotNil(t, claims)

	// Wait 1 second - should still be valid
	time.Sleep(1 * time.Second)
	claims, err = service.ValidateAccessToken(token)
	require.NoError(t, err, "Should still be valid after 1 second")
	assert.NotNil(t, claims)

	// Wait 2 more seconds (total 3 seconds) - should expire
	time.Sleep(2 * time.Second)
	claims, err = service.ValidateAccessToken(token)
	assert.Error(t, err, "Should be expired after 3 seconds")
	assert.Nil(t, claims)
}

// Test 17: Claims Extraction - All Fields
func TestClaims_AllFieldsExtraction(t *testing.T) {
	cfg := config.JWTConfig{
		Secret:        "test-secret-key-for-jwt-testing",
		Algorithm:     "HS256",
		Expiry:        15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	}

	service, err := NewTokenService(cfg)
	require.NoError(t, err)

	// Generate token with specific values
	userID := "user-12345"
	email := "testuser@example.com"
	tenantID := "tenant-67890"
	role := "FINANCE"

	token, err := service.GenerateAccessToken(userID, email, tenantID, role)
	require.NoError(t, err)

	// Validate and extract all claims
	claims, err := service.ValidateAccessToken(token)
	require.NoError(t, err)

	// Verify all custom claims
	assert.Equal(t, userID, claims.UserID, "UserID mismatch")
	assert.Equal(t, email, claims.Email, "Email mismatch")
	assert.Equal(t, tenantID, claims.TenantID, "TenantID mismatch")
	assert.Equal(t, role, claims.Role, "Role mismatch")

	// Verify standard claims
	assert.NotNil(t, claims.ExpiresAt, "Should have ExpiresAt")
	assert.NotNil(t, claims.IssuedAt, "Should have IssuedAt")
	assert.NotNil(t, claims.NotBefore, "Should have NotBefore")

	// Verify timing
	now := time.Now()
	assert.True(t, claims.IssuedAt.Time.Before(now) || claims.IssuedAt.Time.Equal(now), "IssuedAt should be in the past or now")
	assert.True(t, claims.ExpiresAt.Time.After(now), "ExpiresAt should be in the future")
}

// Test 18: Empty/Nil Claims Handling
func TestGenerateToken_EmptyClaims(t *testing.T) {
	cfg := config.JWTConfig{
		Secret:        "test-secret-key-for-jwt-testing",
		Algorithm:     "HS256",
		Expiry:        15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	}

	service, err := NewTokenService(cfg)
	require.NoError(t, err)

	// Generate token with empty claims
	token, err := service.GenerateAccessToken("", "", "", "")
	require.NoError(t, err, "Should allow empty claims")

	// Validate and verify empty claims are preserved
	claims, err := service.ValidateAccessToken(token)
	require.NoError(t, err)
	assert.Empty(t, claims.UserID)
	assert.Empty(t, claims.Email)
	assert.Empty(t, claims.TenantID)
	assert.Empty(t, claims.Role)
}

// Test 19: Refresh Token Contains Minimal Claims (Security Best Practice)
func TestRefreshToken_MinimalClaims(t *testing.T) {
	cfg := config.JWTConfig{
		Secret:        "test-secret-key-for-jwt-testing",
		Algorithm:     "HS256",
		Expiry:        15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	}

	service, err := NewTokenService(cfg)
	require.NoError(t, err)

	// Generate refresh token
	userID := "user-123"
	refreshToken, err := service.GenerateRefreshToken(userID)
	require.NoError(t, err)

	// Parse token manually to check claims structure
	parsedToken, err := jwt.ParseWithClaims(refreshToken, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.Secret), nil
	})
	require.NoError(t, err)

	claims, ok := parsedToken.Claims.(*jwt.RegisteredClaims)
	require.True(t, ok)

	// Verify refresh token only has Subject (userID) and standard claims
	assert.Equal(t, userID, claims.Subject, "Should have Subject claim")
	assert.NotNil(t, claims.ExpiresAt, "Should have ExpiresAt")
	assert.NotNil(t, claims.IssuedAt, "Should have IssuedAt")

	// Verify it's NOT a full Claims struct (no email, tenantID, role)
	fullClaims := &Claims{}
	parsedToken2, err := jwt.ParseWithClaims(refreshToken, fullClaims, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.Secret), nil
	})
	require.NoError(t, err)

	fullClaimsExtracted, ok := parsedToken2.Claims.(*Claims)
	require.True(t, ok)

	// These should be empty in refresh token
	assert.Empty(t, fullClaimsExtracted.UserID, "Refresh token should not have UserID claim")
	assert.Empty(t, fullClaimsExtracted.Email, "Refresh token should not have Email claim")
	assert.Empty(t, fullClaimsExtracted.TenantID, "Refresh token should not have TenantID claim")
	assert.Empty(t, fullClaimsExtracted.Role, "Refresh token should not have Role claim")
}

// Test 20: Concurrent Token Generation (Thread Safety)
func TestConcurrentTokenGeneration(t *testing.T) {
	cfg := config.JWTConfig{
		Secret:        "test-secret-key-for-jwt-testing",
		Algorithm:     "HS256",
		Expiry:        15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	}

	service, err := NewTokenService(cfg)
	require.NoError(t, err)

	// Generate 10 tokens concurrently
	const numTokens = 10
	tokens := make(chan string, numTokens)
	errors := make(chan error, numTokens)

	for i := 0; i < numTokens; i++ {
		go func(id int) {
			token, err := service.GenerateAccessToken(
				"user-"+string(rune(id)),
				"user"+string(rune(id))+"@example.com",
				"tenant-1",
				"ADMIN",
			)
			if err != nil {
				errors <- err
			} else {
				tokens <- token
			}
		}(i)
	}

	// Collect results
	generatedTokens := make([]string, 0, numTokens)
	for i := 0; i < numTokens; i++ {
		select {
		case token := <-tokens:
			generatedTokens = append(generatedTokens, token)
		case err := <-errors:
			t.Fatalf("Token generation failed: %v", err)
		}
	}

	// Verify all tokens are unique
	assert.Len(t, generatedTokens, numTokens, "Should generate all tokens")
	tokenSet := make(map[string]bool)
	for _, token := range generatedTokens {
		assert.False(t, tokenSet[token], "Tokens should be unique")
		tokenSet[token] = true
	}
}

// Test 21: RS256 - Invalid PEM Format
func TestRS256_InvalidPEMFormat(t *testing.T) {
	tempDir := t.TempDir()
	privateKeyPath := filepath.Join(tempDir, "private.pem")
	publicKeyPath := filepath.Join(tempDir, "public.pem")

	// Write invalid PEM files (not proper PEM format)
	err := os.WriteFile(privateKeyPath, []byte("not a valid PEM file"), 0600)
	require.NoError(t, err)
	err = os.WriteFile(publicKeyPath, []byte("not a valid PEM file"), 0644)
	require.NoError(t, err)

	cfg := config.JWTConfig{
		Algorithm:      "RS256",
		PrivateKeyPath: privateKeyPath,
		PublicKeyPath:  publicKeyPath,
		Expiry:         15 * time.Minute,
		RefreshExpiry:  7 * 24 * time.Hour,
	}

	service, err := NewTokenService(cfg)
	assert.Error(t, err, "Should fail with invalid PEM format")
	assert.Nil(t, service)
	assert.Contains(t, err.Error(), "failed to decode private key PEM")
}

// Test 22: RS256 - Invalid Key Data (Not RSA)
func TestRS256_InvalidKeyData(t *testing.T) {
	tempDir := t.TempDir()
	privateKeyPath := filepath.Join(tempDir, "private.pem")
	publicKeyPath := filepath.Join(tempDir, "public.pem")

	// Write valid PEM format but with invalid key data
	invalidPrivateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: []byte("invalid key data"),
	})
	err := os.WriteFile(privateKeyPath, invalidPrivateKeyPEM, 0600)
	require.NoError(t, err)

	invalidPublicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: []byte("invalid key data"),
	})
	err = os.WriteFile(publicKeyPath, invalidPublicKeyPEM, 0644)
	require.NoError(t, err)

	cfg := config.JWTConfig{
		Algorithm:      "RS256",
		PrivateKeyPath: privateKeyPath,
		PublicKeyPath:  publicKeyPath,
		Expiry:         15 * time.Minute,
		RefreshExpiry:  7 * 24 * time.Hour,
	}

	service, err := NewTokenService(cfg)
	assert.Error(t, err, "Should fail with invalid key data")
	assert.Nil(t, service)
	assert.Contains(t, err.Error(), "failed to parse private key")
}

// Test 23: RS256 - Public Key Wrong Type (Not RSA)
func TestRS256_PublicKeyWrongType(t *testing.T) {
	tempDir := t.TempDir()
	privateKeyPath := filepath.Join(tempDir, "private.pem")
	publicKeyPath := filepath.Join(tempDir, "public.pem")

	// Generate valid RSA private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Write valid private key
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})
	err = os.WriteFile(privateKeyPath, privateKeyPEM, 0600)
	require.NoError(t, err)

	// Write invalid public key PEM (valid PEM but invalid PKIX data)
	invalidPublicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: []byte{0x30, 0x00}, // Minimal ASN.1 sequence but not valid PKIX
	})
	err = os.WriteFile(publicKeyPath, invalidPublicKeyPEM, 0644)
	require.NoError(t, err)

	cfg := config.JWTConfig{
		Algorithm:      "RS256",
		PrivateKeyPath: privateKeyPath,
		PublicKeyPath:  publicKeyPath,
		Expiry:         15 * time.Minute,
		RefreshExpiry:  7 * 24 * time.Hour,
	}

	service, err := NewTokenService(cfg)
	assert.Error(t, err, "Should fail with invalid public key data")
	assert.Nil(t, service)
}

// Test 24: RS256 - Public Key PEM Decode Failure
func TestRS256_PublicKeyPEMDecodeFail(t *testing.T) {
	tempDir := t.TempDir()
	privateKeyPath := filepath.Join(tempDir, "private.pem")
	publicKeyPath := filepath.Join(tempDir, "public.pem")

	// Generate valid RSA private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Write valid private key
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})
	err = os.WriteFile(privateKeyPath, privateKeyPEM, 0600)
	require.NoError(t, err)

	// Write invalid public key (not PEM format)
	err = os.WriteFile(publicKeyPath, []byte("not a valid PEM file for public key"), 0644)
	require.NoError(t, err)

	cfg := config.JWTConfig{
		Algorithm:      "RS256",
		PrivateKeyPath: privateKeyPath,
		PublicKeyPath:  publicKeyPath,
		Expiry:         15 * time.Minute,
		RefreshExpiry:  7 * 24 * time.Hour,
	}

	service, err := NewTokenService(cfg)
	assert.Error(t, err, "Should fail with invalid public key PEM")
	assert.Nil(t, service)
	assert.Contains(t, err.Error(), "failed to decode public key PEM")
}

// Test 25: RS256 - Algorithm Confusion with HS256 Token
func TestRS256_AlgorithmConfusionWithHS256(t *testing.T) {
	// Create temporary directory for test keys
	tempDir := t.TempDir()
	privateKeyPath := filepath.Join(tempDir, "private.pem")
	publicKeyPath := filepath.Join(tempDir, "public.pem")

	// Generate test RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Write private key
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})
	err = os.WriteFile(privateKeyPath, privateKeyPEM, 0600)
	require.NoError(t, err)

	// Write public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err)
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})
	err = os.WriteFile(publicKeyPath, publicKeyPEM, 0644)
	require.NoError(t, err)

	// Create RS256 service
	rs256Cfg := config.JWTConfig{
		Algorithm:      "RS256",
		PrivateKeyPath: privateKeyPath,
		PublicKeyPath:  publicKeyPath,
		Expiry:         15 * time.Minute,
		RefreshExpiry:  7 * 24 * time.Hour,
	}

	rs256Service, err := NewTokenService(rs256Cfg)
	require.NoError(t, err)

	// Create HS256 token (algorithm confusion attack)
	hs256Claims := Claims{
		UserID:   "attacker",
		Email:    "attacker@example.com",
		TenantID: "stolen-tenant",
		Role:     "OWNER",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	hs256Token := jwt.NewWithClaims(jwt.SigningMethodHS256, hs256Claims)
	hs256TokenString, err := hs256Token.SignedString([]byte("attacker-secret"))
	require.NoError(t, err)

	// Try to validate HS256 token with RS256 service - should reject
	claims, err := rs256Service.ValidateAccessToken(hs256TokenString)
	assert.Error(t, err, "Should reject HS256 token when expecting RS256")
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "unexpected signing method")
}
