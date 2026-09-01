-- name: CreateReport :one
INSERT INTO reports (inspection_id, type, status, generated_by_id)
VALUES ($1, $2, 'PENDING', sqlc.narg('generated_by_id'))
RETURNING *;

-- name: ListReportsByInspection :many
SELECT * FROM reports WHERE inspection_id = $1 ORDER BY created_at DESC;

-- name: GetReport :one
SELECT * FROM reports WHERE id = $1;

-- name: MarkReportGenerating :exec
UPDATE reports SET status = 'GENERATING', updated_at = now() WHERE id = $1;

-- name: MarkReportReady :exec
UPDATE reports SET status = 'READY', storage_key = $2, file_name = $3, size_bytes = $4, checksum = $5, generated_at = now(), updated_at = now() WHERE id = $1;

-- name: MarkReportFailed :exec
UPDATE reports SET status = 'FAILED', error = $2, updated_at = now() WHERE id = $1;

-- name: GetInspectionForReport :one
SELECT
  i.*,
  p.name AS product_name,
  pc.name AS category_name,
  m.name AS manufacturer_name,
  u1.full_name AS inspector_name, u1.email AS inspector_email,
  u2.full_name AS reviewer_name, u2.email AS reviewer_email
FROM inspections i
LEFT JOIN products p ON p.id = i.product_id
LEFT JOIN product_categories pc ON pc.id = p.category_id
LEFT JOIN manufacturers m ON m.id = p.manufacturer_id
JOIN users u1 ON u1.id = i.inspector_id
LEFT JOIN users u2 ON u2.id = i.reviewer_id
WHERE i.id = $1;
