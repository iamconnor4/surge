package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/iamconnor4/surge/internal/domain"
	"github.com/iamconnor4/surge/internal/platform/postgres/db"
)

type User struct {
	queries *db.Queries
}

func NewUser(queries *db.Queries) *User {
	return &User{queries: queries}
}

func (s *User) UserByID(ctx context.Context, id domain.UserID) (domain.User, error) {
	dbUser, err := s.queries.GetUserById(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, domain.ErrUserNotFound
		}

		return domain.User{}, fmt.Errorf("get user: %w", err)
	}

	return domain.User{
		ID:        dbUser.ID,
		Email:     dbUser.Email,
		FirstName: dbUser.FirstName,
		LastName:  dbUser.LastName,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
	}, nil
}
