package main

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/StanislavYaroslavtsev/url-shortener/config"
	"github.com/StanislavYaroslavtsev/url-shortener/internal/cache"
	"github.com/StanislavYaroslavtsev/url-shortener/internal/http/handler"
	"github.com/StanislavYaroslavtsev/url-shortener/internal/repository"
	"github.com/StanislavYaroslavtsev/url-shortener/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg := config.GetConfig()

	var repo repository.LinkRepository

	if cfg.Database.UsePostgres {
		pgRepo, err := repository.NewPostgresRepository(cfg)
		if err != nil {
			slog.Error("Failed to connect to postgres", "error", err)
			panic(err)
		}

		slog.Info("Using PostgreSQL repository")
		repo = pgRepo
	} else {
		slog.Info("Using in-memory repository")
		repo = repository.NewInMemoryRepository()
	}

	memCache := cache.NewInMemoryCache(24*time.Hour, 1*time.Minute)
	svc := service.NewLinkService(repo, memCache)

	router := chi.NewRouter()
	h := handler.NewHandler(svc, cfg)

	// Middleware
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Timeout(60 * time.Second))

	// Routes
	router.Post("/shorten", h.ShortenURL)
	router.Get("/{id}", h.RedirectURL)

	server := &http.Server{
		Addr:    h.Config.Server.Host + ":" + strconv.Itoa(h.Config.Server.Port),
		Handler: router,
	}

	slog.Info("Starting server",
		"address", server.Addr,
	)

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("Server failed", "error", err)
	}
}
