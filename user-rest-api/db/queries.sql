-- name: CreateUser :one
INSERT INTO users (first_name, last_name, email, phone, age, status)
VALUES ($1, $2, $3, $4, $5, $6)
    RETURNING *;

-- name: GetUser :one
SELECT * FROM users WHERE user_id = $1;

-- name: ListUsers :many
SELECT * FROM users;

-- name: UpdateUser :one
UPDATE users
SET first_name = $2, last_name = $3, email = $4, phone = $5, age = $6, status = $7
WHERE user_id = $1
    RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE user_id = $1;
