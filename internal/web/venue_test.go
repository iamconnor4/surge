package web

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/iamconnor4/surge/internal/domain"
)

func TestHandleVenueByID(t *testing.T) {
	t.Parallel()

	id, err := domain.ParseVenueID(testUUIDv7)
	if err != nil {
		t.Fatalf("parse test venue ID: %v", err)
	}

	tests := []struct {
		name       string
		path       string
		result     domain.Venue
		serviceErr error
		wantStatus int
		wantBody   string
	}{
		{
			name: "success", path: "/venues/" + testUUIDv7,
			result:     domain.Venue{ID: id, Name: "Royal Albert Hall"},
			wantStatus: http.StatusOK, wantBody: `"name":"Royal Albert Hall"`,
		},
		{
			name: "invalid ID", path: "/venues/not-a-uuid",
			wantStatus: http.StatusBadRequest, wantBody: `"error":"Invalid venue ID"`,
		},
		{
			name: "not found", path: "/venues/" + testUUIDv7, serviceErr: domain.ErrVenueNotFound,
			wantStatus: http.StatusNotFound, wantBody: `"error":"Venue not found"`,
		},
		{
			name: "service failure", path: "/venues/" + testUUIDv7, serviceErr: errors.New("database unavailable"),
			wantStatus: http.StatusInternalServerError, wantBody: `"error":"The server encountered a problem`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := testServer(Dependencies{
				Venues: venueManagerFunc(func(_ context.Context, gotID domain.VenueID) (domain.Venue, error) {
					if gotID != id {
						t.Errorf("VenueByID() ID = %s, want %s", gotID, id)
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
		})
	}
}
