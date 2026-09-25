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

func TestHandleEventSeatByID(t *testing.T) {
	t.Parallel()

	eventID, err := domain.ParseEventID(testEventID)
	if err != nil {
		t.Fatalf("parse test event ID: %v", err)
	}

	venueSeatID, err := domain.ParseVenueSeatID(testVenueSeatID)
	if err != nil {
		t.Fatalf("parse test venue seat ID: %v", err)
	}

	id, err := domain.ParseEventSeatID(testEventSeatID)
	if err != nil {
		t.Fatalf("parse test event seat ID: %v", err)
	}

	tests := []struct {
		name            string
		path            string
		result          domain.EventSeat
		serviceErr      error
		wantStatus      int
		wantBody        string
		wantServiceCall bool
	}{
		{
			name: "success", path: "/event-seats/" + testEventSeatID,
			result: domain.EventSeat{
				ID: id, EventID: eventID, VenueSeatID: venueSeatID, PricePence: 8000, IsAvailable: true,
			},
			wantStatus: http.StatusOK, wantBody: `"pricePence":8000`, wantServiceCall: true,
		},
		{
			name: "invalid ID", path: "/event-seats/not-a-uuid",
			wantStatus: http.StatusBadRequest, wantBody: `"error":"Invalid event seat ID"`,
		},
		{
			name: "not found", path: "/event-seats/" + testEventSeatID,
			serviceErr: domain.ErrEventSeatNotFound,
			wantStatus: http.StatusNotFound, wantBody: `"error":"Event seat not found"`, wantServiceCall: true,
		},
		{
			name: "service failure", path: "/event-seats/" + testEventSeatID,
			serviceErr: errors.New("database unavailable"),
			wantStatus: http.StatusInternalServerError, wantBody: `"error":"The server encountered a problem`, wantServiceCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			called := false
			server := testServer(Dependencies{
				EventSeats: fakeEventSeatManager{
					eventSeatByID: func(_ context.Context, gotID domain.EventSeatID) (domain.EventSeat, error) {
						called = true
						if gotID != id {
							t.Errorf("EventSeatByID() ID = %s, want %s", gotID, id)
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

func TestHandleCreateEventSeats(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, time.September, 25, 12, 0, 0, 0, time.UTC)

	eventID, err := domain.ParseEventID(testEventID)
	if err != nil {
		t.Fatalf("parse test event ID: %v", err)
	}

	venueSectionID, err := domain.ParseVenueSectionID(testVenueSectionID)
	if err != nil {
		t.Fatalf("parse test venue section ID: %v", err)
	}

	venueSeatID, err := domain.ParseVenueSeatID(testVenueSeatID)
	if err != nil {
		t.Fatalf("parse test venue seat ID: %v", err)
	}

	venueSeatID2, err := domain.ParseVenueSeatID(testVenueSeatID2)
	if err != nil {
		t.Fatalf("parse second test venue seat ID: %v", err)
	}

	id, err := domain.ParseEventSeatID(testEventSeatID)
	if err != nil {
		t.Fatalf("parse test event seat ID: %v", err)
	}

	id2, err := domain.ParseEventSeatID(testEventSeatID2)
	if err != nil {
		t.Fatalf("parse second test event seat ID: %v", err)
	}

	createdEventSeats := []domain.EventSeat{
		{
			ID: id, EventID: eventID, VenueSeatID: venueSeatID, PricePence: 8000, IsAvailable: true, CreatedAt: createdAt,
		},
		{
			ID: id2, EventID: eventID, VenueSeatID: venueSeatID2, PricePence: 8000, IsAvailable: true, CreatedAt: createdAt,
		},
	}

	validBody := `{"sections":[{"venueSectionId":"` + testVenueSectionID + `","pricePence":8000}]}`
	validInput := domain.CreateEventSeatsInput{
		EventID: eventID,
		Sections: []domain.CreateEventSeatSectionInput{
			{VenueSectionID: venueSectionID, PricePence: 8000},
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
		wantInput       domain.CreateEventSeatsInput
	}{
		{
			name:            "success",
			path:            "/events/" + testEventID + "/seats",
			body:            validBody,
			wantStatus:      http.StatusCreated,
			wantBody:        `"seats":[{"id":"` + testEventSeatID + `"`,
			wantServiceCall: true,
			wantInput:       validInput,
		},
		{
			name:       "invalid event ID",
			path:       "/events/not-a-uuid/seats",
			body:       validBody,
			wantStatus: http.StatusBadRequest,
			wantBody:   `"error":"Invalid event ID"`,
		},
		{
			name:       "invalid JSON",
			path:       "/events/" + testEventID + "/seats",
			body:       `{"sections":`,
			wantStatus: http.StatusBadRequest,
			wantBody:   `"error":"Invalid request body"`,
		},
		{
			name:       "invalid venue section ID",
			path:       "/events/" + testEventID + "/seats",
			body:       `{"sections":[{"venueSectionId":"not-a-uuid","pricePence":8000}]}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   `"error":"Invalid venue section ID"`,
		},
		{
			name:            "event not found",
			path:            "/events/" + testEventID + "/seats",
			body:            validBody,
			serviceErr:      domain.ErrEventNotFound,
			wantStatus:      http.StatusNotFound,
			wantBody:        `"error":"Event not found"`,
			wantServiceCall: true,
			wantInput:       validInput,
		},
		{
			name:            "venue section not found",
			path:            "/events/" + testEventID + "/seats",
			body:            validBody,
			serviceErr:      domain.ErrVenueSectionNotFound,
			wantStatus:      http.StatusNotFound,
			wantBody:        `"error":"Venue section not found"`,
			wantServiceCall: true,
			wantInput:       validInput,
		},
		{
			name:            "venue seat not found",
			path:            "/events/" + testEventID + "/seats",
			body:            validBody,
			serviceErr:      domain.ErrVenueSeatNotFound,
			wantStatus:      http.StatusNotFound,
			wantBody:        `"error":"Venue seat not found"`,
			wantServiceCall: true,
			wantInput:       validInput,
		},
		{
			name:            "venue mismatch",
			path:            "/events/" + testEventID + "/seats",
			body:            validBody,
			serviceErr:      domain.ErrEventSeatVenueMismatch,
			wantStatus:      http.StatusUnprocessableEntity,
			wantBody:        `"error":"One or more venue sections do not belong to this event's venue"`,
			wantServiceCall: true,
			wantInput:       validInput,
		},
		{
			name:            "price below zero",
			path:            "/events/" + testEventID + "/seats",
			body:            validBody,
			serviceErr:      domain.ErrEventSeatPricePenceBelowZero,
			wantStatus:      http.StatusUnprocessableEntity,
			wantBody:        `"error":"One or more event seat prices are below zero"`,
			wantServiceCall: true,
			wantInput:       validInput,
		},
		{
			name:            "duplicate section",
			path:            "/events/" + testEventID + "/seats",
			body:            validBody,
			serviceErr:      domain.ErrEventSeatDuplicateSection,
			wantStatus:      http.StatusUnprocessableEntity,
			wantBody:        `"error":"The same venue section was given more than once"`,
			wantServiceCall: true,
			wantInput:       validInput,
		},
		{
			name:            "sections required",
			path:            "/events/" + testEventID + "/seats",
			body:            `{"sections":[]}`,
			serviceErr:      domain.ErrEventSeatSectionsRequired,
			wantStatus:      http.StatusUnprocessableEntity,
			wantBody:        `"error":"At least one venue section is required"`,
			wantServiceCall: true,
			wantInput:       domain.CreateEventSeatsInput{EventID: eventID, Sections: []domain.CreateEventSeatSectionInput{}},
		},
		{
			name:            "sections limit exceeded",
			path:            "/events/" + testEventID + "/seats",
			body:            validBody,
			serviceErr:      domain.ErrEventSeatSectionsLimitExceeded,
			wantStatus:      http.StatusUnprocessableEntity,
			wantBody:        `"error":"Too many venue sections"`,
			wantServiceCall: true,
			wantInput:       validInput,
		},
		{
			name:            "event seat already exists",
			path:            "/events/" + testEventID + "/seats",
			body:            validBody,
			serviceErr:      domain.ErrEventSeatAlreadyExists,
			wantStatus:      http.StatusConflict,
			wantBody:        `"error":"One or more event seats already exist"`,
			wantServiceCall: true,
			wantInput:       validInput,
		},
		{
			name:            "service failure",
			path:            "/events/" + testEventID + "/seats",
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
				EventSeats: fakeEventSeatManager{
					createEventSeats: func(_ context.Context, input domain.CreateEventSeatsInput) ([]domain.EventSeat, error) {
						called = true
						if !reflect.DeepEqual(input, tt.wantInput) {
							t.Errorf("CreateEventSeats() input = %#v, want %#v", input, tt.wantInput)
						}

						return createdEventSeats, tt.serviceErr
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
