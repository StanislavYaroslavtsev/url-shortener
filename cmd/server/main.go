package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/StanislavYaroslavtsev/url-shortener/config"
	"github.com/StanislavYaroslavtsev/url-shortener/internal/cache"
	"github.com/StanislavYaroslavtsev/url-shortener/internal/http/handler"
	"github.com/StanislavYaroslavtsev/url-shortener/internal/repository"
	"github.com/StanislavYaroslavtsev/url-shortener/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg, err := config.Init()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	repo, err := initRepository(cfg.Database)
	if err != nil {
		slog.Error("Failed to initialize repository", "error", err)
		os.Exit(1)
	}

	memCache := cache.NewInMemoryCache(cfg.Cache.TTL, cfg.Cache.CleanupInterval)

	svc := service.NewLinkService(repo, memCache)
	h := handler.NewHandler(svc, cfg.App.BaseURL)

	router := chi.NewRouter()

	// Middleware
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Timeout(cfg.Server.HandlerTimeout))

	// Routes
	router.Post("/shorten", h.ShortenURL)
	router.Get("/{id}", h.RedirectURL)

	addr := cfg.Server.Host + ":" + strconv.Itoa(cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	errChan := make(chan error, 1)

	go func() {
		slog.Info("Starting server", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		slog.Error("Server failed", "error", err)
	case <-quit:
		slog.Info("Shutting down server...")
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err = server.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	if err = memCache.Close(); err != nil {
		slog.Error("Failed to close cache", "error", err)
	} else {
		slog.Info("Cache closed")
	}

	if err = repo.Close(); err != nil {
		slog.Error("Failed to close repository", "error", err)
	} else {
		slog.Info("Database connection closed")
	}

	slog.Info("Server stopped")
}

func initRepository(cfg config.DatabaseConfig) (repository.LinkRepository, error) {
	if cfg.UsePostgres {
		repo, err := repository.NewPostgresRepository(cfg)

		if err != nil {
			return nil, err
		}

		slog.Info("Using PostgreSQL repository")
		return repo, nil
	}

	slog.Info("Using in-memory repository")
	return repository.NewInMemoryRepository(), nil
}
