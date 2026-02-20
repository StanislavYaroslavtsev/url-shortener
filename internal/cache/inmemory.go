package cache

import (
	"context"
	"fmt"
	"sync"
)

type InMemoryCache struct {
	mu    sync.RWMutex
	store map[string]string
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{
		store: make(map[string]string),
	}
}

func (cache *InMemoryCache) Set(ctx context.Context, key, value string) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	cache.store[key] = value
	return nil
}

func (cache *InMemoryCache) Get(ctx context.Context, key string) (string, error) {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	val, ok := cache.store[key]
	if !ok {
		return "", fmt.Errorf("key '%s' not found", key)
	}
	return val, nil
}
