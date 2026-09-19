-- name: GetUserById :one
SELECT id, email, first_name, last_name, created_at, updated_at
FROM users WHERE id = $1;

-- name: CreateUser :exec
INSERT INTO users (
    id,
    email,
    first_name,
    last_name,
    created_at
)
VALUES ($1, $2, $3, $4, $5);
