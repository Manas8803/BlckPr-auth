-- name: CreateUser :one
INSERT INTO users(
    firstname, middlename, lastname, email, password, otp, created_at,updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP,CURRENT_TIMESTAMP
)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: UpdateUserByEmail :exec
UPDATE users SET is_verified = TRUE WHERE email = $1;
