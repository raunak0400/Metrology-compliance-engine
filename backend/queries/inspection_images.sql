-- name: CreateInspectionImage :one
INSERT INTO inspection_images (
  inspection_id, storage_key, original_filename, mime_type, size_bytes, kind, status, checksum
) VALUES (
  $1, $2, sqlc.narg('original_filename'), $3, sqlc.narg('size_bytes'), 'ORIGINAL', 'UPLOADED', sqlc.narg('checksum')
)
RETURNING *;

-- name: GetInspectionImage :one
SELECT * FROM inspection_images WHERE id = $1;

-- name: ListOriginalImagesByInspection :many
SELECT id FROM inspection_images WHERE inspection_id = $1 AND kind = 'ORIGINAL';

-- name: UpdateInspectionImageStatus :exec
UPDATE inspection_images SET status = $2, updated_at = now() WHERE id = $1;

-- name: MarkInspectionImageProcessed :exec
UPDATE inspection_images SET status = 'PROCESSED', mongo_ocr_doc_id = $2, updated_at = now() WHERE id = $1;

-- name: SetInspectionImageScaleReference :exec
UPDATE inspection_images SET scale_reference = $2, updated_at = now() WHERE id = $1;
