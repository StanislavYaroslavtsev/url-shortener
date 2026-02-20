package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/StanislavYaroslavtsev/url-shortener/internal/entity"
)

type InMemoryRepository struct {
	mu    sync.RWMutex
	links map[string]entity.Link
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		links: make(map[string]entity.Link),
	}
}

func (repo *InMemoryRepository) Save(ctx context.Context, url, code, userID string) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	if _, exists := repo.links[code]; exists {
		return fmt.Errorf("code already exists: %s", code)
	}

	repo.links[code] = entity.Link{
		URL:       url,
		Code:      code,
		UserID:    userID,
		CreatedAt: time.Now(),
	}
	return nil
}

func (repo *InMemoryRepository) Get(ctx context.Context, code string) (string, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	record, exists := repo.links[code]
	if !exists {
		return "", fmt.Errorf("link not found for code: %s", code)
	}

	return record.URL, nil
}

func (repo *InMemoryRepository) Delete(ctx context.Context, code string) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	delete(repo.links, code)
	return nil
}
