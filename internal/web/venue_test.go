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

func TestHandleVenueByID(t *testing.T) {
	t.Parallel()

	id, err := domain.ParseVenueID(testVenueID)
	if err != nil {
		t.Fatalf("parse test venue ID: %v", err)
	}

	tests := []struct {
		name            string
		path            string
		result          domain.Venue
		serviceErr      error
		wantStatus      int
		wantBody        string
		wantServiceCall bool
	}{
		{
			name: "success", path: "/venues/" + testVenueID,
			result:     domain.Venue{ID: id, Name: "Royal Albert Hall"},
			wantStatus: http.StatusOK, wantBody: `"name":"Royal Albert Hall"`, wantServiceCall: true,
		},
		{
			name: "invalid ID", path: "/venues/not-a-uuid",
			wantStatus: http.StatusBadRequest, wantBody: `"error":"Invalid venue ID"`,
		},
		{
			name: "not found", path: "/venues/" + testVenueID, serviceErr: domain.ErrVenueNotFound,
			wantStatus: http.StatusNotFound, wantBody: `"error":"Venue not found"`, wantServiceCall: true,
		},
		{
			name: "service failure", path: "/venues/" + testVenueID, serviceErr: errors.New("database unavailable"),
			wantStatus: http.StatusInternalServerError, wantBody: `"error":"The server encountered a problem`, wantServiceCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			called := false
			server := testServer(Dependencies{
				Venues: fakeVenueManager{
					venueByID: func(_ context.Context, gotID domain.VenueID) (domain.Venue, error) {
						called = true
						if gotID != id {
							t.Errorf("VenueByID() ID = %s, want %s", gotID, id)
						}
						return tt.result, tt.serviceErr
					},
				},
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

func TestHandleCreateVenue(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, time.September, 19, 12, 0, 0, 0, time.UTC)
	id, err := domain.ParseVenueID(testVenueID)
	if err != nil {
		t.Fatalf("parse test venue ID: %v", err)
	}

	createdVenue := domain.Venue{
		ID:        id,
		Name:      "Royal Albert Hall",
		CreatedAt: createdAt,
	}

	tests := []struct {
		name            string
		body            string
		serviceErr      error
		wantStatus      int
		wantBody        string
		wantServiceCall bool
		wantInput       domain.CreateVenueInput
	}{
		{
			name:            "success",
			body:            `{"name":"Royal Albert Hall"}`,
			wantStatus:      http.StatusCreated,
			wantBody:        `"id":"` + testVenueID + `"`,
			wantServiceCall: true,
			wantInput:       domain.CreateVenueInput{Name: "Royal Albert Hall"},
		},
		{
			name:       "invalid JSON",
			body:       `{"name":`,
			wantStatus: http.StatusBadRequest,
			wantBody:   `"error":"Invalid request body"`,
		},
		{
			name:            "invalid name",
			body:            `{"name":""}`,
			serviceErr:      domain.ErrVenueInvalidName,
			wantStatus:      http.StatusUnprocessableEntity,
			wantBody:        `"error":"Name is required"`,
			wantServiceCall: true,
			wantInput:       domain.CreateVenueInput{},
		},
		{
			name:            "service failure",
			body:            `{"name":"Royal Albert Hall"}`,
			serviceErr:      errors.New("database unavailable"),
			wantStatus:      http.StatusInternalServerError,
			wantBody:        `"error":"The server encountered a problem`,
			wantServiceCall: true,
			wantInput:       domain.CreateVenueInput{Name: "Royal Albert Hall"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			called := false
			server := testServer(Dependencies{
				Venues: fakeVenueManager{
					createVenue: func(_ context.Context, input domain.CreateVenueInput) (domain.Venue, error) {
						called = true
						if input != tt.wantInput {
							t.Errorf("CreateVenue() input = %#v, want %#v", input, tt.wantInput)
						}

						return createdVenue, tt.serviceErr
					},
				},
			})

			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/venues", strings.NewReader(tt.body))
			server.routes().ServeHTTP(response, request)

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
