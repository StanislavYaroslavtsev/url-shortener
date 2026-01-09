package handler

import (
	"encoding/json"
	"net/http"

	"github.com/StanislavYaroslavtsev/url-shortener/config"
	"github.com/StanislavYaroslavtsev/url-shortener/internal/http/dto"
	"github.com/StanislavYaroslavtsev/url-shortener/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	Service   *service.UrlService
	Config    *config.Config
	Validator *validator.Validate
}

func NewHandler(svc *service.UrlService, config *config.Config) *Handler {
	return &Handler{
		Service:   svc,
		Config:    config,
		Validator: validator.New(),
	}
}

func (h *Handler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	var req dto.ShortenRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.Validator.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	shortKey, err := h.Service.ShortenURL(r.Context(), req.URL, r.RemoteAddr)

	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(dto.ShortenResponse{
		ShortURL: h.Config.App.BaseURL + "/" + shortKey,
	})
}

func (h *Handler) RedirectURL(w http.ResponseWriter, r *http.Request) {
	shortKey := chi.URLParam(r, "id")

	original, err := h.Service.ExpandURL(r.Context(), shortKey)

	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, original, http.StatusFound)
}
