package handler

import (
	"encoding/json"
	"net/http"

	"github.com/StanislavYaroslavtsev/url-shortener/internal/http/dto"
	"github.com/StanislavYaroslavtsev/url-shortener/internal/service"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	Service *service.UrlService
}

func (h *Handler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	var req dto.ShortenRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	shortKey, err := h.Service.ShortenURL(r.Context(), req.URL, r.RemoteAddr)

	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Replace with configurable base URL
	baseURL := "http://localhost:3000"
	_ = json.NewEncoder(w).Encode(dto.ShortenResponse{
		ShortURL: baseURL + shortKey,
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
