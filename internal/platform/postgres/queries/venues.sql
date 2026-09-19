-- name: GetVenueById :one
SELECT id, name, created_at, updated_at
FROM venues WHERE id = $1;