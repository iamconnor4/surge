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

func TestHandleEventByID(t *testing.T) {
	t.Parallel()

	id, err := domain.ParseEventID(testEventID)
	if err != nil {
		t.Fatalf("parse test event ID: %v", err)
	}

	venueID, err := domain.ParseVenueID(testVenueID)
	if err != nil {
		t.Fatalf("parse test venue ID: %v", err)
	}

	startsAt := time.Date(2026, time.December, 20, 19, 30, 0, 0, time.UTC)

	tests := []struct {
		name            string
		path            string
		result          domain.Event
		serviceErr      error
		wantStatus      int
		wantBody        string
		wantServiceCall bool
	}{
		{
			name: "success", path: "/events/" + testEventID,
			result: domain.Event{
				ID: id, VenueID: venueID, Title: "Radiohead Live", Description: "Live at the Royal Albert Hall", StartsAt: startsAt,
			},
			wantStatus: http.StatusOK, wantBody: `"startsAt":"2026-12-20T19:30:00Z"`, wantServiceCall: true,
		},
		{
			name: "invalid ID", path: "/events/not-a-uuid",
			wantStatus: http.StatusBadRequest, wantBody: `"error":"Invalid event ID"`,
		},
		{
			name: "not found", path: "/events/" + testEventID, serviceErr: domain.ErrEventNotFound,
			wantStatus: http.StatusNotFound, wantBody: `"error":"Event not found"`, wantServiceCall: true,
		},
		{
			name: "service failure", path: "/events/" + testEventID, serviceErr: errors.New("database unavailable"),
			wantStatus: http.StatusInternalServerError, wantBody: `"error":"The server encountered a problem`, wantServiceCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			called := false
			server := testServer(Dependencies{
				Events: fakeEventManager{
					eventByID: func(_ context.Context, gotID domain.EventID) (domain.Event, error) {
						called = true
						if gotID != id {
							t.Errorf("EventByID() ID = %s, want %s", gotID, id)
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

func TestHandleCreateEvent(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, time.September, 21, 12, 0, 0, 0, time.UTC)
	startsAt := time.Date(2026, time.December, 20, 19, 30, 0, 0, time.UTC)

	id, err := domain.ParseEventID(testEventID)
	if err != nil {
		t.Fatalf("parse test event ID: %v", err)
	}

	venueID, err := domain.ParseVenueID(testVenueID)
	if err != nil {
		t.Fatalf("parse test venue ID: %v", err)
	}

	createdEvent := domain.Event{
		ID:          id,
		VenueID:     venueID,
		Title:       "Radiohead Live",
		Description: "Live at the Royal Albert Hall",
		StartsAt:    startsAt,
		CreatedAt:   createdAt,
	}

	validBody := `{"title":"Radiohead Live","description":"Live at the Royal Albert Hall","startsAt":"2026-12-20T19:30:00Z"}`
	validInput := domain.CreateEventInput{
		VenueID:     venueID,
		Title:       "Radiohead Live",
		Description: "Live at the Royal Albert Hall",
		StartsAt:    startsAt,
	}

	tests := []struct {
		name            string
		path            string
		body            string
		serviceErr      error
		wantStatus      int
		wantBody        string
		wantServiceCall bool
		wantInput       domain.CreateEventInput
	}{
		{
			name: "success", path: "/venues/" + testVenueID + "/events", body: validBody,
			wantStatus: http.StatusCreated, wantBody: `"id":"` + testEventID + `"`, wantServiceCall: true, wantInput: validInput,
		},
		{
			name: "invalid venue ID", path: "/venues/not-a-uuid/events", body: validBody,
			wantStatus: http.StatusBadRequest, wantBody: `"error":"Invalid venue ID"`,
		},
		{
			name: "invalid JSON", path: "/venues/" + testVenueID + "/events", body: `{"title":`,
			wantStatus: http.StatusBadRequest, wantBody: `"error":"Invalid request body"`,
		},
		{
			name: "invalid title", path: "/venues/" + testVenueID + "/events",
			body:       `{"title":"","description":"Live at the Royal Albert Hall","startsAt":"2026-12-20T19:30:00Z"}`,
			serviceErr: domain.ErrEventInvalidTitle, wantStatus: http.StatusUnprocessableEntity,
			wantBody: `"error":"Title is required"`, wantServiceCall: true,
			wantInput: domain.CreateEventInput{VenueID: venueID, Description: "Live at the Royal Albert Hall", StartsAt: startsAt},
		},
		{
			name: "invalid description", path: "/venues/" + testVenueID + "/events",
			body:       `{"title":"Radiohead Live","description":"","startsAt":"2026-12-20T19:30:00Z"}`,
			serviceErr: domain.ErrEventInvalidDescription, wantStatus: http.StatusUnprocessableEntity,
			wantBody: `"error":"Description is required"`, wantServiceCall: true,
			wantInput: domain.CreateEventInput{VenueID: venueID, Title: "Radiohead Live", StartsAt: startsAt},
		},
		{
			name: "start time required", path: "/venues/" + testVenueID + "/events",
			body:       `{"title":"Radiohead Live","description":"Live at the Royal Albert Hall"}`,
			serviceErr: domain.ErrEventStartsAtRequired, wantStatus: http.StatusUnprocessableEntity,
			wantBody: `"error":"Start time is required"`, wantServiceCall: true,
			wantInput: domain.CreateEventInput{VenueID: venueID, Title: "Radiohead Live", Description: "Live at the Royal Albert Hall"},
		},
		{
			name: "start time in past", path: "/venues/" + testVenueID + "/events", body: validBody,
			serviceErr: domain.ErrEventStartsAtInPast, wantStatus: http.StatusUnprocessableEntity,
			wantBody: `"error":"Start time must be in the future"`, wantServiceCall: true, wantInput: validInput,
		},
		{
			name: "venue not found", path: "/venues/" + testVenueID + "/events", body: validBody,
			serviceErr: domain.ErrVenueNotFound, wantStatus: http.StatusNotFound,
			wantBody: `"error":"Venue not found"`, wantServiceCall: true, wantInput: validInput,
		},
		{
			name: "service failure", path: "/venues/" + testVenueID + "/events", body: validBody,
			serviceErr: errors.New("database unavailable"), wantStatus: http.StatusInternalServerError,
			wantBody: `"error":"The server encountered a problem`, wantServiceCall: true, wantInput: validInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			called := false
			server := testServer(Dependencies{
				Events: fakeEventManager{
					createEvent: func(_ context.Context, input domain.CreateEventInput) (domain.Event, error) {
						called = true
						if input != tt.wantInput {
							t.Errorf("CreateEvent() input = %#v, want %#v", input, tt.wantInput)
						}

						return createdEvent, tt.serviceErr
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
