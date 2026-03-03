package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/StanislavYaroslavtsev/url-shortener/internal/domain"
	"github.com/StanislavYaroslavtsev/url-shortener/internal/http/dto"
	"github.com/StanislavYaroslavtsev/url-shortener/internal/repository"
	"github.com/StanislavYaroslavtsev/url-shortener/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	service  service.LinkServiceInterface
	baseURL  string
	validate *validator.Validate
}

func NewHandler(svc service.LinkServiceInterface, baseURL string) *Handler {
	return &Handler{
		service:  svc,
		baseURL:  baseURL,
		validate: validator.New(),
	}
}

func (h *Handler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	var req dto.ShortenRequest

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
			return
		}

		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	link, err := h.service.Create(r.Context(), req.URL, req.Alias, req.ExpiresAt)

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidLink):
			http.Error(w, "Invalid format", http.StatusBadRequest)
		case errors.Is(err, repository.ErrCodeExists):
			http.Error(w, "Alias already taken", http.StatusConflict)
		default:
			slog.Error("Unexpected error", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	resp := dto.NewShortenResponse(link.Code, h.baseURL)

	data, err := json.Marshal(resp)
	if err != nil {
		slog.Error("failed to marshal response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(data)
}

func (h *Handler) RedirectURL(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "id")

	link, err := h.service.Get(r.Context(), code)

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrLinkExpired):
			http.Error(w, "Link expired", http.StatusGone)
		case errors.Is(err, repository.ErrLinkNotFound):
			http.Error(w, "Link not found", http.StatusNotFound)
		default:
			slog.Error("Unexpected error", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	http.Redirect(w, r, link.URL, http.StatusFound)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	pings := h.service.Ping(r.Context())

	deps := make(map[string]dto.HealthDependency, len(pings))
	overall := "ok"

	for name, err := range pings {
		if err != nil {
			deps[name] = dto.HealthDependency{Status: "unavailable"}
			overall = "unavailable"
		} else {
			deps[name] = dto.HealthDependency{Status: "ok"}
		}
	}

	resp := dto.HealthResponse{
		Status:       overall,
		Dependencies: deps,
	}

	code := http.StatusOK
	if overall != "ok" {
		code = http.StatusServiceUnavailable
	}

	data, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(data)
}
