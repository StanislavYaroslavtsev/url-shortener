package dto

type ShortenRequest struct {
	URL string `json:"url" validate:"required,url"`
}
