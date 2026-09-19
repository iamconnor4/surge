package web

import "net/http"

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /livez", s.handleLiveness)
	mux.HandleFunc("GET /readyz", s.handleReadiness)

	mux.HandleFunc("GET /users/{id}", s.handleUserByID)
	mux.HandleFunc("GET /venues/{id}", s.handleVenueByID)

	return s.logRequests(mux)
}
