package dto

import (
	"fmt"
	"strings"

	"github.com/StanislavYaroslavtsev/url-shortener/internal/domain"
)

type ShortenResponse struct {
	ShortURL string `json:"short_url"`
}

func NewShortenResponse(link *domain.Link, baseURL string) ShortenResponse {
	return ShortenResponse{
		ShortURL: fmt.Sprintf("%s/%s", strings.TrimSuffix(baseURL, "/"), link.Code),
	}
}
