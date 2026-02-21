package service

import (
	"context"
	"crypto/md5"
	"fmt"

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

func (s *LinkService) Create(ctx context.Context, url, userID string) (*domain.Link, error) {
	code := GenerateCode(url)

	link, err := domain.NewLink(url, code, userID)
	if err != nil {
		return nil, err
	}

	if err = s.repo.Save(ctx, link); err != nil {
		return nil, err
	}

	err = s.cache.Set(ctx, code, link)
	return link, err
}

func (s *LinkService) Get(ctx context.Context, code string) (*domain.Link, error) {
	if cached, err := s.cache.Get(ctx, code); err == nil {
		return cached, nil
	}

	link, err := s.repo.Get(ctx, code)
	if err != nil {
		return nil, err
	}

	err = s.cache.Set(ctx, code, link)
	return link, err
}

func GenerateCode(url string) string {
	hash := md5.Sum([]byte(url))
	return fmt.Sprintf("%x", hash)[:6]
}
