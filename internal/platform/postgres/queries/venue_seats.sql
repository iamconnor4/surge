-- name: GetVenueSeatById :one
SELECT id, venue_section_id, row_label, seat_label, created_at, updated_at
FROM venue_seats WHERE id = $1;

-- name: GetVenueSeatsBySectionIds :many
SELECT id, venue_section_id
FROM venue_seats WHERE venue_section_id = ANY(sqlc.slice('venue_section_ids'));

-- name: CreateVenueSeats :copyfrom
-- COPY is used because venue layouts are created
-- in bulk and no database generated values need to be returned.
INSERT INTO venue_seats (
    id,
    venue_section_id,
    row_label,
    seat_label,
    created_at
)
VALUES ($1, $2, $3, $4, $5);
