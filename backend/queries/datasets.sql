-- name: CreateDataset :one
INSERT INTO datasets (name, slug, description, purpose, created_by_id)
VALUES ($1, $2, sqlc.narg('description'), sqlc.narg('purpose'), sqlc.narg('created_by_id'))
RETURNING *;

-- name: ListDatasets :many
SELECT d.*,
  (SELECT count(*) FROM dataset_items di WHERE di.dataset_id = d.id) AS items_count,
  (SELECT count(*) FROM dataset_versions dv WHERE dv.dataset_id = d.id) AS versions_count
FROM datasets d
ORDER BY d.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountDatasets :one
SELECT count(*) FROM datasets;

-- name: GetDataset :one
SELECT * FROM datasets WHERE id = $1;

-- name: CreateDatasetItem :one
INSERT INTO dataset_items (dataset_id, storage_key, source_image_id, mime_type, split, metadata)
VALUES ($1, $2, sqlc.narg('source_image_id'), sqlc.narg('mime_type'), COALESCE(sqlc.narg('split')::dataset_split, 'UNASSIGNED'), sqlc.narg('metadata'))
RETURNING *;

-- name: ListDatasetItems :many
SELECT di.*, (SELECT count(*) FROM dataset_annotations a WHERE a.item_id = di.id) AS annotations_count
FROM dataset_items di
WHERE
  di.dataset_id = $1
  AND (sqlc.narg('split')::dataset_split IS NULL OR di.split = sqlc.narg('split'))
  AND (sqlc.narg('annotation_status')::annotation_status IS NULL OR di.annotation_status = sqlc.narg('annotation_status'))
ORDER BY di.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountDatasetItems :one
SELECT count(*) FROM dataset_items di
WHERE
  di.dataset_id = $1
  AND (sqlc.narg('split')::dataset_split IS NULL OR di.split = sqlc.narg('split'))
  AND (sqlc.narg('annotation_status')::annotation_status IS NULL OR di.annotation_status = sqlc.narg('annotation_status'));

-- name: GetDatasetItem :one
SELECT * FROM dataset_items WHERE id = $1;

-- name: CreateDatasetAnnotation :one
INSERT INTO dataset_annotations (item_id, annotator_id, field_label, classification_label, bbox, value, confidence, status, metadata)
VALUES ($1, sqlc.narg('annotator_id'), sqlc.narg('field_label'), sqlc.narg('classification_label'), sqlc.narg('bbox'), sqlc.narg('value'), sqlc.narg('confidence'), COALESCE(sqlc.narg('status')::annotation_status, 'ANNOTATED'), sqlc.narg('metadata'))
RETURNING *;

-- name: UpdateDatasetItemAnnotationStatus :exec
UPDATE dataset_items SET annotation_status = $2, updated_at = now() WHERE id = $1;
