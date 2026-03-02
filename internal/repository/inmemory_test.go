package repository

import (
	"context"
	"testing"

	"github.com/StanislavYaroslavtsev/url-shortener/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryRepository_Save_SavesLink(t *testing.T) {
	repo := NewInMemoryRepository()
	ctx := context.Background()
	link, err := domain.NewLink("https://google.com", "abc123", nil)
	require.NoError(t, err)

	err = repo.Save(ctx, link)
	require.NoError(t, err)

	saved, err := repo.Get(ctx, "abc123")
	require.NoError(t, err)
	assert.Equal(t, link, saved)
}

func TestInMemoryRepository_Save_CodeExists_ReturnsError(t *testing.T) {
	repo := NewInMemoryRepository()
	ctx := context.Background()
	link, err := domain.NewLink("https://google.com", "abc123", nil)
	require.NoError(t, err)

	err = repo.Save(ctx, link)
	require.NoError(t, err)

	err = repo.Save(ctx, link)
	assert.ErrorIs(t, err, ErrCodeExists)
}

func TestInMemoryRepository_Get_ReturnsLink(t *testing.T) {
	repo := NewInMemoryRepository()
	ctx := context.Background()
	link, err := domain.NewLink("https://google.com", "abc123", nil)
	require.NoError(t, err)

	err = repo.Save(ctx, link)
	require.NoError(t, err)

	saved, err := repo.Get(ctx, "abc123")
	require.NoError(t, err)
	assert.Equal(t, link, saved)
}

func TestInMemoryRepository_Get_NoSuchCode_ReturnsError(t *testing.T) {
	repo := NewInMemoryRepository()
	ctx := context.Background()

	_, err := repo.Get(ctx, "abc123")
	assert.ErrorIs(t, err, ErrLinkNotFound)
}

func TestInMemoryRepository_Delete_RemovesLink(t *testing.T) {
	repo := NewInMemoryRepository()
	ctx := context.Background()
	link, err := domain.NewLink("https://google.com", "abc123", nil)
	require.NoError(t, err)

	err = repo.Save(ctx, link)
	require.NoError(t, err)

	err = repo.Delete(ctx, "abc123")
	require.NoError(t, err)

	_, err = repo.Get(ctx, "abc123")
	assert.ErrorIs(t, err, ErrLinkNotFound)
}

func TestInMemoryRepository_Delete_NoSuchCode_ReturnsError(t *testing.T) {
	repo := NewInMemoryRepository()
	ctx := context.Background()

	err := repo.Delete(ctx, "abc123")
	assert.ErrorIs(t, err, ErrLinkNotFound)
}
