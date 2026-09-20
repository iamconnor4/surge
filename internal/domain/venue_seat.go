package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrVenueSeatNotFound      = errors.New("venue seat not found")
	ErrVenueSeatRowLabel      = errors.New("venue seat row label is invalid")
	ErrVenueSeatSeatLabel     = errors.New("venue seat seat label is invalid")
	ErrVenueSeatAlreadyExists = errors.New("venue seat already exists")
	ErrVenueSeatRequired      = errors.New("at least one venue seat is required")
	ErrVenueSeatLimitExceeded = errors.New("too many venue seats")
)

type CreateVenueSeatsInput struct {
	VenueSectionID VenueSectionID
	Seats          []CreateVenueSeatInput
}

type CreateVenueSeatInput struct {
	RowLabel  string
	SeatLabel string
}

type VenueSeat struct {
	ID             VenueSeatID
	VenueSectionID VenueSectionID
	RowLabel       string
	SeatLabel      string
	CreatedAt      time.Time
	UpdatedAt      *time.Time
}

type VenueSeatID uuid.UUID

func NewVenueSeatID() (VenueSeatID, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return VenueSeatID{}, fmt.Errorf("generate venue seat ID: %w", err)
	}

	return VenueSeatID(id), nil
}

func ParseVenueSeatID(s string) (VenueSeatID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return VenueSeatID{}, fmt.Errorf("parse venue seat ID: %w", err)
	}

	if id.Version() != 7 {
		return VenueSeatID{}, fmt.Errorf("parse venue seat ID: expected UUID v7")
	}

	return VenueSeatID(id), nil
}

func (id VenueSeatID) String() string {
	return uuid.UUID(id).String()
}

func (id VenueSeatID) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

func (id *VenueSeatID) UnmarshalText(data []byte) error {
	parsed, err := ParseVenueSeatID(string(data))
	if err != nil {
		return err
	}

	*id = parsed
	return nil
}
