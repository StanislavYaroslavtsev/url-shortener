package repository

import (
	"context"
	"time"

	"github.com/StanislavYaroslavtsev/url-shortener/services/stats/internal/domain"
)

type StatsResult struct {
	TotalClicks   uint64
	LastClickedAt time.Time
	TopCountries  []CountryStats
}

type CountryStats struct {
	Country string
	Clicks  uint64
}

type EventRepository interface {
	Save(ctx context.Context, event *domain.ClickEvent) error
	GetStats(ctx context.Context, code string) (*StatsResult, error)
	Close() error
}
