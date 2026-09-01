-- name: CreateEvidence :one
INSERT INTO evidence (inspection_id, violation_id, image_id, type, storage_key, bbox, caption, metadata, created_by_id)
VALUES ($1, sqlc.narg('violation_id'), sqlc.narg('image_id'), $2, sqlc.narg('storage_key'), sqlc.narg('bbox'), sqlc.narg('caption'), sqlc.narg('metadata'), sqlc.narg('created_by_id'))
RETURNING *;

-- name: ListEvidenceByInspection :many
SELECT * FROM evidence WHERE inspection_id = $1 ORDER BY created_at DESC;

-- name: ListEvidenceByViolation :many
SELECT * FROM evidence WHERE violation_id = $1 ORDER BY created_at DESC;
