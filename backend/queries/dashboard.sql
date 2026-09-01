-- name: CountAllInspections :one
SELECT count(*) FROM inspections;

-- name: InspectionsGroupByOutcome :many
SELECT outcome, count(*) AS total FROM inspections GROUP BY outcome;

-- name: InspectionsGroupByStatus :many
SELECT status, count(*) AS total FROM inspections GROUP BY status;

-- name: CountOpenViolations :one
SELECT count(*) FROM violations WHERE status = 'OPEN';

-- name: CountAllProducts :one
SELECT count(*) FROM products;

-- name: CountActiveRules :one
SELECT count(*) FROM legal_rules WHERE is_active = true;

-- name: ViolationsGroupBySeverity :many
SELECT severity, count(*) AS total FROM violations GROUP BY severity;

-- name: ViolationsGroupByStatus :many
SELECT status, count(*) AS total FROM violations GROUP BY status;

-- name: TopViolatedRules :many
SELECT v.rule_id, lr.code AS rule_code, lr.title AS rule_title, count(*) AS total
FROM violations v
LEFT JOIN legal_rules lr ON lr.id = v.rule_id
GROUP BY v.rule_id, lr.code, lr.title
ORDER BY total DESC
LIMIT 10;

-- name: RecentInspections :many
SELECT id, code, status, outcome, created_at FROM inspections ORDER BY created_at DESC LIMIT 10;

-- name: CountInspectionsSince :one
SELECT count(*) FROM inspections WHERE created_at >= $1;
