package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/StanislavYaroslavtsev/url-shortener/internal/entity"
)

type MemoryRepository struct {
	mutex sync.RWMutex
	urls  map[string]entity.URL
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		urls: make(map[string]entity.URL),
	}
}

func (repo *MemoryRepository) SaveURL(ctx context.Context, originalURL, shortCode, userID string) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	if _, exists := repo.urls[shortCode]; exists {
		return fmt.Errorf("short code already exists: %s", shortCode)
	}

	repo.urls[shortCode] = entity.URL{
		OriginalURL: originalURL,
		ShortCode:   shortCode,
		UserID:      userID,
		CreatedAt:   time.Now(),
	}
	return nil
}

func (repo *MemoryRepository) GetURL(ctx context.Context, shortCode string) (string, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	record, exists := repo.urls[shortCode]
	if !exists {
		return "", fmt.Errorf("URL not found for code: %s", shortCode)
	}

	return record.OriginalURL, nil
}

func (repo *MemoryRepository) DeleteURL(ctx context.Context, shortCode string) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	delete(repo.urls, shortCode)
	return nil
}
