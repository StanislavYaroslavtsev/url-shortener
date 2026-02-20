package cache

import (
	"context"
	"fmt"
	"sync"
)

type MemoryCache struct {
	mu    sync.RWMutex
	store map[string]string
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		store: make(map[string]string),
	}
}

func (cache *MemoryCache) Set(ctx context.Context, key, value string) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	cache.store[key] = value
	return nil
}

func (cache *MemoryCache) Get(ctx context.Context, key string) (string, error) {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	val, ok := cache.store[key]
	if !ok {
		return "", fmt.Errorf("key '%s' not found", key)
	}
	return val, nil
}
