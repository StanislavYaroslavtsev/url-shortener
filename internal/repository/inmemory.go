package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/StanislavYaroslavtsev/url-shortener/internal/domain"
)

type InMemoryRepository struct {
	mu       sync.RWMutex
	links    map[string]domain.Link
	urlIndex map[string]string
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		links:    make(map[string]domain.Link),
		urlIndex: make(map[string]string),
	}
}

func (repo *InMemoryRepository) Save(ctx context.Context, link *domain.Link) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	if _, exists := repo.links[link.Code]; exists {
		return ErrCodeExists
	}

	repo.links[link.Code] = *link
	repo.urlIndex[link.URL] = link.Code
	return nil
}

func (repo *InMemoryRepository) Get(ctx context.Context, code string) (*domain.Link, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	link, exists := repo.links[code]
	if !exists {
		return nil, ErrLinkNotFound
	}

	return &link, nil
}

func (repo *InMemoryRepository) Delete(ctx context.Context, code string) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	link, exists := repo.links[code]
	if !exists {
		return ErrLinkNotFound
	}

	delete(repo.links, code)
	delete(repo.urlIndex, link.URL)

	return nil
}

func (repo *InMemoryRepository) FindByURL(ctx context.Context, url string) (*domain.Link, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	code, exists := repo.urlIndex[url]
	if !exists {
		return nil, ErrLinkNotFound
	}

	link, exists := repo.links[code]
	if !exists {
		return nil, fmt.Errorf("inconsistent index: url %s points to missing code %s", url, code)
	}

	return &link, nil
}
