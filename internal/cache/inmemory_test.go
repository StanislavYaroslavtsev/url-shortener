package cache

import (
	"context"
	"testing"
	"time"

	"github.com/StanislavYaroslavtsev/url-shortener/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestCache() *InMemoryCache {
	return NewInMemoryCache(1*time.Minute, 1*time.Minute)
}

func TestInMemoryCache_Set_Get_ReturnsLink(t *testing.T) {
	cache := newTestCache()
	t.Cleanup(func() {
		assert.NoError(t, cache.Close())
	})

	link, err := domain.NewLink("https://google.com", "abc123", nil, nil)
	require.NoError(t, err)

	err = cache.Set(context.Background(), "abc123", link)
	require.NoError(t, err)

	result, err := cache.Get(context.Background(), "abc123")
	require.NoError(t, err)
	assert.Equal(t, link, result)
}

func TestInMemoryCache_Get_MissingKey_ReturnsCacheMiss(t *testing.T) {
	cache := newTestCache()
	t.Cleanup(func() {
		assert.NoError(t, cache.Close())
	})

	_, err := cache.Get(context.Background(), "nonexistent")
	assert.ErrorIs(t, err, ErrCacheMiss)
}

func TestInMemoryCache_Get_ExpiredItem_ReturnsCacheExpired(t *testing.T) {
	cache := NewInMemoryCache(1*time.Millisecond, 1*time.Minute)
	t.Cleanup(func() {
		assert.NoError(t, cache.Close())
	})

	link, err := domain.NewLink("https://google.com", "abc123", nil, nil)
	require.NoError(t, err)

	err = cache.Set(context.Background(), "abc123", link)
	require.NoError(t, err)

	time.Sleep(5 * time.Millisecond)

	_, err = cache.Get(context.Background(), "abc123")
	assert.ErrorIs(t, err, ErrCacheExpired)
}

func TestInMemoryCache_Close_ReturnsNoError(t *testing.T) {
	cache := newTestCache()
	assert.NoError(t, cache.Close())
}
