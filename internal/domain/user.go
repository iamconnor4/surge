package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrUserAlreadyExists    = errors.New("user already exists with this email")
	ErrUserInvalidEmail     = errors.New("user email is invalid")
	ErrUserInvalidFirstName = errors.New("user first name is invalid")
	ErrUserInvalidLastName  = errors.New("user last name is invalid")
)

type CreateUserInput struct {
	Email     string
	FirstName string
	LastName  string
}

type User struct {
	ID        UserID
	Email     string
	FirstName string
	LastName  string
	CreatedAt time.Time
	UpdatedAt *time.Time
}

type UserID uuid.UUID

func NewUserID() (UserID, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return UserID{}, fmt.Errorf("generate user ID: %w", err)
	}

	return UserID(id), nil
}

func ParseUserID(s string) (UserID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return UserID{}, fmt.Errorf("parse user ID: %w", err)
	}

	if id.Version() != 7 {
		return UserID{}, fmt.Errorf("parse user ID: expected UUID v7")
	}

	return UserID(id), nil
}

func (id UserID) String() string {
	return uuid.UUID(id).String()
}

func (id UserID) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

func (id *UserID) UnmarshalText(data []byte) error {
	parsed, err := ParseUserID(string(data))
	if err != nil {
		return err
	}

	*id = parsed
	return nil
}
