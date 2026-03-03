package dto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewShortenResponse_BuildsCorrectURL(t *testing.T) {
	resp := NewShortenResponse("abc123", "http://localhost:3000")

	assert.Equal(t, "http://localhost:3000/abc123", resp.ShortURL)
}

func TestNewShortenResponse_TrimsTrailingSlash(t *testing.T) {
	resp := NewShortenResponse("abc123", "http://localhost:3000/")

	assert.Equal(t, "http://localhost:3000/abc123", resp.ShortURL)
}
