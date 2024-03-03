-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: UpdateUser :exec
UPDATE users
SET isverified = TRUE
WHERE email = $1;

-- name: CreateUser :one
INSERT INTO users ( email, password, role, isverified, otp)
VALUES ($1, $2, $3, false, $4)
RETURNING *;
