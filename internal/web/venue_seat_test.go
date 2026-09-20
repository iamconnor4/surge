package web

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/iamconnor4/surge/internal/domain"
)

func TestHandleVenueSeatByID(t *testing.T) {
	t.Parallel()

	venueSectionID, err := domain.ParseVenueSectionID(testVenueSectionID)
	if err != nil {
		t.Fatalf("parse test venue section ID: %v", err)
	}

	id, err := domain.ParseVenueSeatID(testVenueSeatID)
	if err != nil {
		t.Fatalf("parse test venue seat ID: %v", err)
	}

	tests := []struct {
		name            string
		path            string
		result          domain.VenueSeat
		serviceErr      error
		wantStatus      int
		wantBody        string
		wantServiceCall bool
	}{
		{
			name: "success", path: "/venue-seats/" + testVenueSeatID,
			result: domain.VenueSeat{
				ID: id, VenueSectionID: venueSectionID, RowLabel: "A", SeatLabel: "1",
			},
			wantStatus: http.StatusOK, wantBody: `"seatLabel":"1"`, wantServiceCall: true,
		},
		{
			name: "invalid ID", path: "/venue-seats/not-a-uuid",
			wantStatus: http.StatusBadRequest, wantBody: `"error":"Invalid venue seat ID"`,
		},
		{
			name: "not found", path: "/venue-seats/" + testVenueSeatID,
			serviceErr: domain.ErrVenueSeatNotFound,
			wantStatus: http.StatusNotFound, wantBody: `"error":"Venue seat not found"`, wantServiceCall: true,
		},
		{
			name: "service failure", path: "/venue-seats/" + testVenueSeatID,
			serviceErr: errors.New("database unavailable"),
			wantStatus: http.StatusInternalServerError, wantBody: `"error":"The server encountered a problem`, wantServiceCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			called := false
			server := testServer(Dependencies{
				VenueSeats: fakeVenueSeatManager{
					venueSeatByID: func(_ context.Context, gotID domain.VenueSeatID) (domain.VenueSeat, error) {
						called = true
						if gotID != id {
							t.Errorf("VenueSeatByID() ID = %s, want %s", gotID, id)
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

func TestHandleCreateVenueSeats(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, time.September, 20, 12, 0, 0, 0, time.UTC)
	venueSectionID, err := domain.ParseVenueSectionID(testVenueSectionID)
	if err != nil {
		t.Fatalf("parse test venue section ID: %v", err)
	}

	id, err := domain.ParseVenueSeatID(testVenueSeatID)
	if err != nil {
		t.Fatalf("parse test venue seat ID: %v", err)
	}

	id2, err := domain.ParseVenueSeatID(testVenueSeatID2)
	if err != nil {
		t.Fatalf("parse second test venue seat ID: %v", err)
	}

	createdVenueSeats := []domain.VenueSeat{
		{
			ID: id, VenueSectionID: venueSectionID, RowLabel: "A", SeatLabel: "1", CreatedAt: createdAt,
		},
		{
			ID: id2, VenueSectionID: venueSectionID, RowLabel: "A", SeatLabel: "2", CreatedAt: createdAt,
		},
	}

	validBody := `{"seats":[{"rowLabel":"A","seatLabel":"1"},{"rowLabel":"A","seatLabel":"2"}]}`
	validInput := domain.CreateVenueSeatsInput{
		VenueSectionID: venueSectionID,
		Seats: []domain.CreateVenueSeatInput{
			{RowLabel: "A", SeatLabel: "1"},
			{RowLabel: "A", SeatLabel: "2"},
		},
	}

	tests := []struct {
		name            string
		path            string
		body            string
		serviceErr      error
		wantStatus      int
		wantBody        string
		wantServiceCall bool
		wantInput       domain.CreateVenueSeatsInput
	}{
		{
			name:            "success",
			path:            "/venue-sections/" + testVenueSectionID + "/seats",
			body:            validBody,
			wantStatus:      http.StatusCreated,
			wantBody:        `"seats":[{"id":"` + testVenueSeatID + `"`,
			wantServiceCall: true,
			wantInput:       validInput,
		},
		{
			name:       "invalid venue section ID",
			path:       "/venue-sections/not-a-uuid/seats",
			body:       validBody,
			wantStatus: http.StatusBadRequest,
			wantBody:   `"error":"Invalid venue section ID"`,
		},
		{
			name:       "invalid JSON",
			path:       "/venue-sections/" + testVenueSectionID + "/seats",
			body:       `{"seats":`,
			wantStatus: http.StatusBadRequest,
			wantBody:   `"error":"Invalid request body"`,
		},
		{
			name:            "venue section not found",
			path:            "/venue-sections/" + testVenueSectionID + "/seats",
			body:            validBody,
			serviceErr:      domain.ErrVenueSectionNotFound,
			wantStatus:      http.StatusNotFound,
			wantBody:        `"error":"Venue section not found"`,
			wantServiceCall: true,
			wantInput:       validInput,
		},
		{
			name:            "missing row label",
			path:            "/venue-sections/" + testVenueSectionID + "/seats",
			body:            validBody,
			serviceErr:      domain.ErrVenueSeatRowLabel,
			wantStatus:      http.StatusUnprocessableEntity,
			wantBody:        `"error":"One or more venue seats are missing a row label"`,
			wantServiceCall: true,
			wantInput:       validInput,
		},
		{
			name:            "missing seat label",
			path:            "/venue-sections/" + testVenueSectionID + "/seats",
			body:            validBody,
			serviceErr:      domain.ErrVenueSeatSeatLabel,
			wantStatus:      http.StatusUnprocessableEntity,
			wantBody:        `"error":"One or more venue seats are missing a seat label"`,
			wantServiceCall: true,
			wantInput:       validInput,
		},
		{
			name:            "seat already exists",
			path:            "/venue-sections/" + testVenueSectionID + "/seats",
			body:            validBody,
			serviceErr:      domain.ErrVenueSeatAlreadyExists,
			wantStatus:      http.StatusConflict,
			wantBody:        `"error":"One or more venue seats already exist"`,
			wantServiceCall: true,
			wantInput:       validInput,
		},
		{
			name:            "seats required",
			path:            "/venue-sections/" + testVenueSectionID + "/seats",
			body:            `{"seats":[]}`,
			serviceErr:      domain.ErrVenueSeatRequired,
			wantStatus:      http.StatusUnprocessableEntity,
			wantBody:        `"error":"At least one venue seat is required"`,
			wantServiceCall: true,
			wantInput:       domain.CreateVenueSeatsInput{VenueSectionID: venueSectionID, Seats: []domain.CreateVenueSeatInput{}},
		},
		{
			name:            "seat limit exceeded",
			path:            "/venue-sections/" + testVenueSectionID + "/seats",
			body:            validBody,
			serviceErr:      domain.ErrVenueSeatLimitExceeded,
			wantStatus:      http.StatusUnprocessableEntity,
			wantBody:        `"error":"Too many venue seats"`,
			wantServiceCall: true,
			wantInput:       validInput,
		},
		{
			name:            "service failure",
			path:            "/venue-sections/" + testVenueSectionID + "/seats",
			body:            validBody,
			serviceErr:      errors.New("database unavailable"),
			wantStatus:      http.StatusInternalServerError,
			wantBody:        `"error":"The server encountered a problem`,
			wantServiceCall: true,
			wantInput:       validInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			called := false
			server := testServer(Dependencies{
				VenueSeats: fakeVenueSeatManager{
					createVenueSeats: func(_ context.Context, input domain.CreateVenueSeatsInput) ([]domain.VenueSeat, error) {
						called = true
						if !reflect.DeepEqual(input, tt.wantInput) {
							t.Errorf("CreateVenueSeats() input = %#v, want %#v", input, tt.wantInput)
						}

						return createdVenueSeats, tt.serviceErr
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
