package cache

import (
	"context"
	"fmt"
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

func (cache *InMemoryCache) Set(ctx context.Context, code string, link *domain.Link) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	cache.store[code] = Item{
		link:      link,
		expiresAt: time.Now().Add(cache.ttl).Unix(),
	}
	return nil
}

func (cache *InMemoryCache) Get(ctx context.Context, code string) (*domain.Link, error) {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	item, ok := cache.store[code]
	if !ok {
		return nil, fmt.Errorf("code '%s' not found", code)
	}

	if time.Now().Unix() > item.expiresAt {
		return nil, fmt.Errorf("code '%s' expired", code)
	}

	return item.link, nil
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
