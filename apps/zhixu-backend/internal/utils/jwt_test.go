package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateToken(t *testing.T) {
	userID := uint(123)
	token, err := GenerateToken(userID)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestParseToken(t *testing.T) {
	userID := uint(456)
	token, err := GenerateToken(userID)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := ParseToken(token)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, userID, claims.UserID)
}

func TestParseToken_ExpiredToken(t *testing.T) {
	_, err := ParseToken("invalid.token.here")
	assert.Error(t, err)
}

func TestParseToken_EmptyToken(t *testing.T) {
	_, err := ParseToken("")
	assert.Error(t, err)
}

func TestGenerateAndParseToken_RoundTrip(t *testing.T) {
	userID := uint(789)
	token, err := GenerateToken(userID)
	assert.NoError(t, err)

	claims, err := ParseToken(token)
	assert.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
}
