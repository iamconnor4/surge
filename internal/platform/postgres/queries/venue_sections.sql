-- name: GetVenueSectionById :one
SELECT id, venue_id, name, created_at, updated_at
FROM venue_sections WHERE id = $1;

-- name: GetVenueSectionsByIds :many
SELECT id, venue_id
FROM venue_sections WHERE id = ANY(sqlc.slice('venue_section_ids'));

-- name: CreateVenueSection :exec
INSERT INTO venue_sections (
    id,
    venue_id,
    name,
    created_at
)
VALUES ($1, $2, $3, $4);
