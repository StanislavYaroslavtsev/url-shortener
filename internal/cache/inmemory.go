package cache

import (
	"context"
	"sync"
	"time"

	"github.com/StanislavYaroslavtsev/url-shortener/internal/domain"
)

type InMemoryCache struct {
	mu      sync.RWMutex
	store   map[string]Item
	ttl     time.Duration
	cleaner *Cleaner
}

type Item struct {
	link      *domain.Link
	expiresAt time.Time
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

func (cache *InMemoryCache) Set(_ context.Context, code string, link *domain.Link) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	cache.store[code] = Item{
		link:      link,
		expiresAt: time.Now().UTC().Add(cache.ttl),
	}
	return nil
}

func (cache *InMemoryCache) Get(_ context.Context, code string) (*domain.Link, error) {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	item, ok := cache.store[code]
	if !ok {
		return nil, ErrCacheMiss
	}

	if time.Now().UTC().After(item.expiresAt) {
		return nil, ErrCacheExpired
	}

	return item.link, nil
}

func (cache *InMemoryCache) deleteExpired() {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	now := time.Now().UTC()
	for key, item := range cache.store {
		if now.After(item.expiresAt) {
			delete(cache.store, key)
		}
	}
}

func (cache *InMemoryCache) Close() error {
	if cache.cleaner != nil {
		cache.cleaner.Stop()
	}
	return nil
}
