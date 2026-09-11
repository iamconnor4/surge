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
		slog.WarnContext(ctx, "postgres readiness check failed", "error", err)
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
