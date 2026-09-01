-- name: CreateComplianceCheck :one
INSERT INTO compliance_checks (
  inspection_id, rule_id, requirement_id, declaration_id, status,
  actual_value, expected_value, reason, confidence, requires_manual_review, details
) VALUES (
  $1, $2, sqlc.narg('requirement_id'), sqlc.narg('declaration_id'), $3,
  sqlc.narg('actual_value'), sqlc.narg('expected_value'), sqlc.narg('reason'),
  sqlc.narg('confidence'), $4, sqlc.narg('details')
)
RETURNING *;

-- name: DeleteComplianceChecksByInspection :exec
DELETE FROM compliance_checks WHERE inspection_id = $1;

-- name: ListComplianceChecksByInspection :many
SELECT cc.*, lr.code AS rule_code, lr.title AS rule_title
FROM compliance_checks cc
JOIN legal_rules lr ON lr.id = cc.rule_id
WHERE cc.inspection_id = $1
ORDER BY cc.created_at ASC;

-- name: CreateViolation :one
INSERT INTO violations (
  inspection_id, rule_id, check_id, severity, status, title, reason, legal_reference
) VALUES (
  $1, sqlc.narg('rule_id'), sqlc.narg('check_id'), $2, 'OPEN', $3, sqlc.narg('reason'), sqlc.narg('legal_reference')
)
RETURNING *;

-- name: DeleteViolationsByInspection :exec
DELETE FROM violations WHERE inspection_id = $1;

-- name: ListViolationsByInspection :many
SELECT v.*, lr.code AS rule_code, lr.title AS rule_title, lr.act_reference AS rule_act_reference
FROM violations v
LEFT JOIN legal_rules lr ON lr.id = v.rule_id
WHERE v.inspection_id = $1
ORDER BY v.severity DESC, v.created_at ASC;

-- name: GetViolation :one
SELECT * FROM violations WHERE id = $1;

-- name: UpdateViolation :one
UPDATE violations SET
  status = COALESCE(sqlc.narg('status'), status),
  severity = COALESCE(sqlc.narg('severity'), severity),
  review_notes = COALESCE(sqlc.narg('review_notes'), review_notes),
  resolved_by_id = CASE WHEN sqlc.narg('terminal')::boolean THEN sqlc.narg('resolved_by_id') ELSE resolved_by_id END,
  resolved_at = CASE WHEN sqlc.narg('terminal')::boolean THEN now() ELSE resolved_at END,
  updated_at = now()
WHERE id = $1
RETURNING *;
