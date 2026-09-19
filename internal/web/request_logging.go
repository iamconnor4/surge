package web

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/google/uuid"
)

const requestIDHeader = "X-Request-ID"

type requestLogContextKey struct{}

type requestLogAttrs struct {
	attrs []slog.Attr
}

func (a *requestLogAttrs) add(attrs ...slog.Attr) {
	for _, attr := range attrs {
		if attr.Key != "" {
			a.attrs = append(a.attrs, attr)
		}
	}
}

func addRequestLogAttrs(ctx context.Context, attrs ...slog.Attr) {
	requestAttrs, ok := ctx.Value(requestLogContextKey{}).(*requestLogAttrs)
	if !ok {
		return
	}

	requestAttrs.add(attrs...)
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	if r.statusCode != 0 {
		return
	}

	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseRecorder) Write(body []byte) (int, error) {
	if r.statusCode == 0 {
		r.statusCode = http.StatusOK
	}

	n, err := r.ResponseWriter.Write(body)
	r.bytesWritten += int64(n)

	return n, err
}

func (r *responseRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

func (r *responseRecorder) status() int {
	if r.statusCode == 0 {
		return http.StatusOK
	}

	return r.statusCode
}

func (s *Server) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		requestAttrs := &requestLogAttrs{}
		response := &responseRecorder{
			ResponseWriter: w,
		}

		ctx := context.WithValue(
			r.Context(),
			requestLogContextKey{},
			requestAttrs,
		)
		r = r.WithContext(ctx)

		var recovered any
		var requestID string

		defer func() {
			if recovered = recover(); recovered != nil {
				requestAttrs.add(
					slog.String("error_type", "panic"),
					slog.String("error", fmt.Sprint(recovered)),
					slog.String("stack_trace", string(debug.Stack())),
				)

				if recovered != http.ErrAbortHandler && response.statusCode == 0 {
					writeError(response, http.StatusInternalServerError, "The server encountered a problem and could not process your request")
				}
			}

			statusCode := response.status()
			level, outcome := requestOutcome(statusCode, recovered)

			route := r.Pattern
			if route == "" {
				route = "unmatched"
			} else if _, path, found := strings.Cut(route, " "); found {
				route = path
			}

			attrs := []slog.Attr{
				slog.String("service", s.service),
				slog.String("environment", s.environment),
				slog.String("request_id", requestID),
				slog.String("method", r.Method),
				slog.String("route", route),
				slog.Int("status_code", statusCode),
				slog.Int64(
					"duration_ms",
					time.Since(startedAt).Milliseconds(),
				),
				slog.Int64("response_bytes", response.bytesWritten),
				slog.String("outcome", outcome),
			}

			attrs = append(attrs, requestAttrs.attrs...)

			s.logger.LogAttrs(
				ctx,
				level,
				"HTTP request",
				attrs...,
			)

			if recovered == http.ErrAbortHandler {
				panic(http.ErrAbortHandler)
			}
		}()

		requestID = newRequestID()
		w.Header().Set(requestIDHeader, requestID)

		next.ServeHTTP(response, r)
	})
}

func newRequestID() string {
	id, err := uuid.NewV7()
	if err != nil {
		panic(fmt.Errorf("generate request ID: %w", err))
	}

	return id.String()
}

func requestOutcome(statusCode int, recovered any) (slog.Level, string) {
	if recovered != nil || statusCode >= http.StatusInternalServerError {
		return slog.LevelError, "error"
	}

	if statusCode >= http.StatusBadRequest {
		return slog.LevelInfo, "client_error"
	}

	return slog.LevelInfo, "success"
}
