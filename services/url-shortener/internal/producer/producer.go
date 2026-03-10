package producer

import (
	"context"

	"github.com/StanislavYaroslavtsev/url-shortener/services/url-shortener/internal/domain"
)

type ClickProducer interface {
	SendClickEvent(ctx context.Context, link *domain.Link, ip string) error
	Close() error
}
