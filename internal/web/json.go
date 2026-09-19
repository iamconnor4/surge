package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

const maxRequestBodyBytes = 1024 * 1024

type errorResponse struct {
	Error     string `json:"error"`
	RequestID string `json:"requestId"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to write JSON response", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	reqId := w.Header().Get(requestIDHeader)

	writeJSON(w, status, errorResponse{
		Error:     message,
		RequestID: reqId,
	})
}

func readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return fmt.Errorf("decode JSON Body: %w", err)
	}

	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("request body must contain a single JSON value")
		}

		return fmt.Errorf("decode trailing JSON data: %w", err)
	}

	return nil
}
