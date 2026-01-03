package entity

import "time"

type URL struct {
	OriginalURL string
	ShortCode   string
	UserID      string
	CreatedAt   time.Time
}
