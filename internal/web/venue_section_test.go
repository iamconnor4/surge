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

func TestHandleVenueSectionByID(t *testing.T) {
	t.Parallel()

	venueID, err := domain.ParseVenueID(testVenueID)
	if err != nil {
		t.Fatalf("parse test venue ID: %v", err)
	}

	id, err := domain.ParseVenueSectionID(testVenueSectionID)
	if err != nil {
		t.Fatalf("parse test venue section ID: %v", err)
	}

	tests := []struct {
		name            string
		path            string
		result          domain.VenueSection
		serviceErr      error
		wantStatus      int
		wantBody        string
		wantServiceCall bool
	}{
		{
			name: "success", path: "/venue-sections/" + testVenueSectionID,
			result: domain.VenueSection{
				ID: id, VenueID: venueID, Name: "Stalls",
			},
			wantStatus: http.StatusOK, wantBody: `"name":"Stalls"`, wantServiceCall: true,
		},
		{
			name: "invalid ID", path: "/venue-sections/not-a-uuid",
			wantStatus: http.StatusBadRequest, wantBody: `"error":"Invalid venue section ID"`,
		},
		{
			name: "not found", path: "/venue-sections/" + testVenueSectionID,
			serviceErr: domain.ErrVenueSectionNotFound,
			wantStatus: http.StatusNotFound, wantBody: `"error":"Venue section not found"`, wantServiceCall: true,
		},
		{
			name: "service failure", path: "/venue-sections/" + testVenueSectionID,
			serviceErr: errors.New("database unavailable"),
			wantStatus: http.StatusInternalServerError, wantBody: `"error":"The server encountered a problem`, wantServiceCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			called := false
			server := testServer(Dependencies{
				VenueSections: fakeVenueSectionManager{
					venueSectionByID: func(_ context.Context, gotID domain.VenueSectionID) (domain.VenueSection, error) {
						called = true
						if gotID != id {
							t.Errorf("VenueSectionByID() ID = %s, want %s", gotID, id)
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

func TestHandleCreateVenueSection(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, time.September, 20, 12, 0, 0, 0, time.UTC)
	venueID, err := domain.ParseVenueID(testVenueID)
	if err != nil {
		t.Fatalf("parse test venue ID: %v", err)
	}

	id, err := domain.ParseVenueSectionID(testVenueSectionID)
	if err != nil {
		t.Fatalf("parse test venue section ID: %v", err)
	}

	createdVenueSection := domain.VenueSection{
		ID:        id,
		VenueID:   venueID,
		Name:      "Stalls",
		CreatedAt: createdAt,
	}

	tests := []struct {
		name            string
		path            string
		body            string
		serviceErr      error
		wantStatus      int
		wantBody        string
		wantServiceCall bool
		wantInput       domain.CreateVenueSectionInput
	}{
		{
			name:            "success",
			path:            "/venues/" + testVenueID + "/sections",
			body:            `{"name":"Stalls"}`,
			wantStatus:      http.StatusCreated,
			wantBody:        `"id":"` + testVenueSectionID + `"`,
			wantServiceCall: true,
			wantInput:       domain.CreateVenueSectionInput{VenueID: venueID, Name: "Stalls"},
		},
		{
			name:       "invalid venue ID",
			path:       "/venues/not-a-uuid/sections",
			body:       `{"name":"Stalls"}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   `"error":"Invalid venue ID"`,
		},
		{
			name:       "invalid JSON",
			path:       "/venues/" + testVenueID + "/sections",
			body:       `{"name":`,
			wantStatus: http.StatusBadRequest,
			wantBody:   `"error":"Invalid request body"`,
		},
		{
			name:            "venue not found",
			path:            "/venues/" + testVenueID + "/sections",
			body:            `{"name":"Stalls"}`,
			serviceErr:      domain.ErrVenueNotFound,
			wantStatus:      http.StatusNotFound,
			wantBody:        `"error":"Venue not found"`,
			wantServiceCall: true,
			wantInput:       domain.CreateVenueSectionInput{VenueID: venueID, Name: "Stalls"},
		},
		{
			name:            "invalid name",
			path:            "/venues/" + testVenueID + "/sections",
			body:            `{"name":""}`,
			serviceErr:      domain.ErrVenueSectionInvalidName,
			wantStatus:      http.StatusUnprocessableEntity,
			wantBody:        `"error":"Name is required"`,
			wantServiceCall: true,
			wantInput:       domain.CreateVenueSectionInput{VenueID: venueID},
		},
		{
			name:            "service failure",
			path:            "/venues/" + testVenueID + "/sections",
			body:            `{"name":"Stalls"}`,
			serviceErr:      errors.New("database unavailable"),
			wantStatus:      http.StatusInternalServerError,
			wantBody:        `"error":"The server encountered a problem`,
			wantServiceCall: true,
			wantInput:       domain.CreateVenueSectionInput{VenueID: venueID, Name: "Stalls"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			called := false
			server := testServer(Dependencies{
				VenueSections: fakeVenueSectionManager{
					createVenueSection: func(_ context.Context, input domain.CreateVenueSectionInput) (domain.VenueSection, error) {
						called = true
						if input != tt.wantInput {
							t.Errorf("CreateVenueSection() input = %#v, want %#v", input, tt.wantInput)
						}

						return createdVenueSection, tt.serviceErr
					},
				},
			})

			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, tt.path, strings.NewReader(tt.body))
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
