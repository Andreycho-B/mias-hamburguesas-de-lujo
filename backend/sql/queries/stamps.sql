-- name: CreateStampRequest :one
INSERT INTO stamp_requests (
    order_id, user_id, receipt_number, burgers_count
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: GetStampRequestByID :one
SELECT sr.*, o.code as order_code, o.total_amount, u.full_name as customer_name
FROM stamp_requests sr
JOIN orders o ON sr.order_id = o.id
JOIN users u ON sr.user_id = u.id
WHERE sr.id = $1 LIMIT 1;

-- name: ListPendingStampRequests :many
SELECT sr.*, o.code as order_code, o.total_amount, u.full_name as customer_name
FROM stamp_requests sr
JOIN orders o ON sr.order_id = o.id
JOIN users u ON sr.user_id = u.id
WHERE sr.status = 'PENDING'
ORDER BY sr.created_at ASC;

-- name: UpdateStampRequestStatus :one
UPDATE stamp_requests
SET status = $2,
    reviewed_by = $3,
    reviewed_at = NOW()
WHERE id = $1
RETURNING *;

-- name: AddStampLedgerEntry :one
INSERT INTO stamp_ledger (
    user_id, order_id, amount, reason, expires_at
) VALUES (
    $1, $2, $3, $4, NOW() + INTERVAL '6 months'
)
RETURNING *;

-- name: GetUserStampBalance :one
SELECT COALESCE(SUM(amount), 0)::INT as active_stamps
FROM stamp_ledger
WHERE user_id = $1 AND expires_at > NOW();

-- name: ListUserStampHistory :many
SELECT * FROM stamp_ledger
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CreateRedemptionRequest :one
INSERT INTO redemption_requests (
    user_id
) VALUES (
    $1
)
RETURNING *;

-- name: GetPendingRedemptionByUser :one
SELECT * FROM redemption_requests
WHERE user_id = $1 AND status = 'PENDING'
LIMIT 1;

-- name: ListPendingRedemptions :many
SELECT rr.*, u.full_name as customer_name, u.email as customer_email,
       (SELECT COALESCE(SUM(amount), 0)::INT FROM stamp_ledger WHERE user_id = rr.user_id AND expires_at > NOW()) as current_stamps
FROM redemption_requests rr
JOIN users u ON rr.user_id = u.id
WHERE rr.status = 'PENDING'
ORDER BY rr.created_at ASC;

-- name: UpdateRedemptionStatus :one
UPDATE redemption_requests
SET status = $2,
    reviewed_by = $3,
    reviewed_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ExpirePendingStampRequests :exec
UPDATE stamp_requests
SET status = 'EXPIRED'
WHERE status = 'PENDING' AND created_at < NOW() - INTERVAL '60 minutes';
