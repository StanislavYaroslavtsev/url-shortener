package repository

import (
	"context"

	"github.com/StanislavYaroslavtsev/url-shortener/services/stats/internal/domain"
)

type EventRepository interface {
	Save(ctx context.Context, event *domain.ClickEvent) error
	Close() error
}
