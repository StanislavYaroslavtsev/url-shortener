package service

import (
	"context"
	"crypto/md5"
	"fmt"
	"log/slog"

	"github.com/StanislavYaroslavtsev/url-shortener/internal/cache"
	"github.com/StanislavYaroslavtsev/url-shortener/internal/domain"
	"github.com/StanislavYaroslavtsev/url-shortener/internal/repository"
)

type LinkService struct {
	repo  repository.LinkRepository
	cache cache.Cache
}

func NewLinkService(repo repository.LinkRepository, cache cache.Cache) *LinkService {
	return &LinkService{
		repo:  repo,
		cache: cache,
	}
}

func (s *LinkService) Create(ctx context.Context, url string) (*domain.Link, error) {
	if existing, err := s.repo.FindByURL(ctx, url); err == nil {
		_ = s.cache.Set(ctx, existing.Code, existing)
		return existing, nil
	}

	code := GenerateCode(url)

	link, err := domain.NewLink(url, code)
	if err != nil {
		return nil, err
	}

	if err = s.repo.Save(ctx, link); err != nil {
		return nil, err
	}

	if err = s.cache.Set(ctx, code, link); err != nil {
		slog.Warn("Failed to write to cache",
			"code", code,
			"error", err,
		)
	}

	return link, nil
}

func (s *LinkService) Get(ctx context.Context, code string) (*domain.Link, error) {
	if cached, err := s.cache.Get(ctx, code); err == nil {
		return cached, nil
	}

	link, err := s.repo.Get(ctx, code)
	if err != nil {
		return nil, err
	}

	if err = s.cache.Set(ctx, code, link); err != nil {
		slog.Warn("Failed to write to cache",
			"code", code,
			"error", err,
		)
	}

	return link, nil
}

func GenerateCode(url string) string {
	hash := md5.Sum([]byte(url))
	return fmt.Sprintf("%x", hash)[:6]
}
