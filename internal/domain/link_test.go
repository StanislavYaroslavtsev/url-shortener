package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// New Link tests

func TestNewLink_ValidInput_ReturnsLink(t *testing.T) {
	link, err := NewLink("https://google.com", "abc123", nil)

	require.NoError(t, err)
	assert.Equal(t, "https://google.com", link.URL)
	assert.Equal(t, "abc123", link.Code)
	assert.Nil(t, link.ExpiresAt)
	assert.False(t, link.CreatedAt.IsZero())
}

func TestNewLink_InvalidURL_ReturnsError(t *testing.T) {
	_, err := NewLink("not-a-url", "abc123", nil)

	assert.Error(t, err)
}

func TestNewLink_EmptyURL_ReturnsError(t *testing.T) {
	_, err := NewLink("", "abc123", nil)

	assert.Error(t, err)
}

func TestNewLink_EmptyCode_ReturnsError(t *testing.T) {
	_, err := NewLink("https://google.com", "", nil)

	assert.Error(t, err)
}

func TestNewLink_CreatedAt_IsUTC(t *testing.T) {
	link, err := NewLink("https://google.com", "abc123", nil)

	require.NoError(t, err)
	assert.Equal(t, time.UTC, link.CreatedAt.Location())
}

// IsExpired tests

func TestIsExpired_NoExpiration_ReturnsFalse(t *testing.T) {
	link, err := NewLink("https://google.com", "abc123", nil)

	require.NoError(t, err)
	assert.False(t, link.IsExpired())
}

func TestIsExpired_ExpirationInFuture_ReturnsFalse(t *testing.T) {
	expiresAt := time.Now().UTC().Add(1 * time.Hour)
	link, err := NewLink("https://google.com", "abc123", &expiresAt)

	require.NoError(t, err)
	assert.False(t, link.IsExpired())
}

func TestIsExpired_ExpirationInPast_ReturnsTrue(t *testing.T) {
	expiresAt := time.Now().UTC().Add(-1 * time.Hour)
	link, err := NewLink("https://google.com", "abc123", &expiresAt)

	require.NoError(t, err)
	assert.True(t, link.IsExpired())
}
