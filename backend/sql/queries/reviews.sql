-- name: CreateDeliveryReview :one
INSERT INTO delivery_reviews (
    order_id, user_id, rating, comment
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: GetOrderReview :one
SELECT * FROM delivery_reviews
WHERE order_id = $1 LIMIT 1;

-- name: ListRecentReviews :many
SELECT dr.*, u.full_name as customer_name, o.code as order_code
FROM delivery_reviews dr
JOIN users u ON dr.user_id = u.id
JOIN orders o ON dr.order_id = o.id
ORDER BY dr.created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetAverageRating :one
SELECT COALESCE(AVG(rating), 0)::FLOAT as average_rating, COUNT(*)::INT as total_reviews
FROM delivery_reviews;
