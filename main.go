package main

import (
	"context"
	"errors"
	"fmt"
	"log"
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
	memoryRepository := repository.NewMemoryRepository()
	memoryCache := cache.NewMemoryCache()
	svc := service.NewUrlService(memoryRepository, memoryCache)

	ctx := context.Background()
	cfg := config.GetConfig()

	shortCode, err := svc.ShortenURL(ctx, "https://google.com/", "123")
	if err != nil {
		log.Fatalf("Failed to shorten URL: %v", err)
	}
	fmt.Println(shortCode)

	originalURL, err := svc.ExpandURL(ctx, shortCode)
	if err != nil {
		log.Fatalf("Failed to expand URL: %v", err)
	}
	fmt.Println(originalURL)

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

	log.Printf("Starting server on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Server failed: %v", err)
	}
}
