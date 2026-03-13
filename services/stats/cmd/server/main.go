package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/StanislavYaroslavtsev/url-shortener/services/stats/config"
	"github.com/StanislavYaroslavtsev/url-shortener/services/stats/internal/consumer"
	grpcserver "github.com/StanislavYaroslavtsev/url-shortener/services/stats/internal/grpc"
	"github.com/StanislavYaroslavtsev/url-shortener/services/stats/internal/repository"
	pb "github.com/StanislavYaroslavtsev/url-shortener/shared/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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

	grpcSrv := grpc.NewServer()
	pb.RegisterStatsServiceServer(grpcSrv, grpcserver.NewStatsServer(repo))
	if cfg.App.Env == "dev" {
		reflection.Register(grpcSrv)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.Port))
	if err != nil {
		slog.Error("Failed to listen", "error", err)
		os.Exit(1)
	}

	go func() {
		slog.Info("Starting gRPC server", "port", cfg.GRPC.Port)
		if err = grpcSrv.Serve(lis); err != nil {
			slog.Error("gRPC server failed", "error", err)
		}
	}()

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

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.Start(ctx)
	}()

	<-quit
	slog.Info("Shutting down...")
	cancel()
	grpcSrv.GracefulStop()
	wg.Wait()
	slog.Info("Consumer stopped")
}
