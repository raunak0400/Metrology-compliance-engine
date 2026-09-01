-- name: CreateInspection :one
INSERT INTO inspections (code, product_id, category_id, inspector_id, source, notes, metadata)
VALUES ($1, sqlc.narg('product_id'), sqlc.narg('category_id'), $2, sqlc.narg('source'), sqlc.narg('notes'), sqlc.narg('metadata'))
RETURNING *;

-- name: GetInspection :one
SELECT * FROM inspections WHERE id = $1;

-- name: GetInspectionDetail :one
SELECT
  i.*,
  p.name AS product_name,
  pc.name AS product_category_name,
  u1.full_name AS inspector_name, u1.email AS inspector_email,
  u2.full_name AS reviewer_name, u2.email AS reviewer_email,
  (SELECT count(*) FROM extracted_declarations d WHERE d.inspection_id = i.id) AS declarations_count,
  (SELECT count(*) FROM compliance_checks c WHERE c.inspection_id = i.id) AS checks_count,
  (SELECT count(*) FROM violations v WHERE v.inspection_id = i.id) AS violations_count,
  (SELECT count(*) FROM reports r WHERE r.inspection_id = i.id) AS reports_count
FROM inspections i
LEFT JOIN products p ON p.id = i.product_id
LEFT JOIN product_categories pc ON pc.id = p.category_id
JOIN users u1 ON u1.id = i.inspector_id
LEFT JOIN users u2 ON u2.id = i.reviewer_id
WHERE i.id = $1;

-- name: ListInspections :many
SELECT
  i.*,
  p.name AS product_name,
  (SELECT count(*) FROM inspection_images img WHERE img.inspection_id = i.id) AS images_count,
  (SELECT count(*) FROM violations v WHERE v.inspection_id = i.id) AS violations_count,
  (SELECT count(*) FROM extracted_declarations d WHERE d.inspection_id = i.id) AS declarations_count
FROM inspections i
LEFT JOIN products p ON p.id = i.product_id
WHERE
  (sqlc.narg('status')::inspection_status IS NULL OR i.status = sqlc.narg('status'))
  AND (sqlc.narg('outcome')::inspection_outcome IS NULL OR i.outcome = sqlc.narg('outcome'))
  AND (sqlc.narg('inspector_id')::uuid IS NULL OR i.inspector_id = sqlc.narg('inspector_id'))
  AND (sqlc.narg('product_id')::uuid IS NULL OR i.product_id = sqlc.narg('product_id'))
ORDER BY i.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountInspections :one
SELECT count(*) FROM inspections i
WHERE
  (sqlc.narg('status')::inspection_status IS NULL OR i.status = sqlc.narg('status'))
  AND (sqlc.narg('outcome')::inspection_outcome IS NULL OR i.outcome = sqlc.narg('outcome'))
  AND (sqlc.narg('inspector_id')::uuid IS NULL OR i.inspector_id = sqlc.narg('inspector_id'))
  AND (sqlc.narg('product_id')::uuid IS NULL OR i.product_id = sqlc.narg('product_id'));

-- name: UpdateInspection :one
UPDATE inspections SET
  status = COALESCE(sqlc.narg('status'), status),
  outcome = COALESCE(sqlc.narg('outcome'), outcome),
  reviewer_id = COALESCE(sqlc.narg('reviewer_id'), reviewer_id),
  notes = COALESCE(sqlc.narg('notes'), notes),
  metadata = COALESCE(sqlc.narg('metadata'), metadata),
  updated_at = now()
WHERE id = $1
RETURNING *;

-- name: SetInspectionProcessing :exec
UPDATE inspections SET status = 'PROCESSING', started_at = now(), updated_at = now() WHERE id = $1;

-- name: SetInspectionOutcome :exec
UPDATE inspections SET status = $2, outcome = $3, completed_at = now(), updated_at = now() WHERE id = $1;
