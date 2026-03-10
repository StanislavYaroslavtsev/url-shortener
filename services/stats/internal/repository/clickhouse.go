package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/StanislavYaroslavtsev/url-shortener/services/stats/internal/domain"
)

type ClickHouseRepository struct {
	conn clickhouse.Conn
}

func NewClickHouseRepository(addr, database, user, password string) (*ClickHouseRepository, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Database: database,
			Username: user,
			Password: password,
		},
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to clickhouse: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping clickhouse: %w", err)
	}

	return &ClickHouseRepository{conn: conn}, nil
}

func (r *ClickHouseRepository) Save(ctx context.Context, event *domain.ClickEvent) error {
	if err := r.conn.Exec(ctx, `
        INSERT INTO click_events (code, ip, country, clicked_at)
        VALUES (?, ?, ?, ?)`,
		event.Code,
		event.IP,
		event.Country,
		event.ClickedAt,
	); err != nil {
		return fmt.Errorf("failed to save click event: %w", err)
	}

	return nil
}

func (r *ClickHouseRepository) Close() error {
	return r.conn.Close()
}
