-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByIDForUpdate :one
SELECT * FROM users
WHERE id = $1 LIMIT 1 FOR UPDATE;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: CreateOAuthUser :one
INSERT INTO users (
    email, full_name, avatar_url, phone, role
) VALUES (
    $1, $2, $3, $4, 'customer'
)
ON CONFLICT (email) DO UPDATE 
SET full_name = EXCLUDED.full_name,
    avatar_url = EXCLUDED.avatar_url,
    updated_at = NOW()
RETURNING *;

-- name: CreateStaffUser :one
INSERT INTO users (
    email, full_name, phone, role, password_hash
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: UpdateUserLastOrder :exec
UPDATE users
SET last_order_at = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;
