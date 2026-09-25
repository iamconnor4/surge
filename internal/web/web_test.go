package web

import (
	"context"
	"io"
	"log/slog"

	"github.com/iamconnor4/surge/internal/domain"
)

const (
	testUserID         = "01994fa0-7c4f-7d22-8a3e-4e2fdc262e30"
	testVenueID        = "01994fa0-7c4f-7d22-8a3e-4e2fdc262e30"
	testVenueSectionID = "01994fa0-7c4f-7d22-8a3e-4e2fdc262e31"
	testVenueSeatID    = "01994fa0-7c4f-7d22-8a3e-4e2fdc262e32"
	testVenueSeatID2   = "01994fa0-7c4f-7d22-8a3e-4e2fdc262e33"
	testEventID        = "01994fa0-7c4f-7d22-8a3e-4e2fdc262e34"
	testEventSeatID    = "01994fa0-7c4f-7d22-8a3e-4e2fdc262e35"
	testEventSeatID2   = "01994fa0-7c4f-7d22-8a3e-4e2fdc262e36"
)

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

type fakeVenueManager struct {
	venueByID   func(context.Context, domain.VenueID) (domain.Venue, error)
	createVenue func(context.Context, domain.CreateVenueInput) (domain.Venue, error)
}

func (f fakeVenueManager) VenueByID(ctx context.Context, id domain.VenueID) (domain.Venue, error) {
	return f.venueByID(ctx, id)
}

func (f fakeVenueManager) CreateVenue(ctx context.Context, input domain.CreateVenueInput) (domain.Venue, error) {
	return f.createVenue(ctx, input)
}

type fakeVenueSectionManager struct {
	venueSectionByID   func(context.Context, domain.VenueSectionID) (domain.VenueSection, error)
	createVenueSection func(context.Context, domain.CreateVenueSectionInput) (domain.VenueSection, error)
}

func (f fakeVenueSectionManager) VenueSectionByID(ctx context.Context, id domain.VenueSectionID) (domain.VenueSection, error) {
	return f.venueSectionByID(ctx, id)
}

func (f fakeVenueSectionManager) CreateVenueSection(ctx context.Context, input domain.CreateVenueSectionInput) (domain.VenueSection, error) {
	return f.createVenueSection(ctx, input)
}

type fakeVenueSeatManager struct {
	venueSeatByID    func(context.Context, domain.VenueSeatID) (domain.VenueSeat, error)
	createVenueSeats func(context.Context, domain.CreateVenueSeatsInput) ([]domain.VenueSeat, error)
}

func (f fakeVenueSeatManager) VenueSeatByID(ctx context.Context, id domain.VenueSeatID) (domain.VenueSeat, error) {
	return f.venueSeatByID(ctx, id)
}

func (f fakeVenueSeatManager) CreateVenueSeats(ctx context.Context, input domain.CreateVenueSeatsInput) ([]domain.VenueSeat, error) {
	return f.createVenueSeats(ctx, input)
}

type fakeEventManager struct {
	eventByID   func(context.Context, domain.EventID) (domain.Event, error)
	createEvent func(context.Context, domain.CreateEventInput) (domain.Event, error)
}

func (f fakeEventManager) EventByID(ctx context.Context, id domain.EventID) (domain.Event, error) {
	return f.eventByID(ctx, id)
}

func (f fakeEventManager) CreateEvent(ctx context.Context, input domain.CreateEventInput) (domain.Event, error) {
	return f.createEvent(ctx, input)
}

type fakeEventSeatManager struct {
	eventSeatByID    func(context.Context, domain.EventSeatID) (domain.EventSeat, error)
	createEventSeats func(context.Context, domain.CreateEventSeatsInput) ([]domain.EventSeat, error)
}

func (f fakeEventSeatManager) EventSeatByID(ctx context.Context, id domain.EventSeatID) (domain.EventSeat, error) {
	return f.eventSeatByID(ctx, id)
}

func (f fakeEventSeatManager) CreateEventSeats(ctx context.Context, input domain.CreateEventSeatsInput) ([]domain.EventSeat, error) {
	return f.createEventSeats(ctx, input)
}

func testServer(dependencies Dependencies) *Server {
	return &Server{
		dependencies: dependencies,
		logger:       slog.New(slog.NewTextHandler(io.Discard, nil)),
		service:      "surge",
		environment:  "test",
	}
}
