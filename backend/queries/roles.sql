-- name: ListRoles :many
SELECT * FROM roles ORDER BY name ASC;

-- name: GetRoleByID :one
SELECT * FROM roles WHERE id = $1;

-- name: GetRoleByName :one
SELECT * FROM roles WHERE name = $1;

-- name: CreateRole :one
INSERT INTO roles (name, description, permissions, is_system)
VALUES ($1, sqlc.narg('description'), $2, false)
RETURNING *;

-- name: UpdateRole :one
UPDATE roles SET
  description = COALESCE(sqlc.narg('description'), description),
  permissions = COALESCE(sqlc.narg('permissions'), permissions),
  updated_at = now()
WHERE id = $1
RETURNING *;
