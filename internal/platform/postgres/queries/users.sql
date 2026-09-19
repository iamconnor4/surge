-- name: GetUserById :one
SELECT id, email, first_name, last_name, created_at, updated_at
FROM users WHERE id = $1;
