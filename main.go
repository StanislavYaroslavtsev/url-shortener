package main

import (
	"context"
	"fmt"
	"log"

	"github.com/StanislavYaroslavtsev/url-shortener/internal/cache"
	"github.com/StanislavYaroslavtsev/url-shortener/internal/repository"
	"github.com/StanislavYaroslavtsev/url-shortener/internal/service"
)

func main() {
	memoryRepository := repository.NewMemoryRepository()
	memoryCache := cache.NewMemoryCache()
	svc := service.NewUrlService(memoryRepository, memoryCache)

	ctx := context.Background()

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
}
