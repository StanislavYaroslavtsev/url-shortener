package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/StanislavYaroslavtsev/url-shortener/services/stats/config"
	"github.com/StanislavYaroslavtsev/url-shortener/services/stats/internal/consumer"
	"github.com/StanislavYaroslavtsev/url-shortener/services/stats/internal/repository"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Init()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	repo, err := repository.NewClickHouseRepository(
		cfg.ClickHouse.Addr,
		cfg.ClickHouse.Database,
		cfg.ClickHouse.User,
		cfg.ClickHouse.Password,
	)
	if err != nil {
		slog.Error("Failed to connect to ClickHouse", "error", err)
		os.Exit(1)
	}
	defer repo.Close()

	slog.Info("Connected to ClickHouse")

	c := consumer.NewKafkaConsumer(
		cfg.Kafka.Addr,
		cfg.Kafka.Topic,
		cfg.Kafka.GroupID,
		repo,
	)
	defer c.Close()

	slog.Info("Starting Kafka consumer")

	ctx, cancel := context.WithCancel(context.Background())

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go c.Start(ctx)

	<-quit
	slog.Info("Shutting down...")
	cancel()
}
