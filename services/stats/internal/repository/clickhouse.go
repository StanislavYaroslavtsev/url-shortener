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

func (r *ClickHouseRepository) GetStats(ctx context.Context, code string) (*StatsResult, error) {
	var result StatsResult
	var topCountries []CountryStats

	rows, err := r.conn.Query(ctx, `
		SELECT 
			country, 
			COUNT() AS clicks
		FROM click_events
		WHERE code = ?
		GROUP BY country
		ORDER BY clicks DESC
		LIMIT 10`, code)

	if err != nil {
		return nil, fmt.Errorf("failed to query top countries: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var cs CountryStats
		if err = rows.Scan(&cs.Country, &cs.Clicks); err != nil {
			return nil, fmt.Errorf("failed to scan country stats: %w", err)
		}

		topCountries = append(topCountries, cs)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate rows: %w", err)
	}

	err = r.conn.QueryRow(ctx, `
		SELECT 
			COUNT(), 
			MAX(clicked_at)
		FROM click_events
		WHERE code = ?`, code).Scan(&result.TotalClicks, &result.LastClickedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to query total clicks and last clicked at: %w", err)
	}

	result.TopCountries = topCountries
	return &result, nil
}

func (r *ClickHouseRepository) Close() error {
	return r.conn.Close()
}
