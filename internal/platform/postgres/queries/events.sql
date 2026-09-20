-- name: GetEventById :one
SELECT id, venue_id, title, description, starts_at, created_at, updated_at
FROM events WHERE id = $1;

-- name: CreateEvent :exec
INSERT INTO events (
    id,
    venue_id,
    title,
    description,
    starts_at,
    created_at
)
VALUES ($1, $2, $3, $4, $5, $6);
