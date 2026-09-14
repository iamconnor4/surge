package web

import (
	"context"
	"log/slog"
	"net/http"
)

type ReadinessChecker interface {
	Ping(context.Context) error
}

func (s *Server) handleLiveness(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleReadiness(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := s.dependencies.Postgres.Ping(ctx); err != nil {
		addRequestLogAttrs(ctx,
			slog.String("dependency", "postgres"),
			slog.String("error_code", "dependency_unavailable"),
			slog.Any("error", err),
		)
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}

	if err := s.dependencies.Redis.Ping(ctx); err != nil {
		addRequestLogAttrs(ctx,
			slog.String("dependency", "redis"),
			slog.String("error_code", "dependency_unavailable"),
			slog.Any("error", err),
		)
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
