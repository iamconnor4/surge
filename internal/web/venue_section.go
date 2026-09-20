package web

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/iamconnor4/surge/internal/domain"
)

type VenueSectionManager interface {
	VenueSectionByID(ctx context.Context, id domain.VenueSectionID) (domain.VenueSection, error)
	CreateVenueSection(ctx context.Context, input domain.CreateVenueSectionInput) (domain.VenueSection, error)
}

type createVenueSectionRequest struct {
	Name string `json:"name"`
}

type venueSectionResponse struct {
	ID        string     `json:"id"`
	VenueID   string     `json:"venueId"`
	Name      string     `json:"name"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
}

func newVenueSectionResponse(vS domain.VenueSection) venueSectionResponse {
	return venueSectionResponse{
		ID:        vS.ID.String(),
		VenueID:   vS.VenueID.String(),
		Name:      vS.Name,
		CreatedAt: vS.CreatedAt,
		UpdatedAt: vS.UpdatedAt,
	}
}

func (s *Server) handleCreateVenueSection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	venueID, err := domain.ParseVenueID(r.PathValue("venueID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid venue ID")
		return
	}

	var request createVenueSectionRequest
	if err := readJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	venueSection, err := s.dependencies.VenueSections.CreateVenueSection(
		ctx,
		domain.CreateVenueSectionInput{
			VenueID: venueID,
			Name:    request.Name,
		},
	)

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrVenueNotFound):
			writeError(w, http.StatusNotFound, "Venue not found")
		case errors.Is(err, domain.ErrVenueSectionInvalidName):
			writeError(w, http.StatusUnprocessableEntity, "Name is required")
		default:
			addRequestLogAttrs(ctx,
				slog.String("error_code", "venue_section_creation_failed"),
				slog.Any("error", err),
			)

			writeError(w, http.StatusInternalServerError, "The server encountered a problem and could not process your request")
		}

		return
	}

	writeJSON(w, http.StatusCreated, newVenueSectionResponse(venueSection))
}

func (s *Server) handleVenueSectionByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := domain.ParseVenueSectionID(r.PathValue("venueSectionID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid venue section ID")
		return
	}

	venueSection, err := s.dependencies.VenueSections.VenueSectionByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrVenueSectionNotFound) {
			writeError(w, http.StatusNotFound, "Venue section not found")
			return
		}

		addRequestLogAttrs(ctx,
			slog.String("error_code", "venue_section_lookup_failed"),
			slog.Any("error", err),
		)

		writeError(w, http.StatusInternalServerError, "The server encountered a problem and could not process your request")
		return
	}

	writeJSON(w, http.StatusOK, newVenueSectionResponse(venueSection))
}
