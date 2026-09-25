package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEventSeatNotFound              = errors.New("event seat not found")
	ErrEventSeatPricePenceBelowZero   = errors.New("event seat price below zero")
	ErrEventSeatSectionsRequired      = errors.New("event seat sections required")
	ErrEventSeatSectionsLimitExceeded = errors.New("event seat sections limit exceeded")
	ErrEventSeatDuplicateSection      = errors.New("event seat section ID given numerous tiems")
	ErrEventSeatVenueMismatch         = errors.New("venue section does not belong to the event's venue")
	ErrEventSeatAlreadyExists         = errors.New("event seat already exists")
)

type CreateEventSeatsInput struct {
	EventID  EventID
	Sections []CreateEventSeatSectionInput
}

type CreateEventSeatSectionInput struct {
	VenueSectionID VenueSectionID
	PricePence     int64
}

type EventSeat struct {
	ID          EventSeatID
	EventID     EventID
	VenueSeatID VenueSeatID
	PricePence  int64
	IsAvailable bool
	CreatedAt   time.Time
	UpdatedAt   *time.Time
}

type EventSeatID uuid.UUID

func NewEventSeatID() (EventSeatID, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return EventSeatID{}, fmt.Errorf("generate event seat ID: %w", err)
	}

	return EventSeatID(id), nil
}

func ParseEventSeatID(s string) (EventSeatID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return EventSeatID{}, fmt.Errorf("parse event seat ID: %w", err)
	}

	if id.Version() != 7 {
		return EventSeatID{}, fmt.Errorf("parse event seat ID: expected UUID v7")
	}

	return EventSeatID(id), nil
}

func (id EventSeatID) String() string {
	return uuid.UUID(id).String()
}

func (id EventSeatID) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

func (id *EventSeatID) UnmarshalText(data []byte) error {
	parsed, err := ParseEventSeatID(string(data))
	if err != nil {
		return err
	}

	*id = parsed
	return nil
}
