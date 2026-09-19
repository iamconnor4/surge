package web

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type Config struct {
	Address     string
	Service     string
	Environment string
}

type Dependencies struct {
	HealthChecks map[string]Pinger
	Users        UserManager
	Venues       VenueManager
}

type Server struct {
	httpServer      *http.Server
	dependencies    Dependencies
	logger          *slog.Logger
	service         string
	environment     string
	shutdownTimeout time.Duration
}

func New(cfg Config, dependencies Dependencies) *Server {
	logger := slog.Default()

	s := &Server{
		dependencies:    dependencies,
		logger:          logger,
		service:         cfg.Service,
		environment:     cfg.Environment,
		shutdownTimeout: 15 * time.Second,
	}

	s.httpServer = &http.Server{
		Addr:              cfg.Address,
		Handler:           s.routes(),
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelError),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return s
}

func (s *Server) Address() string {
	return s.httpServer.Addr
}

func (s *Server) Serve(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		errCh <- s.httpServer.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return fmt.Errorf("serve HTTP: %w", err)
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			s.shutdownTimeout,
		)
		defer cancel()

		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			closeErr := s.httpServer.Close()

			return errors.Join(
				fmt.Errorf("graceful HTTP shutdown: %w", err),
				closeErr,
			)
		}

		err := <-errCh
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("stop HTTP server: %w", err)
		}

		return nil
	}
}
