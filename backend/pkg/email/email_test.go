package email

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateInvitationToken(t *testing.T) {
	token1, err := GenerateInvitationToken()
	require.NoError(t, err)
	assert.NotEmpty(t, token1)

	// Token should be URL-safe base64 (length = 44 for 32 bytes)
	assert.True(t, len(token1) > 40)

	// Generate second token
	token2, err := GenerateInvitationToken()
	require.NoError(t, err)
	assert.NotEmpty(t, token2)

	// Tokens should be unique
	assert.NotEqual(t, token1, token2)

	// Token should only contain URL-safe characters
	urlSafeChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_="
	for _, char := range token1 {
		assert.True(t, strings.ContainsRune(urlSafeChars, char),
			"Token contains non-URL-safe character: %c", char)
	}
}

func TestGenerateInvitationToken_Uniqueness(t *testing.T) {
	// Generate 100 tokens and ensure all are unique
	tokens := make(map[string]bool)

	for i := 0; i < 100; i++ {
		token, err := GenerateInvitationToken()
		require.NoError(t, err)

		// Check if token already exists (should never happen)
		assert.False(t, tokens[token], "Duplicate token generated: %s", token)

		tokens[token] = true
	}

	// All 100 tokens should be unique
	assert.Equal(t, 100, len(tokens))
}

func TestGenerateInvitationToken_Security(t *testing.T) {
	token, err := GenerateInvitationToken()
	require.NoError(t, err)

	// Token should have sufficient entropy
	// 32 bytes = 256 bits of entropy (cryptographically secure)
	assert.True(t, len(token) >= 40, "Token too short for secure entropy")

	// Token should not be predictable (basic check)
	// Generate 10 tokens and ensure they don't follow a pattern
	var tokens []string
	for i := 0; i < 10; i++ {
		tok, err := GenerateInvitationToken()
		require.NoError(t, err)
		tokens = append(tokens, tok)
	}

	// Check that consecutive tokens are different
	for i := 1; i < len(tokens); i++ {
		assert.NotEqual(t, tokens[i-1], tokens[i])
	}
}

// Benchmark for token generation performance
func BenchmarkGenerateInvitationToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateInvitationToken()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Note: SendInvitationEmail and sendEmailWithRetry tests require SMTP server mock
// These should be implemented with integration tests or using a mock SMTP server
// For now, we test the token generation which is the critical security component
