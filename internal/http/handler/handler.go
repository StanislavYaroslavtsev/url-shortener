package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/StanislavYaroslavtsev/url-shortener/config"
	"github.com/StanislavYaroslavtsev/url-shortener/internal/http/dto"
	"github.com/StanislavYaroslavtsev/url-shortener/internal/repository"
	"github.com/StanislavYaroslavtsev/url-shortener/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	Service *service.LinkService
	Config  *config.Config
}

func NewHandler(svc *service.LinkService, config *config.Config) *Handler {
	return &Handler{
		Service: svc,
		Config:  config,
	}
}

func (h *Handler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	var req dto.ShortenRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	link, err := h.Service.Create(r.Context(), req.URL)

	if err != nil {
		if _, ok := errors.AsType[validator.ValidationErrors](err); ok {
			http.Error(w, "Invalid URL format", http.StatusBadRequest)
			return
		}

		switch {
		case errors.Is(err, repository.ErrCodeExists):
			http.Error(w, "Code already exists, try again", http.StatusConflict)
		default:
			log.Printf("Unexpected error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	resp := dto.NewShortenResponse(link, h.Config.App.BaseURL)
	if err = json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

func (h *Handler) RedirectURL(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "id")

	link, err := h.Service.Get(r.Context(), code)

	if err != nil {
		switch {
		case errors.Is(err, repository.ErrLinkNotFound):
			http.Error(w, "Link not found", http.StatusNotFound)
		default:
			log.Printf("Unexpected error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	http.Redirect(w, r, link.URL, http.StatusFound)
}
