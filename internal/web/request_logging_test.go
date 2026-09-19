package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestLoggingMiddleware(t *testing.T) {
	t.Parallel()

	t.Run("adds a request ID", func(t *testing.T) {
		t.Parallel()

		server := testServer(Dependencies{})
		handler := server.logRequests(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))

		if response.Code != http.StatusNoContent {
			t.Errorf("status = %d, want %d", response.Code, http.StatusNoContent)
		}
		if response.Header().Get(requestIDHeader) == "" {
			t.Error("request ID header is empty")
		}
	})

	t.Run("recovers a panic", func(t *testing.T) {
		t.Parallel()

		server := testServer(Dependencies{})
		handler := server.logRequests(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			panic("boom")
		}))
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))

		if response.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, want %d", response.Code, http.StatusInternalServerError)
		}
		if !strings.Contains(response.Body.String(), `"error":"The server encountered a problem`) {
			t.Errorf("unexpected body: %q", response.Body.String())
		}
	})
}
