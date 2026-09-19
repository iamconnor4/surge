-- name: GetVenueById :one
SELECT id, name, created_at, updated_at
FROM venues WHERE id = $1;

-- name: CreateVenue :exec
INSERT INTO venues (
    id,
    name,
    created_at
)
VALUES ($1, $2, $3);
