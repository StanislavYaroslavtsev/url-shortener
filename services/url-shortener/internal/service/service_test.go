package service

import (
	"context"
	"testing"
	"time"

	appcache "github.com/StanislavYaroslavtsev/url-shortener/services/url-shortener/internal/cache"
	domain2 "github.com/StanislavYaroslavtsev/url-shortener/services/url-shortener/internal/domain"
	"github.com/StanislavYaroslavtsev/url-shortener/services/url-shortener/internal/repository"
	mocks2 "github.com/StanislavYaroslavtsev/url-shortener/services/url-shortener/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestLinkService_Create_ReturnsLink(t *testing.T) {
	repo := mocks2.NewMockLinkRepository(t)
	cache := mocks2.NewMockCache(t)

	repo.EXPECT().Save(mock.Anything, mock.Anything).Return(nil)
	cache.EXPECT().Set(mock.Anything, mock.Anything, mock.Anything).Return(nil)

	svc := NewLinkService(repo, cache)

	link, err := svc.Create(context.Background(), "https://google.com", nil, nil)

	require.NoError(t, err)
	assert.Equal(t, "https://google.com", link.URL)
	assert.NotEmpty(t, link.Code)
}

func TestLinkService_Create_InvalidURL_ReturnsError(t *testing.T) {
	repo := mocks2.NewMockLinkRepository(t)
	cache := mocks2.NewMockCache(t)

	svc := NewLinkService(repo, cache)
	_, err := svc.Create(context.Background(), "not-a-url", nil, nil)

	assert.ErrorIs(t, err, domain2.ErrInvalidLink)
}

func TestLinkService_Create_WithAlias_ReturnsLink(t *testing.T) {
	repo := mocks2.NewMockLinkRepository(t)
	cache := mocks2.NewMockCache(t)

	alias := "my-link"

	repo.EXPECT().Save(mock.Anything, mock.Anything).Return(nil)
	cache.EXPECT().Set(mock.Anything, alias, mock.Anything).Return(nil)

	svc := NewLinkService(repo, cache)

	link, err := svc.Create(context.Background(), "https://google.com", &alias, nil)

	require.NoError(t, err)
	assert.Equal(t, alias, link.Code)
	assert.Equal(t, &alias, link.Alias)
}

func TestLinkService_Create_AliasAlreadyTaken_ReturnsError(t *testing.T) {
	repo := mocks2.NewMockLinkRepository(t)
	cache := mocks2.NewMockCache(t)

	alias := "my-link"

	repo.EXPECT().Save(mock.Anything, mock.Anything).Return(repository.ErrCodeExists)

	svc := NewLinkService(repo, cache)

	_, err := svc.Create(context.Background(), "https://google.com", &alias, nil)

	assert.ErrorIs(t, err, repository.ErrCodeExists)
}

func TestLinkService_Create_InvalidAlias_ReturnsError(t *testing.T) {
	repo := mocks2.NewMockLinkRepository(t)
	cache := mocks2.NewMockCache(t)

	alias := "ab"

	svc := NewLinkService(repo, cache)

	_, err := svc.Create(context.Background(), "https://google.com", &alias, nil)

	assert.ErrorIs(t, err, domain2.ErrInvalidLink)
}

func TestLinkService_Get_ReturnsCachedLink(t *testing.T) {
	repo := mocks2.NewMockLinkRepository(t)
	cache := mocks2.NewMockCache(t)

	link, err := domain2.NewLink("https://google.com", "abc123", nil, nil)
	require.NoError(t, err)

	cache.EXPECT().Get(mock.Anything, "abc123").Return(link, nil)

	svc := NewLinkService(repo, cache)
	result, err := svc.Get(context.Background(), "abc123")

	require.NoError(t, err)
	assert.Equal(t, link, result)
}

func TestLinkService_Get_CacheMiss_ReturnsFromRepo(t *testing.T) {
	repo := mocks2.NewMockLinkRepository(t)
	cache := mocks2.NewMockCache(t)

	link, err := domain2.NewLink("https://google.com", "abc123", nil, nil)
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
	repo := mocks2.NewMockLinkRepository(t)
	cache := mocks2.NewMockCache(t)

	expiresAt := time.Now().UTC().Add(-1 * time.Hour)
	link, err := domain2.NewLink("https://google.com", "abc123", nil, &expiresAt)
	require.NoError(t, err)

	cache.EXPECT().Get(mock.Anything, "abc123").Return(link, nil)

	svc := NewLinkService(repo, cache)
	_, err = svc.Get(context.Background(), "abc123")

	assert.ErrorIs(t, err, domain2.ErrLinkExpired)
}

func TestLinkService_Get_NotFound_ReturnsError(t *testing.T) {
	repo := mocks2.NewMockLinkRepository(t)
	cache := mocks2.NewMockCache(t)

	cache.EXPECT().Get(mock.Anything, "abc123").Return(nil, appcache.ErrCacheMiss)
	repo.EXPECT().Get(mock.Anything, "abc123").Return(nil, repository.ErrLinkNotFound)

	svc := NewLinkService(repo, cache)
	_, err := svc.Get(context.Background(), "abc123")

	assert.ErrorIs(t, err, repository.ErrLinkNotFound)
}
