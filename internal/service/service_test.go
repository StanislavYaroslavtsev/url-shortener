package service

import (
	"context"
	"testing"
	"time"

	appcache "github.com/StanislavYaroslavtsev/url-shortener/internal/cache"
	"github.com/StanislavYaroslavtsev/url-shortener/internal/domain"
	"github.com/StanislavYaroslavtsev/url-shortener/internal/repository"
	"github.com/StanislavYaroslavtsev/url-shortener/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestLinkService_Create_ReturnsLink(t *testing.T) {
	repo := mocks.NewMockLinkRepository(t)
	cache := mocks.NewMockCache(t)

	repo.EXPECT().Save(mock.Anything, mock.Anything).Return(nil)
	cache.EXPECT().Set(mock.Anything, mock.Anything, mock.Anything).Return(nil)

	svc := NewLinkService(repo, cache)

	link, err := svc.Create(context.Background(), "https://google.com", nil)

	require.NoError(t, err)
	assert.Equal(t, "https://google.com", link.URL)
	assert.NotEmpty(t, link.Code)
}

func TestLinkService_Create_InvalidURL_ReturnsError(t *testing.T) {
	repo := mocks.NewMockLinkRepository(t)
	cache := mocks.NewMockCache(t)

	svc := NewLinkService(repo, cache)
	_, err := svc.Create(context.Background(), "not-a-url", nil)

	assert.ErrorIs(t, err, domain.ErrInvalidLink)
}

func TestLinkService_Get_ReturnsCachedLink(t *testing.T) {
	repo := mocks.NewMockLinkRepository(t)
	cache := mocks.NewMockCache(t)

	link, err := domain.NewLink("https://google.com", "abc123", nil)
	require.NoError(t, err)

	cache.EXPECT().Get(mock.Anything, "abc123").Return(link, nil)

	svc := NewLinkService(repo, cache)
	result, err := svc.Get(context.Background(), "abc123")

	require.NoError(t, err)
	assert.Equal(t, link, result)
}

func TestLinkService_Get_CacheMiss_ReturnsFromRepo(t *testing.T) {
	repo := mocks.NewMockLinkRepository(t)
	cache := mocks.NewMockCache(t)

	link, err := domain.NewLink("https://google.com", "abc123", nil)
	require.NoError(t, err)

	cache.EXPECT().Get(mock.Anything, "abc123").Return(nil, appcache.ErrCacheMiss)
	repo.EXPECT().Get(mock.Anything, "abc123").Return(link, nil)
	cache.EXPECT().Set(mock.Anything, "abc123", link).Return(nil)

	svc := NewLinkService(repo, cache)
	result, err := svc.Get(context.Background(), "abc123")

	require.NoError(t, err)
	assert.Equal(t, link, result)
}

func TestLinkService_Get_ExpiredLink_ReturnsError(t *testing.T) {
	repo := mocks.NewMockLinkRepository(t)
	cache := mocks.NewMockCache(t)

	expiresAt := time.Now().UTC().Add(-1 * time.Hour)
	link, err := domain.NewLink("https://google.com", "abc123", &expiresAt)
	require.NoError(t, err)

	cache.EXPECT().Get(mock.Anything, "abc123").Return(link, nil)

	svc := NewLinkService(repo, cache)
	_, err = svc.Get(context.Background(), "abc123")

	assert.ErrorIs(t, err, domain.ErrLinkExpired)
}

func TestLinkService_Get_NotFound_ReturnsError(t *testing.T) {
	repo := mocks.NewMockLinkRepository(t)
	cache := mocks.NewMockCache(t)

	cache.EXPECT().Get(mock.Anything, "abc123").Return(nil, appcache.ErrCacheMiss)
	repo.EXPECT().Get(mock.Anything, "abc123").Return(nil, repository.ErrLinkNotFound)

	svc := NewLinkService(repo, cache)
	_, err := svc.Get(context.Background(), "abc123")

	assert.ErrorIs(t, err, repository.ErrLinkNotFound)
}
