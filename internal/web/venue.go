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

func (s *Server) handleVenueByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := domain.ParseVenueID(r.PathValue("id"))
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
