package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"log/slog"
	"strings"

	"github.com/lmittmann/tint"

	"github.com/iamconnor4/surge/internal/config"
	"github.com/iamconnor4/surge/internal/platform/postgres"
	"github.com/iamconnor4/surge/internal/platform/postgres/db"
	"github.com/iamconnor4/surge/internal/platform/redis"
	"github.com/iamconnor4/surge/internal/service"
	"github.com/iamconnor4/surge/internal/web"
)

func main() {
	// Use production environment JSON logging until
	// environment configuration confirmed.
	setupBootstrapLogger(slog.LevelInfo)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	err := run(ctx)

	stop()

	if err != nil {
		slog.ErrorContext(
			context.Background(),
			"application error",
			"error", err,
		)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	level, err := parseLogLevel(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("parse log level: %w", err)
	}

	setupLogger(cfg.Env, level)

	postgresClient, err := postgres.NewClient(ctx, postgres.Config{
		Host:     cfg.Postgres.Host,
		Port:     cfg.Postgres.Port,
		User:     cfg.Postgres.User,
		Password: cfg.Postgres.Password,
		DB:       cfg.Postgres.DB,
		SSLMode:  cfg.Postgres.SSLMode,
	})
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer postgresClient.Close()

	redisClient, err := redis.NewClient(ctx, redis.Config{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err != nil {
		return fmt.Errorf("connect redis: %w", err)
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			slog.Error("failed to close redis client", "error", err)
		}
	}()

	queries := db.New(postgresClient.Pool)

	userService := service.NewUser(queries)
	venueService := service.NewVenue(queries)
	venueSectionService := service.NewVenueSection(queries)

	server := web.New(
		web.Config{
			Address:     ":" + strconv.Itoa(cfg.Port),
			Service:     "surge",
			Environment: cfg.Env,
		},
		web.Dependencies{
			Users:         userService,
			Venues:        venueService,
			VenueSections: venueSectionService,

			HealthChecks: map[string]web.Pinger{
				"postgres": postgresClient,
				"redis":    redisClient,
			},
		},
	)

	slog.InfoContext(ctx, "starting surge",
		"environment", cfg.Env,
		"address", server.Address(),
	)

	return server.Serve(ctx)
}

func parseLogLevel(levelStr string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(levelStr)) {
	case "debug":
		return slog.LevelDebug, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	case "info":
		return slog.LevelInfo, nil
	default:
		return 0, fmt.Errorf("unsupported log level: %q", levelStr)
	}
}

func setupBootstrapLogger(level slog.Level) {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     level,
		AddSource: level == slog.LevelDebug,
	})

	slog.SetDefault(slog.New(handler))
}

func setupLogger(env string, level slog.Level) {
	var handler slog.Handler
	if env == "production" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:     level,
			AddSource: level == slog.LevelDebug,
		})
	} else {
		handler = tint.NewTextHandler(os.Stdout, &tint.Options{
			Level:      level,
			AddSource:  level == slog.LevelDebug,
			TimeFormat: time.Kitchen,
		})
	}

	slog.SetDefault(slog.New(handler))
}
