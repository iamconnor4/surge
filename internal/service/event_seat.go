package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/iamconnor4/surge/internal/domain"
	"github.com/iamconnor4/surge/internal/platform/postgres/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const maxSectionsPerBatch = 1000

type EventSeat struct {
	queries *db.Queries
}

func NewEventSeat(queries *db.Queries) *EventSeat {
	return &EventSeat{queries: queries}
}

func (s *EventSeat) EventSeatByID(ctx context.Context, id domain.EventSeatID) (domain.EventSeat, error) {
	dbEventSeat, err := s.queries.GetEventSeatById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.EventSeat{}, domain.ErrEventSeatNotFound
		}

		return domain.EventSeat{}, fmt.Errorf("get event seat: %w", err)
	}

	return domain.EventSeat{
		ID:          dbEventSeat.ID,
		EventID:     dbEventSeat.EventID,
		VenueSeatID: dbEventSeat.VenueSeatID,
		PricePence:  dbEventSeat.PricePence,
		IsAvailable: dbEventSeat.IsAvailable,
		CreatedAt:   dbEventSeat.CreatedAt,
		UpdatedAt:   dbEventSeat.UpdatedAt,
	}, nil
}

func (s *EventSeat) CreateEventSeats(ctx context.Context, input domain.CreateEventSeatsInput) ([]domain.EventSeat, error) {
	if len(input.Sections) == 0 {
		return []domain.EventSeat{}, domain.ErrEventSeatSectionsRequired
	}
	if len(input.Sections) > maxSectionsPerBatch {
		return []domain.EventSeat{}, domain.ErrEventSeatSectionsLimitExceeded
	}

	event, err := s.queries.GetEventById(ctx, input.EventID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []domain.EventSeat{}, domain.ErrEventNotFound
		}

		return []domain.EventSeat{}, fmt.Errorf("get event: %w", err)
	}

	sectionIDs := make([]domain.VenueSectionID, len(input.Sections))
	priceBySection := make(map[domain.VenueSectionID]int64, len(input.Sections))
	for i, section := range input.Sections {
		if section.PricePence < 0 {
			return []domain.EventSeat{}, domain.ErrEventSeatPricePenceBelowZero
		}
		if _, exists := priceBySection[section.VenueSectionID]; exists {
			return []domain.EventSeat{}, domain.ErrEventSeatDuplicateSection
		}
		priceBySection[section.VenueSectionID] = section.PricePence
		sectionIDs[i] = section.VenueSectionID
	}

	venueSections, err := s.queries.GetVenueSectionsByIds(ctx, sectionIDs)
	if err != nil {
		return []domain.EventSeat{}, fmt.Errorf("get venue sections: %w", err)
	}

	if len(venueSections) != len(sectionIDs) {
		return []domain.EventSeat{}, domain.ErrVenueSectionNotFound
	}

	for _, section := range venueSections {
		if section.VenueID != event.VenueID {
			return []domain.EventSeat{}, domain.ErrEventSeatVenueMismatch
		}
	}

	venueSeats, err := s.queries.GetVenueSeatsBySectionIds(ctx, sectionIDs)
	if err != nil {
		return []domain.EventSeat{}, fmt.Errorf("get venue seats: %w", err)
	}

	currentTime := time.Now().UTC()
	eventSeats := make([]domain.EventSeat, len(venueSeats))
	eventSeatsParams := make([]db.CreateEventSeatsParams, len(venueSeats))

	for i, row := range venueSeats {
		id, err := domain.NewEventSeatID()
		if err != nil {
			return []domain.EventSeat{}, fmt.Errorf("generate event seat ID: %w", err)
		}

		eventSeats[i] = domain.EventSeat{
			ID:          id,
			EventID:     input.EventID,
			VenueSeatID: row.ID,
			PricePence:  priceBySection[row.VenueSectionID],
			IsAvailable: true,
			CreatedAt:   currentTime,
		}

		eventSeatsParams[i] = db.CreateEventSeatsParams{
			ID:          id,
			EventID:     input.EventID,
			VenueSeatID: row.ID,
			PricePence:  priceBySection[row.VenueSectionID],
			IsAvailable: true,
			CreatedAt:   currentTime,
		}
	}

	inserted, err := s.queries.CreateEventSeats(ctx, eventSeatsParams)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			switch {
			case pgErr.Code == "23505" &&
				pgErr.ConstraintName == "event_seats_event_id_venue_seat_id_key":
				return []domain.EventSeat{}, domain.ErrEventSeatAlreadyExists
			case pgErr.Code == "23503" &&
				pgErr.ConstraintName == "event_seats_event_id_fkey":
				return []domain.EventSeat{}, domain.ErrEventNotFound
			case pgErr.Code == "23503" &&
				pgErr.ConstraintName == "event_seats_venue_seat_id_fkey":
				return []domain.EventSeat{}, domain.ErrVenueSeatNotFound
			}
		}

		return []domain.EventSeat{}, fmt.Errorf("create event seats: %w", err)
	}

	if inserted != int64(len(eventSeats)) {
		return nil, fmt.Errorf("create event seats: inserted %d of %d seats",
			inserted,
			len(eventSeats))
	}

	return eventSeats, nil
}
