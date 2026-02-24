package dto

import (
	"fmt"
	"strings"
)

type ShortenResponse struct {
	ShortURL string `json:"short_url"`
}

func NewShortenResponse(code, baseURL string) ShortenResponse {
	return ShortenResponse{
		ShortURL: fmt.Sprintf("%s/%s", strings.TrimSuffix(baseURL, "/"), code),
	}
}
