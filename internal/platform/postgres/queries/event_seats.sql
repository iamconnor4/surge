-- name: GetEventSeatById :one
SELECT id, event_id, venue_seat_id, price_pence, is_available, created_at, updated_at
FROM event_seats WHERE id = $1;

-- name: CreateEventSeats :copyfrom
INSERT INTO event_seats (
    id,
    event_id,
    venue_seat_id,
    price_pence,
    is_available,
    created_at
)
VALUES ($1, $2, $3, $4, $5, $6);
