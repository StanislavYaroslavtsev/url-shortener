package repository

import (
	"context"

	"github.com/StanislavYaroslavtsev/url-shortener/internal/domain"
)

type LinkRepository interface {
	Save(ctx context.Context, link *domain.Link) error
	Get(ctx context.Context, code string) (*domain.Link, error)
	Delete(ctx context.Context, code string) error
	FindByURL(ctx context.Context, url string) (*domain.Link, error)
}
