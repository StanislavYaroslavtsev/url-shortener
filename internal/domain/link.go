package domain

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type Link struct {
	URL       string `validate:"required,url"`
	Code      string `validate:"required"`
	ExpiresAt *time.Time
	CreatedAt time.Time
}

func (l *Link) Validate() error {
	if err := validate.Struct(l); err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidLink, err.Error())
	}

	return nil
}

func NewLink(url, code string, expiresAt *time.Time) (*Link, error) {
	link := &Link{
		URL:       url,
		Code:      code,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now().UTC().Truncate(time.Microsecond),
	}

	if err := link.Validate(); err != nil {
		return nil, err
	}

	return link, nil
}

func (l *Link) IsExpired() bool {
	if l.ExpiresAt == nil {
		return false
	}
	return time.Now().UTC().After(*l.ExpiresAt)
}
