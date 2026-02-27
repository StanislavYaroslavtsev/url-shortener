package dto

import "time"

type ShortenRequest struct {
	URL       string     `json:"url" validate:"required,url"`
	ExpiresAt *time.Time `json:"expires_at"`
}
