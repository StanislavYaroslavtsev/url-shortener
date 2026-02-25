package repository

import (
	"context"
	"sync"

	"github.com/StanislavYaroslavtsev/url-shortener/internal/domain"
)

type InMemoryRepository struct {
	mu    sync.RWMutex
	links map[string]domain.Link
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		links: make(map[string]domain.Link),
	}
}

func (repo *InMemoryRepository) Save(_ context.Context, link *domain.Link) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	if _, exists := repo.links[link.Code]; exists {
		return ErrCodeExists
	}

	repo.links[link.Code] = *link
	return nil
}

func (repo *InMemoryRepository) Get(_ context.Context, code string) (*domain.Link, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	link, exists := repo.links[code]
	if !exists {
		return nil, ErrLinkNotFound
	}

	return &link, nil
}

func (repo *InMemoryRepository) Delete(_ context.Context, code string) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	_, exists := repo.links[code]
	if !exists {
		return ErrLinkNotFound
	}

	delete(repo.links, code)
	return nil
}
