package web

import (
	"context"
	"io"
	"log/slog"

	"github.com/iamconnor4/surge/internal/domain"
)

const testUUIDv7 = "01994fa0-7c4f-7d22-8a3e-4e2fdc262e30"

type fakeUserManager struct {
	userByID   func(context.Context, domain.UserID) (domain.User, error)
	createUser func(context.Context, domain.CreateUserInput) (domain.User, error)
}

func (f fakeUserManager) UserByID(ctx context.Context, id domain.UserID) (domain.User, error) {
	return f.userByID(ctx, id)
}

func (f fakeUserManager) CreateUser(ctx context.Context, input domain.CreateUserInput) (domain.User, error) {
	return f.createUser(ctx, input)
}

type venueManagerFunc func(context.Context, domain.VenueID) (domain.Venue, error)

func (f venueManagerFunc) VenueByID(ctx context.Context, id domain.VenueID) (domain.Venue, error) {
	return f(ctx, id)
}

func testServer(dependencies Dependencies) *Server {
	return &Server{
		dependencies: dependencies,
		logger:       slog.New(slog.NewTextHandler(io.Discard, nil)),
		service:      "surge",
		environment:  "test",
	}
}
