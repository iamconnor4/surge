package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEventNotFound           = errors.New("event not found")
	ErrEventInvalidTitle       = errors.New("event title is invalid")
	ErrEventInvalidDescription = errors.New("event description is invalid")
	ErrEventStartsAtInPast     = errors.New("event start time is in the past")
	ErrEventStartsAtRequired   = errors.New("event start time is required")
)

type CreateEventInput struct {
	VenueID     VenueID
	Title       string
	Description string
	StartsAt    time.Time
}

type Event struct {
	ID          EventID
	VenueID     VenueID
	Title       string
	Description string
	StartsAt    time.Time
	CreatedAt   time.Time
	UpdatedAt   *time.Time
}

type EventID uuid.UUID

func NewEventID() (EventID, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return EventID{}, fmt.Errorf("generate event ID: %w", err)
	}

	return EventID(id), nil
}

func ParseEventID(s string) (EventID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return EventID{}, fmt.Errorf("parse event ID: %w", err)
	}

	if id.Version() != 7 {
		return EventID{}, fmt.Errorf("parse event ID: expected UUID v7")
	}

	return EventID(id), nil
}

func (id EventID) String() string {
	return uuid.UUID(id).String()
}

func (id EventID) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

func (id *EventID) UnmarshalText(data []byte) error {
	parsed, err := ParseEventID(string(data))
	if err != nil {
		return err
	}

	*id = parsed
	return nil
}
