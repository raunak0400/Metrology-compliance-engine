-- name: CreateMLJob :one
INSERT INTO ml_jobs (inspection_id, image_id, type, status, queue_name)
VALUES ($1, sqlc.narg('image_id'), $2, 'QUEUED', $3)
RETURNING *;

-- name: SetMLJobAsynqID :exec
UPDATE ml_jobs SET bull_job_id = $2, updated_at = now() WHERE id = $1;

-- name: SetMLJobRunning :exec
UPDATE ml_jobs SET status = 'RUNNING', started_at = now(), updated_at = now() WHERE id = $1;

-- name: SetMLJobSucceeded :exec
UPDATE ml_jobs SET status = 'SUCCEEDED', finished_at = now(), result_mongo_id = $2, updated_at = now() WHERE id = $1;

-- name: SetMLJobFailed :exec
UPDATE ml_jobs SET status = 'FAILED', finished_at = now(), error = $2, attempts = attempts + 1, updated_at = now() WHERE id = $1;

-- name: ListMLJobsByInspection :many
SELECT * FROM ml_jobs WHERE inspection_id = $1 ORDER BY created_at DESC;
