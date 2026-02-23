package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/StanislavYaroslavtsev/url-shortener/config"
	"github.com/StanislavYaroslavtsev/url-shortener/internal/domain"
	_ "github.com/lib/pq"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(cfg *config.Config) (*PostgresRepository, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres connection: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	return &PostgresRepository{db: db}, nil
}

func (r *PostgresRepository) Save(ctx context.Context, link *domain.Link) error {
	query := `INSERT INTO links (code, url, created_at) VALUES ($1, $2, $3)`
	_, err := r.db.ExecContext(ctx, query, link.Code, link.URL, link.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to save link: %w", err)
	}
	return nil
}

func (r *PostgresRepository) Get(ctx context.Context, code string) (*domain.Link, error) {
	query := `SELECT code, url, created_at FROM links WHERE code = $1`
	row := r.db.QueryRowContext(ctx, query, code)

	var link domain.Link
	if err := row.Scan(&link.Code, &link.URL, &link.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrLinkNotFound
		}
		return nil, fmt.Errorf("failed to get link: %w", err)
	}

	return &link, nil
}

func (r *PostgresRepository) Delete(ctx context.Context, code string) error {
	query := `DELETE FROM links WHERE code = $1`
	result, err := r.db.ExecContext(ctx, query, code)
	if err != nil {
		return fmt.Errorf("failed to delete link: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrLinkNotFound
	}

	return nil
}

func (r *PostgresRepository) FindByURL(ctx context.Context, url string) (*domain.Link, error) {
	query := `SELECT code, url, created_at FROM links WHERE url = $1`
	row := r.db.QueryRowContext(ctx, query, url)

	var link domain.Link
	if err := row.Scan(&link.Code, &link.URL, &link.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrLinkNotFound
		}
		return nil, fmt.Errorf("failed to find link by url: %w", err)
	}

	return &link, nil
}

func (r *PostgresRepository) Close() error {
	return r.db.Close()
}
