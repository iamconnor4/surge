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

type Event struct {
	queries *db.Queries
}

func NewEvent(queries *db.Queries) *Event {
	return &Event{queries: queries}
}

func (s *Event) EventByID(ctx context.Context, id domain.EventID) (domain.Event, error) {
	dbEvent, err := s.queries.GetEventById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Event{}, domain.ErrEventNotFound
		}

		return domain.Event{}, fmt.Errorf("get event: %w", err)
	}

	return domain.Event{
		ID:          dbEvent.ID,
		VenueID:     dbEvent.VenueID,
		Title:       dbEvent.Title,
		Description: dbEvent.Description,
		StartsAt:    dbEvent.StartsAt,
		CreatedAt:   dbEvent.CreatedAt,
		UpdatedAt:   dbEvent.UpdatedAt,
	}, nil
}

func (s *Event) CreateEvent(ctx context.Context, input domain.CreateEventInput) (domain.Event, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return domain.Event{}, domain.ErrEventInvalidTitle
	}

	if input.Description == "" {
		return domain.Event{}, domain.ErrEventInvalidDescription
	}

	now := time.Now().UTC()
	startsAt := input.StartsAt.UTC()

	if startsAt.IsZero() {
		return domain.Event{}, domain.ErrEventStartsAtRequired
	}

	if startsAt.Before(now) {
		return domain.Event{}, domain.ErrEventStartsAtInPast
	}

	id, err := domain.NewEventID()
	if err != nil {
		return domain.Event{}, fmt.Errorf("generate event ID: %w", err)
	}

	event := domain.Event{
		ID:          id,
		VenueID:     input.VenueID,
		Title:       input.Title,
		Description: input.Description,
		StartsAt:    startsAt,
		CreatedAt:   now,
	}

	err = s.queries.CreateEvent(ctx, db.CreateEventParams{
		ID:          event.ID,
		VenueID:     event.VenueID,
		Title:       event.Title,
		Description: event.Description,
		StartsAt:    event.StartsAt,
		CreatedAt:   event.CreatedAt,
	})

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) &&
			pgErr.Code == "23503" &&
			pgErr.ConstraintName == "events_venue_id_fkey" {
			return domain.Event{}, domain.ErrVenueNotFound
		}
		return domain.Event{}, fmt.Errorf("create event: %w", err)
	}

	return event, nil
}
