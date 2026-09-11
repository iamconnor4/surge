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
		return fmt.Errorf("failed to load config: %w", err)
	}

	setupLogger(cfg.LogLevel)

	db, err := postgres.NewClient(ctx, postgres.Config{
		Host:     cfg.Postgres.Host,
		Port:     cfg.Postgres.Port,
		User:     cfg.Postgres.User,
		Password: cfg.Postgres.Password,
		DB:       cfg.Postgres.DB,
		SSLMode:  cfg.Postgres.SSLMode,
	})
	if err != nil {
		return fmt.Errorf("initialise database: %w", err)
	}
	defer db.Close()

	server := web.New(
		web.Config{
			Address: ":" + strconv.Itoa(cfg.Port),
		},
		web.Dependencies{
			Postgres: db,
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
