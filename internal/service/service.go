package service

import (
	"context"
	"crypto/md5"
	"fmt"

	"github.com/StanislavYaroslavtsev/url-shortener/internal/cache"
	"github.com/StanislavYaroslavtsev/url-shortener/internal/repository"
)

type UrlService struct {
	repo  repository.URLRepository
	cache cache.Cache
}

func NewUrlService(repo repository.URLRepository, cache cache.Cache) *UrlService {
	return &UrlService{
		repo:  repo,
		cache: cache,
	}
}

func (s *UrlService) ShortenURL(ctx context.Context, originalURL, userID string) (string, error) {
	shortCode := GenerateShortCode(originalURL)

	err := s.repo.SaveURL(ctx, originalURL, shortCode, userID)
	if err != nil {
		return "", err
	}

	err = s.cache.Set(ctx, shortCode, originalURL)
	return shortCode, err
}

func (s *UrlService) ExpandURL(ctx context.Context, shortCode string) (string, error) {
	if cached, err := s.cache.Get(ctx, shortCode); err == nil {
		return cached, nil
	}

	originalURL, err := s.repo.GetURL(ctx, shortCode)
	if err != nil {
		return "", err
	}

	err = s.cache.Set(ctx, shortCode, originalURL)
	return originalURL, err
}

func GenerateShortCode(url string) string {
	hash := md5.Sum([]byte(url))
	return fmt.Sprintf("%x", hash)[:6]
}
