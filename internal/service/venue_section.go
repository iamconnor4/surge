package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/iamconnor4/surge/internal/domain"
	"github.com/iamconnor4/surge/internal/platform/postgres/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type VenueSection struct {
	queries *db.Queries
}

func NewVenueSection(queries *db.Queries) *VenueSection {
	return &VenueSection{queries: queries}
}

func (s *VenueSection) VenueSectionByID(ctx context.Context, id domain.VenueSectionID) (domain.VenueSection, error) {
	dbVenueSection, err := s.queries.GetVenueSectionById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.VenueSection{}, domain.ErrVenueSectionNotFound
		}

		return domain.VenueSection{}, fmt.Errorf("get venue section: %w", err)
	}

	return domain.VenueSection{
		ID:        dbVenueSection.ID,
		VenueID:   dbVenueSection.VenueID,
		Name:      dbVenueSection.Name,
		CreatedAt: dbVenueSection.CreatedAt,
		UpdatedAt: dbVenueSection.UpdatedAt,
	}, nil
}

func (s *VenueSection) CreateVenueSection(ctx context.Context, input domain.CreateVenueSectionInput) (domain.VenueSection, error) {
	input.Name = strings.TrimSpace(input.Name)

	if input.Name == "" {
		return domain.VenueSection{}, domain.ErrVenueSectionInvalidName
	}

	id, err := domain.NewVenueSectionID()
	if err != nil {
		return domain.VenueSection{}, fmt.Errorf("generate venue section ID: %w", err)
	}

	venueSection := domain.VenueSection{
		ID:        id,
		VenueID:   input.VenueID,
		Name:      input.Name,
		CreatedAt: time.Now().UTC(),
	}

	err = s.queries.CreateVenueSection(ctx, db.CreateVenueSectionParams{
		ID:        venueSection.ID,
		VenueID:   venueSection.VenueID,
		Name:      venueSection.Name,
		CreatedAt: venueSection.CreatedAt,
	})

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) &&
			pgErr.Code == "23503" &&
			pgErr.ConstraintName == "venue_sections_venue_id_fkey" {
			return domain.VenueSection{}, domain.ErrVenueNotFound
		}

		return domain.VenueSection{}, fmt.Errorf("create venue section: %w", err)
	}

	return venueSection, nil
}
