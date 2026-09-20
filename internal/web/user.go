package web

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/iamconnor4/surge/internal/domain"
)

type UserManager interface {
	UserByID(ctx context.Context, id domain.UserID) (domain.User, error)
	CreateUser(ctx context.Context, input domain.CreateUserInput) (domain.User, error)
}

type createUserRequest struct {
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type userResponse struct {
	ID        string     `json:"id"`
	Email     string     `json:"email"`
	FirstName string     `json:"firstName"`
	LastName  string     `json:"lastName"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
}

func newUserResponse(u domain.User) userResponse {
	return userResponse{
		ID:        u.ID.String(),
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var request createUserRequest

	if err := readJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := s.dependencies.Users.CreateUser(
		ctx,
		domain.CreateUserInput{
			Email:     request.Email,
			FirstName: request.FirstName,
			LastName:  request.LastName,
		},
	)

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserAlreadyExists):
			writeError(w, http.StatusConflict, "A user with this email already exists")
		case errors.Is(err, domain.ErrUserInvalidEmail):
			writeError(w, http.StatusUnprocessableEntity, "Email is required")
		case errors.Is(err, domain.ErrUserInvalidFirstName):
			writeError(w, http.StatusUnprocessableEntity, "First name is required")
		case errors.Is(err, domain.ErrUserInvalidLastName):
			writeError(w, http.StatusUnprocessableEntity, "Last name is required")
		default:
			addRequestLogAttrs(ctx,
				slog.String("error_code", "user_creation_failed"),
				slog.Any("error", err),
			)

			writeError(w, http.StatusInternalServerError, "The server encountered a problem and could not process your request")
		}

		return
	}

	writeJSON(w, http.StatusCreated, newUserResponse(user))
}

func (s *Server) handleUserByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := domain.ParseUserID(r.PathValue("userID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	user, err := s.dependencies.Users.UserByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "User not found")
			return
		}

		addRequestLogAttrs(ctx,
			slog.String("error_code", "user_lookup_failed"),
			slog.Any("error", err),
		)

		writeError(w, http.StatusInternalServerError, "The server encountered a problem and could not process your request")
		return
	}

	writeJSON(w, http.StatusOK, newUserResponse(user))
}
