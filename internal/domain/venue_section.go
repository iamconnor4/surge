package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrVenueSectionNotFound    = errors.New("venue section not found")
	ErrVenueSectionInvalidName = errors.New("venue section name is invalid")
)

type CreateVenueSectionInput struct {
	VenueID VenueID
	Name    string
}

type VenueSection struct {
	ID        VenueSectionID
	VenueID   VenueID
	Name      string
	CreatedAt time.Time
	UpdatedAt *time.Time
}

type VenueSectionID uuid.UUID

func NewVenueSectionID() (VenueSectionID, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return VenueSectionID{}, fmt.Errorf("generate venue section ID: %w", err)
	}

	return VenueSectionID(id), nil
}

func ParseVenueSectionID(s string) (VenueSectionID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return VenueSectionID{}, fmt.Errorf("parse venue section ID: %w", err)
	}

	if id.Version() != 7 {
		return VenueSectionID{}, fmt.Errorf("parse venue section ID: expected UUID v7")
	}

	return VenueSectionID(id), nil
}

func (id VenueSectionID) String() string {
	return uuid.UUID(id).String()
}

func (id VenueSectionID) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

func (id *VenueSectionID) UnmarshalText(data []byte) error {
	parsed, err := ParseVenueSectionID(string(data))
	if err != nil {
		return err
	}

	*id = parsed
	return nil
}
