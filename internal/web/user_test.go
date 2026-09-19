package web

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/iamconnor4/surge/internal/domain"
)

func TestHandleUserByID(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, time.September, 19, 12, 0, 0, 0, time.UTC)
	id, err := domain.ParseUserID(testUUIDv7)
	if err != nil {
		t.Fatalf("parse test user ID: %v", err)
	}

	tests := []struct {
		name            string
		path            string
		result          domain.User
		serviceErr      error
		wantStatus      int
		wantBody        string
		wantServiceCall bool
	}{
		{
			name: "success",
			path: "/users/" + testUUIDv7,
			result: domain.User{
				ID: id, Email: "connor@example.com", FirstName: "Connor", LastName: "Smith", CreatedAt: createdAt,
			},
			wantStatus: http.StatusOK, wantBody: `"email":"connor@example.com"`, wantServiceCall: true,
		},
		{
			name: "invalid ID", path: "/users/not-a-uuid",
			wantStatus: http.StatusBadRequest, wantBody: `"error":"Invalid user ID"`,
		},
		{
			name: "not found", path: "/users/" + testUUIDv7, serviceErr: domain.ErrUserNotFound,
			wantStatus: http.StatusNotFound, wantBody: `"error":"User not found"`, wantServiceCall: true,
		},
		{
			name: "service failure", path: "/users/" + testUUIDv7, serviceErr: errors.New("database unavailable"),
			wantStatus: http.StatusInternalServerError, wantBody: `"error":"The server encountered a problem`, wantServiceCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			called := false
			server := testServer(Dependencies{
				Users: userManagerFunc(func(_ context.Context, gotID domain.UserID) (domain.User, error) {
					called = true
					if gotID != id {
						t.Errorf("UserByID() ID = %s, want %s", gotID, id)
					}
					return tt.result, tt.serviceErr
				}),
			})

			response := httptest.NewRecorder()
			server.routes().ServeHTTP(response, httptest.NewRequest(http.MethodGet, tt.path, nil))

			if response.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", response.Code, tt.wantStatus)
			}
			if !strings.Contains(response.Body.String(), tt.wantBody) {
				t.Errorf("body = %q, want it to contain %q", response.Body.String(), tt.wantBody)
			}
			if called != tt.wantServiceCall {
				t.Errorf("service called = %t, want %t", called, tt.wantServiceCall)
			}
		})
	}
}
