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
	memRepo := repository.NewInMemoryRepository()
	memCache := cache.NewInMemoryCache(24*time.Hour, 1*time.Minute)
	svc := service.NewLinkService(memRepo, memCache)

	ctx := context.Background()
	cfg := config.GetConfig()

	link, err := svc.Create(ctx, "https://google.com/", "123")
	if err != nil {
		log.Fatalf("Failed to shorten URL: %v", err)
	}
	fmt.Println(link.Code)

	receivedLink, err := svc.Get(ctx, link.Code)
	if err != nil {
		log.Fatalf("Failed to get URL: %v", err)
	}
	fmt.Println(receivedLink.URL)

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
