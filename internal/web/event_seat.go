package web

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/iamconnor4/surge/internal/domain"
)

type EventSeatManager interface {
	EventSeatByID(ctx context.Context, id domain.EventSeatID) (domain.EventSeat, error)
	CreateEventSeats(ctx context.Context, input domain.CreateEventSeatsInput) ([]domain.EventSeat, error)
}

type createEventSeatsRequest struct {
	Sections []createEventSeatSectionRequest `json:"sections"`
}

type createEventSeatSectionRequest struct {
	VenueSectionID string `json:"venueSectionId"`
	PricePence     int64  `json:"pricePence"`
}

type eventSeatsResponse struct {
	Seats []eventSeatResponse `json:"seats"`
}

type eventSeatResponse struct {
	ID          string     `json:"id"`
	EventID     string     `json:"eventId"`
	VenueSeatID string     `json:"venueSeatId"`
	PricePence  int64      `json:"pricePence"`
	IsAvailable bool       `json:"isAvailable"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty"`
}

func newEventSeatsResponse(seats []domain.EventSeat) eventSeatsResponse {
	eventSeats := make([]eventSeatResponse, len(seats))

	for i, seat := range seats {
		eventSeats[i] = newEventSeatResponse(seat)
	}

	return eventSeatsResponse{
		Seats: eventSeats,
	}
}

func newEventSeatResponse(seat domain.EventSeat) eventSeatResponse {
	return eventSeatResponse{
		ID:          seat.ID.String(),
		EventID:     seat.EventID.String(),
		VenueSeatID: seat.VenueSeatID.String(),
		PricePence:  seat.PricePence,
		IsAvailable: seat.IsAvailable,
		CreatedAt:   seat.CreatedAt,
		UpdatedAt:   seat.UpdatedAt,
	}
}

func (s *Server) handleCreateEventSeats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	eventID, err := domain.ParseEventID(r.PathValue("eventID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid event ID")
		return
	}

	var request createEventSeatsRequest
	if err := readJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	sections := make([]domain.CreateEventSeatSectionInput, len(request.Sections))
	for i, section := range request.Sections {
		venueSectionID, err := domain.ParseVenueSectionID(section.VenueSectionID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid venue section ID")
			return
		}

		sections[i] = domain.CreateEventSeatSectionInput{
			VenueSectionID: venueSectionID,
			PricePence:     section.PricePence,
		}
	}

	eventSeats, err := s.dependencies.EventSeats.CreateEventSeats(
		ctx,
		domain.CreateEventSeatsInput{
			EventID:  eventID,
			Sections: sections,
		},
	)

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrEventNotFound):
			writeError(w, http.StatusNotFound, "Event not found")

		case errors.Is(err, domain.ErrVenueSectionNotFound):
			writeError(w, http.StatusNotFound, "Venue section not found")

		case errors.Is(err, domain.ErrVenueSeatNotFound):
			writeError(w, http.StatusNotFound, "Venue seat not found")

		case errors.Is(err, domain.ErrEventSeatVenueMismatch):
			writeError(w, http.StatusUnprocessableEntity, "One or more venue sections do not belong to this event's venue")

		case errors.Is(err, domain.ErrEventSeatPricePenceBelowZero):
			writeError(w, http.StatusUnprocessableEntity, "One or more event seat prices are below zero")

		case errors.Is(err, domain.ErrEventSeatDuplicateSection):
			writeError(w, http.StatusUnprocessableEntity, "The same venue section was given more than once")

		case errors.Is(err, domain.ErrEventSeatSectionsRequired):
			writeError(w, http.StatusUnprocessableEntity, "At least one venue section is required")

		case errors.Is(err, domain.ErrEventSeatSectionsLimitExceeded):
			writeError(w, http.StatusUnprocessableEntity, "Too many venue sections")

		case errors.Is(err, domain.ErrEventSeatAlreadyExists):
			writeError(w, http.StatusConflict, "One or more event seats already exist")

		default:
			addRequestLogAttrs(ctx,
				slog.String("error_code", "event_seat_creation_failed"),
				slog.Any("error", err),
			)

			writeError(w, http.StatusInternalServerError, "The server encountered a problem and could not process your request")
		}

		return
	}

	writeJSON(w, http.StatusCreated, newEventSeatsResponse(eventSeats))
}

func (s *Server) handleEventSeatByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := domain.ParseEventSeatID(r.PathValue("eventSeatID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid event seat ID")
		return
	}

	eventSeat, err := s.dependencies.EventSeats.EventSeatByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrEventSeatNotFound) {
			writeError(w, http.StatusNotFound, "Event seat not found")
			return
		}

		addRequestLogAttrs(ctx,
			slog.String("error_code", "event_seat_lookup_failed"),
			slog.Any("error", err),
		)

		writeError(w, http.StatusInternalServerError, "The server encountered a problem and could not process your request")
		return
	}

	writeJSON(w, http.StatusOK, newEventSeatResponse(eventSeat))
}
