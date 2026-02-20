package entity

import "time"

type Link struct {
	URL       string
	Code      string
	UserID    string
	CreatedAt time.Time
}
