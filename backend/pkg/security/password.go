package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"

	"backend/internal/config"
)

// PasswordHasher handles password hashing and verification using Argon2id
// Reference: BACKEND-IMPLEMENTATION.md lines 56-69
type PasswordHasher struct {
	config config.Argon2Config
}

// NewPasswordHasher creates a new password hasher with configuration
func NewPasswordHasher(cfg config.Argon2Config) *PasswordHasher {
	return &PasswordHasher{
		config: cfg,
	}
}

// HashPassword hashes a password using Argon2id
// Returns a hash in format: $argon2id$v=19$m=65536,t=3,p=4$salt$hash
// This format is compatible with standard Argon2 implementations
func (h *PasswordHasher) HashPassword(password string) (string, error) {
	// Generate a cryptographically secure random salt
	salt := make([]byte, h.config.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Hash the password using Argon2id
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		h.config.Iterations,
		h.config.Memory,
		h.config.Parallelism,
		h.config.KeyLength,
	)

	// Encode the hash in standard format
	// Format: $argon2id$v=19$m=65536,t=3,p=4$salt$hash
	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		h.config.Memory,
		h.config.Iterations,
		h.config.Parallelism,
		encodedSalt,
		encodedHash,
	), nil
}

// VerifyPassword compares a plain password with a hashed password
// Returns true if they match, false otherwise
// Uses constant-time comparison to prevent timing attacks
func (h *PasswordHasher) VerifyPassword(password, hashedPassword string) (bool, error) {
	// Parse the hash string to extract parameters
	params, salt, hash, err := h.parseHash(hashedPassword)
	if err != nil {
		return false, fmt.Errorf("failed to parse hash: %w", err)
	}

	// Hash the password with the same parameters
	compareHash := argon2.IDKey(
		[]byte(password),
		salt,
		params.iterations,
		params.memory,
		params.parallelism,
		params.keyLength,
	)

	// Use constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare(hash, compareHash) == 1 {
		return true, nil
	}

	return false, nil
}

// hashParams holds parameters extracted from a hash string
type hashParams struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	keyLength   uint32
}

// parseHash parses a hash string and extracts parameters, salt, and hash
func (h *PasswordHasher) parseHash(encodedHash string) (*hashParams, []byte, []byte, error) {
	// Expected format: $argon2id$v=19$m=65536,t=3,p=4$salt$hash
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return nil, nil, nil, fmt.Errorf("invalid hash format")
	}

	if parts[1] != "argon2id" {
		return nil, nil, nil, fmt.Errorf("only argon2id is supported")
	}

	// Parse version (should be 19)
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse version: %w", err)
	}
	if version != argon2.Version {
		return nil, nil, nil, fmt.Errorf("incompatible argon2 version: %d", version)
	}

	// Parse parameters
	params := &hashParams{}
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d",
		&params.memory,
		&params.iterations,
		&params.parallelism,
	); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse parameters: %w", err)
	}

	// Decode salt
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to decode salt: %w", err)
	}

	// Decode hash
	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to decode hash: %w", err)
	}

	params.keyLength = uint32(len(hash))

	return params, salt, hash, nil
}
