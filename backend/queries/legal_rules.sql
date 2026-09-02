-- name: CreateLegalRule :one
INSERT INTO legal_rules (
  code, title, description, act_reference, section, severity_default,
  effective_from, effective_to, is_active, params, created_by_id
) VALUES (
  $1, $2, sqlc.narg('description'), sqlc.narg('act_reference'), sqlc.narg('section'),
  COALESCE(sqlc.narg('severity_default')::violation_severity, 'MEDIUM'),
  sqlc.narg('effective_from'), sqlc.narg('effective_to'),
  COALESCE(sqlc.narg('is_active')::boolean, true), sqlc.narg('params'), sqlc.narg('created_by_id')
)
RETURNING *;

-- name: GetLegalRule :one
SELECT * FROM legal_rules WHERE id = $1;

-- name: ListLegalRules :many
SELECT r.*, (SELECT count(*) FROM rule_requirements rq WHERE rq.rule_id = r.id) AS requirements_count
FROM legal_rules r
WHERE
  (sqlc.narg('is_active')::boolean IS NULL OR r.is_active = sqlc.narg('is_active'))
  AND (
    sqlc.narg('search')::text IS NULL
    OR r.code ILIKE '%' || sqlc.narg('search') || '%'
    OR r.title ILIKE '%' || sqlc.narg('search') || '%'
  )
ORDER BY r.code ASC
LIMIT $1 OFFSET $2;

-- name: CountLegalRules :one
SELECT count(*) FROM legal_rules r
WHERE
  (sqlc.narg('is_active')::boolean IS NULL OR r.is_active = sqlc.narg('is_active'))
  AND (
    sqlc.narg('search')::text IS NULL
    OR r.code ILIKE '%' || sqlc.narg('search') || '%'
    OR r.title ILIKE '%' || sqlc.narg('search') || '%'
  );

-- name: UpdateLegalRule :one
UPDATE legal_rules SET
  title = COALESCE(sqlc.narg('title'), title),
  description = COALESCE(sqlc.narg('description'), description),
  act_reference = COALESCE(sqlc.narg('act_reference'), act_reference),
  section = COALESCE(sqlc.narg('section'), section),
  severity_default = COALESCE(sqlc.narg('severity_default'), severity_default),
  effective_from = COALESCE(sqlc.narg('effective_from'), effective_from),
  effective_to = COALESCE(sqlc.narg('effective_to'), effective_to),
  is_active = COALESCE(sqlc.narg('is_active'), is_active),
  params = COALESCE(sqlc.narg('params'), params),
  updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpsertLegalRuleByCode :one
-- Mirrors one rule from the Python rule engine's registry (GET /rules) into
-- the local legal_rules table, purely so compliance_checks/violations have a
-- foreign key to point at. This never drives evaluation — the Python service
-- is the sole source of the actual ruleset; this row is a local mirror.
INSERT INTO legal_rules (code, title, description, act_reference, severity_default, is_active)
VALUES ($1, $2, sqlc.narg('description'), sqlc.narg('act_reference'), $3, true)
ON CONFLICT (code) DO UPDATE SET
  title = EXCLUDED.title,
  description = COALESCE(EXCLUDED.description, legal_rules.description),
  act_reference = COALESCE(EXCLUDED.act_reference, legal_rules.act_reference),
  severity_default = EXCLUDED.severity_default,
  is_active = true,
  updated_at = now()
RETURNING *;

-- name: ListApplicableRules :many
SELECT * FROM legal_rules
WHERE is_active = true
AND (
  sqlc.narg('category_id')::uuid IS NULL
  OR EXISTS (SELECT 1 FROM rule_categories rc WHERE rc.rule_id = legal_rules.id AND rc.category_id = sqlc.narg('category_id'))
  OR NOT EXISTS (SELECT 1 FROM rule_categories rc2 WHERE rc2.rule_id = legal_rules.id)
)
ORDER BY code;

-- name: CreateRuleRequirement :one
INSERT INTO rule_requirements (
  rule_id, key, declaration_type, operator, expected, unit, severity, message, is_mandatory, order_index
) VALUES (
  $1, $2, sqlc.narg('declaration_type'), $3, sqlc.narg('expected'), sqlc.narg('unit'),
  sqlc.narg('severity'), sqlc.narg('message'), $4, $5
)
RETURNING *;

-- name: DeleteRuleRequirements :exec
DELETE FROM rule_requirements WHERE rule_id = $1;

-- name: ListRuleRequirements :many
SELECT * FROM rule_requirements WHERE rule_id = $1 ORDER BY order_index ASC;

-- name: CreateRuleCategory :exec
INSERT INTO rule_categories (rule_id, category_id) VALUES ($1, $2)
ON CONFLICT (rule_id, category_id) DO NOTHING;

-- name: DeleteRuleCategories :exec
DELETE FROM rule_categories WHERE rule_id = $1;

-- name: ListRuleCategories :many
SELECT rc.rule_id, c.id, c.code, c.name
FROM rule_categories rc
JOIN product_categories c ON c.id = rc.category_id
WHERE rc.rule_id = $1;
