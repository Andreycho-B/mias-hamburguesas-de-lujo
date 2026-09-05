-- name: CreateOrder :one
INSERT INTO orders (
    code, user_id, order_type, status, subtotal, delivery_fee, discount_amount, total_amount,
    delivery_name, delivery_phone, delivery_address, delivery_notes, payment_method, cash_change_amount
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
)
RETURNING *;

-- name: CreateOrderItem :one
INSERT INTO order_items (
    order_id, product_id, quantity, unit_price, customizations
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetOrderByID :one
SELECT * FROM orders
WHERE id = $1 LIMIT 1;

-- name: GetOrderByCode :one
SELECT * FROM orders
WHERE code = $1 LIMIT 1;

-- name: GetOrderItems :many
SELECT oi.*, p.name as product_name, p.bread_type, p.image_url
FROM order_items oi
JOIN products p ON oi.product_id = p.id
WHERE oi.order_id = $1;

-- name: UpdateOrderStatus :one
UPDATE orders
SET status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ListOrdersByUserID :many
SELECT * FROM orders
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListPendingInStoreOrders :many
SELECT * FROM orders
WHERE order_type = 'in_store' AND status = 'PENDING_PAYMENT'
ORDER BY created_at ASC;

-- name: ListActiveDeliveryOrders :many
SELECT * FROM orders
WHERE order_type = 'delivery' AND status IN ('IN_PREPARATION', 'ON_THE_WAY')
ORDER BY created_at ASC;

-- name: ListRecentOrdersForStaff :many
SELECT * FROM orders
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountBurgersInOrder :one
SELECT COALESCE(SUM(oi.quantity), 0)::INT as total_burgers
FROM order_items oi
JOIN products p ON oi.product_id = p.id
JOIN categories c ON p.category_id = c.id
WHERE oi.order_id = $1 AND c.slug = 'hamburguesas';
