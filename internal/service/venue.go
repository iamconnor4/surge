package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/iamconnor4/surge/internal/domain"
	"github.com/iamconnor4/surge/internal/platform/postgres/db"
)

type Venue struct {
	queries *db.Queries
}

func NewVenue(queries *db.Queries) *Venue {
	return &Venue{queries: queries}
}

func (s *Venue) VenueByID(ctx context.Context, id domain.VenueID) (domain.Venue, error) {
	dbVenue, err := s.queries.GetVenueById(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Venue{}, domain.ErrVenueNotFound
		}

		return domain.Venue{}, fmt.Errorf("get venue: %w", err)
	}

	return domain.Venue{
		ID:        dbVenue.ID,
		Name:      dbVenue.Name,
		CreatedAt: dbVenue.CreatedAt,
		UpdatedAt: dbVenue.UpdatedAt,
	}, nil
}
