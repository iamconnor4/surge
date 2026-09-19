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

func (s *Server) handleUserByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := domain.ParseUserID(r.PathValue("id"))
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
