-- name: GetVenueSectionById :one
SELECT id, venue_id, name, created_at, updated_at
FROM venue_sections WHERE id = $1;

-- name: CreateVenueSection :exec
INSERT INTO venue_sections (
    id,
    venue_id,
    name,
    created_at
)
VALUES ($1, $2, $3, $4);
