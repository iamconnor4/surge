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

const maxVenueSeatsPerBatch = 1000

type VenueSeat struct {
	queries *db.Queries
}

func NewVenueSeat(queries *db.Queries) *VenueSeat {
	return &VenueSeat{queries: queries}
}

func (s *VenueSeat) VenueSeatByID(ctx context.Context, id domain.VenueSeatID) (domain.VenueSeat, error) {
	dbVenueSeat, err := s.queries.GetVenueSeatById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.VenueSeat{}, domain.ErrVenueSeatNotFound
		}

		return domain.VenueSeat{}, fmt.Errorf("get venue seat: %w", err)
	}

	return domain.VenueSeat{
		ID:             dbVenueSeat.ID,
		VenueSectionID: dbVenueSeat.VenueSectionID,
		RowLabel:       dbVenueSeat.RowLabel,
		SeatLabel:      dbVenueSeat.SeatLabel,
		CreatedAt:      dbVenueSeat.CreatedAt,
		UpdatedAt:      dbVenueSeat.UpdatedAt,
	}, nil
}

func (s *VenueSeat) CreateVenueSeats(ctx context.Context, input domain.CreateVenueSeatsInput) ([]domain.VenueSeat, error) {
	if len(input.Seats) == 0 {
		return []domain.VenueSeat{}, domain.ErrVenueSeatRequired
	}
	if len(input.Seats) > maxVenueSeatsPerBatch {
		return []domain.VenueSeat{}, domain.ErrVenueSeatLimitExceeded
	}

	venueSeats := make([]domain.VenueSeat, len(input.Seats))
	venueSeatsParams := make([]db.CreateVenueSeatsParams, len(input.Seats))

	currentTime := time.Now().UTC()
	for i, seat := range input.Seats {
		rowLabel := strings.TrimSpace(seat.RowLabel)
		if rowLabel == "" {
			return []domain.VenueSeat{}, domain.ErrVenueSeatRowLabel
		}

		seatLabel := strings.TrimSpace(seat.SeatLabel)
		if seatLabel == "" {
			return []domain.VenueSeat{}, domain.ErrVenueSeatSeatLabel
		}

		id, err := domain.NewVenueSeatID()
		if err != nil {
			return []domain.VenueSeat{}, fmt.Errorf("generate venue seat ID: %w", err)
		}

		venueSeats[i] = domain.VenueSeat{
			ID:             id,
			VenueSectionID: input.VenueSectionID,
			RowLabel:       rowLabel,
			SeatLabel:      seatLabel,
			CreatedAt:      currentTime,
		}

		venueSeatsParams[i] = db.CreateVenueSeatsParams{
			ID:             id,
			VenueSectionID: input.VenueSectionID,
			RowLabel:       rowLabel,
			SeatLabel:      seatLabel,
			CreatedAt:      currentTime,
		}
	}

	inserted, err := s.queries.CreateVenueSeats(ctx, venueSeatsParams)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			switch {
			case pgErr.Code == "23505" &&
				pgErr.ConstraintName == "venue_seats_venue_section_id_row_label_seat_label_key":
				return []domain.VenueSeat{}, domain.ErrVenueSeatAlreadyExists
			case pgErr.Code == "23503" &&
				pgErr.ConstraintName == "venue_seats_venue_section_id_fkey":
				return []domain.VenueSeat{}, domain.ErrVenueSectionNotFound
			}
		}

		return []domain.VenueSeat{}, fmt.Errorf("create venue seats: %w", err)
	}

	if inserted != int64(len(venueSeats)) {
		return nil, fmt.Errorf("create venue seats: inserted %d of %d seats",
			inserted,
			len(venueSeats))
	}

	return venueSeats, nil
}
