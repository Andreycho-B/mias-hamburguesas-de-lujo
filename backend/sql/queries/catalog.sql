-- name: ListActiveCategories :many
SELECT * FROM categories
WHERE is_active = TRUE
ORDER BY sort_order ASC;

-- name: ListAllCategories :many
SELECT * FROM categories
ORDER BY sort_order ASC;

-- name: ListActiveProducts :many
SELECT p.*, c.name as category_name, c.slug as category_slug
FROM products p
JOIN categories c ON p.category_id = c.id
WHERE p.is_active = TRUE
ORDER BY c.sort_order ASC, p.id ASC;

-- name: ListProductsByCategory :many
SELECT * FROM products
WHERE category_id = $1 AND is_active = TRUE
ORDER BY id ASC;

-- name: GetProductBySlug :one
SELECT p.*, c.name as category_name
FROM products p
JOIN categories c ON p.category_id = c.id
WHERE p.slug = $1 AND p.is_active = TRUE
LIMIT 1;

-- name: GetProductByID :one
SELECT * FROM products
WHERE id = $1 LIMIT 1;

-- name: ToggleProductStock :one
UPDATE products
SET is_in_stock = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CreateProduct :one
INSERT INTO products (
    category_id, slug, name, description, price, bread_type, image_url, is_bestseller, is_chef_choice, is_season, is_in_stock
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
RETURNING *;

-- name: UpdateProduct :one
UPDATE products
SET category_id = $2,
    slug = $3,
    name = $4,
    description = $5,
    price = $6,
    bread_type = $7,
    image_url = $8,
    is_bestseller = $9,
    is_chef_choice = $10,
    is_season = $11,
    is_in_stock = $12,
    updated_at = NOW()
WHERE id = $1
RETURNING *;
