package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

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
		if errors.Is(err, pgx.ErrNoRows) {
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

func (s *User) CreateUser(ctx context.Context, input domain.CreateUserInput) (domain.User, error) {
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.FirstName = strings.TrimSpace(input.FirstName)
	input.LastName = strings.TrimSpace(input.LastName)

	if input.Email == "" {
		return domain.User{}, domain.ErrUserInvalidEmail
	}
	if input.FirstName == "" {
		return domain.User{}, domain.ErrUserInvalidFirstName
	}
	if input.LastName == "" {
		return domain.User{}, domain.ErrUserInvalidLastName
	}

	id, err := domain.NewUserID()
	if err != nil {
		return domain.User{}, fmt.Errorf("generate user ID: %w", err)
	}

	user := domain.User{
		ID:        id,
		Email:     input.Email,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		CreatedAt: time.Now().UTC(),
	}

	err = s.queries.CreateUser(ctx, db.CreateUserParams{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: user.CreatedAt,
	})

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) &&
			pgErr.ConstraintName == "users_email_key" {
			return domain.User{}, domain.ErrUserAlreadyExists
		}

		return domain.User{}, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}
