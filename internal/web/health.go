package web

import (
	"context"
	"log/slog"
	"net/http"
)

type Pinger interface {
	Ping(context.Context) error
}

func (s *Server) handleLiveness(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleReadiness(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	for name, pinger := range s.dependencies.HealthChecks {
		if err := pinger.Ping(ctx); err != nil {
			addRequestLogAttrs(ctx,
				slog.String("dependency", name),
				slog.String("error_code", "dependency_unavailable"),
				slog.Any("error", err),
			)

			writeError(w, http.StatusServiceUnavailable, "Service unavailable")
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
