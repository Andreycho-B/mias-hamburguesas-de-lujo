-- name: CreateAuditLog :one
INSERT INTO admin_audit_logs (
    user_id, action, details, ip_address
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: ListAuditLogs :many
SELECT al.*, u.full_name as user_name, u.role as user_role
FROM admin_audit_logs al
JOIN users u ON al.user_id = u.id
ORDER BY al.created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetTopSellingProducts :many
SELECT p.id, p.name, p.slug, p.price, p.image_url,
       COALESCE(SUM(oi.quantity), 0)::INT as units_sold,
       COALESCE(SUM(oi.quantity * oi.unit_price), 0)::INT as total_revenue
FROM products p
LEFT JOIN order_items oi ON p.id = oi.product_id
LEFT JOIN orders o ON oi.order_id = o.id AND o.status = 'PAID'
GROUP BY p.id, p.name, p.slug, p.price, p.image_url
ORDER BY units_sold DESC
LIMIT $1;

-- name: GetSalesMetrics :one
SELECT 
    COUNT(CASE WHEN order_type = 'in_store' AND status = 'PAID' THEN 1 END)::INT as in_store_orders,
    COUNT(CASE WHEN order_type = 'delivery' AND status IN ('DELIVERED', 'ON_THE_WAY') THEN 1 END)::INT as delivery_orders,
    COALESCE(SUM(CASE WHEN status IN ('PAID', 'DELIVERED') THEN total_amount ELSE 0 END), 0)::INT as total_revenue,
    COALESCE(AVG(CASE WHEN status IN ('PAID', 'DELIVERED') THEN total_amount END), 0)::INT as average_ticket
FROM orders;
