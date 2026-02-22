package domain

import (
	"time"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type Link struct {
	URL       string `validate:"required,url"`
	Code      string `validate:"required"`
	CreatedAt time.Time
}

func (l *Link) Validate() error {
	return validate.Struct(l)
}

func NewLink(url, code string) (*Link, error) {
	link := &Link{
		URL:       url,
		Code:      code,
		CreatedAt: time.Now(),
	}

	if err := link.Validate(); err != nil {
		return nil, err
	}

	return link, nil
}
