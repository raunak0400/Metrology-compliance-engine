-- name: CreateUser :one
INSERT INTO users (email, password_hash, full_name, phone)
VALUES ($1, $2, $3, sqlc.narg('phone'))
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: SetLastLogin :exec
UPDATE users SET last_login_at = now(), updated_at = now() WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM users
WHERE
  (sqlc.narg('is_active')::boolean IS NULL OR is_active = sqlc.narg('is_active'))
  AND (
    sqlc.narg('search')::text IS NULL
    OR email ILIKE '%' || sqlc.narg('search') || '%'
    OR full_name ILIKE '%' || sqlc.narg('search') || '%'
  )
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT count(*) FROM users
WHERE
  (sqlc.narg('is_active')::boolean IS NULL OR is_active = sqlc.narg('is_active'))
  AND (
    sqlc.narg('search')::text IS NULL
    OR email ILIKE '%' || sqlc.narg('search') || '%'
    OR full_name ILIKE '%' || sqlc.narg('search') || '%'
  );

-- name: UpdateUser :one
UPDATE users SET
  full_name = COALESCE(sqlc.narg('full_name'), full_name),
  phone = COALESCE(sqlc.narg('phone'), phone),
  is_active = COALESCE(sqlc.narg('is_active'), is_active),
  updated_at = now()
WHERE id = $1
RETURNING *;

-- name: ListRolesForUser :many
SELECT r.id, r.name, r.permissions
FROM roles r
JOIN user_roles ur ON ur.role_id = r.id
WHERE ur.user_id = $1
ORDER BY r.name;

-- name: DeleteUserRoles :exec
DELETE FROM user_roles WHERE user_id = $1;

-- name: InsertUserRole :exec
INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)
ON CONFLICT (user_id, role_id) DO NOTHING;
