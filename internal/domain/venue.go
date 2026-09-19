package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrVenueNotFound = errors.New("venue not found")
)

type Venue struct {
	ID        VenueID
	Name      string
	CreatedAt time.Time
	UpdatedAt *time.Time
}

type VenueID uuid.UUID

func NewVenueID() (VenueID, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return VenueID{}, fmt.Errorf("generate venue ID: %w", err)
	}

	return VenueID(id), nil
}

func ParseVenueID(s string) (VenueID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return VenueID{}, fmt.Errorf("parse venue ID: %w", err)
	}

	if id.Version() != 7 {
		return VenueID{}, fmt.Errorf("parse venue ID: expected UUID v7")
	}

	return VenueID(id), nil
}

func (id VenueID) String() string {
	return uuid.UUID(id).String()
}

func (id VenueID) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

func (id *VenueID) UnmarshalText(data []byte) error {
	parsed, err := ParseVenueID(string(data))
	if err != nil {
		return err
	}

	*id = parsed
	return nil
}
