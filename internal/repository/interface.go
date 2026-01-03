package repository

import "context"

type URLRepository interface {
	SaveURL(ctx context.Context, originalURL, shortCode, userID string) error
	GetURL(ctx context.Context, shortCode string) (string, error)
	DeleteURL(ctx context.Context, shortCode string) error
}
