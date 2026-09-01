-- name: ListDeclarationsByInspection :many
SELECT * FROM extracted_declarations WHERE inspection_id = $1 ORDER BY type ASC;

-- name: GetDeclaration :one
SELECT * FROM extracted_declarations WHERE id = $1;

-- name: UpdateDeclarationManual :one
UPDATE extracted_declarations SET
  normalized_value = COALESCE(sqlc.narg('normalized_value'), normalized_value),
  numeric_value = COALESCE(sqlc.narg('numeric_value'), numeric_value),
  unit = COALESCE(sqlc.narg('unit'), unit),
  type = COALESCE(sqlc.narg('type'), type),
  source = 'MANUAL',
  is_edited_by_reviewer = true,
  edited_by_id = sqlc.narg('edited_by_id'),
  updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteMLDeclarationsForImage :exec
DELETE FROM extracted_declarations
WHERE inspection_id = $1 AND source_image_id = $2 AND source IN ('OCR', 'ML_EXTRACTION');

-- name: CreateDeclaration :one
INSERT INTO extracted_declarations (
  inspection_id, source_image_id, type, normalized_value, raw_value, unit,
  numeric_value, confidence, bbox, source, mongo_ref_id
) VALUES (
  $1, $2, $3, sqlc.narg('normalized_value'), sqlc.narg('raw_value'), sqlc.narg('unit'),
  sqlc.narg('numeric_value'), sqlc.narg('confidence'), sqlc.narg('bbox'), $4, sqlc.narg('mongo_ref_id')
)
RETURNING *;

-- name: ListDeclarationsWithScaleByInspection :many
SELECT d.id, d.type, d.bbox, img.scale_reference
FROM extracted_declarations d
LEFT JOIN inspection_images img ON img.id = d.source_image_id
WHERE d.inspection_id = $1;
