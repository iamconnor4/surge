package web

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/iamconnor4/surge/internal/domain"
)

type VenueManager interface {
	VenueByID(ctx context.Context, id domain.VenueID) (domain.Venue, error)
	CreateVenue(ctx context.Context, input domain.CreateVenueInput) (domain.Venue, error)
}

type createVenueRequest struct {
	Name string `json:"name"`
}

type venueResponse struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
}

func newVenueResponse(v domain.Venue) venueResponse {
	return venueResponse{
		ID:        v.ID.String(),
		Name:      v.Name,
		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
	}
}

func (s *Server) handleCreateVenue(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var request createVenueRequest

	if err := readJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	venue, err := s.dependencies.Venues.CreateVenue(
		ctx,
		domain.CreateVenueInput{
			Name: request.Name,
		},
	)

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrVenueInvalidName):
			writeError(w, http.StatusUnprocessableEntity, "Name is required")
		default:
			addRequestLogAttrs(ctx,
				slog.String("error_code", "venue_creation_failed"),
				slog.Any("error", err),
			)

			writeError(w, http.StatusInternalServerError, "The server encountered a problem and could not process your request")
		}

		return
	}

	writeJSON(w, http.StatusCreated, newVenueResponse(venue))
}

func (s *Server) handleVenueByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := domain.ParseVenueID(r.PathValue("venueID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid venue ID")
		return
	}

	venue, err := s.dependencies.Venues.VenueByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrVenueNotFound) {
			writeError(w, http.StatusNotFound, "Venue not found")
			return
		}

		addRequestLogAttrs(ctx,
			slog.String("error_code", "venue_lookup_failed"),
			slog.Any("error", err),
		)

		writeError(w, http.StatusInternalServerError, "The server encountered a problem and could not process your request")
		return
	}

	writeJSON(w, http.StatusOK, newVenueResponse(venue))
}
