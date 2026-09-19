package web

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to write JSON response", "error", err)
	}
}

type errorResponse struct {
	Error     string `json:"error"`
	RequestID string `json:"requestId"`
}

func writeError(w http.ResponseWriter, status int, message string) {
	reqId := w.Header().Get("X-Request-Id")

	writeJSON(w, status, errorResponse{
		Error:     message,
		RequestID: reqId,
	})
}
