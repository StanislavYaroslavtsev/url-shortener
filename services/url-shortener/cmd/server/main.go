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
	"time"

	"github.com/StanislavYaroslavtsev/url-shortener/services/url-shortener/config"
	cache2 "github.com/StanislavYaroslavtsev/url-shortener/services/url-shortener/internal/cache"
	"github.com/StanislavYaroslavtsev/url-shortener/services/url-shortener/internal/http/handler"
	repository2 "github.com/StanislavYaroslavtsev/url-shortener/services/url-shortener/internal/repository"
	"github.com/StanislavYaroslavtsev/url-shortener/services/url-shortener/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
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

	var appCache cache2.Cache

	if cfg.Cache.UseRedis {
		redisCache, err := cache2.NewRedisCache(cfg.Cache.RedisAddr, cfg.Cache.TTL)
		if err != nil {
			slog.Error("Failed to connect to redis", "error", err)
			os.Exit(1)
		}

		slog.Info("Using Redis cache")
		appCache = redisCache
	} else {
		slog.Info("Using in-memory cache")
		appCache = cache2.NewInMemoryCache(cfg.Cache.TTL, cfg.Cache.CleanupInterval)
	}

	svc := service.NewLinkService(repo, appCache)
	h := handler.NewHandler(svc, cfg.App.BaseURL)

	router := chi.NewRouter()

	// Middleware
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Timeout(cfg.Server.HandlerTimeout))
	router.Use(httprate.LimitByIP(100, time.Minute))

	// Routes
	router.Post("/shorten", h.ShortenURL)
	router.Get("/{id}", h.RedirectURL)
	router.Get("/health", h.Health)

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

	if err = appCache.Close(); err != nil {
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

func initRepository(cfg config.DatabaseConfig) (repository2.LinkRepository, error) {
	if cfg.UsePostgres {
		repo, err := repository2.NewPostgresRepository(cfg)

		if err != nil {
			return nil, err
		}

		slog.Info("Using PostgreSQL repository")
		return repo, nil
	}

	slog.Info("Using in-memory repository")
	return repository2.NewInMemoryRepository(), nil
}
