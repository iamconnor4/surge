package web

import "net/http"

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /livez", s.handleLiveness)
	mux.HandleFunc("GET /readyz", s.handleReadiness)

	mux.HandleFunc("POST /users", s.handleCreateUser)
	mux.HandleFunc("GET /users/{userID}", s.handleUserByID)

	mux.HandleFunc("POST /venues", s.handleCreateVenue)
	mux.HandleFunc("GET /venues/{venueID}", s.handleVenueByID)

	mux.HandleFunc("POST /venues/{venueID}/sections", s.handleCreateVenueSection)
	mux.HandleFunc("GET /venue-sections/{venueSectionID}", s.handleVenueSectionByID)

	mux.HandleFunc("POST /venue-sections/{venueSectionID}/seats", s.handleCreateVenueSeats)
	mux.HandleFunc("GET /venue-seats/{venueSeatID}", s.handleVenueSeatByID)

	return s.logRequests(mux)
}
