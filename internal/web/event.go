package web

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/iamconnor4/surge/internal/domain"
)

type EventManager interface {
	EventByID(ctx context.Context, id domain.EventID) (domain.Event, error)
	CreateEvent(ctx context.Context, input domain.CreateEventInput) (domain.Event, error)
}

type createEventRequest struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartsAt    time.Time `json:"startsAt"`
}

type eventResponse struct {
	ID          string     `json:"id"`
	VenueID     string     `json:"venueId"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	StartsAt    time.Time  `json:"startsAt"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty"`
}

func newEventResponse(event domain.Event) eventResponse {
	return eventResponse{
		ID:          event.ID.String(),
		VenueID:     event.VenueID.String(),
		Title:       event.Title,
		Description: event.Description,
		StartsAt:    event.StartsAt,
		CreatedAt:   event.CreatedAt,
		UpdatedAt:   event.UpdatedAt,
	}
}

func (s *Server) handleCreateEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	venueID, err := domain.ParseVenueID(r.PathValue("venueID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid venue ID")
		return
	}

	var request createEventRequest
	if err := readJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	event, err := s.dependencies.Events.CreateEvent(
		ctx,
		domain.CreateEventInput{
			VenueID:     venueID,
			Title:       request.Title,
			Description: request.Description,
			StartsAt:    request.StartsAt,
		},
	)

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrEventInvalidTitle):
			writeError(w, http.StatusUnprocessableEntity, "Title is required")

		case errors.Is(err, domain.ErrEventInvalidDescription):
			writeError(w, http.StatusUnprocessableEntity, "Description is required")

		case errors.Is(err, domain.ErrEventStartsAtRequired):
			writeError(w, http.StatusUnprocessableEntity, "Start time is required")

		case errors.Is(err, domain.ErrEventStartsAtInPast):
			writeError(w, http.StatusUnprocessableEntity, "Start time must be in the future")

		case errors.Is(err, domain.ErrVenueNotFound):
			writeError(w, http.StatusNotFound, "Venue not found")

		default:
			addRequestLogAttrs(ctx,
				slog.String("error_code", "event_creation_failed"),
				slog.Any("error", err),
			)

			writeError(w, http.StatusInternalServerError, "The server encountered a problem and could not process your request")
		}

		return
	}

	writeJSON(w, http.StatusCreated, newEventResponse(event))
}

func (s *Server) handleEventByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := domain.ParseEventID(r.PathValue("eventID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid event ID")
		return
	}

	event, err := s.dependencies.Events.EventByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrEventNotFound) {
			writeError(w, http.StatusNotFound, "Event not found")
			return
		}

		addRequestLogAttrs(ctx,
			slog.String("error_code", "event_lookup_failed"),
			slog.Any("error", err),
		)

		writeError(w, http.StatusInternalServerError, "The server encountered a problem and could not process your request")
		return
	}

	writeJSON(w, http.StatusOK, newEventResponse(event))
}
