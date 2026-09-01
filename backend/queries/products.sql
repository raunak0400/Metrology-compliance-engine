-- name: CreateProduct :one
INSERT INTO products (
  name, sku, brand, barcode, category_id, manufacturer_id,
  net_quantity_declared, mrp_declared, attributes, created_by_id
) VALUES (
  $1, sqlc.narg('sku'), sqlc.narg('brand'), sqlc.narg('barcode'), $2, sqlc.narg('manufacturer_id'),
  sqlc.narg('net_quantity_declared'), sqlc.narg('mrp_declared'), sqlc.narg('attributes'), sqlc.narg('created_by_id')
)
RETURNING *;

-- name: GetProductByID :one
SELECT
  p.*,
  pc.name AS category_name, pc.code AS category_code,
  m.name AS manufacturer_name, m.country AS manufacturer_country
FROM products p
JOIN product_categories pc ON pc.id = p.category_id
LEFT JOIN manufacturers m ON m.id = p.manufacturer_id
WHERE p.id = $1;

-- name: ListProducts :many
SELECT
  p.*,
  pc.name AS category_name, pc.code AS category_code,
  m.name AS manufacturer_name, m.country AS manufacturer_country
FROM products p
JOIN product_categories pc ON pc.id = p.category_id
LEFT JOIN manufacturers m ON m.id = p.manufacturer_id
WHERE
  (sqlc.narg('category_id')::uuid IS NULL OR p.category_id = sqlc.narg('category_id'))
  AND (sqlc.narg('manufacturer_id')::uuid IS NULL OR p.manufacturer_id = sqlc.narg('manufacturer_id'))
  AND (
    sqlc.narg('search')::text IS NULL
    OR p.name ILIKE '%' || sqlc.narg('search') || '%'
    OR p.brand ILIKE '%' || sqlc.narg('search') || '%'
    OR p.sku ILIKE '%' || sqlc.narg('search') || '%'
  )
ORDER BY p.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountProducts :one
SELECT count(*) FROM products p
WHERE
  (sqlc.narg('category_id')::uuid IS NULL OR p.category_id = sqlc.narg('category_id'))
  AND (sqlc.narg('manufacturer_id')::uuid IS NULL OR p.manufacturer_id = sqlc.narg('manufacturer_id'))
  AND (
    sqlc.narg('search')::text IS NULL
    OR p.name ILIKE '%' || sqlc.narg('search') || '%'
    OR p.brand ILIKE '%' || sqlc.narg('search') || '%'
    OR p.sku ILIKE '%' || sqlc.narg('search') || '%'
  );

-- name: UpdateProduct :one
UPDATE products SET
  name = COALESCE(sqlc.narg('name'), name),
  sku = COALESCE(sqlc.narg('sku'), sku),
  brand = COALESCE(sqlc.narg('brand'), brand),
  barcode = COALESCE(sqlc.narg('barcode'), barcode),
  category_id = COALESCE(sqlc.narg('category_id'), category_id),
  manufacturer_id = COALESCE(sqlc.narg('manufacturer_id'), manufacturer_id),
  net_quantity_declared = COALESCE(sqlc.narg('net_quantity_declared'), net_quantity_declared),
  mrp_declared = COALESCE(sqlc.narg('mrp_declared'), mrp_declared),
  attributes = COALESCE(sqlc.narg('attributes'), attributes),
  updated_at = now()
WHERE id = $1
RETURNING *;

-- name: GetCategoryByID :one
SELECT * FROM product_categories WHERE id = $1;

-- name: GetCategoryByCode :one
SELECT * FROM product_categories WHERE code = $1;

-- name: UpsertCategory :one
INSERT INTO product_categories (name, code, description)
VALUES ($1, $2, sqlc.narg('description'))
ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name, description = EXCLUDED.description, updated_at = now()
RETURNING *;
