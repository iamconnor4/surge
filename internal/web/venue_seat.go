package web

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/iamconnor4/surge/internal/domain"
)

type VenueSeatManager interface {
	VenueSeatByID(ctx context.Context, id domain.VenueSeatID) (domain.VenueSeat, error)
	CreateVenueSeats(ctx context.Context, input domain.CreateVenueSeatsInput) ([]domain.VenueSeat, error)
}

type createVenueSeatsRequest struct {
	Seats []createVenueSeatRequest `json:"seats"`
}

type createVenueSeatRequest struct {
	RowLabel  string `json:"rowLabel"`
	SeatLabel string `json:"seatLabel"`
}

type venueSeatsResponse struct {
	Seats []venueSeatResponse `json:"seats"`
}

type venueSeatResponse struct {
	ID             string     `json:"id"`
	VenueSectionID string     `json:"venueSectionId"`
	RowLabel       string     `json:"rowLabel"`
	SeatLabel      string     `json:"seatLabel"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      *time.Time `json:"updatedAt,omitempty"`
}

func newVenueSeatsResponse(seats []domain.VenueSeat) venueSeatsResponse {
	venueSeats := make([]venueSeatResponse, len(seats))

	for i, seat := range seats {
		venueSeats[i] = newVenueSeatResponse(seat)
	}

	return venueSeatsResponse{
		Seats: venueSeats,
	}
}

func newVenueSeatResponse(seat domain.VenueSeat) venueSeatResponse {
	return venueSeatResponse{
		ID:             seat.ID.String(),
		VenueSectionID: seat.VenueSectionID.String(),
		RowLabel:       seat.RowLabel,
		SeatLabel:      seat.SeatLabel,
		CreatedAt:      seat.CreatedAt,
		UpdatedAt:      seat.UpdatedAt,
	}
}

func (s *Server) handleCreateVenueSeats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	venueSectionID, err := domain.ParseVenueSectionID(r.PathValue("venueSectionID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid venue section ID")
		return
	}

	var request createVenueSeatsRequest
	if err := readJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	venueSeatsInput := make([]domain.CreateVenueSeatInput, len(request.Seats))
	for i, seat := range request.Seats {
		venueSeatsInput[i] = domain.CreateVenueSeatInput{
			RowLabel:  seat.RowLabel,
			SeatLabel: seat.SeatLabel,
		}
	}

	venueSeats, err := s.dependencies.VenueSeats.CreateVenueSeats(
		ctx,
		domain.CreateVenueSeatsInput{
			VenueSectionID: venueSectionID,
			Seats:          venueSeatsInput,
		},
	)

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrVenueSectionNotFound):
			writeError(w, http.StatusNotFound, "Venue section not found")

		case errors.Is(err, domain.ErrVenueSeatRowLabel):
			writeError(w, http.StatusUnprocessableEntity, "One or more venue seats are missing a row label")

		case errors.Is(err, domain.ErrVenueSeatSeatLabel):
			writeError(w, http.StatusUnprocessableEntity, "One or more venue seats are missing a seat label")

		case errors.Is(err, domain.ErrVenueSeatAlreadyExists):
			writeError(w, http.StatusConflict, "One or more venue seats already exist")

		case errors.Is(err, domain.ErrVenueSeatRequired):
			writeError(w, http.StatusUnprocessableEntity, "At least one venue seat is required")

		case errors.Is(err, domain.ErrVenueSeatLimitExceeded):
			writeError(w, http.StatusUnprocessableEntity, "Too many venue seats")

		default:
			addRequestLogAttrs(ctx,
				slog.String("error_code", "venue_seat_creation_failed"),
				slog.Any("error", err),
			)

			writeError(w, http.StatusInternalServerError, "The server encountered a problem and could not process your request")
		}

		return
	}

	writeJSON(w, http.StatusCreated, newVenueSeatsResponse(venueSeats))
}

func (s *Server) handleVenueSeatByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := domain.ParseVenueSeatID(r.PathValue("venueSeatID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid venue seat ID")
		return
	}

	venueSeat, err := s.dependencies.VenueSeats.VenueSeatByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrVenueSeatNotFound) {
			writeError(w, http.StatusNotFound, "Venue seat not found")
			return
		}

		addRequestLogAttrs(ctx,
			slog.String("error_code", "venue_seat_lookup_failed"),
			slog.Any("error", err),
		)

		writeError(w, http.StatusInternalServerError, "The server encountered a problem and could not process your request")
		return
	}

	writeJSON(w, http.StatusOK, newVenueSeatResponse(venueSeat))
}
