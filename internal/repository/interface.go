package repository

import "context"

type LinkRepository interface {
	Save(ctx context.Context, url, code, userID string) error
	Get(ctx context.Context, code string) (string, error)
	Delete(ctx context.Context, code string) error
}
