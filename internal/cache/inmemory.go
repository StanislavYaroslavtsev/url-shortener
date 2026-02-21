package cache

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type InMemoryCache struct {
	mu      sync.RWMutex
	store   map[string]Item
	ttl     time.Duration
	cleaner *Cleaner
}

type Item struct {
	value     string
	expiresAt int64
}

func NewInMemoryCache(ttl time.Duration, cleanupInterval time.Duration) *InMemoryCache {
	cache := &InMemoryCache{
		store: make(map[string]Item),
		ttl:   ttl,
	}

	cleaner := NewCleaner(cleanupInterval, cache.deleteExpired)
	cleaner.Start()
	cache.cleaner = cleaner

	return cache
}

func (cache *InMemoryCache) Set(ctx context.Context, key, value string) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	cache.store[key] = Item{
		value:     value,
		expiresAt: time.Now().Add(cache.ttl).Unix(),
	}
	return nil
}

func (cache *InMemoryCache) Get(ctx context.Context, key string) (string, error) {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	item, ok := cache.store[key]
	if !ok {
		return "", fmt.Errorf("key '%s' not found", key)
	}

	if time.Now().Unix() > item.expiresAt {
		return "", fmt.Errorf("key '%s' expired", key)
	}

	return item.value, nil
}

func (cache *InMemoryCache) deleteExpired() {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	now := time.Now().Unix()
	for key, item := range cache.store {
		if now > item.expiresAt {
			delete(cache.store, key)
		}
	}
}

func (cache *InMemoryCache) Close() {
	if cache.cleaner != nil {
		cache.cleaner.Stop()
	}
}
