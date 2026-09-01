-- name: CreateAuditLog :exec
INSERT INTO audit_logs (actor_id, action, entity_type, entity_id, old_state, new_state, ip, user_agent, request_id)
VALUES (sqlc.narg('actor_id'), $1, $2, sqlc.narg('entity_id'), sqlc.narg('old_state'), sqlc.narg('new_state'), sqlc.narg('ip'), sqlc.narg('user_agent'), sqlc.narg('request_id'));

-- name: ListAuditLogs :many
SELECT * FROM audit_logs
WHERE
  (sqlc.narg('entity_type')::text IS NULL OR entity_type = sqlc.narg('entity_type'))
  AND (sqlc.narg('entity_id')::text IS NULL OR entity_id = sqlc.narg('entity_id'))
  AND (sqlc.narg('actor_id')::uuid IS NULL OR actor_id = sqlc.narg('actor_id'))
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountAuditLogs :one
SELECT count(*) FROM audit_logs
WHERE
  (sqlc.narg('entity_type')::text IS NULL OR entity_type = sqlc.narg('entity_type'))
  AND (sqlc.narg('entity_id')::text IS NULL OR entity_id = sqlc.narg('entity_id'))
  AND (sqlc.narg('actor_id')::uuid IS NULL OR actor_id = sqlc.narg('actor_id'));
