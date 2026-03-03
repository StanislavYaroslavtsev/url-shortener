package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/StanislavYaroslavtsev/url-shortener/services/url-shortener/internal/cache"
	domain2 "github.com/StanislavYaroslavtsev/url-shortener/services/url-shortener/internal/domain"
	repository2 "github.com/StanislavYaroslavtsev/url-shortener/services/url-shortener/internal/repository"
)

type LinkServiceInterface interface {
	Create(ctx context.Context, url string, alias *string, expiresAt *time.Time) (*domain2.Link, error)
	Get(ctx context.Context, code string) (*domain2.Link, error)
	Ping(ctx context.Context) map[string]error
}

type LinkService struct {
	repo  repository2.LinkRepository
	cache cache.Cache
}

func NewLinkService(repo repository2.LinkRepository, cache cache.Cache) *LinkService {
	return &LinkService{
		repo:  repo,
		cache: cache,
	}
}

const maxCodeGenerationAttempts = 10

func (s *LinkService) Create(ctx context.Context, url string, alias *string, expiresAt *time.Time) (*domain2.Link, error) {
	if alias != nil {
		link, err := domain2.NewLink(url, *alias, alias, expiresAt)
		if err != nil {
			return nil, err
		}

		if err = s.repo.Save(ctx, link); err != nil {
			return nil, err
		}

		if err = s.cache.Set(ctx, *alias, link); err != nil {
			slog.Warn("Failed to write to cache", "code", *alias, "error", err)
		}

		return link, nil
	}

	for range maxCodeGenerationAttempts {
		code, err := generateCode()
		if err != nil {
			return nil, err
		}

		link, err := domain2.NewLink(url, code, nil, expiresAt)
		if err != nil {
			return nil, err
		}

		if err = s.repo.Save(ctx, link); err != nil {
			if errors.Is(err, repository2.ErrCodeExists) {
				continue
			}
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

	return nil, fmt.Errorf("failed to generate unique code after %d attempts", maxCodeGenerationAttempts)
}

func (s *LinkService) Get(ctx context.Context, code string) (*domain2.Link, error) {
	if cached, err := s.cache.Get(ctx, code); err == nil {
		if cached.IsExpired() {
			return nil, domain2.ErrLinkExpired
		}
		return cached, nil
	}

	link, err := s.repo.Get(ctx, code)
	if err != nil {
		return nil, err
	}

	if link.IsExpired() {
		return nil, domain2.ErrLinkExpired
	}

	if err = s.cache.Set(ctx, code, link); err != nil {
		slog.Warn("Failed to write to cache",
			"code", code,
			"error", err,
		)
	}

	return link, nil
}

func (s *LinkService) Ping(ctx context.Context) map[string]error {
	return map[string]error{
		"database": s.repo.Ping(ctx),
		"cache":    s.cache.Ping(ctx),
	}
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateCode() (string, error) {
	result := make([]byte, 6)
	if _, err := rand.Read(result); err != nil {
		return "", fmt.Errorf("failed to generate random code: %w", err)
	}

	for i := range result {
		result[i] = charset[result[i]%byte(len(charset))]
	}

	return string(result), nil
}
