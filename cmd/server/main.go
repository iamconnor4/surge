package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"log/slog"
	"strings"

	"github.com/iamconnor4/surge/internal/config"
	"github.com/iamconnor4/surge/internal/platform/postgres"
	"github.com/iamconnor4/surge/internal/platform/redis"
	"github.com/iamconnor4/surge/internal/web"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	err := run(ctx)

	stop()

	if err != nil {
		fmt.Fprintf(os.Stderr, "application error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config failed to load: %w", err)
	}

	setupLogger(cfg.LogLevel)

	postgresClient, err := postgres.NewClient(ctx, postgres.Config{
		Host:     cfg.Postgres.Host,
		Port:     cfg.Postgres.Port,
		User:     cfg.Postgres.User,
		Password: cfg.Postgres.Password,
		DB:       cfg.Postgres.DB,
		SSLMode:  cfg.Postgres.SSLMode,
	})
	if err != nil {
		return fmt.Errorf("postgres client failed to initalise: %w", err)
	}
	defer postgresClient.Close()

	redisClient, err := redis.NewClient(ctx, redis.Config{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err != nil {
		return fmt.Errorf("redis client failed to initialise: %w", err)
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			slog.Error("failed to close redis client", "error", err)
		}
	}()

	server := web.New(
		web.Config{
			Address: ":" + strconv.Itoa(cfg.Port),
		},
		web.Dependencies{
			Postgres: postgresClient,
			Redis:    redisClient,
		},
	)

	slog.InfoContext(ctx, "starting surge",
		"environment", cfg.Env,
		"address", server.Address(),
	)

	return server.Serve(ctx)
}

func setupLogger(levelStr string) {
	var level slog.Level
	switch strings.ToLower(levelStr) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     level,
		AddSource: level == slog.LevelDebug,
	})

	slog.SetDefault(slog.New(handler))

}
